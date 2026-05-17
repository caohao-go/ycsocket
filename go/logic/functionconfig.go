// 日常副本配置模块 - 开启条件、次数、奖励
package logic

import (
	"context"
	"fmt"
	"strings"

	"server_golang/common/util"
	"server_golang/repo/info"
)

// 日常副本类型常量
const (
	DailyFunctionCoin       = 1 // 金币副本
	DailyFunctionExp        = 2 // 经验副本
	DailyFunctionHero       = 3 // 英雄副本
	DailyFunctionGod        = 4 // 神器副本
	DailyFunctionFuwen      = 5 // 符文副本
	DailyFunctionEndless    = 6 // 无尽试炼
	DailyFunctionExpedition = 7 // 英雄远征
)

// 日常副本ID映射
var FunctionIDs = map[int][]int{
	DailyFunctionCoin:       {20011, 20012, 20013, 20014, 20015, 20016, 20017, 20018, 20019},
	DailyFunctionExp:        {20021, 20022, 20023, 20024, 20025, 20026, 20027, 20028, 20029},
	DailyFunctionHero:       {20031, 20032, 20033, 20034, 20035, 20036, 20037, 20038, 20039},
	DailyFunctionGod:        {20041, 20042, 20043, 20044, 20045, 20046, 20047, 20048, 20049},
	DailyFunctionFuwen:      {20051, 20052, 20053, 20054, 20055, 20056, 20057, 20058, 20059},
	DailyFunctionExpedition: {20201, 20202, 20203},
}

type FunctionConfig struct {
	Id         int            `json:"id"`
	CopyId     int            `json:"copy_id"`
	Name       string         `json:"name"`
	OpenType   []util.TypeNum `json:"open_type"`
	CdTime     int            `json:"cd_time"`
	FreeCount  int            `json:"free_count"`
	CostType   int            `json:"cost_type"`
	BasicCost  int            `json:"basic_cost"`
	IncreaCost int            `json:"increa_cost"`
	LimitCost  int            `json:"limit_cost"`
	Layer      int            `json:"layer"`
	Reward     []util.TypeNum `json:"reward"`
}

// 日常副本全局数据
var (
	// 副本配置 copy_id => data
	FunctionConfigs map[int]FunctionConfig
	// copy_id => function_id 映射
	CopyIDsID map[int]int
)

// InitFunctionConfig 初始化日常副本配置
func InitFunctionConfig(ctx context.Context) {
	FunctionConfigs = make(map[int]FunctionConfig)
	CopyIDsID = make(map[int]int)

	fcRows, err := info.GetAllFunctionConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("init function_config error: %v", err))
	}

	// 获取奖励
	rewardRows, err := info.GetCheckpointDaliyRewardMap(ctx)
	if err != nil {
		panic(fmt.Errorf("init checkpoint_daliy_reward error: %v", err))
	}

	for _, fc := range fcRows {
		tmp := FunctionConfig{}
		parts := strings.Split(fc.Name, "-")
		tmp.Name = parts[len(parts)-1]
		tmp.CdTime = fc.CdTime
		tmp.FreeCount = fc.FreeCount
		tmp.CostType = fc.CostType
		tmp.BasicCost = fc.BasicCost
		tmp.IncreaCost = fc.IncreaCost
		tmp.LimitCost = fc.LimitCost
		tmp.Layer = fc.Layer
		tmp.OpenType = util.ToTypeNums(fc.OpenType)

		rewardRow, ok := rewardRows[fc.CopyId]
		if ok {
			tmp.Reward = util.ToTypeNums(rewardRow.Reward)
		}

		FunctionConfigs[fc.CopyId] = tmp
	}

	// 建立copy_id => function_id映射
	for id, copyIDs := range FunctionIDs {
		for _, copyID := range copyIDs {
			CopyIDsID[copyID] = id
		}
	}
}
