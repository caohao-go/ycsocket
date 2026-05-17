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

// VIP/充值系统

// VipBuyInfoAction vip界面信息
func (c *ShinelightController) VipBuyInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	vipLv := c.Params.GetIntE("viplv")
	if vipLv < 0 || vipLv > 13 {
		return c.ResponseError(649424, "vip等级错误")
	}

	userGrade := model.GetUserAttr(userID)

	userVipContent := model.GetUserVipContents(userID)

	// VIP奖品
	rewards := logic.TequanRewards[vipLv]
	// VIP购买价格
	vipBuy := logic.VipBuy[vipLv]
	// VIP特权礼包购买状态
	vipBuyLibao := userVipContent.GetMapE("vip_buy_libao")
	status := vipBuyLibao.GetIntE(vipLv)

	result := types.Map{
		"vip_buy": vipBuy,
		"status":  status,
		"rewards": rewards,
	}

	// 至尊月卡状态
	dateLeft := model.GetRongyaoCardExpire(ctx, userID, 2)
	if dateLeft == 0 {
		result["yueka_status"] = 0
		result["rewards_yueka"] = []interface{}{}
	} else {
		result["yueka_status"] = 1
		date := model.GetVipZhizunLingqu(ctx, userID)
		if date == 2 {
			result["yueka_status"] = 2
		}
	}

	// 至尊月卡专属奖励
	rewardsYueka := logic.ZhizunZhuanshuRewards[vipLv]
	result["rewards_yueka"] = rewardsYueka

	// 累计充值
	leijiChong := userVipContent.GetIntE("leiji_xiaofei")
	curVipLevel := userGrade.GetIntE("vip_level")
	nextNeed := 0
	if cfg, ok := logic.VipConfigs[curVipLevel+1]; ok {
		nextNeed = cfg["rmb_tatal"]
	}

	result["vip_level"] = curVipLevel
	result["next_vip_level"] = curVipLevel + 1
	result["leiji_chong"] = leijiChong
	result["next_need"] = nextNeed
	result["need"] = nextNeed - leijiChong

	return c.ResponseSuccessToMe(result)
}

// VipBuyzhizunLingquAction vip特权至尊月卡领取
func (c *ShinelightController) VipBuyzhizunLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	vipLv := c.Params.GetIntE("viplv")
	lockKey := fmt.Sprintf("vip_buyzhizun_lingqu%d", userID)

	if err := lock.Lock(lockKey, 2); err != nil {
		return c.ResponseError(7456, err.Error())
	}
	defer lock.Unlock(lockKey)

	userGrade := model.GetUserAttr(userID)
	curVipLevel := userGrade.GetIntE("vip_level")

	// VIP专属奖励
	rewardsYueka, ok := logic.ZhizunZhuanshuRewards[curVipLevel]
	if !ok || len(rewardsYueka) == 0 {
		return c.ResponseError(7456, "vip等级错误")
	}

	// 今日是否已领取
	date := model.GetVipZhizunLingqu(ctx, userID)
	if date == 2 {
		return c.ResponseError(74561, "已经领取奖励了")
	}

	// VIP等级检查
	if curVipLevel < vipLv {
		return c.ResponseError(745612, "vip等级不够")
	}

	// 设置领取状态
	model.SetVipZhizunLingqu(ctx, userID)

	// 给予奖励

	model.GiveReward(userID, rewardsYueka...)
	return c.ResponseSuccessToMe(types.Map{"rewards_yueka": rewardsYueka})
}

// VipBuyLingquAction vip特权礼包购买
func (c *ShinelightController) VipBuyLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	vipLv := c.Params.GetIntE("viplv")
	lockKey := fmt.Sprintf("vip_buy_lingqu%d", userID)

	if err := lock.Lock(lockKey, 2); err != nil {
		return c.ResponseError(7456, err.Error())
	}
	defer lock.Unlock(lockKey)

	// 验证等级
	vipBuyConfig, ok := logic.VipBuy[vipLv]
	if !ok || len(vipBuyConfig) == 0 {
		return c.ResponseError(7456, "输入等级错误")
	}

	userGrade := model.GetUserAttr(userID)
	curVipLevel := userGrade.GetIntE("vip_level")
	if vipLv > curVipLevel {
		return c.ResponseError(64942, "vip等级不够")
	}

	userVipContent := model.GetUserVipContents(userID)
	vipBuyLibao := userVipContent.GetMapE("vip_buy_libao")
	status := vipBuyLibao.GetIntE(vipLv)
	if status != 0 {
		return c.ResponseError(64943, "已经购买了")
	}

	// 购买价格（VipBuy[vipLv][0].Num 为折扣价）
	price := vipBuyConfig[0].Num

	if item.NotEnough(userID, 2, price) {
		return c.ResponseError(543671, "砖石不够")
	}

	// 奖励
	rewards := logic.TequanRewards[vipLv]
	model.GiveReward(userID, rewards...)

	// 扣除钻石
	item.Sub(userID, 2, price)

	// 标记已购买
	vipBuyLibao[types.ToString(vipLv)] = 1
	userVipContent["vip_buy_libao"] = vipBuyLibao
	model.UpdateUserVipContents(userID, userVipContent)

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 28, 1)

	return c.ResponseSuccessToMe(types.Map{"reward": rewards})
}

// LeijiChongzhiAction 每日累计充值
func (c *ShinelightController) LeijiChongzhiAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := model.GetVipContentsCurDay(ctx, userID)

	// 距离下一天0点的剩余时间
	now := time.Now()
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 1, 0, now.Location())
	leftTime := tomorrow.Unix() - now.Unix()

	leijiXiaofei := data.GetIntE("leiji_xiaofei")
	leijiChongLibao := data.GetMapE("leiji_chong_libao")
	lingqu := leijiChongLibao.GetMapE("lingqu")

	leiji := make([]types.Map, 0, len(logic.LejiChongCount))
	for _, chong := range logic.LejiChongCount {
		status := 0
		if leijiXiaofei >= chong {
			status = 1
			if lingqu.GetIntE(chong) != 0 {
				status = 2
			}
		}
		leiji = append(leiji, types.Map{
			"money":   chong,
			"status":  status,
			"rewards": logic.LejiChongzhiRewards[chong],
		})
	}

	return c.ResponseSuccessToMe(types.Map{
		"left_time":   leftTime,
		"leiji_chong": leijiXiaofei,
		"leiji":       leiji,
	})
}

// LeijiChongzhiLingquAction 每日累计充值领取
func (c *ShinelightController) LeijiChongzhiLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	typ := c.Params.GetIntE("type") // 50 100 328 648

	// 获取奖励配置
	rewards, ok := logic.LejiChongzhiRewards[typ]
	if !ok || len(rewards) == 0 {
		return c.ResponseError(543672, "奖励类型错误")
	}

	// 每天的累计充值额度
	data := model.GetVipContentsCurDay(ctx, userID)
	leijiXiaofei := data.GetIntE("leiji_xiaofei")
	if leijiXiaofei < typ {
		return c.ResponseError(543672, "充值金额不够")
	}

	// 领取状态检查
	leijiChongLibao := data.GetMapE("leiji_chong_libao")
	lingqu := leijiChongLibao.GetMapE("lingqu")
	if lingqu.GetIntE(typ) != 0 {
		return c.ResponseError(543672, "已经领取了该奖励")
	}

	// 给予奖励

	model.GiveReward(userID, rewards...)

	// 设置领取状态
	lingqu[types.ToString(typ)] = 1
	leijiChongLibao["lingqu"] = lingqu
	data["leiji_chong_libao"] = leijiChongLibao
	model.SetVipContentsCurDay(ctx, userID, data)

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

// JiTianAction 积天豪礼
func (c *ShinelightController) JiTianAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userVipContents := model.GetUserVipContents(userID)
	jitianData := userVipContents.GetMapE("jitian")
	leijiDays := jitianData.GetIntE("leiji")
	lingquData := jitianData.GetMapE("lingqu")

	jitian := make([]types.Map, 0, 15)
	for i := 1; i <= 15; i++ {
		if leijiDays >= i {
			lingqu := 1 // 可领取
			if lingquData.GetIntE(i) != 0 {
				lingqu = 2 // 已领取
			}
			jitian = append(jitian, types.Map{"days": i, "status": lingqu})
		} else {
			jitian = append(jitian, types.Map{"days": i, "status": 0})
		}
	}

	return c.ResponseSuccessToMe(types.Map{"jitian": jitian})
}

// JiTianLingquAction 积天豪礼领取
func (c *ShinelightController) JiTianLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	typ := c.Params.GetIntE("type") // 1-15

	// 奖励配置
	rewards, ok := logic.JitianHaoliReward[typ]
	if !ok || len(rewards) == 0 {
		return c.ResponseError(543672, "奖励类型错误")
	}

	userVipContents := model.GetUserVipContents(userID)
	jitianData := userVipContents.GetMapE("jitian")

	// 充值天数检查
	leijiDays := jitianData.GetIntE("leiji")
	if leijiDays < typ {
		return c.ResponseError(5436721, "充值天数不满足条件")
	}

	// 是否已领取
	lingquData := jitianData.GetMapE("lingqu")
	if lingquData.GetIntE(typ) != 0 {
		return c.ResponseError(5436721, "已经领取了该奖励")
	}

	// 给予奖励
	model.GiveReward(userID, rewards...)

	// 标记已领取
	lingquData[types.ToString(typ)] = 1
	jitianData["lingqu"] = lingquData
	userVipContents["jitian"] = jitianData
	model.UpdateUserVipContents(userID, userVipContents)

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

// YuekaAction 荣耀月卡和至尊月卡
func (c *ShinelightController) YuekaAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	typ := c.Params.GetIntE("type") // 1-荣耀月卡 2-至尊月卡
	if typ < 1 || typ > 2 {
		return c.ResponseError(649421, "月卡错误")
	}

	result := types.Map{}
	jihuo := 0

	// 获取月卡是否被激活
	dateLeft := model.GetRongyaoCardExpire(ctx, userID, typ)
	if dateLeft == 0 {
		// 当日累计充值
		leijiChongzhi := model.GetRongyaoCardAmt(ctx, userID, typ)
		result["leiji"] = leijiChongzhi
		jihuo = 0
	} else {
		// 剩余多少时间
		result["date_left"] = dateLeft
		status := model.GetRongyaoCardLingqu(ctx, userID, typ)
		result["status"] = status
		jihuo = 1
	}
	result["jihuo"] = jihuo

	return c.ResponseSuccessToMe(result)
}

// YuekalingquAction 荣耀月卡和至尊月卡领取
func (c *ShinelightController) YuekalingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	typ := c.Params.GetIntE("type") // 1-荣耀月卡 2-至尊月卡
	if typ < 1 || typ > 2 {
		return c.ResponseError(649421, "月卡错误")
	}

	// 获取月卡是否被激活/到期时间
	dateLeft := model.GetRongyaoCardExpire(ctx, userID, typ)
	if dateLeft == 0 {
		return c.ResponseError(6494211, "月卡未激活")
	}

	// 奖励
	var reward []util.TypeNum
	if typ == 1 {
		reward = []util.TypeNum{{Type: 2, Num: 120}}
	} else {
		reward = []util.TypeNum{{Type: 2, Num: 200}}
	}

	// 给予奖励

	model.GiveReward(userID, reward...)

	// 设置领取状态
	model.SetRongyaoCardLingqu(ctx, userID, typ)

	return c.ResponseSuccessToMe(types.Map{
		"reward":    reward,
		"date_left": dateLeft,
	})
}

// ChongzhiListAction 充值列表
func (c *ShinelightController) ChongzhiListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)

	userVipContents := model.GetUserVipContents(userID)

	firstChong := userVipContents.GetMapE("first_chong")
	list := make([]types.Map, 0, len(logic.ChongzhiList))
	for _, v := range logic.ChongzhiList {
		sendDouble := 1 - firstChong.GetIntE(v)
		list = append(list, types.Map{"money": v, "send_double": sendDouble})
	}

	// 累计充值
	leijiChong := userVipContents.GetIntE("leiji_xiaofei")
	curVipLevel := userGrade.GetIntE("vip_level")
	nextNeed := 0
	if cfg, ok := logic.VipConfigs[curVipLevel+1]; ok {
		nextNeed = cfg["rmb_tatal"]
	}

	return c.ResponseSuccessToMe(types.Map{
		"vip_level":      curVipLevel,
		"next_vip_level": curVipLevel + 1,
		"leiji_chong":    leijiChong,
		"next_need":      nextNeed,
		"need":           nextNeed - leijiChong,
		"list":           list,
	})
}

// ShouchongAction 6元/100元首充
func (c *ShinelightController) ShouchongAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	typ := c.Params.GetIntE("type") // 6-首充6元 100-首充100元
	if typ != 6 && typ != 100 {
		return c.ResponseError(23252, "当天已经领取")
	}

	result, _ := getShouchongData(ctx, userID, typ)
	return c.ResponseSuccessToMe(result)
}

// ShouchonglingquAction 首充领取
func (c *ShinelightController) ShouchonglingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	typ := c.Params.GetIntE("type") // 6-首充6元 100-首充100元
	if typ != 6 && typ != 100 {
		return c.ResponseError(23252, "当天已经领取")
	}

	result, userVipContent := getShouchongData(ctx, userID, typ)

	status := result.GetIntE("status")
	if status == 1 { // 可以领取
		lingTimes := result.GetIntE("ling_times")
		rewardDayList := result["reward_day"].([][]util.TypeNum)
		if lingTimes >= len(rewardDayList) {
			return c.ResponseError(23252, "当天已经领取")
		}
		rewards := rewardDayList[lingTimes]

		model.GiveReward(userID, rewards...)

		// 更新 VIP 内容
		shouChong := userVipContent.GetMapE("shou_chong")
		typData := shouChong.GetMapE(typ)
		typData["ling_date"] = types.ToIntE(time.Now().Format("20060102"))
		typData["ling_times"] = lingTimes + 1
		shouChong[types.ToString(typ)] = typData
		userVipContent["shou_chong"] = shouChong
		model.UpdateUserVipContents(userID, userVipContent)

		return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
	}

	return c.ResponseError(23252, "当天已经领取")
}

// getShouchongData 获取首充数据（辅助函数）
func getShouchongData(ctx context.Context, userID int64, num int) (types.Map, types.Map) {

	userVipContent := model.GetUserVipContents(userID)

	result := types.Map{}
	leijiChong := userVipContent.GetIntE("leiji_xiaofei")
	shouChong := userVipContent.GetMapE("shou_chong")
	typData := shouChong.GetMapE(num)

	lingTimes := typData.GetIntE("ling_times")
	lastLingquDate := typData.GetIntE("ling_date")

	result["leiji_chong"] = leijiChong
	result["ling_times"] = lingTimes
	result["last_lingqu_date"] = lastLingquDate

	status := 0 // 0-不满足 1-可以领取 2-不可领取
	todayInt := types.ToIntE(time.Now().Format("20060102"))
	if leijiChong >= num {
		if lingTimes < 3 && lastLingquDate != todayInt {
			status = 1
		} else {
			status = 2
		}
	}
	result["status"] = status

	// 奖励列表
	rewardDay := make([]interface{}, 0, 3)
	if num == 6 {
		rewardDay = append(rewardDay, logic.Shouchong6Day1)
		rewardDay = append(rewardDay, logic.Shouchong6Day2)
		rewardDay = append(rewardDay, logic.Shouchong6Day3)
	} else {
		rewardDay = append(rewardDay, logic.Shouchong100Day1)
		rewardDay = append(rewardDay, logic.Shouchong100Day2)
		rewardDay = append(rewardDay, logic.Shouchong100Day3)
	}
	result["reward_day"] = rewardDay

	return result, userVipContent
}

// DayShouchongAction 每日180钻石首充
func (c *ShinelightController) DayShouchongAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	result, _ := getDayShouchongData(ctx, userID)
	return c.ResponseSuccessToMe(result)
}

// DayShouchonglingquAction 每日首充领取
func (c *ShinelightController) DayShouchonglingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	result, userVipContentDay := getDayShouchongData(ctx, userID)

	status := result.GetIntE("status")
	if status == 1 { // 可以领取
		rewards := result["rewards"].([]util.TypeNum)

		model.GiveReward(userID, rewards...)

		userVipContentDay["chong_180_ling"] = 1
		model.SetVipContentsCurDay(ctx, userID, userVipContentDay)

		return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
	}

	return c.ResponseError(23252, "当天已经领取")
}

// getDayShouchongData 获取每日首充数据（辅助函数）
func getDayShouchongData(ctx context.Context, userID int64) (types.Map, types.Map) {
	userVipContentDay := model.GetVipContentsCurDay(ctx, userID)

	result := types.Map{}
	leijiChong := userVipContentDay.GetIntE("leiji_xiaofei")
	result["leiji_chong"] = leijiChong

	status := 0 // 0-不满足 1-未领取 2-已经领取
	if leijiChong >= 18 {
		if userVipContentDay.GetIntE("chong_180_ling") != 1 {
			status = 1
		} else {
			status = 2
		}
	}
	result["status"] = status
	result["rewards"] = logic.DayShouchong180Day1

	return result, userVipContentDay
}

// JijinAction 成长基金
func (c *ShinelightController) JijinAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userVipContent := model.GetUserVipContents(userID)

	jijin := userVipContent.GetMapE("jijin")
	jijinStatus := jijin.GetIntE("status") // 0-未购买 1-已购买
	jijinLingqu := jijin.GetMapE("lingqu")

	result := types.Map{
		"status": jijinStatus,
		"lingqu": []types.Map{},
	}

	// 基金等级列表（10,20,30,...,150）
	jijinLvs := []int{10, 20, 30, 40, 50, 60, 70, 80, 90, 100, 110, 120, 130, 140, 150}

	if jijinStatus != 0 {
		userGrade := model.GetUserAttr(userID)
		lv := userGrade.GetIntE("lv")
		result["lv"] = lv

		lingquList := make([]types.Map, 0, len(jijinLvs))
		for _, jijinLv := range jijinLvs {
			if lv < jijinLv {
				lingquList = append(lingquList, types.Map{"lv": jijinLv, "status": 0})
			} else if jijinLingqu.GetIntE(jijinLv) != 0 {
				lingquList = append(lingquList, types.Map{"lv": jijinLv, "status": 2})
			} else {
				lingquList = append(lingquList, types.Map{"lv": jijinLv, "status": 1})
			}
		}
		result["lingqu"] = lingquList
	} else {
		lingquList := make([]types.Map, 0, len(jijinLvs))
		for _, jijinLv := range jijinLvs {
			lingquList = append(lingquList, types.Map{"lv": jijinLv, "status": 0})
		}
		result["lingqu"] = lingquList
	}

	return c.ResponseSuccessToMe(result)
}

// JijinLingquAction 成长基金领取
func (c *ShinelightController) JijinLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	lingquLv := c.Params.GetIntE("lv")

	userVipContent := model.GetUserVipContents(userID)

	jijin := userVipContent.GetMapE("jijin")
	jijinStatus := jijin.GetIntE("status")
	if jijinStatus == 0 {
		return c.ResponseError(1244, "还未购买基金")
	}

	rewards, ok := logic.JijinRewards[lingquLv]
	if !ok || len(rewards) == 0 {
		return c.ResponseError(7455, "输入等级错误")
	}

	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")
	if lv < lingquLv {
		return c.ResponseError(5347, "等级不够")
	}

	jijinLingqu := jijin.GetMapE("lingqu")
	if jijinLingqu.GetIntE(lingquLv) != 0 {
		return c.ResponseError(3256, "已经领取")
	}

	// 给予奖励
	model.GiveReward(userID, rewards...)

	// 标记已领取
	jijinLingqu[types.ToString(lingquLv)] = 1
	jijin["lingqu"] = jijinLingqu
	userVipContent["jijin"] = jijin
	model.UpdateUserVipContents(userID, userVipContent)

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

// YueduLibaoAction 月度超值礼包
func (c *ShinelightController) YueduLibaoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	_, monEndTime := logic.CurrentMonthZhouqi()
	userVipContent := model.GetVipContentsLimitMon(ctx, userID, 0)
	yueduLibao := userVipContent.GetMapE("yuedu_libao")
	limit := yueduLibao.GetMapE("limit")

	ret := make([]types.Map, 0, len(logic.YueduLibaoLimit))
	for k, v := range logic.YueduLibaoLimit {
		num := v - limit.GetIntE(k)
		if num < 0 {
			num = 0
		}
		ret = append(ret, types.Map{"price": k, "num": num})
	}

	now := time.Now().Unix()
	leftTime := monEndTime - now

	return c.ResponseSuccessToMe(types.Map{
		"end_time":  monEndTime,
		"left_time": leftTime,
		"limits":    ret,
	})
}

// DayLibaoAction 每日礼包
func (c *ShinelightController) DayLibaoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userVipContent := model.GetVipContentsCurDay(ctx, userID)
	dayLibao := userVipContent.GetMapE("day_libao")
	limit := dayLibao.GetMapE("limit")

	ret := make([]types.Map, 0, len(logic.DayLibaoLimit))
	for k, v := range logic.DayLibaoLimit {
		num := v - limit.GetIntE(k)
		if num < 0 {
			num = 0
		}
		ret = append(ret, types.Map{"price": k, "num": num})
	}

	return c.ResponseSuccessToMe(types.Map{"limits": ret})
}

// WeekLibaoAction 每周限购礼包
func (c *ShinelightController) WeekLibaoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	_, weekEndTime := logic.CurrentWeekZhouqi()
	userVipContent := model.GetVipContentsLimitWeek(ctx, userID, 0)
	weekLibao := userVipContent.GetMapE("week_libao")
	limit := weekLibao.GetMapE("limit")

	ret := make([]types.Map, 0, len(logic.WeekLibaoLimit))
	for k, v := range logic.WeekLibaoLimit {
		num := v - limit.GetIntE(k)
		if num < 0 {
			num = 0
		}
		ret = append(ret, types.Map{"price": k, "num": num})
	}

	now := time.Now().Unix()
	leftTime := weekEndTime - now

	return c.ResponseSuccessToMe(types.Map{
		"end_time":  weekEndTime,
		"left_time": leftTime,
		"limits":    ret,
	})
}

// QianggouAction 礼包抢购
func (c *ShinelightController) QianggouAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	_, weekEndTime := logic.CurrentWeekZhouqi()
	userVipContent := model.GetVipContentsLimitWeek(ctx, userID, 0)
	qianggou := userVipContent.GetMapE("qianggou")
	limit := qianggou.GetMapE("limit")

	ret := make([]types.Map, 0, len(logic.LibaoQianggouLimit))
	for k, v := range logic.LibaoQianggouLimit {
		num := v - limit.GetIntE(k)
		if num < 0 {
			num = 0
		}
		ret = append(ret, types.Map{"price": k, "num": num})
	}

	now := time.Now().Unix()
	leftTime := weekEndTime - now

	return c.ResponseSuccessToMe(types.Map{
		"end_time":  weekEndTime,
		"left_time": leftTime,
		"limits":    ret,
	})
}

// YiyuanlibaoAction 一元礼包
func (c *ShinelightController) YiyuanlibaoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userVipContent := model.GetUserVipContents(userID)
	yiyuanlibao := userVipContent.GetMapE("yiyuanlibao")

	now := time.Now().Unix()
	timestamp := yiyuanlibao.GetInt64E("timestamp")
	buyFlag := yiyuanlibao.GetIntE("buy_flag")

	ret := types.Map{}
	if now > timestamp+3600 || buyFlag == 1 {
		ret["flag"] = 0 // 不可以购买
	} else {
		ret["flag"] = 1 // 可以购买
		ret["left_time"] = timestamp + 3600 - now
		ret["rewards"] = logic.YiyuanLibaoRewards
	}

	return c.ResponseSuccessToMe(ret)
}

// AnychonglibaoAction 任意充值
func (c *ShinelightController) AnychonglibaoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userVipContent := model.GetUserVipContents(userID)
	anychong := userVipContent.GetMapE("anychong")

	now := time.Now().Unix()
	timestamp := anychong.GetInt64E("timestamp")

	ret := types.Map{
		"status":    anychong.GetIntE("status"), // 0-未购买 1-已购买 2-已领取
		"left_time": timestamp + 7200 - now,
		"rewards":   logic.AnychongRewards,
	}

	return c.ResponseSuccessToMe(ret)
}

// AnychonglingquAction 任意充值领取
func (c *ShinelightController) AnychonglingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userVipContent := model.GetUserVipContents(userID)
	anychong := userVipContent.GetMapE("anychong")

	anychongStatus := anychong.GetIntE("status")
	if anychongStatus == 0 {
		return c.ResponseError(4235, "还未购买")
	}
	if anychongStatus == 2 {
		return c.ResponseError(4326, "已经领取")
	}

	// 标记已领取
	anychong["status"] = 2
	userVipContent["anychong"] = anychong
	model.UpdateUserVipContents(userID, userVipContent)

	// 给予奖励
	model.GiveReward(userID, logic.AnychongRewards...)

	return c.ResponseSuccessToMe(types.Map{"rewards": logic.AnychongRewards})
}

// TequanShopAction 特权商城
func (c *ShinelightController) TequanShopAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	ret := types.Map{}

	// 先知礼包本周限购剩余个数
	weekContent := model.GetVipContentsCurWeek(ctx, userID)
	xianzhi := weekContent.GetMapE("xianzhi")
	xianLimit := xianzhi.GetIntE("limit")
	if xianLimit > 0 {
		ret["xian_left_num"] = 0
	} else {
		ret["xian_left_num"] = 1
	}

	// 快速作战本月剩余个数
	monContent := model.GetVipContentsCurMon(ctx, userID)
	kuaisu := monContent.GetMapE("kuaisu")
	kuaisuLimit := kuaisu.GetIntE("limit")
	if kuaisuLimit > 0 {
		ret["kuaisu_left_num"] = 0
	} else {
		ret["kuaisu_left_num"] = 1
	}

	// 远航特权

	userVipContent := model.GetUserVipContents(userID)
	ret["voyage_high_vip"] = userVipContent.GetIntE("voyage_high_vip") // 0-未购买 1-已购买
	ret["voyage_hao_vip"] = userVipContent.GetIntE("voyage_hao_vip")   // 0-未购买 1-已购买

	return c.ResponseSuccessToMe(ret)
}

// TequanBuyAction 特权商城-远航高级特权/远航豪华特权购买
func (c *ShinelightController) TequanBuyAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	typ := c.Params.GetIntE("type") // 3-远航高级特权 4-远航豪华特权

	userVipContent := model.GetUserVipContents(userID)

	var key string
	if typ == 3 {
		key = "voyage_high_vip"
		if userVipContent.GetIntE(key) != 0 {
			return c.ResponseError(666231, "远航高级特权已经购买")
		}
		userVipContent[key] = 1
	} else if typ == 4 {
		key = "voyage_hao_vip"
		if userVipContent.GetIntE(key) != 0 {
			return c.ResponseError(666231, "远航高级特权已经购买")
		}
		userVipContent[key] = 1
	} else {
		return c.ResponseError(666231, "type error")
	}

	// 价格从配置获取
	shopShell, ok := logic.TequanShopShell[typ]
	if !ok || len(shopShell) == 0 {
		return c.ResponseError(666231, "type error")
	}
	price := shopShell[0].Num

	if item.NotEnough(userID, 2, price) {
		return c.ResponseError(666666, "钻石不够")
	}

	// 给予奖励
	model.GiveReward(userID, logic.TequanShop[typ]...)

	// 减少货币
	item.Sub(userID, 2, price)

	// 更新 VIP 内容
	model.UpdateUserVipContents(userID, userVipContent)

	return c.ResponseSuccessToMe(types.Map{"rewards": logic.TequanShop[typ]})
}

// MonJijinAction 月基金
func (c *ShinelightController) MonJijinAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	typ := c.Params.GetIntE("type") // 128, 328

	ret := types.Map{}
	// 至尊月卡状态
	zhizun := model.GetRongyaoCardExpire(ctx, userID, 2)
	if zhizun == 0 {
		ret["zhizun"] = 0
	} else {
		ret["zhizun"] = 1
	}

	userVipContent := model.GetUserVipContents(userID)
	monJijin := userVipContent.GetMapE("mon_jijin")
	jijin := monJijin.GetMapE(typ)

	jijinStatus := jijin.GetIntE("status")
	jijinDays := jijin.GetIntE("days")
	lastLingDate := jijin.GetIntE("last_ling_date")

	ret["status"] = jijinStatus // 0-未购买 1-已购买
	ret["days"] = jijinDays
	ret["lingqu"] = 0

	todayInt := types.ToIntE(time.Now().Format("20060102"))
	if jijinStatus != 0 && lastLingDate >= todayInt {
		ret["lingqu"] = 1 // 今日已领取
	}

	// 奖励（days+1 天的奖励）
	nextDay := jijinDays + 1
	if typ == 128 {
		ret["rewards"] = logic.Yue128JijinReward[nextDay]
	} else {
		ret["rewards"] = logic.Yue328JijinReward[nextDay]
	}

	return c.ResponseSuccessToMe(ret)
}

// MonJijinLingAction 月基金领取
func (c *ShinelightController) MonJijinLingAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	typ := c.Params.GetIntE("type") // 128, 328

	userVipContent := model.GetUserVipContents(userID)
	monJijin := userVipContent.GetMapE("mon_jijin")
	jijin := monJijin.GetMapE(typ)

	jijinStatus := jijin.GetIntE("status")
	jijinDays := jijin.GetIntE("days")
	lastLingDate := jijin.GetIntE("last_ling_date")
	todayInt := types.ToIntE(time.Now().Format("20060102"))

	lingqu := 0
	if jijinStatus != 0 && lastLingDate >= todayInt {
		lingqu = 1 // 今日已领取
	}

	// 奖励（days+1 天的奖励）
	nextDay := jijinDays + 1
	var rewards []util.TypeNum
	if typ == 128 {
		rewards = logic.Yue128JijinReward[nextDay]
	} else {
		rewards = logic.Yue328JijinReward[nextDay]
	}

	// 可以领取
	if jijinStatus != 0 && lingqu == 0 {
		jijinDays++
		jijin["days"] = jijinDays
		jijin["last_ling_date"] = todayInt

		// 给予奖励
		model.GiveReward(userID, rewards...)

		// 全部领完（30天），重置状态
		if jijinDays == 30 {
			jijin["status"] = 0
			jijin["days"] = 0
			jijin["last_ling_date"] = 0
		}
		monJijin[types.ToString(typ)] = jijin
		userVipContent["mon_jijin"] = monJijin
		model.UpdateUserVipContents(userID, userVipContent)
	}

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}
