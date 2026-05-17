package controller

import (
	"context"

	"server_golang/common/types"
	"server_golang/logic"
	"server_golang/model"
)

// 成就任务

func (c *ShinelightController) AchieveTaskAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	if userGrade.GetIntE("lv") < 6 {
		return c.ResponseError(666666, "等级不够")
	}

	data := model.GetUserAchieveTaskList(ctx, userID)
	// 过滤已领取的任务（status == 2）
	for k, v := range data {
		if v.Status == 2 {
			delete(data, k)
		}
	}

	return c.ResponseSuccessToMe(types.Map{"task": data})
}

func (c *ShinelightController) AchieveTaskLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	taskID := c.Params.GetIntE("task_id")

	// 获取任务配置信息（对应 PHP: Task::$achieve_datas[$task_id]）
	taskConfig, ok := logic.AchieveDatas[taskID]
	if !ok || taskConfig == nil {
		return c.ResponseError(666667, "任务id错误")
	}

	result := types.Map{"rewards": taskConfig.Reward}

	// 获取玩家的任务信息（对应 PHP: RedisProxy::get_user_achieve_task($userId, $task_id)）
	userTaskData := model.GetUserAchieveTask(ctx, userID, taskConfig.Type, taskID)
	status := userTaskData.Status

	if status == 2 {
		return c.ResponseError(666668, "任务已经领取")
	} else if status == 0 {
		return c.ResponseError(666668, "任务未完成")
	} else if status == 1 {
		// 给予奖励（对应 PHP: $this->shinelight_model->give_rewards($userId, $task_datas['reward'])）
		model.GiveReward(userID, taskConfig.Reward...)
		// 设置领取状态（对应 PHP: RedisProxy::lingqu_user_achieve_reward($userId, $task_id)）
		model.LingquUserAchieveReward(ctx, userID, taskID, taskConfig.Type)
	}

	// 引导任务（对应 PHP: if ($task_id == 1000 || $task_id == 1001 || $task_id == 1002)）
	if taskID == 1000 || taskID == 1001 || taskID == 1002 {
		model.GuideTaskHandle(ctx, userID, 47, 1)
	}

	return c.ResponseSuccessToMe(result)
}
