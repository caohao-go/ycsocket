package controller

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"server_golang/common/lock"
	"server_golang/common/util"
	"server_golang/repo/mem/item"
	"server_golang/repo/table"

	"server_golang/common/types"
	"server_golang/logic"
	"server_golang/model"
)

// 英雄系统

func (c *ShinelightController) UserHeroAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := model.GetUserHeroList(ctx, userID)

	ids := make([]int, 0, len(data))
	for _, v := range data {
		ids = append(ids, v.Id)
	}

	// 如果前端传了 type 参数，使用 GetFightHeros 来确定已出战英雄
	typ := c.Params.GetStringE("type")

	heroAttrs := model.GetUserHeroAttrByIDs(ctx, ids, userID, false)
	ret := make([]types.Map, 0, len(heroAttrs))
	for _, v := range heroAttrs {
		heroID := v.HeroInfo
		ret = append(ret, types.Map{
			"id":          v.Id,
			"hero_id":     heroID,
			"property":    logic.HeroProperty(heroID),
			"star":        v.Star,
			"stage":       v.Stage,
			"lv":          v.Lv,
			"hp":          v.Hp,
			"atk":         v.Atk,
			"fight_point": v.FightPoint,
		})
	}

	// 收集已出战英雄ID
	inBattleHeroIDs := make(map[int]bool)
	var fightHeros *model.FightHerosData
	var userPositions []*model.UserPosition
	if typ != "" {
		// 有 type 参数时，使用 GetFightHeros 获取对应类型的出战英雄
		fightHeros = model.GetFightHeros(ctx, userID, typ)
		if fightHeros != nil && len(fightHeros.Heros) > 0 {
			for heroID := range fightHeros.Heros {
				inBattleHeroIDs[heroID] = true
			}
		}
	} else {
		// 无 type 参数时，使用原有逻辑（所有阵型中出战的英雄）
		var e error
		userPositions, e = model.GetUserPosition(ctx, userID)
		if e != nil {
			return c.ResponseError(99910131, e.Error())
		}
		for _, pos := range userPositions {
			for heroID := range pos.HeroPos {
				inBattleHeroIDs[heroID] = true
			}
		}
	}

	// 排序：出战英雄优先，然后按 star 降序、lv 降序、hero_id 升序、id 升序
	sort.SliceStable(ret, func(i, j int) bool {
		iID := types.ToIntE(ret[i]["id"])
		jID := types.ToIntE(ret[j]["id"])
		iInBattle := inBattleHeroIDs[iID]
		jInBattle := inBattleHeroIDs[jID]

		// 出战英雄排在前面
		if iInBattle != jInBattle {
			return iInBattle
		}

		// 同组内按 star 降序
		iStar := types.ToIntE(ret[i]["star"])
		jStar := types.ToIntE(ret[j]["star"])
		if iStar != jStar {
			return iStar > jStar
		}

		// star 相同按 lv 降序
		iLv := types.ToIntE(ret[i]["lv"])
		jLv := types.ToIntE(ret[j]["lv"])
		if iLv != jLv {
			return iLv > jLv
		}

		// lv 相同按 hero_id 升序，同一种英雄排在一起
		iHeroID := types.ToIntE(ret[i]["hero_id"])
		jHeroID := types.ToIntE(ret[j]["hero_id"])
		if iHeroID != jHeroID {
			return iHeroID < jHeroID
		}

		// hero_id 也相同按 id 升序，确保顺序完全稳定
		return iID < jID
	})

	// 转换 position 格式
	position := make(map[int]types.Map)
	if typ != "" {
		// 有 type 参数时，返回对应类型的阵型信息
		if fightHeros != nil && len(fightHeros.Heros) > 0 {
			heroMap := make(map[int]int)
			for heroID, pos := range fightHeros.Heros {
				heroMap[pos] = heroID
			}
			position[0] = types.Map{
				"position":          fightHeros.Position,
				"pos":               fightHeros.Position,
				"hero":              heroMap,
				"combination":       101,
				"total_fight_point": 0,
			}
		}
	} else {
		// 无 type 参数时，使用原有逻辑
		for _, pos := range userPositions {
			posType := pos.PosType
			heroMap := make(map[int]int)

			for k, v := range pos.HeroPos {
				heroMap[v] = k
			}

			// 使用 GetUserFightPoint 计算阵型总战力（包含组合加成，与PHP端及主界面一致）
			totalFightPoint := model.GetUserFightPoint(ctx, userID, posType)

			position[posType] = types.Map{
				"position":          pos.Position,
				"pos":               pos.Position,
				"hero":              heroMap,
				"combination":       101,
				"total_fight_point": totalFightPoint,
			}
		}
	}

	return c.ResponseSuccessToMe(types.Map{"list": ret, "position": position})
}

// 保存阵型
func (c *ShinelightController) SavePostionAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	posType := c.Params.GetIntE("pos_type")
	position := c.Params.GetIntE("position")

	posData := map[int]int{}
	for i := 1; i <= 5; i++ {
		hero := c.Params.GetIntE(fmt.Sprintf("pos%d_hero", i))
		pos := c.Params.GetIntE(fmt.Sprintf("pos%d_pos", i))
		if hero > 0 {
			posData[hero] = pos
		}
	}

	// 校验不能有重复英雄
	heroUniq := make(map[int]int)
	for hero := range posData {
		heroUniq[hero]++
		if heroUniq[hero] >= 2 {
			return c.ResponseError(666666, "不能出战相同的英雄")
		}
	}

	idData := model.GetUserPositionByID(ctx, userID, posType)

	data := table.UserPosition{
		UserId:   userID,
		PosType:  posType,
		Position: position,
	}

	if idData != nil {
		data.Id = idData.Id
	}

	data.Pos1Pos = c.Params.GetIntE("pos1_pos")
	data.Pos1Hero = c.Params.GetIntE("pos1_hero")
	data.Pos2Pos = c.Params.GetIntE("pos2_pos")
	data.Pos2Hero = c.Params.GetIntE("pos2_hero")
	data.Pos3Pos = c.Params.GetIntE("pos3_pos")
	data.Pos3Hero = c.Params.GetIntE("pos3_hero")
	data.Pos4Pos = c.Params.GetIntE("pos4_pos")
	data.Pos4Hero = c.Params.GetIntE("pos4_hero")
	data.Pos5Pos = c.Params.GetIntE("pos5_pos")
	data.Pos5Hero = c.Params.GetIntE("pos5_hero")

	model.ReplaceUserPosition(ctx, &data)

	// 副本阵型缓存
	if posType == 1 {
		model.SetFightHeros(ctx, userID, "copy_fight", posData, position)
	}

	// 收集英雄ID
	heroIDs := make([]int, 0)
	for hero := range posData {
		if hero > 0 {
			heroIDs = append(heroIDs, hero)
		}
	}

	// 引导任务
	if len(heroIDs) >= 1 {
		model.GuideTaskHandle(ctx, userID, 2, 1)
		model.GuideTaskHandle(ctx, userID, 11, len(heroIDs))
		if position != 101 {
			model.GuideTaskHandle(ctx, userID, 58, 1)
		}
	}

	// 计算总战斗力
	totalFightPoint := 0
	herosDetail := model.GetUserHeroAttrByIDs(ctx, heroIDs, userID, true)
	for _, v := range herosDetail {
		totalFightPoint += v.FightPoint
	}

	// 更新主阵型战力总排行
	if posType == 1 {
		model.SetFightPointRank(ctx, userID)
	}

	return c.ResponseSuccessToMe(types.Map{"fight_point": totalFightPoint})
}

// 获取他人英雄属性
func (c *ShinelightController) OtherUserHeroPropAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	otherUserID := c.Params.GetInt64E("other_userid")
	heroID := c.Params.GetIntE("id")

	heroAttrs := model.GetUserHeroAttrWithSkillByIDs(ctx, []int{heroID}, otherUserID, false)
	if heroAttrs == nil || heroAttrs[heroID] == nil {
		return c.ResponseError(6801, "hero not find")
	}

	return c.ResponseSuccessToMe(types.ObjectToMap(heroAttrs[heroID]))
}

// 获取英雄属性
func (c *ShinelightController) UserHeroPropAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetIntE("id")

	heroInfos := model.GetUserHeroAttrWithSkillByIDs(ctx, []int{heroID}, userID, false)
	if len(heroInfos) == 0 {
		return c.ResponseError(6801, "hero not find")
	}

	return c.ResponseSuccessToMe(heroInfos[heroID].ToMap())
}

// 英雄升级
func (c *ShinelightController) UpHeroLvAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetIntE("id")

	// 锁 key 与 PHP 一致：up_hero_lv{uid}
	lock.Lock(fmt.Sprintf("up_hero_lv%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("up_hero_lv%d", userID))

	userHero := model.GetUserHeroByID(ctx, heroID)
	if userHero.UserId != userID {
		return c.ResponseError(76319, "英雄不存在")
	}

	lv := userHero.Lv
	nextLv := lv + 1
	star := userHero.Star
	stage := userHero.Stage

	lvStageMax := logic.GetHeroMaxLv(star, stage)
	if lv >= lvStageMax {
		return c.ResponseError(76319, "已经达到最大等级")
	}

	// 先升级（与 PHP 一致：失败返回 99 system error）
	if err := model.UpdateUserHero(ctx, userID, heroID, 1, 0, 0); err != nil {
		return c.ResponseError(99, "system error")
	}

	// 检查资源
	heroUpLv := logic.HeroLvDatas[lv]
	uplvExpNeed := heroUpLv.Exp
	uplvGoldNeed := heroUpLv.Gold

	if item.NotEnough(userID, 7, uplvExpNeed) {
		model.UpdateUserHero(ctx, userID, heroID, -1, 0, 0)
		return c.ResponseError(64526, "经验不够")
	}
	if item.NotEnough(userID, 1, uplvGoldNeed) {
		model.UpdateUserHero(ctx, userID, heroID, -1, 0, 0)
		return c.ResponseError(666666, "金币不够")
	}

	item.Sub(userID, 7, uplvExpNeed)
	item.Sub(userID, 1, uplvGoldNeed)

	cost := []util.TypeNum{
		{Type: 7, Num: uplvExpNeed},
		{Type: 1, Num: uplvGoldNeed},
	}

	heroInfo := model.GetUserHeroAttrWithSkillByIDs(ctx, []int{heroID}, userID, false)
	if len(heroInfo) == 0 {
		return c.ResponseError(4434234, "未找到英雄")
	}

	// 引导任务
	if nextLv < 60 {
		model.GuideTaskHandle(ctx, userID, 69, nextLv)
		model.GuideTaskHandle(ctx, userID, 42, nextLv)
		model.GuideTaskHandle(ctx, userID, 17, nextLv)
		model.GuideTaskHandle(ctx, userID, 6, nextLv)
	}

	return c.ResponseSuccessToMe(types.Map{"cost": cost, "lv": nextLv, "attr": heroInfo[heroID]})
}

// 英雄升阶详情
func (c *ShinelightController) UpHeroStageInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetIntE("id")

	userHero := model.GetUserHeroByID(ctx, heroID)
	stage := userHero.Stage

	stageData := logic.HeroStageDatas[stage]
	stageStoneNeed := stageData.StageStone
	stageGoldNeed := stageData.StageGold

	cost := []util.TypeNum{
		{Type: 20201, Num: stageStoneNeed},
		{Type: 1, Num: stageGoldNeed},
	}

	myStage := item.Total(userID, 20201)

	// 预览下一阶属性
	userHero.Stage++
	heroAttr := model.GetHeroAttrCore(ctx, userHero, userID, 0, 0, 0)

	star := userHero.Star
	skillSet := logic.GetSkill(userHero.HeroId, star, stage+1)
	var activeKill interface{}
	if skillSet != nil {
		lv := userHero.Lv
		if lv == 30 && stage+1 == 1 && len(skillSet.Skills) > 1 {
			activeKill = skillSet.Skills[1]
		} else if lv == 50 && stage+1 == 3 && len(skillSet.Skills) > 2 {
			activeKill = skillSet.Skills[2]
		} else if lv == 80 && stage+1 == 5 && len(skillSet.Skills) > 3 {
			activeKill = skillSet.Skills[3]
		}
	}

	attr := types.Map{
		"lv":     heroAttr.Lv,
		"lv_max": heroAttr.LvMax,
		"atk":    heroAttr.Atk,
		"hp":     heroAttr.Hp,
		"def":    heroAttr.Def,
		"speed":  heroAttr.Speed,
	}

	return c.ResponseSuccessToMe(types.Map{
		"cost":        cost,
		"my_stage":    myStage,
		"attr":        attr,
		"active_kill": activeKill,
	})
}

// 英雄升阶
func (c *ShinelightController) UpHeroStageAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetIntE("id")

	userHero := model.GetUserHeroByID(ctx, heroID)

	lv := userHero.Lv
	stage := userHero.Stage
	star := userHero.Star
	nextStage := stage + 1

	if userHero.UserId != userID {
		return c.ResponseError(7642, "hero error")
	}
	if stage >= star || stage >= 6 {
		return c.ResponseError(7688, "已经达到最大等阶")
	}
	lvStageMax := logic.GetHeroMaxLv(star, stage)
	if lv < lvStageMax {
		return c.ResponseError(7633, "等级不够，不能升阶")
	}

	stageData := logic.HeroStageDatas[stage]
	stageStoneNeed := stageData.StageStone
	stageGoldNeed := stageData.StageGold

	if item.NotEnough(userID, 20201, stageStoneNeed) {
		return c.ResponseError(645326, "进阶石不够")
	}
	if item.NotEnough(userID, 1, stageGoldNeed) {
		return c.ResponseError(666666, "金币不够")
	}

	// 先升阶（与 PHP 一致：失败返回 99 system error）
	if err := model.UpdateUserHero(ctx, userID, heroID, 0, 1, 0); err != nil {
		return c.ResponseError(99, "system error")
	}
	item.Sub(userID, 20201, stageStoneNeed)
	item.Sub(userID, 1, stageGoldNeed)

	cost := []util.TypeNum{
		{Type: 20201, Num: stageStoneNeed},
		{Type: 1, Num: stageGoldNeed},
	}

	userHero.Stage = nextStage
	heroAttr := model.GetHeroAttrCore(ctx, userHero, userID, 0, 0, 0)

	model.GuideTaskHandle(ctx, userID, 43, 1)

	return c.ResponseSuccessToMe(types.Map{"cost": cost, "stage": nextStage, "attr": heroAttr})
}

// 英雄升星信息
func (c *ShinelightController) HeroStarUpInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userHeros := model.GetUserHeroList(ctx, userID)
	data := logic.GetStarUpList(userHeros)
	return c.ResponseSuccessToMe(types.Map{"list": data})
}

// 英雄升星详情（融合神殿）
func (c *ShinelightController) HeroStarUpDetailARongheshendianAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")

	userHeros := model.GetUserHeroList(ctx, userID)
	userHeroInfo := model.GetUserHeroByID(ctx, id)
	if userHeroInfo.UserId != userID {
		return c.ResponseError(42345, "hero error")
	}

	data := logic.GetStarUpDetail(userHeroInfo, userHeros)
	delete(data, "user_id")
	delete(data, "fit")
	delete(data, "fu")
	delete(data, "updatetime")

	// 获取出战和远航英雄
	pos1 := model.GetUserPositionByID(ctx, userID, 1)
	pos2 := model.GetUserPositionByID(ctx, userID, 2)
	posHeroIDs := collectPositionHeroIDs(pos1, pos2)
	voyageHero := model.GetVoyageHero(ctx, userID)

	markNeedStatus(data, posHeroIDs, voyageHero)

	heroAttr := model.GetHeroAttrCore(ctx, userHeroInfo, userID, 0, 0, 0)
	data["hp"] = heroAttr.Hp
	data["atk"] = heroAttr.Atk
	data["def"] = heroAttr.Def
	data["speed"] = heroAttr.Speed

	star := data.GetIntE("star")
	if star <= 5 {
		nextAttr := logic.HeroStarUpAttrDatas[star+1]
		data["hp_add"] = fmt.Sprintf("%d%%", int64(nextAttr.Hp))
		data["atk_add"] = fmt.Sprintf("%d%%", int64(nextAttr.Atk))
		data["has_stage_stone"] = item.Total(userID, 20201)
		if sc, ok := logic.HeroStarConsumeDatas[star]; ok {
			data["need_stage_stone"] = sc.StageStone
			data["lv_max"] = sc.LvMax
		}
		if scNext, ok := logic.HeroStarConsumeDatas[star+1]; ok {
			data["next_lv_max"] = scNext.LvMax
		}
	}

	heroIDVal := data.GetIntE("hero_id")
	stage := data.GetIntE("stage")
	skillsNew := logic.GetSkillBaseInfo(heroIDVal, star+1, stage)
	nextStar := star + 1
	if nextStar == 6 || nextStar == 5 {
		data["skills"] = skillsNew
	}

	return c.ResponseSuccessToMe(data)
}

// 英雄升星详情（面板）
func (c *ShinelightController) HeroStarUpDetailAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")

	userHeros := model.GetUserHeroList(ctx, userID)
	userHeroInfo := model.GetUserHeroByID(ctx, id)
	if userHeroInfo.UserId != userID {
		return c.ResponseError(42345, "hero error")
	}

	// 去掉本体
	filtered := make([]*table.UserHero, 0)
	for _, v := range userHeros {
		if v.Id != userHeroInfo.Id {
			filtered = append(filtered, v)
		}
	}

	data := logic.GetStarUpDetail(userHeroInfo, filtered)
	delete(data, "user_id")
	delete(data, "fit")
	delete(data, "fu")
	delete(data, "updatetime")

	pos1 := model.GetUserPositionByID(ctx, userID, 1)
	pos2 := model.GetUserPositionByID(ctx, userID, 2)
	posHeroIDs := collectPositionHeroIDs(pos1, pos2)
	voyageHero := model.GetVoyageHero(ctx, userID)
	markNeedStatus(data, posHeroIDs, voyageHero)

	heroAttr := model.GetHeroAttrCore(ctx, userHeroInfo, userID, 0, 0, 0)
	data["hp"] = heroAttr.Hp
	data["atk"] = heroAttr.Atk
	data["def"] = heroAttr.Def
	data["speed"] = heroAttr.Speed

	star := data.GetIntE("star")
	if star >= 5 {
		// 去掉 needs[0]
		if needs, err := types.ToMapArray(data["needs"], ""); err == nil && len(needs) > 1 {
			data["needs"] = needs[1:]
		}
		data["has_stage_stone"] = item.Total(userID, 20201)
		if sc, ok := logic.HeroStarConsumeDatas[star]; ok {
			data["need_stage_stone"] = sc.StageStone
			data["lv_max"] = sc.LvMax
		}
		if scNext, ok := logic.HeroStarConsumeDatas[star+1]; ok {
			data["next_lv_max"] = scNext.LvMax
		}
	}

	heroIDVal := data.GetIntE("hero_id")
	stage := data.GetIntE("stage")
	skillsOld := logic.GetSkillBaseInfo(heroIDVal, star, stage)
	skillsNew := logic.GetSkillBaseInfo(heroIDVal, star+1, stage)

	nextStar := star + 1
	var skillLv interface{}
	switch nextStar {
	case 6:
		skillLv = skillsNew
	case 7:
		skillLv = getSkillDiff(skillsOld, skillsNew, 0)
	case 8:
		skillLv = getSkillDiff(skillsOld, skillsNew, 1)
	case 9:
		skillLv = getSkillDiff(skillsOld, skillsNew, 2)
	case 10:
		skillLv = getSkillDiff(skillsOld, skillsNew, 3)
	}
	data["skills"] = skillLv

	return c.ResponseSuccessToMe(data)
}

// 英雄升星
func (c *ShinelightController) HeroStarUpAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	needsStr := c.Params.GetStringE("needs")
	needsParts := strings.Split(needsStr, ",")
	if len(needsParts) == 0 || needsStr == "" {
		return c.ResponseError(45345, "input error")
	}

	lock.Lock(fmt.Sprintf("hero_star_up%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("hero_star_up%d", userID))

	mainID := types.ToIntE(needsParts[0])
	userHeros := model.GetUserHeroList(ctx, userID)
	userHeroInfo := model.GetUserHeroByID(ctx, mainID)
	if userHeroInfo.UserId != userID {
		return c.ResponseError(5675, "hero error")
	}

	// 收集所有 need IDs
	needsMap := make(map[int]bool)
	for _, s := range needsParts {
		nid := types.ToIntE(s)
		needsMap[nid] = true
	}

	data := logic.GetStarUpDetail(userHeroInfo, userHeros)

	// 验证是否满足升星条件
	if needs, err := types.ToMapArray(data["needs"], ""); err == nil {
		for _, val := range needs {
			num := val.GetIntE("num")
			if hasList, ok := val["has"].([]*table.UserHero); ok {
				for _, has := range hasList {
					hasID := has.Id
					if needsMap[hasID] {
						num--
						delete(needsMap, hasID)
					}
					if num <= 0 {
						break
					}
				}
			}
			if num > 0 {
				return c.ResponseError(56175, "英雄数量不够")
			}
		}
	}

	star := data.GetIntE("star")
	var needStageStone int
	if star >= 6 {
		hasStageStone := item.Total(userID, 20201)
		if sc, ok := logic.HeroStarConsumeDatas[star]; ok {
			needStageStone = sc.StageStone
		}
		if hasStageStone < needStageStone {
			return c.ResponseError(39281, "进阶石不够")
		}
	}

	heroIDVal := userHeroInfo.HeroId
	beforeAttr := logic.GetHeroLvAttr(heroIDVal, star, userHeroInfo.Stage, userHeroInfo.Lv)
	afterAttr := logic.GetHeroLvAttr(heroIDVal, star+1, userHeroInfo.Stage, userHeroInfo.Lv)

	skillsOld := logic.GetSkillBaseInfo(heroIDVal, star, data.GetIntE("stage"))
	skillsNew := logic.GetSkillBaseInfo(heroIDVal, star+1, data.GetIntE("stage"))

	nextStar := star + 1
	var skillLv interface{}
	switch nextStar {
	case 5, 6:
		skillLv = skillsNew
	case 7:
		skillLv = getSkillDiff(skillsOld, skillsNew, 0)
	case 8:
		skillLv = getSkillDiff(skillsOld, skillsNew, 1)
	case 9:
		skillLv = getSkillDiff(skillsOld, skillsNew, 2)
	case 10:
		skillLv = getSkillDiff(skillsOld, skillsNew, 3)
	}

	// 收集被消耗英雄的装备返还
	sacrificeIDs := make([]int, 0)
	for _, s := range needsParts[1:] {
		nid := types.ToIntE(s)
		if nid > 0 {
			sacrificeIDs = append(sacrificeIDs, nid)
		}
	}

	returns := make([]util.TypeNum, 0)
	unloads := make([]util.TypeNum, 0)
	for _, sid := range sacrificeIDs {
		sacrificeHero := model.GetUserHeroByID(ctx, sid)
		if sacrificeHero == nil || sacrificeHero.UserId != userID {
			return c.ResponseError(666832, "英雄不存在")
		}

		// 返还资源（与 PHP heroStarUpAction 一致）
		sacStage := sacrificeHero.Stage
		sacLv := sacrificeHero.Lv

		// stage > 0 时返还阶段资源（type=2: $hero_star_up_return_datas[2][$stage]）
		if sacStage > 0 {
			if stageReturns, ok := logic.HeroStarUpReturnDatas[2][sacStage]; ok {
				for _, r := range stageReturns {
					returns = append(returns, util.TypeNum{Type: r.Type, Num: r.Num})
				}
			}
		}

		// lv > 1 时返还等级资源（type=3: $hero_star_up_return_datas[3][$lv]）
		if sacLv > 1 {
			if lvReturns, ok := logic.HeroStarUpReturnDatas[3][sacLv]; ok {
				for _, r := range lvReturns {
					returns = append(returns, util.TypeNum{Type: r.Type, Num: r.Num})
				}
			}
		}

		// 装备返还（与 PHP getEquipmentReturn($user_hero['fit'], $user_hero['fu']) 一致）
		equipReturn := logic.GetEquipmentReturn(sacrificeHero.Fit, sacrificeHero.Fu, sacrificeHero.Lv, sacrificeHero.Star)
		if len(equipReturn) > 0 {
			unloads = append(unloads, equipReturn...)
		}
	}

	// 扣减进阶石
	if star > 5 && needStageStone > 0 {
		item.Sub(userID, 20201, needStageStone)
	}

	model.GiveReward(userID, returns...)
	model.GiveReward(userID, unloads...)

	// 升星
	model.UpdateUserHero(ctx, userID, mainID, 0, 0, 1)

	// 删除被消耗的英雄
	if len(sacrificeIDs) > 0 {
		model.DeleteUserHeroByIDs(ctx, userID, sacrificeIDs)
	}

	// 获取升星后属性
	heroInfo := model.GetUserHeroAttrWithSkillByIDs(ctx, []int{mainID}, userID, false)

	if len(heroInfo) == 0 {
		return c.ResponseError(6668332, "未找到英雄")
	}

	// 成就任务
	// 9星英雄（taskID=15001, type 从配置映射，PHP 对应 achieve_types[15001]）
	if nextStar == 9 {
		typ15001 := logic.AchieveTypes[15001]
		task := model.GetUserAchieveTask(ctx, userID, typ15001, 15001)
		if task.Status == 0 {
			if cfg := logic.AchieveDatas[15001]; cfg != nil {
				newStatus, beforeStatus := model.SetUserAchieveTaskNum(ctx, userID, 15001, typ15001, 1, 0, cfg.Num, cfg.ExtraNum)
				if beforeStatus == 0 && newStatus == 1 {
					model.RedpointSend(ctx, userID, 1, 0, 1, 1)
				}
			}
		}
	}
	// 10星英雄（taskID=15002, type 从配置映射，PHP 对应 achieve_types[15002]）
	if nextStar == 10 {
		typ15002 := logic.AchieveTypes[15002]
		task := model.GetUserAchieveTask(ctx, userID, typ15002, 15002)
		if task.Status == 0 {
			if cfg := logic.AchieveDatas[15002]; cfg != nil {
				newStatus, beforeStatus := model.SetUserAchieveTaskNum(ctx, userID, 15002, typ15002, 1, 0, cfg.Num, cfg.ExtraNum)
				if beforeStatus == 0 && newStatus == 1 {
					model.RedpointSend(ctx, userID, 1, 0, 1, 1)
				}
			}
		}
	}
	// 5星英雄个数
	if nextStar == 5 {
		model.IncrFivestarNum(userID, 1)
		fiveNum := model.GetFivestarNum(userID)
		model.AchieveTaskHandle(ctx, userID, 6, fiveNum, 4101, 4106)
	}
	// 6星英雄个数
	if nextStar == 6 {
		model.IncrSixstarNum(userID, 1)
		sixNum := model.GetSixstarNum(userID)
		model.AchieveTaskHandle(ctx, userID, 7, sixNum, 4201, 4206)
	}

	return c.ResponseSuccessToMe(types.Map{
		"before_attr":     beforeAttr,
		"after_attr":      afterAttr,
		"returns":         returns,
		"unloads":         unloads,
		"hero_attr_skill": heroInfo[mainID].Skills,
		"skill":           skillLv,
	})
}

// ---- 英雄系统辅助函数 ----

// collectPositionHeroIDs 收集出战英雄ID
func collectPositionHeroIDs(pos1, pos2 *model.UserPosition) map[int]bool {
	ids := make(map[int]bool)
	if pos1 != nil {
		for id := range pos1.HeroPos {
			ids[id] = true
		}
	}
	if pos2 != nil {
		for id := range pos2.HeroPos {
			ids[id] = true
		}
	}

	return ids
}

// sortVoyageList 远航列表排序（对齐 PHP voyage_rank）
// 排序优先级: fin(完成>未完成) > status(已接>未接) > voyage_id(高品质ID大排前面)
// sortVoyageList 远航列表排序
// 优先级：已完成(fin=1)排最前 > 未接取(status=0)排中间 > 已接未完成排最后
// 同优先级内按 voyage_id 降序（高品质优先）
func sortVoyageList(list []types.Map) {
	if len(list) <= 1 {
		return
	}
	for i := 0; i < len(list)-1; i++ {
		for j := i + 1; j < len(list); j++ {
			a, b := list[i], list[j]
			orderA := voyageSortOrder(a)
			orderB := voyageSortOrder(b)

			swap := false
			if orderA == orderB {
				// 同优先级按 voyage_id 降序（高品质优先）
				if a.GetIntE("voyage_id") < b.GetIntE("voyage_id") {
					swap = true
				}
			} else if orderA > orderB {
				swap = true
			}

			if swap {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

// voyageSortOrder 返回排序权重，数值越小越靠前
// 0: 已完成待领取（fin=1）
// 1: 未接取（status=0）
// 2: 已接取进行中（status=1, fin=0）
func voyageSortOrder(v types.Map) int {
	if v.GetIntE("fin") == 1 {
		return 0
	}
	if v.GetIntE("status") == 0 {
		return 1
	}
	return 2
}

// markNeedStatus 标记升星材料英雄的出战/远航状态
func markNeedStatus(data types.Map, posHeroIDs map[int]bool, voyageHero []string) {
	voyageSet := make(map[string]bool)
	for _, v := range voyageHero {
		voyageSet[v] = true
	}

	type HeroStatus struct {
		*table.UserHero
		Status int `json:"stutas"`
	}

	if needs, err := types.ToMapArray(data["needs"], ""); err == nil {
		for nk, val := range needs {
			if hasList, ok := val["has"].([]*table.UserHero); ok {
				tmp := make([]*HeroStatus, len(hasList))

				for hk, has := range hasList {
					hasID := has.Id
					tmp[hk] = &HeroStatus{UserHero: has}
					if posHeroIDs[hasID] {
						tmp[hk].Status = 1
					} else if voyageSet[types.ToString(hasID)] {
						tmp[hk].Status = 2
					} else {
						tmp[hk].Status = 0
					}
				}
				needs[nk]["has"] = tmp
			}
		}
		data["needs"] = needs
	}
}

// getSkillDiff 获取新旧技能对比
func getSkillDiff(skillsOld, skillsNew *logic.SkillBaseSet, idx int) types.Map {
	var oldSkill, newSkill interface{}
	if skillsOld != nil && idx < len(skillsOld.Skills) {
		oldSkill = skillsOld.Skills[idx]
	}
	if skillsNew != nil && idx < len(skillsNew.Skills) {
		newSkill = skillsNew.Skills[idx]
	}
	return types.Map{"old_skill": oldSkill, "new_skill": newSkill}
}
