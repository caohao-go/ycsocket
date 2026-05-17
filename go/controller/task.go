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
)

// 日常任务

func (c *ShinelightController) TaskAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	typ := c.Params.GetIntE("type")

	taskList := model.GetDailyTask(ctx, userID, typ)

	result := types.Map{"task": taskList}

	if typ == 1 {
		active := model.GetDailyTaskActiveNum(ctx, userID)
		// 优化：使用 GetDailyTaskActiveAll 一次 HGETALL 替代 4 次 HGET（与 PHP getDailyTaskActiveAll 一致）
		activeAllData := model.GetDailyTaskActiveAll(ctx, userID)
		cnts := []int{25, 50, 75, 100}
		activeRet := make([]types.Map, 0, len(cnts))
		for _, cnt := range cnts {
			lingqu := activeAllData.GetIntE(cnt)
			activeRet = append(activeRet, types.Map{"num": cnt, "lingqu": lingqu})
		}
		result["active"] = active
		result["active_lingqu"] = activeRet
	}

	return c.ResponseSuccessToMe(result)
}

func (c *ShinelightController) TaskLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	typ := c.Params.GetIntE("type")
	id := c.Params.GetIntE("id")

	lock.Lock(fmt.Sprintf("taskLingqu%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("taskLingqu%d", userID))

	taskDetail := model.GetDailyTaskByID(ctx, userID, typ, id)
	if taskDetail == nil {
		return c.ResponseError(9731, "任务不存在")
	}

	finishCount := taskDetail.GetIntE("finish_count")
	maxCount := taskDetail.GetIntE("max_count")
	if finishCount < maxCount {
		return c.ResponseError(9731, "任务未完成")
	}

	lingqu := taskDetail.GetIntE("lingqu")
	if lingqu > 0 {
		return c.ResponseError(9331, "已领取")
	}

	rewards, _ := taskDetail["task_reward"].([]util.TypeNum)
	if len(rewards) > 0 {
		model.GiveReward(userID, rewards...)
	}

	if typ == 1 {
		model.SetDailyTaskLingqu(ctx, userID, typ, id)
		active := taskDetail.GetIntE("active")
		model.AddDailyTaskActiveNum(ctx, userID, active)
	} else if typ == 2 {
		// 持久任务领取（对应 PHP $this->shinelight_model->setTaskLingqu($userId, $id)）
		model.SetContentTaskLingqu(ctx, userID, id)
	}

	// 引导任务（对应 PHP taskLingquAction 第 7026-7038 行）
	if typ == 1 {
		switch id {
		case 10001:
			model.GuideTaskHandle(ctx, userID, 24, 1)
		case 10003:
			model.GuideTaskHandle(ctx, userID, 26, 1)
		case 10004:
			model.GuideTaskHandle(ctx, userID, 52, 1)
		case 10010:
			model.GuideTaskHandle(ctx, userID, 66, 1)
		case 10006:
			model.GuideTaskHandle(ctx, userID, 87, 1)
		}
	}

	return c.ResponseSuccessToMe(types.Map{})
}

func (c *ShinelightController) TaskActiveLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	num := c.Params.GetIntE("num")

	if model.GetDailyTaskActiveLingquByID(ctx, userID, num) > 0 {
		return c.ResponseError(9331, "已领取")
	}

	rewards := logic.GetActiveRewardByNum(num)
	if rewards == nil {
		return c.ResponseError(9975, "找不到 active datas")
	}

	model.GiveReward(userID, rewards...)
	model.SetDailyTaskActiveLingqu(ctx, userID, num)

	return c.ResponseSuccessToMe(types.Map{})
}

// 周任务

// TaskInfoAction 周任务信息（对应 PHP taskInfoAction 第 8222-8232 行）
func (c *ShinelightController) TaskInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	task := c.Params.GetStringE("task")

	ret := model.GetWeekTask(ctx, userID, task)
	ret["left_time"] = util.LeftTimeToNextMonday()

	// 排行（对应 PHP $this->shinelight_model->getMyRank($task."_score_".date('w'), $userId)）
	w := int(time.Now().Weekday())
	rankKey := fmt.Sprintf("%s_score_%d", task, w)
	ret["my_rank"] = model.GetMyRank(ctx, rankKey, userID)

	return c.ResponseSuccessToMe(types.Map{"data": ret})
}

// WeekTaskLingquAction 周任务领取（对应 PHP weekTaskLingquAction 第 8235-8254 行）
func (c *ShinelightController) WeekTaskLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	task := c.Params.GetStringE("task")
	number := c.Params.GetIntE("number")

	taskDetail := model.GetWeekTaskDetail(ctx, userID, task, number)
	status := taskDetail.GetIntE("status")

	if status == 0 {
		return c.ResponseError(6492, "任务未完成")
	} else if status == 2 {
		return c.ResponseError(6242, "已经领取")
	}

	// 标记领取（对应 PHP RedisProxy::setTaskLingqu($userId, $task, $number, 7)）
	model.SetTaskLingqu(ctx, userID, task, number, 7)

	// 发放奖励

	rewards, _ := taskDetail["rewards"].([]util.TypeNum)
	if len(rewards) > 0 {
		model.GiveReward(userID, rewards...)
	}

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

func (c *ShinelightController) TaskScoreRankAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	typ := c.Params.GetStringE("type")
	weekday := time.Now().Weekday()
	rankKey := fmt.Sprintf("%s_score_%d", typ, weekday)

	data := types.Map{
		"my_rank":   model.GetMyRank(ctx, rankKey, userID),
		"my_score":  model.GetMyRankScore(ctx, rankKey, userID),
		"rank_list": model.GetRankList(ctx, rankKey, true, 0, 99),
	}
	return c.ResponseSuccessToMe(data)
}

func (c *ShinelightController) LingquIGiftAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 检查是否已领取（与 PHP lingquIGiftAction 一致：$user_grade['i_gift'] != 0 返回错误码 64932）
	userGrade := model.GetUserAttr(userID)

	if userGrade == nil {
		return c.ResponseError(32113, "user_grade is empty")
	}
	if userGrade.GetIntE("i_gift") != 0 {
		return c.ResponseError(64932, "已经领取")
	}

	// 标记已领取（与 PHP 一致：updateUsersGrade($userId, ['i_gift' => 1])）
	model.SetIGift(userID)

	// 从配置读取奖励（与 PHP 一致：Checkpointreward::$datas[5][0]['rewards']）
	var rewards []util.TypeNum
	if cpData, ok := logic.CheckpointRewardDatas[5]; ok {
		if rewardData, ok := cpData[0]; ok {
			rewards = rewardData
		}
	}
	// 兜底：如果配置读取失败，使用默认值
	if len(rewards) == 0 {
		rewards = []util.TypeNum{{Type: 2, Num: 100}}
	}

	model.GiveReward(userID, rewards...)
	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}
