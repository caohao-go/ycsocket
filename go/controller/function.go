package controller

import (
	"context"
	"fmt"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// 日常副本

func (c *ShinelightController) AllFunctionAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")

	data := make([]types.Map, 0)
	funcs := []struct {
		name    string
		needLv  int
		openNum int // open_type.num（与 PHP allFunctionAction 一致）
	}{
		{"日常副本", 15, 18}, {"无尽试炼", 25, 25}, {"英雄远征", 30, 30}, {"星河神殿", 35, 35},
	}
	for _, f := range funcs {
		if lv < f.needLv {
			data = append(data, types.Map{
				"name": f.name, "status": 0,
				"open_type": []types.Map{{"type": 1, "num": f.openNum}},
			})
		} else {
			data = append(data, types.Map{"name": f.name, "status": 1})
		}
	}
	return c.ResponseSuccessToMe(types.Map{"list": data})
}

// 获取日常副本详情
func (c *ShinelightController) GetFunctionAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")
	if id == 0 {
		return c.ResponseError(32116, "id is empty")
	}

	userGrade := model.GetUserAttr(userID)
	if userGrade == nil {
		return c.ResponseError(32113, "user_grade is empty")
	}
	userGrade["fight_point"] = model.GetUserFightPoint(ctx, userID, 1)

	functionFights := model.GetUsersContent(ctx, userID, "function_fights")
	data := model.GetFunctionByID(ctx, userID, id, userGrade, functionFights, -1)
	data.VipLevel = userGrade.GetIntE("vip_level")
	return c.ResponseSuccessToMe(types.ObjectToMap(data))
}

// 完成日常副本
func (c *ShinelightController) FinishFunctionAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	copyID := c.Params.GetIntE("copy_id")
	if copyID == 0 {
		return c.ResponseError(32181, "copy_id is empty")
	}

	userGrade := model.GetUserAttr(userID)
	if userGrade == nil {
		return c.ResponseError(3211, "user_grade is empty")
	}

	lock.Lock(fmt.Sprintf("finishFunction%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("finishFunction%d", userID))

	functionFights := model.GetUsersContent(ctx, userID, "function_fights")
	functionConfig := model.GetFunctionByCopyID(ctx, userID, copyID, userGrade, functionFights)
	if functionConfig == nil {
		return c.ResponseError(32511, "function_config is empty")
	}

	status := functionConfig.Status
	if status == 0 {
		return c.ResponseError(32281, "未满足开启条件")
	}
	if status == 2 {
		return c.ResponseError(32281, "扫荡次数不足")
	}

	basicCost := functionConfig.BasicCost
	costType := functionConfig.CostType
	if basicCost > 0 {
		if item.NotEnough(userID, costType, basicCost) {
			return c.ResponseError(666666, "货币不够")
		}
	}

	ret := types.Map{"tiaozhan": 0}

	// 扫荡或挑战模式（与 PHP finishFunctionAction 一致）
	if status != 3 { // 非挑战模式 = 扫荡
		// 扫荡直接扣次数/扣费/发奖
		model.IncrRedisUserFunctionTimes(ctx, userID, logic.CopyIDsID[copyID])
		item.Sub(userID, costType, basicCost)
		model.GiveReward(userID, functionConfig.Reward...)
	} else {
		ret["tiaozhan"] = 1

		// 获取用户剧情阵型英雄
		myHeroAttrs := model.GetUserPositionWithHeroAttrs(ctx, userID, 1)
		if myHeroAttrs == nil || len(myHeroAttrs) == 0 {
			return c.ResponseError(3241, "请先设置剧情英雄")
		}

		// 获取对手英雄（根据副本layer配置，怪物英雄从expedition_hero表）
		layer := functionConfig.Layer
		opHeros := logic.GetExpeditionHeros(layer)
		if opHeros == nil || len(opHeros) == 0 {
			return c.ResponseError(32512, fmt.Sprintf("副本对手数据缺失(layer=%d)", layer))
		}

		// 对手是怪物，用hero信息直接从配置表计算属性（不查user_hero表）
		opHeroDetail := model.GetFightHeroAttrWithSkill(ctx, opHeros)

		// 执行战斗模拟（使用完整战斗引擎，与PK/试练塔等一致）
		fight := logic.NewFight(myHeroAttrs, opHeroDetail)
		winner, fightResult := fight.FightExec(15)

		if winner == "P1" {
			ret["success"] = 1
		} else {
			ret["success"] = 0
		}

		// 构建战斗展示数据
		ret["my_hero"] = logic.GetBaseFromHero(myHeroAttrs)
		ret["opp_hero"] = logic.GetBaseFromHero(opHeroDetail)
		ret["fight_result"] = fightResult

		// 挑战成功才更新 function_fights、扣次数、扣费、发奖（与 PHP 一致：if ($ret['tiaozhan']==0 || $ret['success'])）
		if winner == "P1" {
			functionFights[types.ToString(copyID)] = 1
			model.UpdateUsersContent(ctx, userID, "function_fights", functionFights)

			model.IncrRedisUserFunctionTimes(ctx, userID, logic.CopyIDsID[copyID])
			item.Sub(userID, costType, basicCost)
			model.GiveReward(userID, functionConfig.Reward...)
		}
	}

	// 日常任务
	model.SetDailyTaskFinish(ctx, userID, 10011, 1)

	// 成就任务
	model.IncrFinishFunctionNum(userID)
	finishNum := model.GetFinishFunctionNum(userID)
	model.AchieveTaskHandle(ctx, userID, 10, finishNum, 7001, 7007)

	// 引导任务
	if copyID >= 20011 && copyID <= 20019 {
		model.GuideTaskHandle(ctx, userID, 75, 1)
	} else if copyID >= 20021 && copyID <= 20029 {
		model.GuideTaskHandle(ctx, userID, 77, 1)
	} else if copyID >= 20031 && copyID <= 20039 {
		model.GuideTaskHandle(ctx, userID, 79, 1)
	}

	return c.ResponseSuccessToMe(ret)
}
