package achieve

import (
	"context"
	"fmt"
	"sync"

	"git.code.oa.com/pcg-csd/trpc-ext/redis"
	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
)

// AchieveTask 成就任务结构体
type AchieveTask struct {
	UserId   int64 `json:"user_id"`
	TaskId   int   `json:"task_id"`
	Type     int   `json:"type"`
	Num      int   `json:"num"`
	ExtraNum int   `json:"extra_num"`
	Status   int   `json:"status"`
}

// userAchieves 用户成就数据：uid → typ → taskID → AchieveTask
var userAchieves = map[int64]map[int]map[int]AchieveTask{}
var userAchieveMutex = sync.RWMutex{}

// initUserAchieve 将用户成就数据从 pika 加载到内存
func initUserAchieve(uid int64) {
	userAchieveMutex.RLock()
	if _, ok := userAchieves[uid]; ok {
		userAchieveMutex.RUnlock()
		return
	}
	userAchieveMutex.RUnlock()

	ctx := context.Background()
	key := pikaKey(uid)

	// 从 pika 加载到内存
	data, err := repo.RedisHGetAll(ctx, key)
	if err != nil && !redis.IsNil(err) {
		panic(err)
	}

	userAchieveMutex.Lock()
	defer userAchieveMutex.Unlock()

	if len(data) == 0 {
		// 初始化用户属性
		initData := types.Map{config.UserId: uid}
		err = repo.RedisHMSet(ctx, key, initData)
		if err != nil {
			panic(err)
		}
		userAchieves[uid] = make(map[int]map[int]AchieveTask)
	} else {
		// 解析 pika 数据到新结构
		achieveMap := make(map[int]map[int]AchieveTask)
		prefix := config.UserAchieve + "_"
		prefixLen := len(prefix)

		for field, val := range data {
			if len(field) <= prefixLen || field[:prefixLen] != prefix {
				continue
			}

			var task AchieveTask
			json.Unmarshal(val, &task)
			if task.TaskId > 0 && task.Type > 0 {
				if achieveMap[task.Type] == nil {
					achieveMap[task.Type] = make(map[int]AchieveTask)
				}
				achieveMap[task.Type][task.TaskId] = task
			}
		}
		userAchieves[uid] = achieveMap
	}
}

// SetTask 设置用户成就任务数据
func SetTask(uid int64, typ, taskID int, task AchieveTask) {
	initUserAchieve(uid)

	userAchieveMutex.Lock()
	if userAchieves[uid][typ] == nil {
		userAchieves[uid][typ] = make(map[int]AchieveTask)
	}
	userAchieves[uid][typ][taskID] = task
	userAchieveMutex.Unlock()

	// 异步写入 pika
	go func(task AchieveTask) {
		ctx := context.Background()
		field := config.UserAchieve + "_" + types.ToString(typ) + "_" + types.ToString(taskID)
		err := repo.RedisHSet(ctx, pikaKey(uid), field, json.Marshal(task))
		if err != nil {
			log.Errorf(ctx, -1, "SetTask failed for uid=%d typ=%d taskID=%d, err=%v", uid, typ, taskID, err)
		}
	}(task)
}

// GetTask 获取用户指定成就任务数据
func GetTask(uid int64, typ, taskID int) (AchieveTask, bool) {
	initUserAchieve(uid)

	userAchieveMutex.RLock()
	defer userAchieveMutex.RUnlock()

	if typMap, ok := userAchieves[uid][typ]; ok {
		if task, ok := typMap[taskID]; ok {
			return task, true
		}
	}
	return AchieveTask{}, false
}

// GetTasksByType 获取用户指定类型的所有成就任务
func GetTasksByType(uid int64, typ int) map[int]AchieveTask {
	initUserAchieve(uid)

	userAchieveMutex.RLock()
	defer userAchieveMutex.RUnlock()

	ret := make(map[int]AchieveTask)
	if typMap, ok := userAchieves[uid][typ]; ok {
		for taskID, task := range typMap {
			ret[taskID] = task
		}
	}
	return ret
}

// pikaKey 返回用户成就在 pika 中的 hash key
func pikaKey(uid int64) string {
	return fmt.Sprintf("%s_%d", config.UserAchieve, uid)
}
