package controller

import (
	"context"
	"fmt"
	"math"
	"time"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// 挂机系统

func (c *ShinelightController) OnHookAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 锁 key 对齐 PHP：on_hook{uid}
	lock.Lock(fmt.Sprintf("on_hook%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("on_hook%d", userID))

	userGrade := model.GetUserAttr(userID)
	timeUp := logic.GetVipInfo(userGrade, "offline_time")
	onHookMin, onHookHour := model.GetOnHookTime(ctx, userID, timeUp, false)

	copyLv := userGrade.GetIntE("copy")
	vipLv := userGrade.GetIntE("vip_level")
	rewardData := logic.GetCopyRewards(copyLv, vipLv)
	rewardRandData := logic.GetRandRewards(copyLv, onHookHour, onHookMin)

	rewards := make([]util.TypeNum, 0)
	if minRewards, ok := rewardData["min"]; ok {
		for _, r := range minRewards {
			if onHookMin > 0 {
				rewards = append(rewards, util.TypeNum{
					Type: r.Type, Num: r.Num * onHookMin,
				})
			}
		}
	}
	if hourRewards, ok := rewardData["hour"]; ok {
		for _, r := range hourRewards {
			if onHookHour > 0 {
				rewards = append(rewards, util.TypeNum{
					Type: r.Type, Num: r.Num * onHookHour,
				})
			}
		}
	}

	dataHero := model.GetUserPositionByID(ctx, userID, 1)

	return c.ResponseSuccessToMe(types.Map{
		"on_hook_min": onHookMin, "rewards": rewards,
		"reward_data": rewardData, "reward_rand_data": rewardRandData,
		"data_hero": dataHero,
	})
}

// 领取挂机奖励
func (c *ShinelightController) HookLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 锁 key 对齐 PHP：hook_lingqu{uid}
	lock.Lock(fmt.Sprintf("hook_lingqu%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("hook_lingqu%d", userID))

	userGrade := model.GetUserAttr(userID)
	exp := item.Exp(userID)
	beforeLv := userGrade.GetIntE("lv")

	result := types.Map{
		"lv":       beforeLv,
		"exp":      exp,
		"need_exp": logic.GetLvUpdateExp(beforeLv),
	}

	timeUp := logic.GetVipInfo(userGrade, "offline_time")
	onHookMin, onHookHour := model.GetOnHookTime(ctx, userID, timeUp, true)

	copyLv := userGrade.GetIntE("copy")
	vipLv := userGrade.GetIntE("vip_level")
	rewardData := logic.GetCopyRewards(copyLv, vipLv)
	// 随机奖励
	rewardRandData := logic.GetRandRewards(copyLv, onHookHour, onHookMin)

	rewards := make([]util.TypeNum, 0)
	var voyage int
	for _, r := range rewardData["min"] {
		if onHookMin > 0 {
			num := r.Num * onHookMin
			rewards = append(rewards, util.TypeNum{
				Type: r.Type, Num: num,
			})
			if r.Type == 6 {
				voyage = num
			}
		}
	}
	for _, r := range rewardData["hour"] {
		if onHookHour > 0 {
			rewards = append(rewards, util.TypeNum{
				Type: r.Type, Num: r.Num * onHookHour,
			})
		}
	}

	result["on_hook_min"] = onHookMin
	result["rewards"] = rewards
	result["reward_data"] = rewardData
	result["reward_rand_data"] = rewardRandData

	// 给予挂机的随机掉落奖励
	model.GiveReward(userID, rewardRandData...)
	// 给予挂机固定奖励
	// 远航道具上限
	voyageNum := item.Total(userID, 6)
	if voyageNum > 22000 {
		for k, r := range rewards {
			if r.Type == 6 {
				rewards[k].Num = 0
			}
		}
	}
	if voyageNum+voyage > 22000 {
		for k, r := range rewards {
			if r.Type == 6 {
				rewards[k].Num = 22000 - voyageNum
			}
		}
	}

	model.GiveReward(userID, rewards...)

	userGrade = model.GetUserAttr(userID)
	afterLv := userGrade.GetIntE("lv")
	addLv := afterLv - beforeLv
	addZhuan := 0
	if addLv > 0 {
		// 与 PHP 一致：$add_zhuan = ceil($user_grade['lv'] / 10) * 10 * $add_lv
		addZhuan = int(math.Ceil(float64(afterLv)/10.0)) * 10 * addLv
		item.AddZuan(userID, addZhuan)
	}

	model.GuideTaskHandle(ctx, userID, 14, 1)
	model.AchieveTaskHandle(ctx, userID, 1, afterLv, 1000, 1038)

	// 对齐 PHP：after_exp 使用循环中保存的 rest_exp（而非重新查询）
	result["add_zhuan"] = addZhuan
	result["after_lv"] = afterLv
	result["after_exp"] = item.Exp(userID)
	result["after_need_exp"] = logic.GetLvUpdateExp(afterLv)

	return c.ResponseSuccessToMe(result)
}

// 快速战斗次数信息
func (c *ShinelightController) FastHookCountAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	copyLv := userGrade.GetIntE("copy")
	vipLv := userGrade.GetIntE("vip_level")

	onHookMin := 120
	onHookHour := 2
	rewardData := logic.GetCopyRewards(copyLv, vipLv)

	rewards := make([]util.TypeNum, 0)
	for _, r := range rewardData["min"] {
		rewards = append(rewards, util.TypeNum{
			Type: r.Type, Num: r.Num * onHookMin,
		})
	}
	for _, r := range rewardData["hour"] {
		rewards = append(rewards, util.TypeNum{
			Type: r.Type, Num: r.Num * onHookHour,
		})
	}

	// 随机奖励合并到 rewards
	rewardRandData := logic.GetRandRewards(copyLv, onHookHour, onHookMin)
	for _, r := range rewardRandData {
		rewards = append(rewards, util.TypeNum{Type: r.Type, Num: r.Num})
	}

	// 快速作战本月剩余个数
	vipContent := model.GetVipContentsCurMon(ctx, userID)
	kuaisuLeftNum := 0
	if vc, err := types.ToMap(vipContent["kuaisu"], ""); err == nil && vc != nil {
		kuaisuLeftNum = vc.GetIntE("limit")
	}

	result := types.Map{}

	if kuaisuLeftNum == 0 {
		// 普通用户
		totalCount := 4
		totalFreeCount := 1
		usedCount := model.GetFastUsedCount(ctx, userID)
		freeCount := 0
		buyCount := 0
		if usedCount >= totalFreeCount {
			freeCount = 0
		} else {
			freeCount = totalFreeCount - usedCount
		}
		if usedCount < totalFreeCount {
			buyCount = 3
		} else {
			buyCount = totalCount - usedCount
		}
		result["free_count"] = freeCount
		result["buy_count"] = buyCount
		result["buy_type"] = 2
		switch buyCount {
		case 3:
			result["buy_num"] = 50
		case 2:
			result["buy_num"] = 100
		case 1:
			result["buy_num"] = 200
		}
	} else {
		// 快速战斗特权
		totalCount := 11
		totalFreeCount := 3
		usedCount := model.GetFastUsedCount(ctx, userID)
		freeCount := 0
		buyCount := 0
		if usedCount >= totalFreeCount {
			freeCount = 0
		} else {
			freeCount = totalFreeCount - usedCount
		}
		if usedCount < totalFreeCount {
			buyCount = 8
		} else {
			buyCount = totalCount - usedCount
		}
		result["free_count"] = freeCount
		result["buy_count"] = buyCount
		result["buy_type"] = 2
		switch {
		case buyCount == 8:
			result["buy_num"] = 50
		case buyCount == 7:
			result["buy_num"] = 100
		case buyCount <= 6 && buyCount >= 1:
			result["buy_num"] = 200
		}
	}

	result["rewards"] = rewards
	result["tequan"] = kuaisuLeftNum

	return c.ResponseSuccessToMe(result)
}

// 快速战斗领取
func (c *ShinelightController) FastHookLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 锁 key 对齐 PHP：fast_hook_lingqu{uid}
	lock.Lock(fmt.Sprintf("fast_hook_lingqu%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("fast_hook_lingqu%d", userID))

	userGrade := model.GetUserAttr(userID)
	exp := item.Exp(userID)
	beforeLv := userGrade.GetIntE("lv")

	result := types.Map{
		"lv": beforeLv, "exp": exp,
		"need_exp": logic.GetLvUpdateExp(beforeLv),
	}

	// 快速作战次数检查（与 PHP 一致）
	vipContent := model.GetVipContentsCurMon(ctx, userID)
	kuaisuLeftNum := 0
	if vc, err := types.ToMap(vipContent["kuaisu"], ""); err == nil && vc != nil {
		kuaisuLeftNum = vc.GetIntE("limit")
	}

	needMoney := 0
	var totalCount, totalFreeCount int
	var isTequan bool

	if kuaisuLeftNum == 0 {
		// 普通用户
		totalCount = 4
		totalFreeCount = 1
		isTequan = false
	} else {
		// VIP 特权用户
		totalCount = 11
		totalFreeCount = 3
		isTequan = true
	}

	usedCount := model.GetFastUsedCount(ctx, userID)
	var freeCount, buyCount int
	if !isTequan {
		if usedCount >= totalFreeCount {
			freeCount = 0
		} else {
			freeCount = totalFreeCount - usedCount
		}
		if usedCount < totalFreeCount {
			buyCount = 3
		} else {
			buyCount = totalCount - usedCount
		}
	} else {
		if usedCount >= totalFreeCount {
			freeCount = 0
		} else {
			freeCount = totalFreeCount - usedCount
		}
		if usedCount < totalFreeCount {
			buyCount = 8
		} else {
			buyCount = totalCount - usedCount
		}
	}

	if freeCount == 0 {
		if buyCount <= 0 {
			return c.ResponseError(91975, "次数用完")
		}
		// 计算需要的钻石
		if isTequan {
			switch buyCount {
			case 8:
				needMoney = 50
			case 7:
				needMoney = 100
			default:
				if buyCount >= 1 && buyCount <= 6 {
					needMoney = 200
				}
			}
		} else {
			switch buyCount {
			case 3:
				needMoney = 50
			case 2:
				needMoney = 100
			case 1:
				needMoney = 200
			}
		}
		// 检查钻石是否足够
		if needMoney > 0 {
			if item.NotEnough(userID, 2, needMoney) {
				return c.ResponseError(666666, "货币不够")
			}
		}
	}

	// 新手引导（等级 > 2 才扣次数）
	if beforeLv > 2 {
		model.IncrFastUsedCount(ctx, userID)
	}

	onHookMin := 120
	onHookHour := 2
	copyLv := userGrade.GetIntE("copy")
	vipLv := userGrade.GetIntE("vip_level")
	rewardData := logic.GetCopyRewards(copyLv, vipLv)
	rewardRandData := logic.GetRandRewards(copyLv, 2, 0)

	rewards := make([]util.TypeNum, 0)
	var voyage int
	for _, r := range rewardData["min"] {
		num := r.Num * onHookMin
		rewards = util.Merge(rewards, []util.TypeNum{{Type: r.Type, Num: num}})
		if r.Type == 6 {
			voyage = num
		}
	}
	for _, r := range rewardData["hour"] {
		rewards = util.Merge(rewards, []util.TypeNum{{Type: r.Type, Num: r.Num * onHookHour}})
	}

	result["rewards"] = rewards
	result["reward_rand_data"] = rewardRandData

	// 航海币上限检查（22000）
	voyageNum := item.Total(userID, 6)
	if voyageNum > 22000 {
		for k, v := range rewards {
			if v.Type == 6 {
				rewards[k].Num = 0
			}
		}
	} else if voyageNum+voyage > 22000 {
		for k, v := range rewards {
			if v.Type == 6 {
				rewards[k].Num = 22000 - voyageNum
			}
		}
	}

	// 发放奖励
	model.GiveReward(userID, rewards...)
	// 给予挂机的随机掉落奖励
	model.GiveReward(userID, rewardRandData...)

	userGrade = model.GetUserAttr(userID)
	afterLv := userGrade.GetIntE("lv")
	addLv := afterLv - beforeLv
	addZhuan := 0
	if addLv > 0 {
		// 与 PHP 一致：$add_zhuan = ceil($user_grade['lv'] / 10) * 10 * $add_lv
		addZhuan = int(math.Ceil(float64(afterLv)/10.0)) * 10 * addLv
		item.AddZuan(userID, addZhuan)
	}

	model.AchieveTaskHandle(ctx, userID, 1, afterLv, 1000, 1038)

	// 扣除钻石（与 PHP 一致，在发放奖励之后扣除）
	if needMoney > 0 {
		item.Sub(userID, 2, needMoney)
	}

	// 日常任务: 快速作战（对应 PHP setDalilyTaskFinish($userId, 10005, 1)）
	model.SetDailyTaskFinish(ctx, userID, 10005, 1)
	// 周任务: 快速作战（对应 PHP Task::finishWeekTask($userId, "quick_battle")）
	model.IncrTaskFinishNumStr(ctx, userID, "quick_battle", 7)
	// 快速作战排行分数（对应 PHP incrRankScore(RANK_TYPE_QUICK_BATTLE_SCORE."_".date('w'), $userId, 1, 86400*8)）
	w := int(time.Now().Weekday())
	model.IncrRankScore(ctx, fmt.Sprintf("%s_%d", config.RankQuickBattleScore, w), userID, 1, 86400*8)
	// 引导任务（对应 PHP guildTaskHandle($userId, $finished_guide_id, 12, 1)）
	model.GuideTaskHandle(ctx, userID, 12, 1)

	// 对齐 PHP：after_exp 使用循环中保存的 rest_exp（而非重新查询）
	result["add_zhuan"] = addZhuan
	result["after_lv"] = afterLv
	result["after_exp"] = item.Exp(userID)
	result["after_need_exp"] = logic.GetLvUpdateExp(afterLv)
	result["rewards"] = rewards

	return c.ResponseSuccessToMe(result)
}
