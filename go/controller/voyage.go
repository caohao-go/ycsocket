package controller

import (
	"context"
	"fmt"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// 远航系统

func (c *ShinelightController) VoyageInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")

	// 远航令牌(item_id=6)
	exp := item.Total(userID, 6)
	// 每日免费刷新次数
	freeCnt := model.GetVoyageFreeCnt(ctx, userID)
	// 远航券(item_id=21101)
	voyageQuan := item.Total(userID, 21101)

	voyageList := model.GetVoyageList(ctx, userID, lv)

	// 排序：已完成优先 > 未接取其次 > 已接未完成最后
	sortVoyageList(voyageList)

	return c.ResponseSuccessToMe(types.Map{
		"exp":         exp,
		"free_cnt":    freeCnt,
		"voyage_quan": voyageQuan,
		"voyage_list": voyageList,
	})
}

// 派遣远航（对齐 PHP 原版）
func (c *ShinelightController) BeatVoyageAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")
	herosStr := c.Params.GetStringE("heros")

	// 解析英雄ID列表
	var heros []int
	if herosStr != "" {
		heros = types.SplitInt(herosStr, ",")
	}
	if len(heros) == 0 {
		return c.ResponseError(4527, "没有选英雄")
	}

	// 每次固定消耗2000远航令牌(item_id=6)
	const perSubNum = 2000
	if item.NotEnough(userID, 6, perSubNum) {
		return c.ResponseError(999999, "远航令不够")
	}
	item.Sub(userID, 6, perSubNum)

	// 执行派遣
	data := model.BeatVoyage(ctx, userID, id, heros)

	// 成就任务：橙色远航计数
	itemColection := data.GetIntE("item_colection")
	if itemColection == 19 { // 橙色
		model.IncrVoyageChengNum(userID)
		chengNum := model.GetVoyageChengNum(userID)
		model.AchieveTaskHandle(ctx, userID, 12, chengNum, 9001, 9007)
	} else if itemColection == 20 { // 红色
		model.IncrVoyageHongNum(userID)
		hongNum := model.GetVoyageHongNum(userID)
		model.AchieveTaskHandle(ctx, userID, 13, hongNum, 9101, 9107)
	}

	// 日常任务：远航订单3次

	model.SetDailyTaskFinish(ctx, userID, 10010, 1)

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 61, 1)
	model.GuideTaskHandleIncr(ctx, userID, 65, 1)

	return c.ResponseSuccessToMe(types.Map{})
}

// 远航中的英雄（对齐 PHP 原版）
func (c *ShinelightController) VoyageIngHerosAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	data := model.GetVoyageHero(ctx, userID)
	return c.ResponseSuccessToMe(types.Map{"hero_ids": data})
}

// 领取远航奖励（对齐 PHP 原版）
func (c *ShinelightController) LingquVoyageAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")

	// 分布式锁防并发
	lock.Lock(fmt.Sprintf("lingquVoyage%d", userID), 10)
	defer lock.Unlock("lingquVoyage" + types.ToString(userID))

	datas := model.GetVoyageList(ctx, userID, 0)
	if len(datas) == 0 {
		return c.ResponseError(55555, "没有远航任务")
	}

	ok, itemsID, num := model.LingquVoyage(ctx, userID, id)
	if !ok {
		return c.ResponseError(4302, "不满足条件")
	}

	item.Add(userID, itemsID, num, nil)

	// 释放英雄（清理过期的锁记录）
	model.GetVoyageHero(ctx, userID)

	return c.ResponseSuccessToMe(types.Map{"items_id": itemsID, "num": num})
}

// 加速完成远航（对齐 PHP 原版）
func (c *ShinelightController) AccelerateVoyageAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")

	voyageList := model.GetVoyageList(ctx, userID, 0)
	accelerateConsume := 0
	found := false
	for _, v := range voyageList {
		if v.GetIntE("status") == 1 && v.GetIntE("id") == id {
			found = true

			// accelerate_num 已被 GetVoyageList 按剩余时间比例折算，直接使用
			accelerateConsume = v.GetIntE("accelerate_num")

			// 释放英雄（对齐 PHP：无论 accelerateConsume 是否 > 0 都释放）
			beatHeros, ok, e := v.GetInt64Array("beat_heros")
			if ok && e == nil {
				model.DelVoyageHero(ctx, userID, beatHeros...)
			}

			// 扣钻石（对齐 PHP：accelerateConsume == 0 时相当于免费加速）
			if accelerateConsume > 0 {
				if item.NotEnough(userID, 2, accelerateConsume) {
					return c.ResponseError(666667, "钻石不够")
				}
				item.Sub(userID, 2, accelerateConsume)
			}
			break
		}
	}

	if !found {
		return c.ResponseError(4302, "不可加速")
	}

	// 执行加速：回拨 beat_time 使 left_time<=0
	model.AccelerateVoyage(ctx, userID, id)

	// 清理英雄锁
	model.GetVoyageHero(ctx, userID)

	return c.ResponseSuccessToMe(types.Map{})
}

// 刷新远航列表（对齐 PHP 原版）
func (c *ShinelightController) RefreshVoyageAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")

	// 执行刷新
	model.RefreshVoyageList(ctx, userID, lv)

	// 刷新付费逻辑（三级优先级）：免费次数 > 远航券 > 钻石30个
	if model.GetVoyageFreeCnt(ctx, userID) > 0 {
		model.IncrVoyageFreeCnt(ctx, userID)
	} else if item.Total(userID, 21101) > 0 {
		item.Sub(userID, 21101, 1)
	} else if item.NotEnough(userID, 2, 30) {
		return c.ResponseError(666666, "钻石不够")
	} else {
		item.Sub(userID, 2, 30)
	}

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 62, 1)

	return c.ResponseSuccessToMe(types.Map{})
}
