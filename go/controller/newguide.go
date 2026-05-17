package controller

import (
	"context"
	"fmt"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// 礼包码/十连抽/引导

// UseLibaoCodeAction 领取礼包码
// 对应 PHP: Shinelight::useLibaoCodeAction
func (c *ShinelightController) UseLibaoCodeAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	code := c.Params.GetStringE("code")
	if code == "" {
		return c.ResponseError(4236, "礼包码为空")
	}

	libaoRewards, errCode := model.UseLibaoCode(ctx, userID, code)
	if errCode == -1 {
		return c.ResponseError(42236, "礼包码不存在")
	}
	if errCode == -2 {
		return c.ResponseError(42236, "礼包码已被使用")
	}
	if errCode == -3 {
		return c.ResponseError(42236, "奖励不存在")
	}

	// 发放奖励
	if rewardsMap, ok := libaoRewards.(types.Map); ok {
		if rewardsList, ok := rewardsMap["rewards"]; ok {
			model.GiveReward(userID, rewardsList.([]util.TypeNum)...)
		}
		return c.ResponseSuccessToMe(rewardsMap)
	}

	return c.ResponseSuccessToMe(types.Map{})
}

func (c *ShinelightController) Lingqu10chouAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	lock.Lock(fmt.Sprintf("lingqu10chou%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("lingqu10chou%d", userID))

	if model.GetLingqu10Chou(userID) > 0 {
		return c.ResponseError(31234, "已经领过")
	}

	item.Add(userID, 20901, 10, nil)

	model.SetLingqu10Chou(userID)
	return c.ResponseSuccessToMe(types.Map{})
}

func (c *ShinelightController) Begin10chouAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	lock.Lock(fmt.Sprintf("begin10chou%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("begin10chou%d", userID))

	// 检查是否已经抽过
	if model.GetBegin10Chou(userID) > 0 {
		return c.ResponseError(31234, "已经抽过")
	}

	// 扣除10张基础召唤券
	if item.NotEnough(userID, 20901, 10) {
		return c.ResponseError(666666, "召唤券不够")
	}

	// 开局10连抽固定英雄列表 [[hero_id, star], ...]
	begin10Heros := [][]int{
		{1201, 5}, {1205, 4}, {2202, 4}, {3401, 3}, {4405, 3},
		{2506, 3}, {1213, 3}, {4504, 3}, {2103, 3}, {2104, 3},
	}

	for _, h := range begin10Heros {
		model.InsertNewUserHero(ctx, userID, h[0], h[1])
	}

	model.SetBegin10Chou(userID)
	item.Sub(userID, 20901, 10)

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 1, 1)

	return c.ResponseSuccessToMe(types.Map{"heros": begin10Heros})
}

// 新人礼包

// GetNewGiftAction 获取新手礼包信息
func (c *ShinelightController) GetNewGiftAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	num := userGrade.GetIntE("new_gift")

	giftTime := 0
	if num < 12 {
		onlineTime := logic.CheckpointOnlineTime[3][num]
		giftTime = model.GetNewGiftTime(ctx, userID, onlineTime)
	}

	// 获取奖励配置
	rewards := logic.CheckpointRewardDatas[3][num]

	// 7天登录信息
	day7 := userGrade.GetIntE("day7")
	todayLingqu := model.GetTodayNew7dayGift(ctx, userID)

	return c.ResponseSuccessToMe(types.Map{
		"num":            num,
		"time":           giftTime,
		"rewards":        rewards,
		"new_7day_login": day7,
		"today_lingqu":   todayLingqu,
	})
}

// LingquNew7dayLoginAction 领取7天登录礼包
func (c *ShinelightController) LingquNew7dayLoginAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	day7 := userGrade.GetIntE("day7")
	if day7 > 6 {
		return c.ResponseError(8653, "已经领完")
	}

	todayLingqu := model.GetTodayNew7dayGift(ctx, userID)
	if todayLingqu != 0 {
		return c.ResponseError(6653, "今天已经领取")
	}

	model.SetTodayNew7dayGift(ctx, userID)

	// 获取奖励配置 checkpoint_type=4
	rewards := logic.CheckpointRewardDatas[4][day7]
	model.GiveReward(userID, rewards...)

	ret := model.IncrUserDay7(userID)
	if ret < 0 {
		model.UnsetTodayNew7dayGift(ctx, userID)
		return c.ResponseError(99, "system error")
	}

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 31, 1)

	return c.ResponseSuccessToMe(types.Map{})
}

// LingquNewGiftAction 领取新手礼包
func (c *ShinelightController) LingquNewGiftAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	num := userGrade.GetIntE("new_gift")

	onlineTime := logic.CheckpointOnlineTime[3][num]
	giftTime := model.GetNewGiftTime(ctx, userID, onlineTime)
	if giftTime > 0 {
		return c.ResponseError(4893, "领取时间未到")
	}

	model.IncrUserNewGift(userID)
	model.ResetNewGiftTime(ctx, userID)

	// 获取奖励配置 checkpoint_type=3
	rewards := logic.CheckpointRewardDatas[3][num]

	model.GiveReward(userID, rewards...)

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 14, 1)
	model.GuideTaskHandle(ctx, userID, 15, 1)

	return c.ResponseSuccessToMe(types.Map{})
}

func (c *ShinelightController) GetCurrentGuideAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")
	data := model.GetCurrentGuideTask(ctx, userID, lv)
	return c.ResponseSuccessToMe(data)
}

func (c *ShinelightController) GuideTaskLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	taskID := c.Params.GetIntE("task_id")

	// 获取任务配置信息
	taskData := logic.GuideDatas[taskID]
	if taskData == nil {
		return c.ResponseError(666667, "任务id错误")
	}

	result := types.Map{"rewards": taskData.Reward}

	// 获取玩家的任务信息
	userTaskData := model.GetUserGuideTaskByTaskID(ctx, userID, taskID)

	userStatus := 0
	userNum := 0
	if userTaskData != nil {
		userStatus = userTaskData.Status
		userNum = userTaskData.Num
	}

	if userStatus == 2 {
		return c.ResponseError(666668, "任务已经领取")
	} else if userStatus == 1 || (userStatus == 0 && userNum >= taskData.Num) {
		// 给予奖励
		model.GiveReward(userID, taskData.Reward...)

		// 设置状态
		model.LingquUserGuideReward(ctx, userID, taskID)
	} else if userStatus == 0 {
		return c.ResponseError(666668, "任务未完成")
	}

	return c.ResponseSuccessToMe(result)
}
