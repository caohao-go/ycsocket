package controller

import (
	"context"
	"fmt"
	"sort"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
	"server_golang/repo/table"
)

// ClimbLingquAction 领取爬塔特殊关卡奖励 — 对齐 PHP climbLingquAction
func (c *ShinelightController) ClimbLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	layer := c.Params.GetIntE("layer")

	lock.Lock(fmt.Sprintf("climb_lingqu%d", userID), 2)
	defer lock.Unlock(fmt.Sprintf("climb_lingqu%d", userID))

	rewards := model.GetClimbLingquRewards(ctx, userID, layer)
	if rewards == nil {
		return c.ResponseError(4678, "没有可领取奖励")
	}

	model.LingquClimbLingquRewards(ctx, userID, layer)

	if len(rewards) > 0 {
		model.GiveReward(userID, rewards...)
	}

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

func (c *ShinelightController) ClimbtowerInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	crossedLayer := model.GetUsersContentInt(ctx, userID, "climbtower_layer")
	rewards := logic.GetClimbtowerRewards(crossedLayer)

	// 补充每层的最快通关和最低战力记录
	for k, v := range rewards {
		layer := v.Layer
		record := model.GetUserClimbtowerRecord(ctx, layer)
		if record != nil {
			rewards[k].LowestFightPoint = record.LowestFightPoint
			rewards[k].FastUserId = record.FastUserId
			rewards[k].FastNickname = record.FastNickname
		} else {
			rewards[k].LowestFightPoint = 0
			rewards[k].FastUserId = 0
			rewards[k].FastNickname = ""
		}
	}

	saodang := model.GetClimbtowerSaodang(ctx, userID)
	rankList := model.GetRankList(ctx, config.RankClimbtower, true, 0, 2)

	// 已领取的奖励
	lingquMap := model.GetAllClimbLingquRewards(ctx, userID)
	lingqu := make([]types.Map, 0)
	for layer, lingquRewards := range lingquMap {
		lingqu = append(lingqu, types.Map{
			"layer":   types.ToIntE(layer),
			"rewards": lingquRewards,
		})
	}

	// 设置红点
	model.SetTodayPoint(userID, 3, 1)

	return c.ResponseSuccessToMe(types.Map{
		"crossed_layer": crossedLayer,
		"rewards":       rewards,
		"saodang":       saodang,
		"rank_list":     rankList,
		"lingqu":        lingqu,
	})
}

// ======================== 战斗辅助方法 ========================

// 根据英雄ID列表获取完整英雄属性，返回 []pk.Hero（战斗专用）
func (c *ShinelightController) getFightHeroByPosMap(ctx context.Context,
	userID int64, posMap map[int]int) []*logic.Hero {
	var heroIDs = []int{}

	for id := range posMap {
		heroIDs = append(heroIDs, id)
	}

	objs := model.GetUserHeroAttrWithSkillByIDs(ctx, heroIDs, userID, true)

	tmp := []*logic.Hero{}

	// 设置 pos（从 fightHeros 参数里解析的阵位）
	for _, obj := range objs {
		if pos, ok := posMap[obj.Id]; ok {
			obj.Pos = pos
		}
		tmp = append(tmp, obj)
	}

	return tmp
}

// setStartHP 设置起始血量（连续战斗时保留上次剩余血量）
func (c *ShinelightController) setStartHP(objs []*logic.Hero, startHP map[int]int) {
	if startHP == nil {
		return
	}
	for i := range objs {
		if hpVal, existsInMap := startHP[objs[i].Id]; existsInMap {
			hpNum := types.ToIntE(hpVal)
			maxHP := objs[i].Hp
			if hpNum > 0 && hpNum < maxHP {
				objs[i].CurrentHP = hpNum
			}
		}
	}
}

// extractHeroIDs 从英雄 Hero 数组中提取所有英雄ID
func (c *ShinelightController) extractHeroIDs(objs []*logic.Hero) []int {
	ids := []int{}
	for _, o := range objs {
		if o.Id > 0 {
			ids = append(ids, o.Id)
		}
	}
	return ids
}

// extractHeroIDsFromMap 从英雄Map数组中提取所有英雄ID（返回[]int64）
// 兼容 id 和 hero_id 两种字段名（敌方NPC用 hero_id，己方英雄用 id）
func (c *ShinelightController) extractHeroIDsFromMap(heros []types.Map) []int64 {
	ids := make([]int64, 0, len(heros))
	for _, h := range heros {
		id := int64(types.ToIntE(h["id"]))
		if id <= 0 {
			id = int64(types.ToIntE(h["hero_id"]))
		}
		if id > 0 {
			ids = append(ids, id)
		}
	}
	return ids
}

// mapToSlice 将 map[int64]types.Map 转为有序的 []types.Map
func (c *ShinelightController) mapToSlice(m map[int64]types.Map) []types.Map {
	ret := make([]types.Map, 0, len(m))
	// 按ID排序保证顺序一致
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		ret = append(ret, m[k])
	}
	return ret
}

// getHelpCount 统计支援英雄数量
func (c *ShinelightController) getHelpCount(userID int64, objs []*logic.Hero) int {
	count := 0
	for _, o := range objs {
		if o.UserId != 0 && o.UserId != userID {
			count++
		}
	}
	return count
}

// CrossClimbtowerAction 爬塔战斗挑战（crossClimbtower）— 完整战斗流程
func (c *ShinelightController) CrossClimbtowerAction(ctx context.Context) *Result {
	userID, _, userInfo, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 加锁 — 对齐 PHP RedisProxy::lock("cross_climbtower" . $userId)
	lock.Lock(fmt.Sprintf("crossClimbtower%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("crossClimbtower%d", userID))

	currentLayer := model.GetUsersContentInt(ctx, userID, "climbtower_layer")
	if currentLayer == 0 {
		currentLayer = 1
	}
	layer := currentLayer + 1

	fightHerosParam := c.Params.GetStringE("fight_heros")
	position := c.Params.GetIntE("position")

	var myHeroAttrs []*logic.Hero
	if fightHerosParam != "" {
		heroIDs, heroPos := util.ToPosHeros(fightHerosParam)

		objs := model.GetUserHeroAttrWithSkillByIDs(ctx, heroIDs, userID, true)
		for _, obj := range objs {
			if pos, ok := heroPos[obj.Id]; ok {
				obj.Pos = pos
			}
			myHeroAttrs = append(myHeroAttrs, obj)
		}

		model.SetFightHeros(ctx, userID, "climb", heroPos, position)
	}

	if len(myHeroAttrs) == 0 {
		return c.ResponseError(6422, "请选择英雄")
	}

	opHerosRaw := logic.GetClimbtowerHeros(layer)
	if opHerosRaw == nil || len(opHerosRaw) == 0 {
		return c.ResponseError(6422, fmt.Sprintf("第%d层数据不存在", layer))
	}

	layerInfo, ok := logic.ClimbtowerData[layer]
	if !ok {
		return c.ResponseError(6422, fmt.Sprintf("第%d层配置不存在", layer))
	}

	oppHeroAttrs := logic.BuildBossDetail(logic.ClimbtowerMonsterDatas, opHerosRaw)

	// 战斗模拟
	fight := logic.NewFight(myHeroAttrs, oppHeroAttrs)
	winner, fightResult := fight.FightExec(15)

	// 从战斗结果提取胜负（参考 CopyFightAction 模式）
	win := 0
	if winner == "P1" {
		win = 1
	}

	data := types.Map{
		"layer":        layer,
		"my_hero":      myHeroAttrs,
		"opp_hero":     oppHeroAttrs,
		"fight_result": fightResult,
		"position":     layerInfo.Position,
		"combination":  layerInfo.Combination,
		"win":          win,
	}

	if win == 1 {
		currentLayer := model.GetUsersContentInt(ctx, userID, "climbtower_layer")
		if layer > currentLayer {
			model.IncrUsersContentInt(ctx, userID, "climbtower_layer", 1)
			model.SetRankScore(ctx, config.RankClimbtower, userID, float64(layer), 0)
		}
		// 给予奖励 — datas[1] 通关奖励 + datas[7] 首通奖励，合并后返回
		var allRewards []util.TypeNum
		if rd, ok := logic.CheckpointRewardDatas[1]; ok {
			if layerReward, ok := rd[layer]; ok {
				model.GiveReward(userID, layerReward...)
				allRewards = append(allRewards, layerReward...)
			}
		}
		if rd7, ok := logic.CheckpointRewardDatas[7]; ok {
			if layerReward7, ok := rd7[layer]; ok {
				model.GiveReward(userID, layerReward7...)
				allRewards = append(allRewards, layerReward7...)
			}
		}
		if allRewards != nil {
			data["rewards"] = allRewards
		} else {
			data["rewards"] = []util.TypeNum{}
		}

		// 通关记录
		record := model.GetUserClimbtowerRecord(ctx, layer)
		replaceRecord := &table.UserClimbtowerRecord{Layer: layer}
		if record == nil || record.FastUserId == 0 {
			replaceRecord.FastUserId = userID
			replaceRecord.FastNickname = userInfo.Nickname
		} else {
			replaceRecord.FastUserId = record.FastUserId
			replaceRecord.FastNickname = record.FastNickname
		}
		// 计算英雄总战力
		fightHeroTotalFightPoint := 0
		for _, h := range myHeroAttrs {
			fightHeroTotalFightPoint += h.FightPoint
		}
		if record == nil || record.LowestUserId == 0 || fightHeroTotalFightPoint < record.LowestFightPoint {
			replaceRecord.LowestUserId = userID
			replaceRecord.LowestNickname = userInfo.Nickname
			replaceRecord.LowestFightPoint = fightHeroTotalFightPoint
		} else {
			replaceRecord.LowestUserId = record.LowestUserId
			replaceRecord.LowestNickname = record.LowestNickname
			replaceRecord.LowestFightPoint = record.LowestFightPoint
		}
		model.ReplaceUserClimbtowerRecord(ctx, layer, replaceRecord)

		// 特殊奖励：每10层设置可领取奖励 datas[6]
		if layer%10 == 0 {
			if rd6, ok := logic.CheckpointRewardDatas[6]; ok {
				if layerReward6, ok := rd6[layer]; ok {
					model.SetClimbLingquRewards(ctx, userID, layer, layerReward6)
				}
			}
		}

		model.SetDailyTaskFinish(ctx, userID, 10006, 1)

		model.GuideTaskHandle(ctx, userID, 82, 1)
		model.GuideTaskHandleIncr(ctx, userID, 86, 1)
	} else {
		data["rewards"] = []util.TypeNum{}
	}

	return c.ResponseSuccessToMe(data)
}

// 爬塔已通关层再战（AlreadycrossClimbtower）
func (c *ShinelightController) AlreadyCrossClimbtowerAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	layer := c.Params.GetIntE("layer")
	if layer <= 0 {
		return c.ResponseError(4678, "layer is empty")
	}

	// 加锁 — 对齐 PHP RedisProxy::lock("AlreadycrossClimbtower" . $userId)
	lock.Lock(fmt.Sprintf("AlreadycrossClimbtower%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("AlreadycrossClimbtower%d", userID))

	// 校验关卡层数：不能超过已通关层数
	currentLayer := model.GetUsersContentInt(ctx, userID, "climbtower_layer")
	if layer > currentLayer {
		return c.ResponseError(6422, "关卡数错误")
	}

	// 校验扫荡次数
	saodang := model.GetClimbtowerSaodang(ctx, userID)
	if saodang.FreeNum <= 0 && saodang.BuyNum <= 0 {
		return c.ResponseError(4688, "没有扫荡次数")
	}

	// 校验货币是否足够
	if saodang.BasicCost > 0 {
		if item.NotEnough(userID, saodang.CostType, saodang.BasicCost) {
			return c.ResponseError(666666, "货币不够")
		}
	}

	fightHerosParam := c.Params.GetStringE("fight_heros")
	position := c.Params.GetIntE("position")

	var myHeroAttrs []*logic.Hero
	if fightHerosParam != "" {
		heroIDs, heroPos := util.ToPosHeros(fightHerosParam)

		objs := model.GetUserHeroAttrWithSkillByIDs(ctx, heroIDs, userID, true)
		for _, obj := range objs {
			if pos, ok := heroPos[obj.Id]; ok {
				obj.Pos = pos
			}
			myHeroAttrs = append(myHeroAttrs, obj)
		}
		model.SetFightHeros(ctx, userID, "climb", heroPos, position)
	}
	if len(myHeroAttrs) == 0 {
		return c.ResponseError(6422, "请选择英雄")
	}

	opHerosRaw := logic.GetClimbtowerHeros(layer)
	if opHerosRaw == nil || len(opHerosRaw) == 0 {
		return c.ResponseError(6422, fmt.Sprintf("第%d层数据不存在", layer))
	}

	layerInfo, ok := logic.ClimbtowerData[layer]
	if !ok {
		return c.ResponseError(6422, fmt.Sprintf("第%d层配置不存在", layer))
	}

	oppHeroAttrs := logic.BuildBossDetail(logic.ClimbtowerMonsterDatas, opHerosRaw)

	fight := logic.NewFight(myHeroAttrs, oppHeroAttrs)
	winner, fightResult := fight.FightExec(15)

	// 从战斗结果提取胜负
	win := 0
	if winner == "P1" {
		win = 1
	}

	data := types.Map{
		"success":      win,
		"my_hero":      myHeroAttrs,
		"opp_hero":     oppHeroAttrs,
		"fight_result": fightResult,
		"position":     layerInfo.Position,
		"combination":  layerInfo.Combination,
	}

	// 挑战成功
	if win == 1 {
		// 给予奖励 — 已通关重打只给 datas[1] 的奖励
		if rd, ok := logic.CheckpointRewardDatas[1]; ok {
			if layerReward, ok := rd[layer]; ok {
				model.GiveReward(userID, layerReward...)
				data["rewards"] = layerReward
			}
		}
		// 扣除扫荡次数
		model.DecrClimbtowerSaodang(ctx, userID)
		// 日常任务
		model.SetDailyTaskFinish(ctx, userID, 10006, 1)
		// 扣除费用
		if saodang.BasicCost > 0 {
			item.Sub(userID, saodang.CostType, saodang.BasicCost)
		}
	}

	return c.ResponseSuccessToMe(data)
}

// 爬塔排行榜
func (c *ShinelightController) ClimbtowerRankAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)

	data := types.Map{
		"my_rank":     model.GetMyRank(ctx, config.RankClimbtower, userID),
		"my_layer":    model.GetMyRankScore(ctx, config.RankClimbtower, userID),
		"lv":          userGrade.GetIntE("lv"),
		"fight_point": model.GetUserFightPoint(ctx, userID, 1),
		"rank_list":   model.GetRankList(ctx, config.RankClimbtower, true, 0, 99, "lv", "fight_point", "vip_level"),
	}
	return c.ResponseSuccessToMe(data)
}

// 爬塔扫荡预览（不消耗次数）
func (c *ShinelightController) ClimbSaodangInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	layer := c.Params.GetIntE("layer")
	if layer <= 0 {
		return c.ResponseError(4678, "layer is empty")
	}

	// 获取该层的扫荡信息
	data := logic.GetClimbSaodang(layer)
	if data == nil {
		data = make(types.Map)
	}

	// 补充首通奖励（与 PHP 一致：CheckpointRewardDatas[7][layer] 直接是 []util.TypeNum）
	firstRewards := logic.CheckpointRewardDatas[7][layer]
	if firstRewards == nil {
		firstRewards = []util.TypeNum{}
	}
	data["first_rewards"] = firstRewards

	// 补充最快通关和最低战力记录
	record := model.GetUserClimbtowerRecord(ctx, layer)
	if record != nil {
		fastNick := string(record.FastNickname)
		lowestNick := string(record.LowestNickname)
		if fastNick == "" {
			fastNick = types.ToString(record.FastUserId)
		}
		if lowestNick == "" {
			lowestNick = types.ToString(record.LowestFightPoint)
		}
		data["fast_user"] = fastNick
		data["lowest_user"] = lowestNick
	} else {
		data["fast_user"] = ""
		data["lowest_user"] = ""
	}

	crossedLayer := model.GetUsersContentInt(ctx, userID, "climbtower_layer")
	data["crossed_layer"] = crossedLayer

	// 日常任务, 试练塔3次 — 对齐 PHP setDalilyTaskFinish($userId, 10006, 1)
	model.SetDailyTaskFinish(ctx, userID, 10006, 1)

	return c.ResponseSuccessToMe(data)
}

func (c *ShinelightController) ClimbtowerSaodangAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	layer := c.Params.GetIntE("layer")
	if layer <= 0 {
		return c.ResponseError(4678, "layer is empty")
	}

	lock.Lock(fmt.Sprintf("climbtowerSaodang%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("climbtowerSaodang%d", userID))

	saodang := model.GetClimbtowerSaodang(ctx, userID)

	if saodang.FreeNum <= 0 && saodang.BuyNum <= 0 {
		return c.ResponseError(4688, "没有扫荡次数")
	}
	if saodang.BasicCost > 0 {
		if item.NotEnough(userID, saodang.CostType, saodang.BasicCost) {
			return c.ResponseError(666666, "货币不够")
		}
	}

	// 获取扫荡奖励
	rewardsData := logic.GetClimbtowerRewards(layer)
	rewards := make([]util.TypeNum, 0)
	for _, v := range rewardsData {
		if v.Layer == layer {
			rewards = v.SaodangRewards
		}
	}
	model.GiveReward(userID, rewards...)
	model.DecrClimbtowerSaodang(ctx, userID)

	if saodang.BasicCost > 0 {
		item.Sub(userID, 2, saodang.BasicCost)
	}

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}
