package controller

import (
	"context"
	"fmt"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/table"
)

// 无尽试炼信息
func (c *ShinelightController) EndlessInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)

	data := make(types.Map)

	// 检查用户等级
	if userGrade.GetIntE("lv") < 25 {
		data["status"] = 0
		return c.ResponseSuccessToMe(data)
	}

	data["status"] = 1
	data["my_rank"] = model.GetMyRank(ctx, config.RankEndlessLayer, userID)
	data["rank_list"] = model.GetRankList(ctx, config.RankEndlessLayer, true, 0, 2)

	endlessLayer := model.GetUsersContentInt(ctx, userID, "endless_layer")

	ret := model.GetEndlessInfo(ctx, userID, endlessLayer)

	// 获取首通奖励状态
	firstCrossRaw := model.GetUserTongguanReward(ctx, userID, 1)
	firstCrossStatus := make([]types.Map, 0, len(firstCrossRaw))
	for _, v := range firstCrossRaw {
		m := types.Map{
			"copy":   v.Copy,
			"status": v.Status,
			"reward": logic.GetFirstCrossLayerReward(v.Copy),
		}
		firstCrossStatus = append(firstCrossStatus, m)
	}
	data["first_cross_status"] = firstCrossStatus

	for k, v := range ret {
		data[k] = v
	}

	return c.ResponseSuccessToMe(data)
}

// 无尽试炼战斗
func (c *ShinelightController) EndlessFightAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	layer := c.Params.GetIntE("layer")
	addition := c.Params.GetIntE("addition") // 增益: 1生命上限 2攻击 3防御...
	position := c.Params.GetIntE("position")
	status := c.Params.GetIntE("status") // 1=本次挑战的调用
	_, fightHeros := util.ToPosHeros(c.Params.GetArrayE("fight_heros"))

	lock.Lock(fmt.Sprintf("endlessFight%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("endlessFight%d", userID))

	userGrade := model.GetUserAttr(userID)
	if userGrade.GetIntE("lv") < 25 {
		return c.ResponseError(677542, "等级不够")
	}

	endlessLayer := model.GetUsersContentInt(ctx, userID, "endless_layer")
	endlessInfo := model.GetEndlessInfo(ctx, userID, endlessLayer)

	if status != 1 {
		model.SetEndlessThistimeAlreadyLayer(ctx, userID, endlessInfo.GetIntE("already_layer"))
	}
	oldAlreadyLayer := model.GetEndlessThistimeAlreadyLayer(ctx, userID)

	// 检查战斗英雄
	cacheFightHeros := model.GetEndlessFightHeros(ctx, userID)
	if cacheFightHeros == nil && len(fightHeros) == 0 {
		return c.ResponseError(677742, "请选择出战英雄")
	}

	// 是否第一次战斗
	model.HsetEndlessTodayLayer(ctx, userID, "first_fight", 0)

	var p1HP map[int]int
	var lastSuccess int

	if cacheFightHeros != nil {
		position = types.ToIntE(cacheFightHeros.Position)
		fightHeros = cacheFightHeros.Heros
		if addition == 0 {
			addition = types.ToIntE(cacheFightHeros.Addition)
		}
		p1HP = cacheFightHeros.P1HP
		lastSuccess = types.ToIntE(cacheFightHeros.Success)
		if lastSuccess == 0 {
			return c.ResponseError(9000, "挑战已结束")
		}
	} else {
		model.SetFightHeros(ctx, userID, "endless_copy", fightHeros, position)
	}

	// 获取己方英雄详情
	fightHeroDetails := c.getFightHeroByPosMap(ctx, userID, fightHeros)
	helpCount := c.getHelpCount(userID, fightHeroDetails)
	if helpCount > 1 {
		return c.ResponseError(7643, "只允许一个支援英雄上阵")
	}

	// 起始血量 + 增益
	c.setStartHP(fightHeroDetails, p1HP)
	logic.AdditionHandle(fightHeroDetails, addition)

	// 对手
	opHeros := logic.GetEndlessHeros(layer)
	opHeroDetail := logic.BuildBossDetail(logic.EndlessMonsterDatas, opHeros)

	// 战斗
	fightResult := logic.NewFight(fightHeroDetails, opHeroDetail)
	winner, retResult := fightResult.FightExec(15)
	success := 0
	if winner == "P1" {
		success = 1
	}

	model.SetEndlessFightHeros(ctx, userID, fightHeros, p1HP, position, addition, success)

	data := make(types.Map)

	if success == 1 {
		model.HsetEndlessTodayLayer(ctx, userID, "current_layer", layer)
		alreadyLayer := types.ToIntE(endlessInfo["already_layer"])
		startLayer := types.ToIntE(endlessInfo["start_layer"])

		if layer > alreadyLayer {
			model.HsetEndlessTodayLayer(ctx, userID, "already_layer", layer)
			rewards := logic.GetCrossLayerReward(layer, startLayer)
			model.GiveReward(userID, rewards...)
		}

		todayCross := types.ToIntE(endlessInfo["today_cross_layer"])
		if (layer - startLayer) >= todayCross {
			model.HsetEndlessTodayLayer(ctx, userID, "today_cross_layer", layer-startLayer)
		}

		if layer > endlessLayer {
			model.IncrUsersContentInt(ctx, userID, "endless_layer", 1)
			model.SetRankScore(ctx, config.RankEndlessLayer, userID, float64(layer), 86400*30)
			nextFirstLayer := logic.GetNextFirstLayer(endlessLayer)
			if layer == nextFirstLayer {
				model.ReplaceUserTongguanReward(ctx, userID, 1, layer, 1)
				data["next_first_layer_reward"] = logic.GetFirstCrossLayerReward(nextFirstLayer)
			}
		}
	}

	// 最新通关信息
	endlessLayerNew := model.GetUsersContentInt(ctx, userID, "endless_layer")
	endlessInfoNew := model.GetEndlessInfo(ctx, userID, endlessLayerNew)

	reset := types.ToIntE(endlessInfo["reset"])
	if reset == 0 {
		data["cross_reward"] = endlessInfoNew["cross_reward"]
	} else {
		if layer <= oldAlreadyLayer {
			data["cross_reward"] = []util.TypeNum{{Type: 1, Num: 0}, {Type: 7, Num: 0}}
		} else {
			data["cross_reward"] = logic.GetLeijiCrossLayerReward(layer, oldAlreadyLayer)
		}
	}
	data["today_cross_layer"] = endlessInfoNew["today_cross_layer"]
	data["current_layer"] = endlessInfoNew["current_layer"]
	data["already_layer"] = endlessInfoNew["already_layer"]

	myHero := logic.GetBaseFromHero(fightHeroDetails)
	oppHero := logic.GetBaseFromHero(opHeroDetail)

	data["success"] = success
	data["my_hero"] = myHero
	data["opp_hero"] = oppHero
	data["fight_result"] = retResult
	data["addition"] = addition

	// 日常任务
	model.SetDailyTaskFinish(ctx, userID, 10009, 1)

	model.GuideTaskHandle(ctx, userID, 111, 1)

	return c.ResponseSuccessToMe(data)
}

// 无尽试炼排行榜
func (c *ShinelightController) EndlessRankAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	data := types.Map{
		"my_rank":     model.GetMyRank(ctx, config.RankEndlessLayer, userID),
		"my_layer":    model.GetMyRankScore(ctx, config.RankEndlessLayer, userID),
		"lv":          userGrade.GetIntE("lv"),
		"fight_point": model.GetUserFightPoint(ctx, userID, 1),
		"rank_list":   model.GetRankList(ctx, config.RankEndlessLayer, true, 0, 99, "lv", "fight_point"),
	}
	return c.ResponseSuccessToMe(data)
}

// 无尽试炼助战英雄列表 — 对齐 PHP endlessHelpListAction
func (c *ShinelightController) EndlessHelpListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := c.endlessHelpInfo(ctx, userID)
	return c.ResponseSuccessToMe(data)
}

// 保存助战英雄 — 对齐 PHP endlessHelpAddAction
func (c *ShinelightController) EndlessHelpAddAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetIntE("hero_id")

	model.ReplaceUserEndlessHelpHero(ctx, &table.UserEndlessHelpHero{
		UserId: userID,
		HeroId: heroID,
	})

	return c.ResponseSuccessToMe(types.Map{})
}

// endlessHelpInfo 获取无尽试炼助战信息（对齐 PHP private function endlessHelpInfo）
func (c *ShinelightController) endlessHelpInfo(ctx context.Context, userID int64) types.Map {

	// 获取好友列表
	friends := model.GetFriendsList(ctx, userID, 2)
	if len(friends) == 0 {
		friends = []int64{userID}
	} else {
		friends = append(friends, userID)
	}

	// 获取好友+自己的助战英雄
	helpHeros := model.GetUserEndlessHelpHero(ctx, friends)
	if len(helpHeros) == 0 {
		return types.Map{"help_me": []interface{}{}, "i_help_other": []interface{}{}}
	}

	// 收集英雄ID，获取英雄属性
	heroIDs := make([]int, 0, len(helpHeros))
	for _, h := range helpHeros {
		heroIDs = append(heroIDs, h.HeroId)
	}
	helpHeroDetails := model.GetUserHeroAttrByIDs(ctx, heroIDs, 0, false)

	// 批量获取用户信息
	userInfos := model.GetUsersWithDetail(ctx, friends, 1)

	// 获取已选择的助战英雄
	helpChoose := model.GetEndlessHelpChoose(ctx, userID)

	helpMe := make([]types.Map, 0)
	iHelpOther := make([]types.Map, 0)

	for _, heroID := range heroIDs {
		val, ok := helpHeroDetails[heroID]
		if !ok {
			continue
		}
		heroUserID := val.UserId
		tmp := types.Map{
			"id":          val.Id,
			"user_id":     heroUserID,
			"hero_id":     val.HeroInfo,
			"star":        val.Star,
			"stage":       val.Stage,
			"lv":          val.Lv,
			"fight_point": val.FightPoint,
			"choosed":     0,
		}
		if helpChoose == val.Id {
			tmp["choosed"] = 1
		}
		if info, ok := userInfos[heroUserID]; ok {
			tmp["nickname"] = info.GetStringE("nickname")
		}

		if heroUserID == userID {
			iHelpOther = append(iHelpOther, tmp)
		} else {
			helpMe = append(helpMe, tmp)
		}
	}

	return types.Map{"help_me": helpMe, "i_help_other": iHelpOther}
}

// 选择无尽试炼助战英雄 — 对齐 PHP endlessChooseHelpAction
func (c *ShinelightController) EndlessChooseHelpAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetStringE("hero_id")

	model.SetEndlessHelpChoose(ctx, userID, heroID)

	return c.ResponseSuccessToMe(types.Map{})
}

// 无尽首通奖励领取
func (c *ShinelightController) EndlessFirstLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	layer := c.Params.GetIntE("layer")

	lock.Lock(fmt.Sprintf("endlessFirstLingqu%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("endlessFirstLingqu%d", userID))

	data := model.GetUserTongguanReward(ctx, userID, 1)

	var targetReward *table.UserTongguanReward
	for _, v := range data {
		if v.Copy == layer {
			targetReward = v
			break
		}
	}

	if targetReward == nil || targetReward.Status == 0 {
		return c.ResponseError(6777421, "此关卡还未通关")
	}
	if targetReward.Status == 2 {
		return c.ResponseError(6777421, "本关奖励已经领取")
	}

	model.DeleteUserTongguanReward(ctx, userID, 1, layer)

	// 给予首通奖励（根据层数动态计算，对齐 PHP Endless::get_first_cross_layer_reward）
	rewards := logic.GetFirstCrossLayerReward(layer)
	model.GiveReward(userID, rewards...)

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

// 重置今日无尽试炼
func (c *ShinelightController) ResetTodayEndlessAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	endlessLayer := model.GetUsersContentInt(ctx, userID, "endless_layer")
	endlessInfo := model.GetEndlessInfo(ctx, userID, endlessLayer)
	fightHeros := model.GetEndlessFightHeros(ctx, userID)

	startLayer := types.ToIntE(endlessInfo["start_layer"])
	model.HsetEndlessTodayLayer(ctx, userID, "current_layer", startLayer)
	model.ResetEndlessFightHeros(ctx, userID)

	model.HsetEndlessTodayLayer(ctx, userID, "first_fight", 0)
	model.HsetEndlessTodayLayer(ctx, userID, "reset", 1)

	data := types.Map{}
	if fightHeros != nil {
		data["position"] = fightHeros.Position
		heroObjs := c.getFightHeroByPosMap(ctx, userID, fightHeros.Heros)

		totalFightPoint := 0
		var heroList []types.Map
		for _, hero := range heroObjs {
			heroList = append(heroList, types.Map{
				"id":          hero.Id,
				"hero_id":     hero.HeroInfo,
				"star":        hero.Star,
				"stage":       hero.Stage,
				"lv":          hero.Lv,
				"pos":         hero.Pos,
				"fight_point": hero.FightPoint,
			})
			totalFightPoint += hero.FightPoint
		}

		data["heros"] = heroList
		data["total_fight_point"] = totalFightPoint
	}

	return c.ResponseSuccessToMe(data)
}
