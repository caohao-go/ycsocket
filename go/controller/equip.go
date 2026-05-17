package controller

import (
	"context"
	"fmt"
	"math/rand"
	"reflect"
	"strings"

	"server_golang/common/json"
	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// 装备系统

func (c *ShinelightController) DressEquipAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetIntE("hero_id")
	equipParam := c.Params.GetStringE("equip")
	var equips map[string]int
	json.Unmarshal(equipParam, &equips)
	if len(equips) == 0 {
		return c.ResponseError(64636, "input error")
	}

	userHero := model.GetUserHeroByID(ctx, heroID)
	if userHero == nil || userHero.UserId != userID {
		return c.ResponseError(43241, "hero id error")
	}

	lock.Lock(fmt.Sprintf("dressEquip%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("dressEquip%d", userID))

	var fit map[string]int
	fitStr := userHero.Fit
	json.Unmarshal(fitStr, &fit)
	if fit == nil {
		fit = make(map[string]int)
	}

	validPos := map[string]bool{"weapon": true, "dress": true, "shoes": true, "head": true}
	beforeFits := make([]int, 0)

	for pos, equipItemID := range equips {
		if !validPos[pos] {
			return c.ResponseError(64326, "input error")
		}
		cnt := item.Total(userID, equipItemID)
		if cnt == 0 {
			return c.ResponseError(64326, "装备不存在")
		}
		if oldID, ok := fit[pos]; ok && oldID > 0 {
			beforeFits = append(beforeFits, oldID)
		}
		fit[pos] = equipItemID
	}

	if err := model.UpdateUserHeroFit(ctx, userID, heroID, fit); err != nil {
		return c.ResponseError(99, "system error")
	}

	for _, equipItemID := range equips {
		item.Sub(userID, equipItemID, 1)
	}
	for _, oldID := range beforeFits {
		item.Add(userID, oldID, 1, nil)
	}

	model.GuideTaskHandle(ctx, userID, 8, 1)

	return c.ResponseSuccessToMe(types.Map{})
}

// 卸载装备
func (c *ShinelightController) UnloadEquipAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetIntE("hero_id")
	posParam := c.Params.GetStringE("pos")
	poses := strings.Split(posParam, ",")
	if len(poses) == 0 || posParam == "" {
		return c.ResponseError(64736, "input error")
	}

	userHero := model.GetUserHeroByID(ctx, heroID)
	if userHero == nil || userHero.UserId != userID {
		return c.ResponseError(43241, "hero id error")
	}

	lock.Lock(fmt.Sprintf("unloadEquip%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("unloadEquip%d", userID))

	var fit map[string]int
	json.Unmarshal(userHero.Fit, &fit)
	if fit == nil {
		fit = make(map[string]int)
	}

	validPos := map[string]bool{"weapon": true, "dress": true, "shoes": true, "head": true}
	unloadItems := make([]int, 0)

	for _, pos := range poses {
		pos = strings.TrimSpace(pos)
		if !validPos[pos] {
			return c.ResponseError(64326, "input error")
		}
		if oldID, ok := fit[pos]; ok && oldID > 0 {
			unloadItems = append(unloadItems, oldID)
			delete(fit, pos)
		}
	}

	if err := model.UpdateUserHeroFit(ctx, userID, heroID, fit); err != nil {
		return c.ResponseError(99, "system error")
	}

	for _, itemID := range unloadItems {
		item.Add(userID, itemID, 1, nil)
	}

	return c.ResponseSuccessToMe(types.Map{})
}

// 装备合成
func (c *ShinelightController) MergeEquipAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	itemID := c.Params.GetIntE("item_id")
	num := c.Params.GetIntE("num")

	mergeInfo, ok := logic.MergeDatas[itemID]
	if !ok {
		return c.ResponseError(93429, "物品不能参与合成")
	}

	lock.Lock(fmt.Sprintf("mergeEquip%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("mergeEquip%d", userID))

	myNum := item.Total(userID, itemID)
	if myNum < num {
		return c.ResponseError(9429, "数量不够")
	}

	updateNum := num / 3
	consumeCoin := types.ToIntE(mergeInfo.ConsumeCoin)
	needMoney := updateNum * consumeCoin
	if item.NotEnough(userID, 1, needMoney) {
		return c.ResponseError(666666, "金币不够")
	}

	item.Sub(userID, 1, needMoney)
	item.Sub(userID, itemID, updateNum*3)
	finalID := types.ToIntE(mergeInfo.FinalID)
	item.Add(userID, finalID, updateNum, nil)

	// 引导任务
	if itemID >= 30101 && itemID <= 30199 {
		model.GuideTaskHandle(ctx, userID, 27, 1)
	} else if itemID >= 30201 && itemID <= 30299 {
		model.GuideTaskHandle(ctx, userID, 72, 1)
	}

	return c.ResponseSuccessToMe(types.Map{"need_money": needMoney, "add_num": updateNum})
}

// 一键装备合成（级联合成到最高级）
func (c *ShinelightController) MergeEquipAllAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	itemID := c.Params.GetIntE("item_id")
	num := c.Params.GetIntE("num")

	_, ok := logic.MergeDatas[itemID]
	if !ok {
		return c.ResponseError(93429, "物品不能参与合成")
	}

	lock.Lock(fmt.Sprintf("mergeEquip%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("mergeEquip%d", userID))

	myNum := item.Total(userID, itemID)
	if myNum < num {
		return c.ResponseError(9429, "数量不够")
	}

	// 第一阶段：纯计算，不做任何物品操作
	type mergeStep struct {
		originalID int
		finalID    int
		consumeNum int // 消耗的原材料数量(updateNum*3)
		produceNum int // 产出的高级装备数量(updateNum)
		costCoin   int // 消耗金币
	}

	var steps []mergeStep
	totalMoney := 0
	rewards := []types.Map{}
	currentID := itemID
	currentNum := num

	for currentNum >= 3 {
		curMerge, curOk := logic.MergeDatas[currentID]
		if !curOk {
			break
		}
		updateNum := currentNum / 3
		consumeCoin := types.ToIntE(curMerge.ConsumeCoin)
		costCoin := updateNum * consumeCoin
		totalMoney += costCoin
		finalID := types.ToIntE(curMerge.FinalID)

		steps = append(steps, mergeStep{
			originalID: currentID,
			finalID:    finalID,
			consumeNum: updateNum * 3,
			produceNum: updateNum,
			costCoin:   costCoin,
		})

		// 合成产物中的余数（中间产物剩余）
		remainder := currentNum % 3
		if remainder > 0 && currentID != itemID {
			rewards = append(rewards, types.Map{"item_id": currentID, "num": remainder})
		}

		currentID = finalID
		currentNum = updateNum
	}

	// 最终产物（不够3个继续合成的那一级）
	if currentNum > 0 {
		rewards = append(rewards, types.Map{"item_id": currentID, "num": currentNum})
	}

	if len(steps) == 0 {
		return c.ResponseError(93429, "数量不足以合成")
	}

	// 校验金币
	if item.NotEnough(userID, 1, totalMoney) {
		return c.ResponseError(666666, "金币不够")
	}

	// 第二阶段：执行所有物品操作
	item.Sub(userID, 1, totalMoney)
	for _, step := range steps {
		item.Sub(userID, step.originalID, step.consumeNum)
		item.Add(userID, step.finalID, step.produceNum, nil)

		// 引导任务
		if step.originalID >= 30101 && step.originalID <= 30199 {
			model.GuideTaskHandle(ctx, userID, 27, 1)
		} else if step.originalID >= 30201 && step.originalID <= 30299 {
			model.GuideTaskHandle(ctx, userID, 72, 1)
		}
	}

	return c.ResponseSuccessToMe(types.Map{"need_money": totalMoney, "rewards": rewards})
}

// 符文系统

// 符文合成
func (c *ShinelightController) MergeRuneAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	itemID := c.Params.GetIntE("item_id")
	num := c.Params.GetIntE("num")

	myNum := item.Total(userID, itemID)
	if num < 1 || num > 5 {
		return c.ResponseError(4129, "数量错误")
	}
	if myNum < num {
		return c.ResponseError(9429, "数量不够")
	}

	mergeMap, ok := logic.RuneMergeDatas[itemID]
	if !ok {
		return c.ResponseError(93429, "配置错误")
	}
	mergeInfo, ok := mergeMap[num]
	if !ok {
		return c.ResponseError(93429, "配置错误")
	}

	lock.Lock(fmt.Sprintf("mergeRune%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("mergeRune%d", userID))

	needMoney := types.ToIntE(mergeInfo.ConsumeCoin)
	if item.NotEnough(userID, 1, needMoney) {
		return c.ResponseError(666666, "金币不够")
	}

	item.Sub(userID, 1, needMoney)
	item.Sub(userID, itemID, num)

	success := logic.MergeRateSuccess(mergeInfo.SuccessRate)
	var get types.Map
	// 合成无论成功失败都增加熔炼值
	item.Add(userID, 21, mergeInfo.FailGet.Num/15, nil)
	if success {
		finalID := types.ToIntE(mergeInfo.FinalID)
		item.Add(userID, finalID, 1, nil)
		get = types.Map{"type": finalID, "num": 1}
	} else {
		failType := types.ToIntE(mergeInfo.FailGet.Type)
		item.Add(userID, failType, mergeInfo.FailGet.Num, nil)
		get = types.Map{"type": failType, "num": mergeInfo.FailGet.Num}
	}

	model.IncrMergeruneNum(userID)
	mergeNum := model.GetMergeruneNum(userID)
	model.AchieveTaskHandle(ctx, userID, 9, mergeNum, 6001, 6007)

	return c.ResponseSuccessToMe(types.Map{"need_money": needMoney, "success": success, "get": get})
}

func (c *ShinelightController) RonglianLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	lock.Lock(fmt.Sprintf("ronglianLingqu%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("ronglianLingqu%d", userID))

	myNum := item.Total(userID, 21)
	if myNum < 1000 {
		return c.ResponseError(94296, "熔炼值不够")
	}
	item.Sub(userID, 21, 1000)

	newFuId := item.Sequence()
	prop, _ := item.Add(userID, 40401, 1, nil, newFuId)
	newFu := types.Map{
		"item_id": 40401,
		"num":     1,
		"id":      newFuId,
		"prop":    prop,
	}

	myNumRonglian := item.Total(userID, 21)

	return c.ResponseSuccessToMe(types.Map{
		"ronglianzhi": myNumRonglian,
		"new_fu":      newFu,
	})
}

// 符文重铸
func (c *ShinelightController) UserRuneResetAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	resetType := c.Params.GetIntE("type") // 0-背包 1-英雄佩戴
	itemID := c.Params.GetIntE("item_id")
	id := c.Params.GetInt64E("id")
	heroID := c.Params.GetIntE("hero_id")
	side := c.Params.GetStringE("side") // left, right

	consumes := item.RuneConsumeDatas[itemID]

	lock.Lock(fmt.Sprintf("userRuneReset%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("userRuneReset%d", userID))

	prop := item.InitRandRuneProp(itemID)
	exchangeID := fmt.Sprintf("%d_%d", userID, rand.Int63())

	var ret types.Map
	var saveData map[string]*util.Fu

	if resetType == 1 {
		userHero := model.GetUserHeroByID(ctx, heroID)
		if userHero == nil {
			return c.ResponseError(3699, "英雄不存在")
		}
		fus := util.ToHeroFus(userHero.Fu, userHero.Lv, userHero.Star)
		if len(fus) == 0 {
			return c.ResponseError(3646, "物品不存在")
		}
		if _, ok := fus[side]; !ok {
			return c.ResponseError(3646, "物品不存在")
		}

		oldFu := fus[side]
		newFu := util.Fu{
			Id:     oldFu.Id,
			ItemId: itemID,
			Unlock: oldFu.Unlock,
			Props:  prop,
		}
		saveData = map[string]*util.Fu{"old": oldFu, "new": &newFu}
		ret = types.Map{"hero_id": heroID, "old": oldFu, "new": newFu, "exchange_id": exchangeID}
	} else {
		itemData := item.GetItemByID(userID, itemID, id)
		if itemData == nil {
			return c.ResponseError(3636, "物品不存在")
		}
		oldFu := util.Fu{
			Id:     id,
			ItemId: itemID,
			Props:  itemData.Prop,
		}

		newFu := util.Fu{
			Id:     oldFu.Id,
			ItemId: oldFu.ItemId,
			Props:  prop,
		}
		saveData = map[string]*util.Fu{"old": &oldFu, "new": &newFu}
		ret = types.Map{"old": oldFu, "new": newFu, "exchange_id": exchangeID}
	}

	// 扣费（对齐 PHP：先校验金币/精华，再 subUserItem，最后 setRuneExchange）
	if len(consumes) >= 1 {
		if item.NotEnough(userID, consumes[0].Type, consumes[0].Num) {
			return c.ResponseError(666666, "金币不够")
		}
	}
	if len(consumes) >= 2 {
		if item.NotEnough(userID, consumes[1].Type, consumes[1].Num) {
			return c.ResponseError(34256, "精华不够")
		}
	}
	for _, cs := range consumes {
		item.Sub(userID, cs.Type, cs.Num)
	}

	model.SetRuneExchange(ctx, userID, exchangeID, saveData)

	return c.ResponseSuccessToMe(ret)
}

// 保存符文重铸结果
func (c *ShinelightController) SaveRuneAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	exchangeID := c.Params.GetStringE("exchange_id")
	save := c.Params.GetIntE("save") // 1-保存 2-取消
	resetType := c.Params.GetIntE("type")
	itemID := c.Params.GetIntE("item_id")
	id := c.Params.GetInt64E("id")
	heroID := c.Params.GetIntE("hero_id")
	side := c.Params.GetStringE("side")

	data := model.GetRuneExchange(ctx, userID, exchangeID)
	if data == nil {
		return c.ResponseError(4325, "操作超时")
	}

	if save == 1 {
		if resetType == 1 {
			userHero := model.GetUserHeroByID(ctx, heroID)
			if userHero == nil {
				return c.ResponseError(3699, "英雄不存在")
			}
			fu := util.ToHeroFus(userHero.Fu, userHero.Lv, userHero.Star)
			if len(fu) == 0 {
				return c.ResponseError(3646, "物品不存在")
			}

			if _, ok := fu[side]; !ok {
				return c.ResponseError(3646, "物品不存在")
			}

			newFu := data["new"]
			tmp := util.ToHeroFus(types.Map{side: newFu}, userHero.Lv, userHero.Star)
			fu[side] = tmp[side]
			model.UpdateUserHeroFu(ctx, userID, heroID, fu)
		} else {
			newFu := data["new"]
			item.UpdatePropByID(userID, itemID, id, newFu.Props)
		}
	} else {
		model.DelRuneExchange(ctx, userID, exchangeID)
	}

	return c.ResponseSuccessToMe(types.Map{})
}

// 返回用户符文
func (c *ShinelightController) UserRuneItemAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := item.GetListByType(userID, 4)

	ret := make(types.Map)
	for _, v := range data {
		ret[v.GetStringE("id")] = v
	}
	return c.ResponseSuccessToMe(types.Map{"list": ret})
}

// 穿戴符文
func (c *ShinelightController) DressRuneAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetIntE("hero_id")
	fuParam := c.Params.GetStringE("fu")
	var dressFus map[string][]interface{}
	json.Unmarshal(fuParam, &dressFus)
	if len(dressFus) == 0 {
		return c.ResponseError(64636, "input error")
	}

	userHero := model.GetUserHeroByID(ctx, heroID)
	if userHero == nil || userHero.UserId != userID {
		return c.ResponseError(43241, "hero id error")
	}

	fus := util.ToHeroFus(userHero.Fu, userHero.Lv, userHero.Star)

	beforeFus := make([]*util.Fu, 0)

	lock.Lock(fmt.Sprintf("dressRune%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("dressRune%d", userID))

	for _, side := range []string{"left", "right"} {
		dressInfo, ok := dressFus[side]
		if !ok || len(dressInfo) < 2 {
			continue
		}
		// 检查符文槽位是否解锁
		if fuSlot, ok2 := fus[side]; !ok2 || fuSlot.Unlock == 0 {
			return c.ResponseError(36460, "符文槽位未解锁")
		}
		dItemID := types.ToIntE(dressInfo[0])
		dID := types.ToInt64E(dressInfo[1])
		data := item.GetItemByID(userID, dItemID, dID)
		if data == nil {
			return c.ResponseError(64326, "符文不存在")
		}
		// 保存旧符文
		if old, ok2 := fus[side]; ok2 {
			beforeFus = append(beforeFus, old)
		}

		fus[side] = &util.Fu{
			Id:     dID,
			ItemId: dItemID,
			Unlock: 1,
			Props:  data.Prop,
		}
	}

	if err := model.UpdateUserHeroFu(ctx, userID, heroID, fus); err != nil {
		return c.ResponseError(99, "system error")
	}

	for _, dressInfo := range dressFus {
		if len(dressInfo) >= 2 {
			dItemID := types.ToIntE(dressInfo[0])
			dID := dressInfo[1]
			item.Sub(userID, dItemID, 1, types.ToInt64E(dID))
		}
	}

	for _, bf := range beforeFus {
		item.Add(userID, bf.ItemId, 1, bf.Props, bf.Id)
	}

	return c.ResponseSuccessToMe(types.Map{})
}

// 卸载符文
func (c *ShinelightController) UnloadRuneAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	heroID := c.Params.GetIntE("hero_id")
	unloadFus := c.Params.GetMapE("fu")

	if len(unloadFus) == 0 {
		return c.ResponseError(64736, "符文是空的")
	}

	userHero := model.GetUserHeroByID(ctx, heroID)
	if userHero == nil || userHero.UserId != userID {
		return c.ResponseError(43241, "hero id error")
	}

	fus := util.ToHeroFus(userHero.Fu, userHero.Lv, userHero.Star)

	beforeFus := make([]*util.Fu, 0)
	for _, side := range []string{"left", "right"} {
		sideVal := unloadFus[side]
		if sideVal == nil || types.IsEmpty(reflect.ValueOf(sideVal)) {
			continue
		}
		if old, ok := fus[side]; ok {
			beforeFus = append(beforeFus, old)
		}
		delete(fus, side)
	}

	lock.Lock(fmt.Sprintf("unloadRune%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("unloadRune%d", userID))

	if err := model.UpdateUserHeroFu(ctx, userID, heroID, fus); err != nil {
		return c.ResponseError(99, "system error")
	}

	for _, bf := range beforeFus {
		item.Add(userID, bf.ItemId, 1, bf.Props, bf.Id)
	}

	return c.ResponseSuccessToMe(types.Map{})
}
