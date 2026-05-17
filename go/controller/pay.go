package controller

import (
	"context"
	"fmt"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/helper"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/model"
	"server_golang/repo/cache"
	"server_golang/repo/table"
)

// PayController 处理支付相关请求
type PayController struct {
	*BaseController
}

func dispatchPay(ctx context.Context, base *BaseController, action string) *Result {
	c := &PayController{BaseController: base}
	switch action {
	case "order":
		return c.Order(ctx)
	case "checkpay":
		return c.CheckPay(ctx)
	case "applepayback":
		return c.ApplePayback(ctx)
	case "h5payback":
		return c.H5Payback(ctx)
	default:
		return base.ResponseError(3, fmt.Sprintf("route error: pay/%s", action))
	}
}

// Order 下单接口，对应 PHP PayController::orderAction
func (c *PayController) Order(ctx context.Context) *Result {
	log.Infof(ctx, "orderAction: %s", helper.CreateLinkstringUrlencode(c.Params))

	userID := c.Params.GetInt64E("userid")
	token := c.Params.GetStringE("token")
	// 0-充值 1-成长基金 2-特权商城 3-每日礼包 4-每周限时礼包 5-月度超值礼包
	tradeType := c.Params.GetIntE("type")
	amount := c.Params.GetIntE("amount")
	pf := c.Params.GetStringE("pf") // 平台 Android / IOS / H5

	zoneID := model.GetUserZoneID(userID)
	sourceUID := model.GetUIDByZoneUserID(userID)
	userInfo, err := model.GetUserAndAuth(ctx, sourceUID, token)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 初始化订单
	tradeData := &table.NewTradeInfo{
		Type:     tradeType,
		ZoneId:   zoneID,
		UserId:   userID,
		OpenId:   userInfo.OpenId,
		TotalFee: amount,
		Ip:       c.IP,
		Pf:       pf,
	}

	insertID, err := model.InsertNewTradeInfo(ctx, tradeData)
	if err != nil || insertID <= 0 {
		return c.ResponseError(98370392, "system error")
	}

	// 预订单缓存（PHP 为 Redis，迁移到 Go 侧内存缓存，TTL 7 天）
	tradeMap := types.Map{
		"type":      tradeType,
		"zone_id":   zoneID,
		"user_id":   userID,
		"open_id":   userInfo.OpenId,
		"total_fee": amount,
		"ip":        c.IP,
		"pf":        pf,
	}
	cache.SetWithTTL(fmt.Sprintf(config.CacheOrder, insertID), tradeMap, 7*86400)

	log.Infof(ctx, "orderActionRet Orderid:%d", insertID)

	return c.ResponseSuccessToMe(types.Map{
		"appid":   userInfo.Appid,
		"orderid": insertID,
	})
}

// CheckPay 支付查询，对应 PHP PayController::checkpayAction
func (c *PayController) CheckPay(ctx context.Context) *Result {
	log.Infof(ctx, "payAction: %s", helper.CreateLinkstringUrlencode(c.Params))

	userID := c.Params.GetInt64E("userid")
	token := c.Params.GetStringE("token")
	orderID := c.Params.GetInt64E("orderid")

	sourceUID := model.GetUIDByZoneUserID(userID)
	userInfo, err := model.GetUserAndAuth(ctx, sourceUID, token)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	tradeInfo, _ := model.GetNewTradeInfoById(ctx, orderID)
	if tradeInfo == nil {
		return c.ResponseError(4323, "订单不存在")
	}

	if tradeInfo.Status == "SUCCESS" {
		return c.ResponseError(8888, "支付已完成")
	}

	// 校验：H5 渠道使用 out_trade_no，其余使用订单 id
	var outTradeNo string
	if tradeInfo.Channel == "H5" {
		outTradeNo = tradeInfo.OutTradeNo
	} else {
		outTradeNo = types.ToString(orderID)
	}

	appid := userInfo.Appid
	openID := userInfo.OpenId
	fdate := tradeInfo.Createtime.Format("20060102")
	checkURL := fmt.Sprintf("https://mprogram.boomegg.cn/wxpay/order?appid=%s&openid=%s&orderid=%s&fdate=%s",
		appid, openID, outTradeNo, fdate)

	checkRetStr, err := helper.HttpsRequest(checkURL)
	if err != nil {
		log.Infof(ctx, "checkPay request error: %v", err)
		return c.ResponseError(999, "交易查询失败")
	}
	log.Infof(ctx, "checkPay: %s", checkRetStr)

	checkRet := json.ToMap(checkRetStr)
	if len(checkRet) == 0 || checkRet.GetIntE("code") != 0 {
		log.Infof(ctx, "交易查询失败: %s", checkRetStr)
		return c.ResponseError(999, "交易查询失败")
	}

	dataList, _ := checkRet["data"].([]any)
	if len(dataList) == 0 {
		log.Infof(ctx, "未查询到订单: %s", checkRetStr)
		return c.ResponseError(1235, "未查询到订单")
	}

	first, _ := dataList[0].(map[string]any)
	firstMap := types.Map(first)

	updateTrade := types.Map{}
	if transactionID := firstMap.GetStringE("transaction_id"); transactionID != "" {
		updateTrade["transaction_id"] = transactionID
	}
	if phpOutTradeNo := firstMap.GetStringE("out_trade_no"); phpOutTradeNo != "" {
		updateTrade["out_trade_no"] = phpOutTradeNo
	}

	status := "FAIL"
	if firstMap.GetIntE("state") == 2 {
		status = "SUCCESS"
	}
	updateTrade["status"] = status

	if err := model.UpdateNewTradeInfoById(ctx, orderID, updateTrade); err != nil {
		log.Errorf(ctx, 0, "updateNewTradeInfo err: %v", err)
	}

	if status == "FAIL" {
		return c.ResponseError(3312, "交易失败")
	}

	// 购买商品发货
	result, err := model.BuyStaff(ctx, userID, tradeInfo.Type, tradeInfo.TotalFee, orderID)
	if err != nil || len(result) == 0 {
		log.Errorf(ctx, 0, "系统错误: trade_info=%+v err=%v", tradeInfo, err)
		return c.ResponseError(99, "系统错误")
	}

	// 仅充值(type=0)允许同一预订单多次购买，其余类型成功后删除预订单缓存
	if tradeInfo.Type != 0 {
		cache.Del(fmt.Sprintf(config.CacheOrder, orderID))
	}

	result["orderid"] = orderID
	return c.ResponseSuccessToMe(result)
}

// ApplePayback 苹果支付回调，对应 PHP PayController::applepaybackAction
func (c *PayController) ApplePayback(ctx context.Context) *Result {
	log.Infof(ctx, "applepaybackAction: %s", helper.CreateLinkstringUrlencode(c.Params))

	tradeInfoStr := c.Params.GetStringE("trade_info")
	tradeInfo := json.ToMap(tradeInfoStr)

	userID := tradeInfo.GetInt64E("user_id")
	orderID := tradeInfo.GetInt64E("mer_order_no")
	amt := tradeInfo.GetIntE("real_mer_amount")
	buyType := tradeInfo.GetIntE("type")

	// 购买商品发货
	result, err := model.BuyStaff(ctx, userID, buyType, amt, orderID)
	if err != nil || len(result) == 0 {
		log.Errorf(ctx, 0, "系统错误: trade_info=%s err=%v", tradeInfoStr, err)
		return c.ResponseError(99, "系统错误")
	}

	result["orderid"] = orderID
	return c.ResponseSuccessToMe(result)
}

// H5Payback H5 支付回调占位（PHP 侧无对应实现，保留空桩维持路由兼容）
func (c *PayController) H5Payback(ctx context.Context) *Result {
	log.Infof(ctx, "h5paybackAction")
	return c.ResponseSuccessToMe(types.Map{})
}
