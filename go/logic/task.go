package logic

import (
	"context"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/repo/info"
)

// 签到奖励全局数据 day => TypeNum
var DailyDatas map[int][]util.TypeNum

// InitDaily 初始化每日签到数据
func InitDaily(ctx context.Context) {
	DailyDatas = make(map[int][]util.TypeNum)

	rows, err := info.GetAllCheckinReward(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init checkin_reward error: %v", err)
		return
	}
	for _, val := range rows {
		DailyDatas[val.CheckinDays] = util.ToTypeNums(val.Reward)
	}
}

// ======================== 引导任务配置 ========================

// GuideData 引导任务配置项（对应 PHP Task::$guide_datas[id]）
type GuideData struct {
	ID     int
	Type   int
	Num    int            // 完成任务需要达到的数量
	Reward []util.TypeNum // 奖励列表 [{type: xx, num: xx}, ...]
	Detail string
}

// GuideDatas 引导任务配置（对应 PHP Task::$guide_datas）
// key 为 task_guide 表的 id
var GuideDatas map[int]*GuideData

// GuideMaxId 引导任务最大ID（对应 PHP Task::$guide_max_id）
var GuideMaxId int

// InitTaskGuide 初始化引导任务配置（从 task_guide 表加载）
func InitTaskGuide(ctx context.Context) {
	GuideDatas = make(map[int]*GuideData)
	GuideMaxId = 0

	rows, err := info.GetAllTaskGuide(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "InitTaskGuide: GetAllTaskGuide error: %v", err)
		return
	}

	for _, row := range rows {
		GuideDatas[row.Id] = &GuideData{
			ID:     row.Id,
			Type:   row.Type,
			Num:    row.Num,
			Reward: util.ToTypeNums(string(row.Reward)),
			Detail: string(row.Detail),
		}
		if row.Id > GuideMaxId {
			GuideMaxId = row.Id
		}
	}

	log.Infof(ctx, "InitTaskGuide: loaded %d task_guide records, max_id=%d", len(rows), GuideMaxId)
}

// ======================== 周任务配置 ========================

// TaskWeeklyDatas 周任务配置（对应 PHP Task::$task_weekly_datas）
// 结构: TaskWeeklyDatas[name][type][number] = []types.Map (rewards)
var TaskWeeklyDatas map[string]map[int]map[int][]util.TypeNum

// InitTaskWeekly 初始化周任务配置（从 task_weekly_config 表加载）
func InitTaskWeekly(ctx context.Context) {
	TaskWeeklyDatas = make(map[string]map[int]map[int][]util.TypeNum)

	rows, err := info.GetAllTaskWeeklyConfig(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "InitTaskWeekly: GetAllTaskWeeklyConfig error: %v", err)
		return
	}

	for _, row := range rows {
		name := row.Name
		typ := row.Type
		number := row.Number

		if TaskWeeklyDatas[name] == nil {
			TaskWeeklyDatas[name] = make(map[int]map[int][]util.TypeNum)
		}
		if TaskWeeklyDatas[name][typ] == nil {
			TaskWeeklyDatas[name][typ] = make(map[int][]util.TypeNum)
		}
		TaskWeeklyDatas[name][typ][number] = util.ToTypeNums(string(row.Reward))
	}

	log.Infof(ctx, "InitTaskWeekly: loaded %d task_weekly_config records", len(rows))
}

// ======================== 日常/成就任务配置 ========================

// TaskConfigData 对应 PHP Task::$datas[$type][$id]
type TaskConfigData struct {
	ID          int
	Type        int
	TaskType    int
	TaskCount   util.TypeNum   // [type, num]
	TaskReward  []util.TypeNum // [{type: xx, num: xx}, ...]
	Active      int
	OnCondition []util.TypeNum // 开启条件
}

// TaskDatas 任务配置（对应 PHP Task::$datas）
// 结构: TaskDatas[type][id] = TaskConfigData
var TaskDatas map[int]map[int]*TaskConfigData

// TaskActiveDatas 活跃度奖励配置（对应 PHP Task::$active_datas）
// 结构: TaskActiveDatas[type][active] = {task_reward}
var TaskActiveDatas map[int]map[int][]util.TypeNum

// InitTaskConfig 初始化任务配置（从 task_config 和 task_active_config 表加载）
func InitTaskConfig(ctx context.Context) {
	// 1. 加载 task_config
	TaskDatas = make(map[int]map[int]*TaskConfigData)
	rows, err := info.GetAllTaskConfig(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "InitTaskConfig: GetAllTaskConfig error: %v", err)
	} else {
		for _, row := range rows {
			id := types.ToIntE(row.Id)
			typ := row.Type

			taskCounts := util.ToTypeNums(string(row.TaskCount))

			if TaskDatas[typ] == nil {
				TaskDatas[typ] = make(map[int]*TaskConfigData)
			}
			TaskDatas[typ][id] = &TaskConfigData{
				ID:          id,
				Type:        typ,
				TaskType:    row.TaskType,
				TaskCount:   taskCounts[0],
				TaskReward:  util.ToTypeNums(string(row.TaskReward)),
				Active:      row.Active,
				OnCondition: util.ToTypeNums(string(row.OnCondition)),
			}
		}
	}

	// 2. 加载 task_active_config
	TaskActiveDatas = make(map[int]map[int][]util.TypeNum)
	activeRows, err := info.GetAllTaskActiveConfig(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "InitTaskConfig: GetAllTaskActiveConfig error: %v", err)
	} else {
		for _, row := range activeRows {
			typ := row.Type
			active := row.Active

			if TaskActiveDatas[typ] == nil {
				TaskActiveDatas[typ] = make(map[int][]util.TypeNum)
			}
			TaskActiveDatas[typ][active] = util.ToTypeNums(string(row.TaskReward))
		}
	}

	log.Infof(ctx, "InitTaskConfig: loaded %d task_config records, %d task_active_config records",
		len(rows), len(activeRows))
}

// GetActiveRewardByNum 获取活跃度奖励（与 PHP Task::$active_datas 一致，从数据库加载）
func GetActiveRewardByNum(num int) []util.TypeNum {
	if TaskActiveDatas[1] != nil {
		if r, ok := TaskActiveDatas[1][num]; ok {
			return r
		}
	}
	return nil
}
