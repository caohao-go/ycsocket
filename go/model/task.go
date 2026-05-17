package model

import (
	"context"
	"fmt"
	"sort"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo"
	"server_golang/repo/mem/content"
	"server_golang/repo/mem/daily"
)

// ---- 日常任务 ----

// SetDailyTaskFinish 设置日常任务完成（与 PHP setDalilyTaskFinish 完全一致）
func SetDailyTaskFinish(ctx context.Context, userID int64, taskID, taskType int) error {
	// 获取任务配置
	if logic.TaskDatas[taskType] == nil || logic.TaskDatas[taskType][taskID] == nil {
		return nil
	}
	val := logic.TaskDatas[taskType][taskID]
	needCount := val.TaskCount.Num

	// 获取当前完成次数
	finishCount := GetDailyTaskCountByID(ctx, userID, taskType, taskID)

	// 如果未完成，增加一次
	if finishCount < needCount {
		IncrDailyTaskCountByID(ctx, userID, taskType, taskID, 1)
	}

	return nil
}

// SetContentTaskLingqu 设置任务已领取
func SetContentTaskLingqu(ctx context.Context, userID int64, id int) error {
	lingqu := content.GetMap(userID, "task_lingqu")
	lingqu[types.ToString(id)] = 1
	content.SetMap(userID, "task_lingqu", lingqu)
	return nil
}

// GetDailyTask 获取日常任务列表（与 PHP Task::getDailyTask 完全一致）
func GetDailyTask(ctx context.Context, userID int64, typ int) []types.Map {
	var countData = types.Map{}
	var lingquData = types.Map{}

	if typ == 1 {
		// type=1：从 Pika 每日 key 读取
		countData = daily.GetAllByPrefix(userID, config.DailyTaskCount+types.ToString(typ))
		lingquData = daily.GetAllByPrefix(userID, config.DailyTaskLingqu+types.ToString(typ))
	} else if typ == 2 || typ == 3 || typ == 4 || typ == 5 {
		// type=2-5：从 MySQL user_contents 读取（与 PHP 一致）
		countData = content.GetMap(userID, "task_finish_count")
		lingquData = content.GetMap(userID, "task_lingqu")
	}

	tasks := make([]types.Map, 0)
	if logic.TaskDatas[typ] == nil {
		return tasks
	}

	for _, val := range logic.TaskDatas[typ] {
		finishCount := countData.GetIntE(val.ID)
		lingqu := lingquData.GetIntE(val.ID)
		maxCount := val.TaskCount.Num

		// 计算状态（与 PHP 一致）
		var st int
		if finishCount < maxCount {
			st = 1 // 未完成
		} else if lingqu > 0 {
			st = 2 // 已领取
		} else {
			st = 0 // 可以领取
		}

		tasks = append(tasks, types.Map{
			"id":           val.ID,
			"max_count":    maxCount,
			"finish_count": finishCount,
			"lingqu":       lingqu,
			"active":       val.Active,
			"st":           st,
			"task_reward":  val.TaskReward,
		})
	}

	// 排序：可领取(st=0)排前面，未完成(st=1)排中间，已领取(st=2)排后面；同类别内按ID升序
	sort.Slice(tasks, func(i, j int) bool {
		si := tasks[i].GetIntE("st")
		sj := tasks[j].GetIntE("st")
		if si != sj {
			return si < sj
		}
		return tasks[i].GetIntE("id") < tasks[j].GetIntE("id")
	})

	return tasks
}

// GetDailyTaskByID 获取单个日常任务详情（与 PHP Task::getDailyTaskById 一致）
func GetDailyTaskByID(ctx context.Context, userID int64, typ, id int) types.Map {
	// 获取配置
	if logic.TaskDatas[typ] == nil || logic.TaskDatas[typ][id] == nil {
		return nil
	}
	val := logic.TaskDatas[typ][id]

	var finishCount, lingqu int
	if typ == 1 {
		v, _ := daily.GetByPrefix(userID, config.DailyTaskCount+types.ToString(typ), id)
		finishCount = types.ToIntE(v)
		v2, _ := daily.GetByPrefix(userID, config.DailyTaskLingqu+types.ToString(typ), id)
		lingqu = types.ToIntE(v2)
	} else if typ == 2 || typ == 3 || typ == 4 || typ == 5 {
		finishCountMap := content.GetMap(userID, "task_finish_count")
		lingquMap := content.GetMap(userID, "task_lingqu")
		finishCount = finishCountMap.GetIntE(id)
		lingqu = lingquMap.GetIntE(id)
	}

	return types.Map{
		"id":           val.ID,
		"max_count":    val.TaskCount.Num,
		"need_count":   val.TaskCount.Num,
		"finish_count": finishCount,
		"lingqu":       lingqu,
		"active":       val.Active,
		"task_reward":  val.TaskReward,
	}
}

// GetWeekTask 获取周任务列表（对应 PHP Task::getWeekTask）
func GetWeekTask(ctx context.Context, userID int64, name string) types.Map {
	ret := types.Map{}
	finishCount := GetTaskFinishNumStr(ctx, userID, name, 7)
	taskLingqu := GetTaskLingqu(ctx, userID, name, 7)
	ret["finish_count"] = finishCount

	statusList := make([]types.Map, 0)
	if logic.TaskWeeklyDatas[name] != nil && logic.TaskWeeklyDatas[name][1] != nil {
		for k, v := range logic.TaskWeeklyDatas[name][1] {
			// 0-未完成 1-未领取 2-已领取
			var status int
			if finishCount < k {
				status = 0
			} else if taskLingqu.GetIntE(k) > 0 {
				status = 2
			} else {
				status = 1
			}
			statusList = append(statusList, types.Map{
				"num": k, "status": status, "rewards": v,
			})
		}
	}
	ret["status"] = statusList
	return ret
}

// GetWeekTaskDetail 获取周任务详情（对应 PHP Task::getWeekTaskDetail）
func GetWeekTaskDetail(ctx context.Context, userID int64, name string, number int) types.Map {
	finishCount := GetTaskFinishNumStr(ctx, userID, name, 7)
	taskLingqu := GetTaskLingquDetail(ctx, userID, name, number, 7)

	// 0-未完成 1-未领取 2-已领取
	var status int
	if finishCount < number {
		status = 0
	} else if taskLingqu == "" {
		status = 1
	} else {
		status = 2
	}

	ret := types.Map{"status": status}
	if logic.TaskWeeklyDatas[name] != nil && logic.TaskWeeklyDatas[name][1] != nil {
		ret["rewards"] = logic.TaskWeeklyDatas[name][1][number]
	}
	return ret
}

// ========================= 每日任务 =========================

func GetDailyTaskCount(ctx context.Context, uid int64, typ int) types.Map {
	return daily.GetAllByPrefix(uid, config.DailyTaskCount+types.ToString(typ))
}

// GetDailyTaskCountByID 获取日常任务计数
func GetDailyTaskCountByID(ctx context.Context, uid int64, typ, id int) int {
	v, _ := daily.GetByPrefix(uid, config.DailyTaskCount+types.ToString(typ), id)
	return types.ToIntE(v)
}

// IncrDailyTaskCountByID 增加日常任务计数
func IncrDailyTaskCountByID(ctx context.Context, uid int64, typ, id int, num int) {
	if num == 0 {
		num = 1
	}
	daily.IncrByPrefix(uid, config.DailyTaskCount+types.ToString(typ), id, int64(num))
}

func GetDailyTaskLingquAll(ctx context.Context, uid int64, typ int) types.Map {
	return daily.GetAllByPrefix(uid, config.DailyTaskLingqu+types.ToString(typ))
}

func GetDailyTaskLingquByID(ctx context.Context, uid int64, typ, id int) int {
	v, _ := daily.GetByPrefix(uid, config.DailyTaskLingqu+types.ToString(typ), id)
	return types.ToIntE(v)
}

// SetDailyTaskLingqu 设置日常任务领取状态
func SetDailyTaskLingqu(ctx context.Context, uid int64, typ, id int) {
	daily.SetByPrefix(uid, config.DailyTaskLingqu+types.ToString(typ), id, "1")
}

// GetDailyTaskActiveNum 获取日常任务活跃度
func GetDailyTaskActiveNum(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyTaskActiv)
	return types.ToIntE(v)
}

// AddDailyTaskActiveNum 增加日常任务活跃度
func AddDailyTaskActiveNum(ctx context.Context, uid int64, num int) {
	if num == 0 {
		num = 1
	}
	daily.Incr(uid, config.DailyTaskActiv, int64(num))
}

// GetDailyTaskActiveAll 获取日常任务活跃度领取状态（一次获取全部）
func GetDailyTaskActiveAll(ctx context.Context, uid int64) types.Map {
	return daily.GetAllByPrefix(uid, config.DailyTaskActivLingqu)
}

// GetDailyTaskActiveLingquByID 获取活跃度奖励领取状态
func GetDailyTaskActiveLingquByID(ctx context.Context, uid int64, id int) int {
	v, _ := daily.GetByPrefix(uid, config.DailyTaskActivLingqu, id)
	return types.ToIntE(v)
}

// SetDailyTaskActiveLingqu 设置活跃度奖励已领取
func SetDailyTaskActiveLingqu(ctx context.Context, uid int64, id int) {
	daily.SetByPrefix(uid, config.DailyTaskActivLingqu, id, "1")
}

// ========================= 任务 =========================

// IncrTaskFinishNumStr 增加任务完成次数（字符串 key）
func IncrTaskFinishNumStr(ctx context.Context, uid int64, taskID string, cdTime int) {
	var k string
	if cdTime == 1 {
		k = fmt.Sprintf(config.DWTaskFinishNum, uid, taskID, util.DateYmd())
	} else {
		k = fmt.Sprintf(config.DWTaskFinishNum, uid, taskID, util.DateW())
	}
	repo.RedisIncr(ctx, k)
	repo.RedisExpire(ctx, k, cdTime*86400)
}

func GetTaskFinishNumStr(ctx context.Context, uid int64, taskID string, cdTime int) int {
	var k string
	if cdTime == 1 {
		k = fmt.Sprintf(config.DWTaskFinishNum, uid, taskID, util.DateYmd())
	} else {
		k = fmt.Sprintf(config.DWTaskFinishNum, uid, taskID, util.DateW())
	}
	v, _ := repo.RedisGet(ctx, k)
	return types.ToIntE(v)
}

func GetActiveNum(ctx context.Context, uid int64, cdTime int) int {
	var k string
	if cdTime == 1 {
		k = fmt.Sprintf(config.DWActiveGetNum, uid, util.DateYmd())
	} else {
		k = fmt.Sprintf(config.DWActiveGetNum, uid, util.DateW())
	}
	v, _ := repo.RedisGet(ctx, k)
	return types.ToIntE(v)
}

func GetTaskLingqu(ctx context.Context, uid int64, taskID string, cdTime int) types.Map {
	var k string
	if cdTime == 1 {
		k = fmt.Sprintf(config.DWTaskLingqu, uid, taskID, util.DateYmd())
	} else {
		k = fmt.Sprintf(config.DWTaskLingqu, uid, taskID, util.DateW())
	}
	v, _ := repo.RedisHGetAll(ctx, k)
	return v
}

func GetTaskLingquDetail(ctx context.Context, uid int64, taskID string, number int, cdTime int) string {
	var k string
	if cdTime == 1 {
		k = fmt.Sprintf(config.DWTaskLingqu, uid, taskID, util.DateYmd())
	} else {
		k = fmt.Sprintf(config.DWTaskLingqu, uid, taskID, util.DateW())
	}
	v, _ := repo.RedisHGet(ctx, k, number)
	return v
}

// SetTaskLingqu 设置任务领取状态
func SetTaskLingqu(ctx context.Context, uid int64, task string, number, cdTime int) {
	var k string
	if cdTime == 1 {
		k = fmt.Sprintf(config.DWTaskLingqu, uid, task, util.DateYmd())
	} else {
		k = fmt.Sprintf(config.DWTaskLingqu, uid, task, util.DateW())
	}
	repo.RedisHSet(ctx, k, number, "1")
	repo.RedisExpire(ctx, k, cdTime*86400)
}
