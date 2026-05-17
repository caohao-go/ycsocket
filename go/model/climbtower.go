package model

import (
	"fmt"

	"golang.org/x/net/context"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/mem/daily"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

type ClimbTowerSaodangResult struct {
	FreeNum   int `json:"free_num"`
	BuyNum    int `json:"buy_num"`
	CostType  int `json:"cost_type"`
	BasicCost int `json:"basic_cost"`
}

// 获取爬塔记录
func GetUserClimbtowerRecord(ctx context.Context, layer int) *table.UserClimbtowerRecord {
	return world.GetUserClimbtowerRecordByLayer(ctx, layer)
}

// 替换爬塔记录
func ReplaceUserClimbtowerRecord(ctx context.Context, layer int, data *table.UserClimbtowerRecord) error {
	return world.ReplaceUserClimbtowerRecord(ctx, data)
}

// ========================= pika =========================

func GetClimbtowerSaodang(ctx context.Context, uid int64) ClimbTowerSaodangResult {
	v, _ := daily.Get(uid, config.DailyClimbtowerSaodang)
	num := 5 - types.ToIntE(v)
	freeNum := num - 3
	if freeNum < 0 {
		freeNum = 0
	}
	buyNum := num - freeNum
	var cost int
	if freeNum > 0 {
		cost = 0
	} else {
		cost = 20 + 10*(3-num)
		if cost > 40 {
			cost = 40
		}
	}
	return ClimbTowerSaodangResult{freeNum, buyNum, 2, cost}
}

func DecrClimbtowerSaodang(ctx context.Context, uid int64) {
	daily.Incr(uid, config.DailyClimbtowerSaodang, 1)
}

func SetClimbLingquRewards(ctx context.Context, uid int64, layer int, rewards []util.TypeNum) {
	repo.RedisHSet(ctx, fmt.Sprintf(config.KeyClimbLingquReward, uid), layer, rewards)
}

func GetAllClimbLingquRewards(ctx context.Context, uid int64) map[string][]util.TypeNum {
	v, _ := repo.RedisHGetAll(ctx, fmt.Sprintf(config.KeyClimbLingquReward, uid))

	ret := map[string][]util.TypeNum{}
	for layer, rewardsStr := range v {
		ret[layer] = util.ToTypeNums(types.ToString(rewardsStr))
	}
	return ret
}

func GetClimbLingquRewards(ctx context.Context, uid int64, layer int) []util.TypeNum {
	v, _ := repo.RedisHGet(ctx, fmt.Sprintf(config.KeyClimbLingquReward, uid), layer)
	if v == "" {
		return nil
	}

	var rewards = []util.TypeNum{}
	json.Unmarshal(v, &rewards)
	return rewards
}

func LingquClimbLingquRewards(ctx context.Context, uid int64, layer int) {
	repo.RedisHDel(ctx, fmt.Sprintf(config.KeyClimbLingquReward, uid), layer)
}
