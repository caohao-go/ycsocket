package model

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/connector"
	"server_golang/logic"
	"server_golang/repo"
	"server_golang/repo/mem/daily"
	"server_golang/repo/user"
)

// ---- 新手引导任务系统 ----

// GuideTaskHandle 引导任务处理（与 PHP guildTaskHandle 一致）
// taskID: 引导任务 ID，num: 完成数量
func GuideTaskHandle(ctx context.Context, userID int64, taskID, num int) {
	// 获取任务配置
	taskData := logic.GuideDatas[taskID]
	if taskData == nil {
		return
	}

	// 获取当前引导最大完成ID，如果已超过，则不做处理
	finishedID := GetUserFinishedGuideID(userID)
	if finishedID >= taskID {
		return
	}

	// 获取玩家任务信息
	userTask := GetUserGuideTaskByTaskID(ctx, userID, taskID)
	if userTask != nil && userTask.Status >= 1 {
		// 已经完成或已领取
		return
	}

	needNum := taskData.Num
	currentNum := num
	if userTask != nil {
		currentNum = userTask.Num + num
	}

	status := 0
	if currentNum >= needNum {
		status = 1
		currentNum = needNum
	}

	SetUserGuideTask(ctx, userID, taskID, &GuideTaskStatus{
		TaskID: taskID,
		Status: status,
		Num:    currentNum,
	})
}

// GuideTaskHandleIncr 引导任务处理（累加模式）
// 与 GuideTaskHandle 相同但累加 num
func GuideTaskHandleIncr(ctx context.Context, userID int64, taskID, num int) {
	// 获取任务配置
	taskData := logic.GuideDatas[taskID]
	if taskData == nil {
		return
	}

	// 获取当前引导最大完成ID
	finishedID := GetUserFinishedGuideID(userID)
	if finishedID >= taskID {
		return
	}

	// 获取玩家任务信息
	userTask := GetUserGuideTaskByTaskID(ctx, userID, taskID)
	if userTask != nil && userTask.Status >= 1 {
		return
	}

	needNum := taskData.Num
	currentNum := num
	if userTask != nil {
		currentNum = userTask.Num + num
	}

	status := 0
	if currentNum >= needNum {
		status = 1
		currentNum = needNum
	}

	SetUserGuideTask(ctx, userID, taskID, &GuideTaskStatus{
		TaskID: taskID,
		Status: status,
		Num:    currentNum,
	})
}

// GetCurrentGuideTask 获取当前引导任务（与 PHP Task::getCurrentGuideTask 一致）
func GetCurrentGuideTask(ctx context.Context, userID int64, lv int) types.Map {
	finishedID := GetUserFinishedGuideID(userID)

	// 获取用户所有引导任务状态
	allTasks := GetUserAllGuideTask(ctx, userID)

	// 从 finishedID+1 开始寻找下一个可做的任务
	for id := finishedID + 1; id <= logic.GuideMaxId; id++ {
		taskData := logic.GuideDatas[id]
		if taskData == nil {
			continue
		}

		userTask := allTasks[id]
		status := 0
		num := 0
		if userTask != nil {
			status = userTask.Status
			num = userTask.Num
		}

		if status < 2 {
			return types.Map{
				"id":     id,
				"status": status,
				"num":    num,
				"need":   taskData.Num,
				"reward": taskData.Reward,
				"detail": taskData.Detail,
			}
		}
	}

	return nil
}

// GuideSend 发送引导任务提醒
func GuideSend(ctx context.Context, uid int64, taskID int, num int, status int) {
	ret := types.Map{
		"c":       "guide_remind",
		"uid":     uid,
		"task_id": taskID,
		"num":     num,
		"status":  status,
	}
	connector.Manager.Send(uid, json.Marshal(ret))
}

// UseLibaoCode 使用礼包码
// 对应 PHP: UserinfoModel::useLibaoCode
// 返回: (奖励数据, 错误码) — 错误码: -1 不存在, -2 已使用, -3 奖励不存在, 0 成功
func UseLibaoCode(ctx context.Context, userID int64, code string) (interface{}, int) {
	// 查询礼包码
	data, err := user.GetUserLibaoCode(ctx, code)
	if err != nil || data == nil {
		return nil, -1
	}

	// 检查是否已被使用
	if data.GetIntE("used") == 1 {
		return nil, -2
	}

	// 获取奖励配置
	rewardsID := data.GetIntE("libao_rewards_id")
	libaoRewards, ok := logic.LibaoRewardsData[rewardsID]
	if !ok || len(libaoRewards.Rewards) == 0 {
		return nil, -3
	}

	// 标记已使用
	user.UpdateUserLibaoCode(ctx, code, types.Map{"use_user_id": userID, "used": 1})

	return types.Map{"rewards": libaoRewards.Rewards}, 0
}

type GuideTaskStatus struct {
	TaskID int `json:"task_id"`
	Status int `json:"status"`
	Num    int `json:"num"`
}

func SetUserGuideTask(ctx context.Context, uid int64, taskID int, data *GuideTaskStatus) {
	k := fmt.Sprintf(config.KeyPreUserGuide, uid)
	repo.RedisHSet(ctx, k, taskID, data)
}

// LingquUserGuideReward 领取引导任务奖励
func LingquUserGuideReward(ctx context.Context, uid int64, taskID int) {
	k := fmt.Sprintf(config.KeyPreUserGuide, uid)
	data := GetUserGuideTaskByTaskID(ctx, uid, taskID)
	if data == nil || data.TaskID == 0 {
		return
	}

	data.Status = 2
	repo.RedisHSet(ctx, k, taskID, data)
	SetUserFinishedGuideID(uid, taskID)
}

// GetUserGuideTaskByTaskID 获取用户引导任务状态
func GetUserGuideTaskByTaskID(ctx context.Context, uid int64, taskID int) *GuideTaskStatus {
	k := fmt.Sprintf(config.KeyPreUserGuide, uid)
	v, _ := repo.RedisHGet(ctx, k, taskID)
	if v == "" {
		return nil
	}
	var ret GuideTaskStatus
	json.Unmarshal(v, &ret)
	return &ret
}

func GetUserAllGuideTask(ctx context.Context, uid int64) map[int]*GuideTaskStatus {
	k := fmt.Sprintf(config.KeyPreUserGuide, uid)
	data, _ := repo.RedisHGetAll(ctx, k)
	ret := make(map[int]*GuideTaskStatus, len(data))
	for taskID, v := range data {
		var item = GuideTaskStatus{}
		json.Unmarshal(v, &item)
		ret[types.ToIntE(taskID)] = &item
	}
	return ret
}

// ========================= 新手礼包 =========================

func GetNewGiftTime(ctx context.Context, uid int64, onlineTime int) int {
	k := fmt.Sprintf(config.KeyNewGiftTime, uid)
	v, _ := repo.RedisGet(ctx, k)
	t := types.ToIntE(v)
	if t == 0 {
		t = int(time.Now().Unix()) + onlineTime
		repo.RedisSet(ctx, k, t, 86400)
	}
	if int(time.Now().Unix()) >= t {
		return 0
	}
	return t - int(time.Now().Unix())
}

// ResetNewGiftTime 重置新手礼包时间
func ResetNewGiftTime(ctx context.Context, uid int64) {
	repo.RedisDel(ctx, fmt.Sprintf(config.KeyNewGiftTime, uid))
}

// ========================= 新手7天礼包 =========================

// GetTodayNew7dayGift 获取今天是否已领取7天登录礼包
func GetTodayNew7dayGift(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyNew7dayGift)
	return types.ToIntE(v)
}

// SetTodayNew7dayGift 设置今天已领取7天登录礼包
func SetTodayNew7dayGift(ctx context.Context, uid int64) {
	daily.Set(uid, config.DailyNew7dayGift, "1")
}

// UnsetTodayNew7dayGift 取消今天已领取7天登录礼包
func UnsetTodayNew7dayGift(ctx context.Context, uid int64) {
	daily.Set(uid, config.DailyNew7dayGift, "0")
}
