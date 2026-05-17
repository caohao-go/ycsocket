package controller

import (
	"context"
	"fmt"

	"server_golang/common/lock"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/repo/mem/item"

	"server_golang/common/types"
	"server_golang/logic"
	"server_golang/model"
)

// 副本系统

func (c *ShinelightController) CopyInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")
	copyLv := userGrade.GetIntE("copy")

	// 获取副本配置（与 PHP copyInfoAction 一致）
	openLv := logic.GetOpenLv(copyLv)
	status := 1
	if openLv > lv {
		status = 0
	}

	return c.ResponseSuccessToMe(types.Map{
		"lv": lv, "copy": copyLv, "status": status, "open_lv": openLv,
	})
}

// 获取上阵英雄信息
func (c *ShinelightController) GetFightHerosAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	typ := c.Params.GetStringE("type")
	if typ == "" {
		typ = "copy_fight"
	}

	fightHeros := model.GetFightHeros(ctx, userID, typ)

	data := types.Map{}

	// 远征模式：读取所有英雄HP（对齐 PHP getFightHerosAction 中 type=='expedition' 的逻辑）
	var startHP = map[int]int{}
	if typ == "expedition" {
		startHP = model.GetAllExpeditionHerosHP(ctx, userID)
		data["all_hero_hp"] = startHP
	}

	if fightHeros != nil {
		data["position"] = fightHeros.Position
		if len(fightHeros.Heros) > 0 {
			heroIDs := make([]int, 0, len(fightHeros.Heros))
			heroPos := make(map[int]int, len(fightHeros.Heros))
			for heroID, pos := range fightHeros.Heros {
				if heroID > 0 {
					heroIDs = append(heroIDs, heroID)
					heroPos[heroID] = pos
				}
			}

			if len(heroIDs) > 0 {
				heroAttrs := model.GetUserHeroAttrByIDs(ctx, heroIDs, userID, true)
				heros := make([]types.Map, 0)
				totalFP := 0
				for _, attr := range heroAttrs {
					id := attr.Id
					currentHP := attr.Hp
					// 远征模式：如果有持久化HP且小于满血值，则使用持久化HP（对齐 PHP）
					if typ == "expedition" {
						savedHP := startHP[id]
						if savedHP > 0 && savedHP < currentHP {
							currentHP = savedHP
						}
					}
					hero := types.Map{
						"id":          id,
						"hero_id":     attr.HeroInfo,
						"star":        attr.Star,
						"stage":       attr.Stage,
						"lv":          attr.Lv,
						"pos":         heroPos[id],
						"hp":          attr.Hp,
						"current_hp":  currentHP,
						"fight_point": attr.FightPoint,
					}
					heros = append(heros, hero)
					totalFP += attr.FightPoint
				}
				data["heros"] = heros
				data["total_fight_point"] = totalFP
			}
		}
	}

	return c.ResponseSuccessToMe(types.Map{"data": data})
}

// Boss 战斗（copyFight）— 与 PHP copyFightAction 完全对齐
func (c *ShinelightController) CopyFightAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	lock.Lock(fmt.Sprintf("copyFight%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("copyFight%d", userID))

	userGrade := model.GetUserAttr(userID)
	beforeLv := userGrade.GetIntE("lv")
	copyLv := userGrade.GetIntE("copy")

	result := types.Map{
		"lv":   beforeLv,
		"copy": copyLv, // PHP 返回初始 copy 值
	}

	// 获取副本配置（与 PHP 一致：status 和 open_lv 字段）
	openLv := logic.GetOpenLv(copyLv)
	status := 1
	if openLv > beforeLv {
		status = 0
	}
	result["status"] = status
	result["open_lv"] = openLv

	// Boss 奖励
	bossReward := logic.GetCopyBossReward(copyLv)

	if status == 0 {
		// PHP 在 status==0 时直接返回错误
		return c.ResponseError(666666, fmt.Sprintf("需要%d级开启", openLv))
	}

	// 获取玩家英雄（阵位信息+完整属性+技能）
	fightHerosParam := c.Params.GetStringE("fight_heros")
	position := c.Params.GetIntE("position")

	// 获取战斗英雄详情
	var myHeroAttrs []*logic.Hero
	if fightHerosParam != "" {
		heroIDs, heroPos := util.ToPosHeros(fightHerosParam)
		heroDetailsMap := model.GetUserHeroAttrWithSkillByIDs(ctx, heroIDs, userID, true)
		for _, heroDetail := range heroDetailsMap {
			if pos, ok := heroPos[heroDetail.Id]; ok {
				heroDetail.Pos = pos
				myHeroAttrs = append(myHeroAttrs, heroDetail)
			}
		}
		// 保存上阵英雄配置
		model.SetFightHeros(ctx, userID, "copy_fight", heroPos, position)
	} else {
		// 从已保存阵位获取
		myHeroAttrs = model.GetUserPositionWithHeroAttrs(ctx, userID, 1)
	}

	// 获取怪物信息
	copyData := logic.CopyDataMap[copyLv]
	var oppHeroAttrs []*logic.Hero
	if copyData != nil {
		oppHeroAttrs = logic.BuildBossDetail(logic.MonsterDatas, copyData.Monster)
	}

	fight := logic.NewFight(myHeroAttrs, oppHeroAttrs)
	winner, fightResult := fight.FightExec(15)

	// 从战斗结果提取胜负
	success := 0
	if winner == "P1" {
		success = 1
	}

	// 生成精简英雄展示数据
	myHero := logic.GetBaseFromHero(myHeroAttrs)
	oppHero := logic.GetBaseFromHero(oppHeroAttrs)

	if success == 1 {
		// 通关奖励
		oldRewards := logic.GetCopyRewards(copyLv, 0)
		newRewards := logic.GetCopyRewards(copyLv+1, 0)
		hookReward := types.Map{}
		if len(oldRewards["min"]) > 0 {
			hookReward["old_1"] = oldRewards["min"][0].Num
		}
		if len(newRewards["min"]) > 0 {
			hookReward["new_1"] = newRewards["min"][0].Num
		}
		if len(oldRewards["min"]) > 1 {
			hookReward["old_3"] = oldRewards["min"][1].Num
		}
		if len(newRewards["min"]) > 1 {
			hookReward["new_3"] = newRewards["min"][1].Num
		}
		if len(oldRewards["min"]) > 2 {
			hookReward["old_7"] = oldRewards["min"][2].Num
		}
		if len(newRewards["min"]) > 2 {
			hookReward["new_7"] = newRewards["min"][2].Num
		}
		if len(oldRewards["min"]) > 3 {
			hookReward["old_6"] = oldRewards["min"][3].Num
		}
		if len(newRewards["min"]) > 3 {
			hookReward["new_6"] = newRewards["min"][3].Num
		}
		result["hook_reward"] = hookReward
		result["reward"] = bossReward

		// 发放奖励
		model.GiveReward(userID, bossReward...)

		// 关卡+1
		model.IncrUserCopy(userID)
		userGrade = model.GetUserAttr(userID)
		afterLv := userGrade.GetIntE("lv")

		newCopy := userGrade.GetIntE("copy")
		model.SetRankScore(ctx, config.RankCopy, userID, float64(newCopy), 0)

		addZhuan := 0
		addLv := afterLv - beforeLv
		if addLv > 0 {
			addZhuan = (afterLv/10 + 1) * 10 * addLv
			item.AddZuan(userID, addZhuan)
		}

		// 成就任务
		model.AchieveTaskHandle(ctx, userID, 1, afterLv, 1000, 1038)

		result["after_lv"] = afterLv
		result["add_zhuan"] = addZhuan
		result["curent_exp"] = item.Exp(userID)
		result["next_exp"] = logic.GetLvUpdateExp(afterLv)
	}

	result["success"] = success
	result["my_hero"] = myHero
	result["opp_hero"] = oppHero
	result["fight_result"] = fightResult

	// 引导任务（与 PHP 一致：对所有 task ID 调用 guideTaskHandle，由内部判断是否触发）
	// PHP: $new_copy = $result['success'] == 1 ? $result['copy'] + 1 : $result['copy'];
	// 注意：PHP 的 $result['copy'] 是初始值，所以 new_copy = copyLv+1(胜利) 或 copyLv(失败)
	newCopy := copyLv
	if success == 1 {
		newCopy = copyLv + 1
	}
	if newCopy <= 51 {
		copyGuideTaskIDs := []int{3, 7, 13, 16, 18, 25, 29, 34, 39, 44, 53, 56, 63, 67, 70, 73, 76, 78, 80, 84, 89, 91, 94, 96, 98, 100, 102, 105, 109, 112}
		for _, tid := range copyGuideTaskIDs {
			model.GuideTaskHandle(ctx, userID, tid, newCopy)
		}
	}

	// 日常任务（仅胜利时）
	if success == 1 {
		model.SetDailyTaskFinish(ctx, userID, 10001, 1)
	}

	return c.ResponseSuccessToMe(result)
}

// 获取副本奖励领取状态（与 PHP copyRewardStatAction 一致）
func (c *ShinelightController) CopyRewardStatAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	currentLayer := userGrade.GetIntE("copy")

	// 获取已领取的关卡列表（数组类型 content）
	copyRewardsStatArr := model.GetUsersContentArray(ctx, userID, "copy_rewards_stat")
	claimedSet := make(map[int]bool)
	for _, v := range copyRewardsStatArr {
		claimedSet[types.ToIntE(v)] = true
	}

	// 遍历 CheckpointRewardDatas[2]（副本关卡奖励），找出可领取的关卡
	lingqu := make([]int, 0)
	if checkpointDatas, ok := logic.CheckpointRewardDatas[2]; ok {
		for checkpoint := range checkpointDatas {
			if checkpoint > currentLayer {
				continue
			}
			if !claimedSet[checkpoint] {
				lingqu = append(lingqu, checkpoint)
			}
		}
	}

	return c.ResponseSuccessToMe(types.Map{
		"copy": currentLayer, "lingqu": lingqu,
	})
}

// 领取副本奖励（与 PHP lingquCopyRewardAction 一致）
func (c *ShinelightController) LingquCopyRewardAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	copyParam := c.Params.GetIntE("copy")

	lock.Lock(fmt.Sprintf("lingquCopyReward%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("lingquCopyReward%d", userID))

	// 检查关卡是否达成
	userGrade := model.GetUserAttr(userID)
	currentCopy := userGrade.GetIntE("copy")
	if copyParam > currentCopy {
		return c.ResponseError(882, "未达成关卡")
	}

	// 获取已领取列表（数组类型）
	claimedList := model.GetUsersContentArray(ctx, userID, "copy_rewards_stat")

	// 检查是否已领取
	for _, v := range claimedList {
		if types.ToIntE(v) == copyParam {
			return c.ResponseError(402, "已领取奖励")
		}
	}

	// 从 CheckpointRewardDatas[2] 查找对应关卡奖励
	var rewards []util.TypeNum
	if checkpointDatas, ok := logic.CheckpointRewardDatas[2]; ok {
		if data, ok := checkpointDatas[copyParam]; ok {
			rewards = data
		}
	}

	if len(rewards) > 0 {
		// 更新已领取列表
		if claimedList == nil {
			claimedList = []interface{}{copyParam}
		} else {
			claimedList = append(claimedList, copyParam)
		}
		model.UpdateUsersContentArray(ctx, userID, "copy_rewards_stat", claimedList)

		// 发放奖励

		model.GiveReward(userID, rewards...)
	}

	// 引导任务（与 PHP 一致：copy <= 21 时触发）
	if copyParam <= 21 {
		copyRewardGuideIDs := []int{4, 19, 45, 71, 85}
		for _, tid := range copyRewardGuideIDs {
			model.GuideTaskHandle(ctx, userID, tid, copyParam)
		}
	}

	return c.ResponseSuccessToMe(types.Map{"success": 1})
}

// 领取次数奖励
func (c *ShinelightController) GetCountRewardAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	copyID := c.Params.GetIntE("copy_id")
	seq := c.Params.GetIntE("reward_seq")

	lock.Lock(fmt.Sprintf("getCountReward%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("getCountReward%d", userID))

	if copyID == 0 {
		return c.ResponseError(321411, "copy_id is empty")
	}

	finish := model.GetRewardFinish(ctx, userID, copyID, seq)
	if finish == 0 {
		return c.ResponseError(321523, "任务未完成")
	}

	status := model.GetRedisRewardStatus(ctx, userID, copyID, seq)
	if status > 0 {
		return c.ResponseError(321556, "任务已领取")
	}

	model.SetRewardStatus(ctx, userID, copyID, seq)

	rewardDatas := logic.GetRewardByCopyID(ctx, copyID)
	if rd, ok := rewardDatas[seq]; ok {
		model.GiveReward(userID, rd.Reward...)
	}

	return c.ResponseSuccessToMe(types.Map{})
}
