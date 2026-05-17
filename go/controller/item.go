package controller

import (
	"context"
	"math/rand"
	"sort"

	"server_golang/common/types"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// 道具系统

func (c *ShinelightController) OpenItemAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	itemID := c.Params.GetIntE("item_id")

	if item.NotEnough(userID, itemID, 1) {
		return c.ResponseError(3567, "数量不够")
	}

	open, _ := item.Open(userID, itemID, 0, 1)
	return c.ResponseSuccessToMe(types.Map{"items": open})
}

// 碎片合成英雄
func (c *ShinelightController) HeroCombineAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	itemID := c.Params.GetIntE("item_id")
	num := c.Params.GetIntE("num")
	if num < 1 {
		num = 1
	}

	itemInfo := item.Table[itemID]

	cnt := model.GetUserCountHero(ctx, userID)
	if cnt+num > 120 {
		return c.ResponseError(4521231, "背包容量不够，请献祭英雄")
	}

	if itemInfo == nil {
		return c.ResponseError(452123, "input item_id is wrong")
	}

	star := itemInfo.Star
	needPer := 20
	if star == 4 {
		needPer = 30
	} else if star >= 5 {
		needPer = 50
	}
	totalNeed := needPer * num

	if item.NotEnough(userID, itemID, totalNeed) {
		return c.ResponseError(54367, "碎片数量不够")
	}

	item.Sub(userID, itemID, totalNeed)

	result := make([]types.Map, 0, num)
	for i := 0; i < num; i++ {
		if len(itemInfo.Open) == 0 {
			continue
		}
		heroID := itemInfo.Open[rand.Intn(len(itemInfo.Open))]
		result = append(result, types.Map{"hero_id": heroID, "star": star})
		model.InsertNewUserHero(ctx, userID, heroID, star)
	}

	// 成就任务 - 4星英雄
	if star == 4 {
		model.IncrFourstarNum(userID, num)
		fourNum := model.GetFourstarNum(userID)
		model.AchieveTaskHandle(ctx, userID, 5, fourNum, 4001, 4005)
	}
	// 成就任务 - 5星英雄
	if star == 5 {
		model.IncrFivestarNum(userID, num)
		fiveNum := model.GetFivestarNum(userID)
		model.AchieveTaskHandle(ctx, userID, 6, fiveNum, 4101, 4106)
	}

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 5, 1)
	if itemID == 502007 {
		model.GuideTaskHandle(ctx, userID, 20, 1)
	}

	return c.ResponseSuccessToMe(types.Map{"heros": result})
}

// 返回用户道具
func (c *ShinelightController) UserItemAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	itemType := c.Params.GetIntE("item_type")
	if itemType < 1 || itemType > 6 {
		return c.ResponseError(9719, "item_type error")
	}

	data := item.GetListByType(userID, itemType)
	if data == nil {
		data = []types.Map{}
	}

	if itemType == 5 {
		for k, v := range data {
			color := v.GetIntE("color")
			num := v.GetIntE("num")
			can := 0
			if (color == 5 && num >= 50) || (color == 4 && num >= 30) || (color == 3 && num >= 20) {
				can = 1
			}
			data[k]["can"] = can
		}
	}

	// 按 item_id 升序排序，相同 item_id 按 num 降序排序
	sort.Slice(data, func(i, j int) bool {
		itemIDi := data[i].GetIntE("item_id")
		itemIDj := data[j].GetIntE("item_id")
		if itemIDi != itemIDj {
			return itemIDi < itemIDj
		}
		return data[i].GetIntE("num") > data[j].GetIntE("num")
	})

	result := types.Map{"list": data}
	if itemType == 4 {
		result["ronglian"] = item.Total(userID, 21)
	}
	return c.ResponseSuccessToMe(result)
}

// 返回用户装备
func (c *ShinelightController) UserEquipItemAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	subType := c.Params.GetIntE("sub_type")
	if subType < 1 || subType > 4 {
		return c.ResponseError(9319, "sub_type error")
	}

	data := item.GetListByType(userID, 3)

	ret := make([]types.Map, 0)
	for _, v := range data {
		iid := v.GetIntE("item_id")
		if info := item.Table[iid]; info != nil && info.TypeSub == subType {
			ret = append(ret, v)
		}
	}
	return c.ResponseSuccessToMe(types.Map{"list": ret})
}
