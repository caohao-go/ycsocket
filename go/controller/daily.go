package controller

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// DailyrewardAction 获取签到信息
func (c *ShinelightController) DailyrewardAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	currentDay := model.GetDailyReward(ctx, userID)
	times := model.GetDailyRewardTimes(ctx, userID)

	data := model.GetVipContentsCurDay(ctx, userID)
	isVip := 0
	if data.GetIntE("leiji_xiaofei") > 0 {
		isVip = 1
	}

	return c.ResponseSuccessToMe(types.Map{
		"current_time": time.Now().Unix(),
		"current_day":  currentDay,
		"times":        times,
		"is_vip":       isVip,
	})
}

// GetdailyrewardAction 领取每日签到奖励
func (c *ShinelightController) GetdailyrewardAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	lockKey := "getdailyreward" + types.ToString(userID)
	if lockErr := lock.Lock(lockKey, 2); lockErr != nil {
		return c.ResponseError(7878, "操作太快，请稍后再试")
	}
	defer lock.Unlock(lockKey)

	days := c.Params.GetIntE("days")
	rewards, ok := logic.DailyDatas[days]
	if !ok || len(rewards) == 0 {
		return c.ResponseError(583241, "天数错误")
	}

	if model.GetDailyRewardTimes(ctx, userID) >= 1 {
		return c.ResponseError(58324, "当天已领取")
	}

	model.IncrDailyReward(ctx, userID)
	model.IncrDailyRewardTimes(ctx, userID)

	model.GiveReward(userID, rewards...)

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 32, 1)

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

// GetvipdailyrewardAction 领取VIP每日签到奖励
func (c *ShinelightController) GetvipdailyrewardAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 每天的累计充值额度
	times := model.GetDailyRewardTimes(ctx, userID)
	if times >= 2 {
		return c.ResponseError(58324, "当天已领取")
	}

	if times <= 0 {
		return c.ResponseError(58324, "请领取普通奖励")
	}

	data := model.GetVipContentsCurDay(ctx, userID)
	if data.GetIntE("leiji_xiaofei") <= 0 {
		return c.ResponseError(583245, "充值任意金额可以再次领取此奖励")
	}

	model.IncrDailyRewardTimes(ctx, userID)

	days := c.Params.GetIntE("days")
	rewards, _ := logic.DailyDatas[days]
	model.GiveReward(userID, rewards...)

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

// 点金手
func (c *ShinelightController) GetDianCoinAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")
	if lv <= 0 {
		lv = 1
	}
	times := 1.0 + float64(logic.GetVipInfo(userGrade, "golden_touch_coin"))/100.0

	ret := []types.Map{
		{"type": 1, "count": model.GetDianCoin(ctx, userID, 1), "price": 0, "coin_num": int(669 * float64(lv) * times)},
		{"type": 2, "count": model.GetDianCoin(ctx, userID, 2), "price": 20, "coin_num": int(1338 * float64(lv) * times)},
		{"type": 3, "count": model.GetDianCoin(ctx, userID, 3), "price": 50, "coin_num": int(3345 * float64(lv) * times)},
	}
	return c.ResponseSuccessToMe(types.Map{"refresh": util.DianCoinRefreshTime(), "data": ret})
}

func (c *ShinelightController) BuyDianCoinAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	typ := c.Params.GetIntE("type")
	if typ < 1 || typ > 3 {
		return c.ResponseError(534, "type error")
	}

	lock.Lock(fmt.Sprintf("buyDianCoin%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("buyDianCoin%d", userID))

	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")
	if lv <= 0 {
		lv = 1
	}
	times := 1.0 + float64(logic.GetVipInfo(userGrade, "golden_touch_coin"))/100.0

	coinBase := []int{669, 1338, 3345}
	coinNum := int(float64(coinBase[typ-1]) * float64(lv) * times)

	priceTypes := []int{0, 20, 50}
	price := priceTypes[typ-1]

	// 检查购买次数
	if model.GetDianCoin(ctx, userID, typ) <= 0 {
		return c.ResponseError(5834, "次数不够")
	}

	// 扣除钻石
	if price > 0 {
		if item.NotEnough(userID, 2, price) {
			return c.ResponseError(666666, "钻石不够")
		}
		item.Sub(userID, 2, price)
	}

	// 获得金币
	item.AddCoin(userID, coinNum)

	// 消耗购买次数
	model.SetDianCoin(ctx, userID, typ)

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 38, 1)

	return c.ResponseSuccessToMe(types.Map{"coin_num": coinNum})
}
