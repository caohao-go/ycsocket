package controller

import (
	"context"
	"fmt"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// 商店系统

func (c *ShinelightController) GetShopAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	shopID := c.Params.GetIntE("shopid")
	if shopID == 0 {
		return c.ResponseError(992123, "shopid is empty")
	}

	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")

	tmp := model.GetShopByID(ctx, userID, shopID, lv)
	data := types.ObjectToMap(tmp)

	if len(tmp.Items) > 0 {
		costType := tmp.Items[0].CostType
		data["cost_type"] = costType
		data["my_cost_count"] = item.Total(userID, costType)
	}

	if shopID == model.PowerType10040 || shopID == model.PowerType10030 {
		data["refresh_times"] = model.GetRefreshTimes(ctx, userID, shopID)
	}
	if shopID == model.PowerType10030 || shopID == model.PowerType10040 {
		var needTime int
		data["user_power"] = model.GetUserPower(ctx, userID, shopID, &needTime)
		data["max_power"] = model.ConfMaxPower[shopID]
		data["need_time"] = needTime
	}
	if shopID == model.PowerType10024 {
		var needTime int
		data["user_power"] = model.GetUserPower(ctx, userID, shopID, &needTime)
		data["need_time"] = needTime
	}

	return c.ResponseSuccessToMe(data)
}

// 卖掉物品
func (c *ShinelightController) SellItemAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	itemID := c.Params.GetIntE("item_id")
	num := c.Params.GetIntE("num")
	if num < 1 {
		num = 1
	}

	itemInfo := item.Table[itemID]
	if itemInfo == nil {
		return c.ResponseError(452123, "input item_id is wrong")
	}

	if item.NotEnough(userID, itemID, num) {
		return c.ResponseError(54367, "数量不够")
	}

	result := make([]types.Map, 0)
	if len(itemInfo.Price) > 0 {
		for _, p := range itemInfo.Price {
			cnt := p.Num * num
			result = append(result, types.Map{"type": p.Type, "num": cnt})
			item.Add(userID, p.Type, cnt, nil)
		}
		item.Sub(userID, itemID, num)
	}

	model.GuideTaskHandle(ctx, userID, 55, 1)

	return c.ResponseSuccessToMe(types.Map{"get": result})
}

// 购买商品（商店）
func (c *ShinelightController) ShopBuyAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")
	num := c.Params.GetIntE("num")
	if num < 1 {
		num = 1
	}
	if id == 0 {
		return c.ResponseError(992123, "id is empty")
	}

	shopItems := logic.GlobalPreShopsItems[id]
	if shopItems == nil {
		return c.ResponseError(995523, "item is empty")
	}

	lock.Lock(fmt.Sprintf("shopbuy%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("shopbuy%d", userID))

	userGrade := model.GetUserAttr(userID)
	shopIDVal := shopItems.ShopId

	// VIP 等级校验
	if shopIDVal == 10010 && id == 5 {
		if userGrade.GetIntE("vip_level") < 2 {
			return c.ResponseError(992526, "vip2以上玩家才能购买")
		}
	}

	// 购买数量上限
	curNum := model.GetRedisShopBuynum(ctx, userID, shopIDVal, id)
	buyLimit := shopItems.BuyLimit
	if buyLimit != 0 && curNum >= buyLimit {
		return c.ResponseError(992523, "商品已卖光")
	}

	// 货币校验
	costType := shopItems.CostType
	price := shopItems.Price
	if item.NotEnough(userID, costType, price*num) {
		return c.ResponseError(666666, "货币不够")
	}

	// 扣减货币
	item.Sub(userID, costType, price*num)

	// 增加商品
	switch shopIDVal {
	case 10010, 10011, 10021, 10022, 10023, 10025, 11000, 11100, 11200, 11300, 11400, 11500:
		item.Add(userID, shopItems.ItemId, shopItems.Number*num, nil)
	case 10024, 10030, 10040:
		item.Add(userID, shopItems.ItemId, shopItems.Number*num, nil)
	}

	// 增加购物计数
	model.AddRedisShopBuynum(ctx, userID, shopIDVal, id, num)

	// 全部卖完刷新
	model.AllByRefresh(ctx, userID, shopIDVal, userGrade.GetIntE("lv"))

	// 公会商店任务
	if shopIDVal == 10023 {
		tasks := model.GetGuildTask(ctx, userID)
		for _, t := range tasks {
			if t.TaskId == 4 {
				finishCount := t.FinishCount
				taskCountLimit := t.TaskCountLimit
				if finishCount < taskCountLimit && finishCount >= 0 {
					model.IncrTaskFinishNumStr(ctx, userID, "guild_4", 1)
					model.IncrUsersContentInt(ctx, userID, "guild_active", 5)
				}
				break
			}
		}
	}

	// 引导任务
	if shopIDVal == 10040 {
		model.GuideTaskHandle(ctx, userID, 9, 1)
	} else if shopIDVal == 10023 {
		model.GuideTaskHandle(ctx, userID, 103, 1)
	} else {
		model.GuideTaskHandle(ctx, userID, 50, 1)
	}

	return c.ResponseSuccessToMe(types.Map{})
}

// 商品购买（直购）
func (c *ShinelightController) ItemBuyAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	itemID := c.Params.GetIntE("item_id")
	num := c.Params.GetIntE("num")
	if num < 1 {
		num = 1
	}
	if itemID == 0 {
		return c.ResponseError(992123, "id is empty")
	}

	// 物品价格表
	itemPrices := map[int]struct{ Type, Num int }{
		21201: {Type: 2, Num: 20},
	}
	priceInfo, ok := itemPrices[itemID]
	if !ok {
		return c.ResponseError(992123, "item not found")
	}

	lock.Lock(fmt.Sprintf("itemBuy%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("itemBuy%d", userID))

	if item.NotEnough(userID, priceInfo.Type, priceInfo.Num*num) {
		return c.ResponseError(666666, "货币不够")
	}

	item.Sub(userID, priceInfo.Type, priceInfo.Num*num)
	item.Add(userID, itemID, num, nil)

	return c.ResponseSuccessToMe(types.Map{})
}

// 刷新商店
func (c *ShinelightController) RefreshShopAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	shopID := c.Params.GetIntE("shopid")

	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")

	shopInfo := model.GetRefreshInfo(ctx, userID, shopID, lv)
	ret := types.Map{"type": 0}

	hasManual := false
	if int(shopInfo.RefreshType.Min) == logic.ShopRefreshTypeManual ||
		int(shopInfo.RefreshType.Max) == logic.ShopRefreshTypeManual {
		hasManual = true
	}

	if hasManual {
		// 体力刷新
		if shopID == model.PowerType10030 || shopID == model.PowerType10040 || shopID == model.PowerType10024 {
			var needTime int
			power := model.GetUserPower(ctx, userID, shopID, &needTime)
			if power > 0 {
				newPower := model.SubUserPower(ctx, userID, shopID, &needTime, 1)
				if newPower != -1 {
					ret["type"] = 1
					ret["max_power"] = model.ConfMaxPower[shopID]
					ret["power"] = newPower
					ret["power_need_time"] = needTime
					model.RefreshShopBuynum(ctx, userID, shopID, lv)
					return c.ResponseSuccessToMe(ret)
				}
			}
		}

		// 货币刷新
		sellType := shopInfo.SellType
		basicCost := shopInfo.BasicCost
		if item.NotEnough(userID, sellType, basicCost) {
			return c.ResponseError(666666, "货币不够")
		}

		// 最大刷新次数
		if shopID == model.PowerType10040 || shopID == model.PowerType10030 {
			refreshTime := model.GetRefreshTimes(ctx, userID, shopID)
			maxRefresh := 20
			if shopID == model.PowerType10040 {
				maxRefresh = 100
			}
			if refreshTime > maxRefresh {
				return c.ResponseError(5478, "超过当天最大刷新次数")
			}
			model.IncrRefreshTimes(ctx, userID, shopID)
		}

		item.Sub(userID, sellType, basicCost)
		model.RefreshShopBuynum(ctx, userID, shopID, lv)
	}

	model.GuideTaskHandle(ctx, userID, 41, 1)

	return c.ResponseSuccessToMe(types.Map{})
}
