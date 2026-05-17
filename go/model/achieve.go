package model

import (
	"context"

	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/repo/mem/achieve"
)

// AchieveTaskHandle 成就任务处理
// 对应 PHP: RedisProxy::achieveTaskHandle
// 遍历 startID~endID 范围内的成就任务，跳过已完成/已领取的，
// 对未完成的调用 SetUserAchieveTaskNum 进行判定，遇到未完成则停止（break）
// 完成时发送红点通知
func AchieveTaskHandle(ctx context.Context, uid int64, typ, num, startID, endID int) {
	achieveTasksDatas := GetUserAchieveTasksByType(ctx, uid, typ)
	for taskID := startID; taskID <= endID; taskID++ {
		// 检查已有数据中是否已完成/已领取
		achieveTask, ok := achieveTasksDatas[taskID]
		if ok {
			// 已完成或已领取的跳过
			if achieveTask.Status >= 1 && achieveTask.Status <= 2 {
				continue
			}
		}

		// 获取任务配置
		achieveConfig := logic.AchieveDatas[taskID]
		if achieveConfig == nil {
			continue
		}

		// 设置完成任务
		configNum := achieveConfig.Num
		configExtraNum := achieveConfig.ExtraNum
		newStatus, beforeStatus := SetUserAchieveTaskNum(ctx, uid, taskID, typ, num, 0, configNum, configExtraNum)

		// 完成时发红点通知（对应 PHP: Redpoint::send）
		if beforeStatus == 0 && newStatus == 1 {
			pk := 0
			if typ == 8 || typ == 16 || typ == 22 {
				pk = 1
			}
			RedpointSend(ctx, uid, 1, pk, 1, 1)
		}

		// 未完成停止查询
		if newStatus == 0 {
			break
		}
	}
}

func GetUserAchieveTaskList(ctx context.Context, uid int64) map[int]*achieve.AchieveTask {
	// 组装数据
	datas := make(map[int]*achieve.AchieveTask)

	for typ := 1; typ <= 22; typ++ {
		tmp := GetUserAchieveTasksByType(ctx, uid, typ)
		for tid, item := range tmp {
			datas[tid] = item
		}
	}

	// 每种任务只显示一个
	taskIDDuan := map[int]util.MinMax{
		0:  {1000, 1038},
		3:  {4001, 4005},
		4:  {4101, 4106},
		5:  {4201, 4206},
		6:  {5001, 5002},
		7:  {6001, 6007},
		8:  {7001, 7007},
		9:  {8001, 8006},
		10: {9001, 9007},
		11: {9101, 9107},
		12: {10001, 10006},
		13: {11001, 11004},
		14: {11101, 11107},
		15: {12001, 12007},
		16: {14001, 14011},
		17: {15001, 15001},
		18: {15002, 15002},
		19: {16001, 16004},
	}

	data := make(map[int]*achieve.AchieveTask)
	for k, rng := range taskIDDuan {
		found := false
		for i := int(rng.Min); i <= int(rng.Max); i++ {
			if task, ok := datas[i]; ok {
				if task.Status == 2 {
					data[k] = task
					continue
				}
				if task.Status == 1 || task.Status == 0 {
					data[k] = task
					found = true
					break
				}
			} else {
				// 与 PHP 行为一致：当 task_id 不存在于 datas 中时，
				// PHP 的 null == 0 为 true，会将其视为未完成任务并 break
				// Go 中需要显式构造默认数据
				data[k] = &achieve.AchieveTask{
					UserId:   uid,
					Status:   0,
					TaskId:   i,
					ExtraNum: 0,
					Num:      0,
				}
				found = true
				break
			}
		}
		if !found {
			if _, ok := data[k]; !ok {
				data[k] = &achieve.AchieveTask{
					UserId:   uid,
					Status:   0,
					TaskId:   int(rng.Min),
					ExtraNum: 0,
					Num:      0,
				}
			}
		}
	}

	return data
}

// SetUserAchieveTaskNum 设置成就任务完成数量，返回任务状态 0-未完成 1-已完成 2-已领取
// 对应 PHP: RedisProxy::setUserAchieveTaskNum
// typ - 成就类型, num - 当前完成数量, extraNum - 额外数据(默认传0)
// configNum - 配置中的目标阈值, configExtraNum - 配置中的额外数据阈值
// 返回: (newStatus, beforeStatus) - 新状态和之前的状态
func SetUserAchieveTaskNum(ctx context.Context, uid int64, taskID, typ, num, extraNum, configNum, configExtraNum int) (int, int) {
	data := GetUserAchieveTask(ctx, uid, typ, taskID)

	if data == nil {
		data = &achieve.AchieveTask{}
		data.TaskId = taskID
		data.UserId = uid
		data.Type = typ
		data.Status = 0
	}

	if data.Status == 2 { // 已经领取
		return 2, 2
	}

	beforeStatus := data.Status

	data.Num = num
	data.ExtraNum = extraNum

	// 判断是否完成，对应 PHP setUserAchieveTaskNum 核心逻辑
	if configExtraNum == 0 {
		if typ == 22 { // 竞技场排名，从小到大（排名越小越好）
			if num <= configNum && data.Status != 1 {
				data.Status = 1
			} else {
				data.Status = 0
			}
		} else {
			if num >= configNum && data.Status != 1 {
				data.Status = 1
			} else {
				data.Status = 0
			}
		}
	} else if configExtraNum > 0 {
		if num >= configNum && data.Status != 1 && extraNum >= configExtraNum {
			data.Status = 1
		} else {
			data.Status = 0
		}
	}

	SetUserAchieveTask(ctx, uid, typ, taskID, data)
	return data.Status, beforeStatus
}

// pika

func SetUserAchieveTask(ctx context.Context, uid int64, typ, taskID int, data *achieve.AchieveTask) {
	task := *data
	task.Type = typ
	task.TaskId = taskID
	achieve.SetTask(uid, typ, taskID, task)
}

func GetUserAchieveTask(ctx context.Context, uid int64, typ, taskID int) *achieve.AchieveTask {
	task, ok := achieve.GetTask(uid, typ, taskID)
	if !ok {
		return &achieve.AchieveTask{
			UserId:   uid,
			TaskId:   taskID,
			Type:     typ,
			Num:      0,
			ExtraNum: 0,
			Status:   0,
		}
	}
	return &task
}

func LingquUserAchieveReward(ctx context.Context, uid int64, taskID, typ int) {
	task, ok := achieve.GetTask(uid, typ, taskID)
	if !ok {
		return
	}
	task.Status = 2
	achieve.SetTask(uid, typ, taskID, task)
}

func GetUserAchieveTasksByType(ctx context.Context, uid int64, typ int) map[int]*achieve.AchieveTask {
	tasks := achieve.GetTasksByType(uid, typ)
	ret := make(map[int]*achieve.AchieveTask, len(tasks))
	for taskID, task := range tasks {
		t := task // 复制避免指针共享
		ret[taskID] = &t
	}
	return ret
}
