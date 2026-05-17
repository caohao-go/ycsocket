package controller

import (
	"context"
	"fmt"
	"strings"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/table"
)

// 献祭系统 — 对齐 PHP `Shinelight::sacrificeInfoAction` / `sacrificeAction`

// SacrificeInfoAction 献祭信息：英雄列表 + 出战阵位英雄ID + 远航英雄ID
func (c *ShinelightController) SacrificeInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := model.GetUserHeroList(ctx, userID)

	// 在竞技场和主线剧情、远航任务中的英雄不能献祭
	positionHero1 := model.GetUserPositionByID(ctx, userID, 1)
	positionHero2 := model.GetUserPositionByID(ctx, userID, 2)
	mergedHeroIDs := make([]int, 0)
	heroIDSet := collectPositionHeroIDs(positionHero1, positionHero2)
	for heroID := range heroIDSet {
		mergedHeroIDs = append(mergedHeroIDs, heroID)
	}

	voyageHero := model.GetVoyageHero(ctx, userID)

	// 每个英雄补充 property 字段（对齐 PHP `Heroinfo::$base_datas[hero_id]['property']`）
	tmp := make([]types.Map, len(data))
	for k, v := range data {
		tmp[k] = types.ObjectToMap(v)
		tmp[k]["property"] = logic.HeroProperty(v.HeroId)
	}

	return c.ResponseSuccessToMe(types.Map{
		"list":        tmp,
		"hero_id":     mergedHeroIDs,
		"voyage_hero": voyageHero,
	})
}

// SacrificeAction 献祭英雄
func (c *ShinelightController) SacrificeAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	idsStr := c.Params.GetStringE("ids")
	idsParts := strings.Split(idsStr, ",")

	lock.Lock(fmt.Sprintf("sacrifice%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("sacrifice%d", userID))

	// 批量预取英雄信息（对齐 PHP getUserHeroByIds + 建 user_hero_map）
	parsedIDs := make([]int, 0, len(idsParts))
	for _, s := range idsParts {
		nid := types.ToIntE(strings.TrimSpace(s))
		parsedIDs = append(parsedIDs, nid)
	}
	heroList := model.GetUserHeroByIDs(ctx, parsedIDs)
	heroMap := make(map[int]*table.UserHero)
	for _, h := range heroList {
		heroMap[h.Id] = h
	}

	rewards := make([]util.TypeNum, 0)
	returns := make([]util.TypeNum, 0)
	unloads := make([]util.TypeNum, 0)

	for _, nid := range parsedIDs {
		userHero, ok := heroMap[nid]
		if !ok || userHero == nil {
			return c.ResponseError(666852, "英雄不存在")
		}
		if userHero.UserId != userID {
			return c.ResponseError(666832, "英雄不存在")
		}

		star := userHero.Star
		stage := userHero.Stage
		lv := userHero.Lv
		heroID := userHero.HeroId

		// 献祭奖励
		property := logic.HeroProperty(heroID)
		sacrificeHeroType := 1
		if property == 4 || property == 5 {
			sacrificeHeroType = 2
		}

		if reward, ok := logic.SacrificeDatas[1][sacrificeHeroType][star]; ok {
			rewards = util.Merge(rewards, reward)
		}

		// 升星返还
		returns = util.Merge(returns, logic.GetReturnItems(star, stage, lv))

		// 装备返还
		unload := logic.GetEquipmentReturn(userHero.Fit, userHero.Fu, userHero.Lv, userHero.Star)
		unloads = util.Merge(unloads, unload)
	}

	// 删除 + 发奖
	model.DeleteUserHeroByIDs(ctx, userID, parsedIDs)
	model.GiveReward(userID, rewards...)
	model.GiveReward(userID, returns...)
	model.GiveReward(userID, unloads...)

	// 成就任务（献祭次数累加）
	model.AddSacrificeNum(userID, len(parsedIDs))
	sacrificeNum := model.GetSacrificeNum(userID)
	model.AchieveTaskHandle(ctx, userID, 14, sacrificeNum, 10001, 10006)

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 33, 1)

	return c.ResponseSuccessToMe(types.Map{
		"rewards": rewards,
		"returns": returns,
		"unloads": unloads,
	})
}
