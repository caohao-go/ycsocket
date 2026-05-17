package logic

import (
	"context"
	"fmt"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/util"
	"server_golang/repo/info"
)

// AchieveTaskConfig 成就任务配置（对应 PHP Task::$achieve_datas[$task_id]）
type AchieveTaskConfig struct {
	ID       int            // 任务ID
	Type     int            // 类型
	Num      int            // 完成所需数量阈值
	ExtraNum int            // 额外数据要求
	Reward   []util.TypeNum // 奖励 [{type: xx, num: xx}, ...]
}

// AchieveDatas 成就任务配置 taskID => AchieveTaskConfig（对应 PHP Task::$achieve_datas）
var AchieveDatas map[int]*AchieveTaskConfig

// AchieveTypes 成就任务类型映射 taskID => type（对应 PHP Task::$achieve_types）
var AchieveTypes map[int]int

// InitAchieve 初始化成就任务配置（从 task_achieve 表加载）
func InitAchieve(ctx context.Context) {
	AchieveDatas = make(map[int]*AchieveTaskConfig)
	AchieveTypes = make(map[int]int)

	rows, err := info.GetAllTaskAchieve(ctx)
	if err != nil {
		panic(fmt.Errorf("InitAchieve: GetAllTaskAchieve error: %v", err))
	}

	for _, row := range rows {
		rewards := util.ToTypeNums(row.Reward)

		AchieveDatas[row.Id] = &AchieveTaskConfig{
			ID:       row.Id,
			Type:     row.Type,
			Num:      row.Num,
			ExtraNum: row.ExtraNum,
			Reward:   rewards,
		}
		AchieveTypes[row.Id] = row.Type
	}

	log.Infof(ctx, "InitAchieve: loaded %d task_achieve records", len(AchieveDatas))
}
