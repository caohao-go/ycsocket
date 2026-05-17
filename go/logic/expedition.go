// expedition 包实现了远征系统
package logic

import (
	"context"

	"server_golang/common/util"
	"server_golang/repo/info"
)

// 宝箱位置常量
var BaoxiangPos = map[int][]int{
	20201: {4, 8, 12, 16, 20},
	20202: {4, 8, 12, 16, 20},
	20203: {4, 8, 12, 16, 20},
}

var MaxPos = map[int]int{20201: 20, 20202: 20, 20203: 20}

var (
	ExpeditionRewards      = make(map[int]map[int][]util.TypeNum)
	ExpeditionTotalRewards = make(map[int][]util.TypeNum)
)

// InitExpedition 初始化远征数据
func InitExpedition(ctx context.Context) {
	initSumRewards(ctx, 20201)
	initSumRewards(ctx, 20202)
	initSumRewards(ctx, 20203)
}

func initSumRewards(ctx context.Context, copyID int) {
	initRewards(ctx, copyID)
	ret := make(map[int]int)
	for _, rewards := range ExpeditionRewards[copyID] {
		for _, r := range rewards {
			ret[r.Type] += r.Num
		}
	}
	for tp, num := range ret {
		ExpeditionTotalRewards[copyID] = append(ExpeditionTotalRewards[copyID],
			util.TypeNum{Type: tp, Num: num})
	}
}

func initRewards(ctx context.Context, copyID int) {
	rows, _ := info.GetCheckpointDaliyRewardByCopyId(ctx, copyID)
	if ExpeditionRewards[copyID] == nil {
		ExpeditionRewards[copyID] = make(map[int][]util.TypeNum)
	}
	for _, row := range rows {
		ExpeditionRewards[copyID][row.Checkpoint] = util.ToTypeNums(row.Reward)
	}
}
