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
	"server_golang/repo/table"
)

// 召唤+英雄转换

func (c *ShinelightController) ZhaohuanInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	var baseTime, highTime int
	basePower := model.GetUserPower(ctx, userID, model.PowerTypeBaseZhaohuanQuan, &baseTime)
	highPower := model.GetUserPower(ctx, userID, model.PowerTypeHighZhaohuanQuan, &highTime)

	data := types.Map{
		"base_zhaohuan_quan":           item.Total(userID, 20901),
		"friends_num":                  item.Total(userID, 10),
		"high_zhaohuan_quan":           item.Total(userID, 21001),
		"zhaohuan_score":               item.Total(userID, 16),
		"base_zhaohuan_free_count":     basePower,
		"high_zhaohuan_free_count":     highPower,
		"base_zhaohuan_free_left_time": baseTime,
		"high_zhaohuan_free_left_time": highTime,
	}
	return c.ResponseSuccessToMe(data)
}

func (c *ShinelightController) HeroZhaoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	typ := c.Params.GetIntE("type")
	// type: 2-基础单次 3-基础10次 4-友情单次 5-友情10次 6-高级单次 7-高级10次 8-积分 9-钻石高级1次 10-钻石高级10次
	if typ < 2 || typ > 10 {
		return c.ResponseError(84238, "类型错误")
	}

	userGrade := model.GetUserAttr(userID)

	if typ == 8 && userGrade.GetIntE("vip_level") < 3 {
		return c.ResponseError(84239, "vip3以上玩家才能使用")
	}

	// 包裹检查
	cnt := model.GetUserCountHero(ctx, userID)
	if (typ == 2 || typ == 4 || typ == 6 || typ == 8 || typ == 9) && cnt >= 120 {
		return c.ResponseError(99, "英雄包裹不够")
	}
	if (typ == 3 || typ == 5 || typ == 7 || typ == 10) && cnt > 110 {
		return c.ResponseError(99, "英雄包裹不够")
	}

	lock.Lock(fmt.Sprintf("heroZhao%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("heroZhao%d", userID))

	free := false
	var costType, costNum int

	switch typ {
	case 2: // 基础召唤单次
		var baseTime int
		power := model.GetUserPower(ctx, userID, model.PowerTypeBaseZhaohuanQuan, &baseTime)
		if power > 0 {
			free = true
			model.SubUserPower(ctx, userID, model.PowerTypeBaseZhaohuanQuan, &baseTime, 1)
		} else {
			costType, costNum = 20901, 1
		}
	case 3: // 基础10次
		costType, costNum = 20901, 10
	case 4: // 友情单次
		costType, costNum = 10, 100
	case 5: // 友情10次
		costType, costNum = 10, 1000
	case 6: // 高级单次
		var highTime int
		power := model.GetUserPower(ctx, userID, model.PowerTypeHighZhaohuanQuan, &highTime)
		if power > 0 {
			free = true
			model.SubUserPower(ctx, userID, model.PowerTypeHighZhaohuanQuan, &highTime, 1)
		} else {
			costType, costNum = 21001, 1
		}
	case 7: // 高级10次
		costType, costNum = 21001, 10
	case 8: // 积分召唤
		costType, costNum = 16, 1000
	case 9: // 钻石高级1次
		costType, costNum = 2, 220
	case 10: // 钻石高级10次
		costType, costNum = 2, 2000
	}

	if !free && costNum > 0 {
		if item.NotEnough(userID, costType, costNum) {
			return c.ResponseError(666666, item.Name(costType)+"不够")
		}
	}

	// 召唤积分（与 PHP 一致：先计算积分，扣除消耗后再加积分）
	zhaohuanScore := map[int]int{2: 1, 3: 10, 4: 2, 5: 20, 6: 20, 7: 200, 8: 0, 9: 20, 10: 200}

	// 随机抽取英雄（与 PHP 一致：通过 getTanbaoItemRand 从 items_collection 随机取物品，再 getOpenItems 打开）
	realType := typ
	isTenDraw := typ == 3 || typ == 5 || typ == 7 || typ == 10

	if isTenDraw {
		// PHP: $type = $type == 10 ? 7 : $type; $rewards = Tanbao::getTanbaoItemRand($type, 10);
		if realType == 10 {
			realType = 7
		}
	} else {
		// PHP: $type = $type == 9 ? 6 : $type; $rewards = Tanbao::getTanbaoItemRand($type, 1, $user_grade['lv']);
		if realType == 9 {
			realType = 6
		}
	}

	var rewards []*table.ItemsCollection
	if isTenDraw {
		rewards = logic.GetTanbaoItemRand(realType, 10, 1)
	} else {
		rewards = logic.GetTanbaoItemRand(realType, 1, userGrade.GetIntE("lv"))
	}

	// PHP: foreach ($rewards as $v) { $open_hero = getOpenItems($v['items_id']); insertNewUserHero(...); $data['reward'][] = $open_hero; }
	rewardData := make([]types.Map, 0, len(rewards))
	for _, v := range rewards {
		itemsID := v.ItemsId
		heroID, star, ok := item.GetOpenHeroItem(itemsID)
		if !ok {
			continue
		}
		model.InsertNewUserHero(ctx, userID, heroID, star)
		rewardData = append(rewardData, types.Map{
			"hero_id":  heroID,
			"star":     star,
			"items_id": itemsID,
			"number":   1,
		})
	}

	// 扣除消耗（与 PHP 一致：扣除在抽取之后）
	if !free && costNum > 0 {
		item.Sub(userID, costType, costNum)
	}

	// 增加召唤积分（与 PHP 一致）
	if score, ok := zhaohuanScore[typ]; ok && score > 0 {
		item.Add(userID, 16, score, nil)
	}

	// 日常任务
	// 修复：十连抽时 PHP 直接调用 incrDailyTaskCountById($userId, 1, 10003, 10) 增加 10，
	// 而非走 setDalilyTaskFinish（该函数每次只增加 1）。
	// 原 Go 代码 SetDailyTaskFinish(ctx, userID, 10003, 10) 中的 10 被误当作 taskType 传入，
	// 导致十连抽只增加 1 次计数。
	if isTenDraw {
		// 十连抽：检查后直接增加 10（与 PHP Shinelight.php:5094-5099 一致）
		finishCount := model.GetDailyTaskCountByID(ctx, userID, 1, 10003)
		if finishCount < 1 { // 10003 任务的 need_count 为 1
			model.IncrDailyTaskCountByID(ctx, userID, 1, 10003, 10)
		}
	} else {
		model.SetDailyTaskFinish(ctx, userID, 10003, 1)
	}

	// 成就任务 - 统计4星和5星英雄个数
	fourCount := 0
	fiveCount := 0
	for _, hero := range rewardData {
		star := hero.GetIntE("star")
		if star == 4 {
			fourCount++
		} else if star >= 5 {
			fiveCount++
		}
	}
	if fourCount > 0 {
		model.IncrFourstarNum(userID, fourCount)
		fourNum := model.GetFourstarNum(userID)
		model.AchieveTaskHandle(ctx, userID, 5, fourNum, 4001, 4005)
	}
	if fiveCount > 0 {
		model.IncrFivestarNum(userID, fiveCount)
		fiveNum := model.GetFivestarNum(userID)
		model.AchieveTaskHandle(ctx, userID, 6, fiveNum, 4101, 4106)
	}

	// 引导任务
	if typ == 6 {
		model.GuideTaskHandle(ctx, userID, 10, 1)
	}

	return c.ResponseSuccessToMe(types.Map{"reward": rewardData})
}

// ======================== 英雄转换 ========================

func (c *ShinelightController) HeroExchangeAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroDBID := c.Params.GetIntE("id")

	userHero := model.GetUserHeroByID(ctx, heroDBID)
	if userHero.UserId != userID {
		return c.ResponseError(43241, "hero id error")
	}
	star := userHero.Star
	if star != 4 && star != 5 {
		return c.ResponseError(43241, "hero_star error")
	}

	cost := 20
	if star == 5 {
		cost = 100
	}
	if item.NotEnough(userID, 20601, cost) {
		return c.ResponseError(999999, "先知精华不足")
	}
	item.Sub(userID, 20601, cost)

	newHeroID := logic.GetRandomSamePropertyHero(userHero.HeroId, star)
	exchangeID := fmt.Sprintf("%d_%d", userID, time.Now().UnixNano())
	model.SetHeroExchange(ctx, userID, exchangeID, heroDBID, newHeroID)

	model.GuideTaskHandle(ctx, userID, 108, 1)

	return c.ResponseSuccessToMe(types.Map{"exchange_id": exchangeID, "new_hero_id": newHeroID})
}

func (c *ShinelightController) SaveHeroExchangeAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	exchangeID := c.Params.GetStringE("exchange_id")
	save := c.Params.GetIntE("save") // 1-保存 2-取消

	var newHeroID int
	if save == 1 {
		oldID, nid := model.GetHeroExchange(ctx, userID, exchangeID)
		if oldID == 0 {
			return c.ResponseError(4325, "操作超时")
		}

		model.UpdateUserHeroID(ctx, userID, oldID, nid)
		newHeroID = nid
	} else {
		model.DelHeroExchange(ctx, userID, exchangeID)
	}

	return c.ResponseSuccessToMe(types.Map{"new_hero_id": newHeroID})
}

// 先知系统

func (c *ShinelightController) XianzhiInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := types.Map{
		"shuijing": item.Total(userID, 21901),
		"jinghua":  item.Total(userID, 20601),
		"jiejing":  item.Total(userID, 19),
	}
	return c.ResponseSuccessToMe(data)
}

func (c *ShinelightController) XianzhiZhaoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	typ := c.Params.GetIntE("type") // 15-风 16-水 17-火 18-光暗
	if typ < 15 || typ > 18 {
		return c.ResponseError(84238, "类型错误")
	}

	shuijing := item.Total(userID, 21901)
	if shuijing < 1 {
		return c.ResponseError(888888, "水晶不够")
	}

	item.Sub(userID, 21901, 1)
	item.Add(userID, 19, 1, nil)
	item.Add(userID, 20601, 30, nil)

	// 随机获得英雄（与 PHP 一致：通过 getTanbaoItemRand 从 items_collection 随机取物品）
	rewards := logic.GetTanbaoItemRand(typ, 1, 1)
	var rewardData []types.Map
	giveRewards := make([]util.TypeNum, 0)
	for _, v := range rewards {
		itemsID := v.ItemsId
		num := v.Number
		rewardData = append(rewardData, types.Map{
			"items_id": itemsID,
			"number":   num,
		})
		giveRewards = append(giveRewards, util.TypeNum{Type: itemsID, Num: num})
	}

	if len(giveRewards) > 0 {
		model.GiveReward(userID, giveRewards...)
	}

	// 成就任务
	model.IncrXianzhiNum(userID)
	xianzhiNum := model.GetXianzhiNum(userID)
	model.AchieveTaskHandle(ctx, userID, 11, xianzhiNum, 8001, 8006)

	model.GuideTaskHandle(ctx, userID, 107, 1)

	return c.ResponseSuccessToMe(types.Map{
		"reward":   rewardData,
		"shuijing": item.Total(userID, 21901),
		"jinghua":  item.Total(userID, 20601),
		"jiejing":  item.Total(userID, 19),
	})
}
