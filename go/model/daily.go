package model

import (
	"context"
	"fmt"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/mem/daily"
)

// ========================= 每日签到 =========================

func GetDailyRewardTimes(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyRewardTimes)
	return types.ToIntE(v)
}

func IncrDailyRewardTimes(ctx context.Context, uid int64) int64 {
	return daily.Incr(uid, config.DailyRewardTimes, 1)
}

func GetDailyReward(ctx context.Context, uid int64) int {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyDailyReward, util.DateM(), uid))
	return types.ToIntE(v)
}

func IncrDailyReward(ctx context.Context, uid int64) int64 {
	k := fmt.Sprintf(config.KeyDailyReward, util.DateM(), uid)
	v, _ := repo.RedisIncr(ctx, k)
	repo.RedisExpire(ctx, k, 31*86400)
	return v
}

// ========================= 点金手 =========================

func GetDianCoin(ctx context.Context, uid int64, typ int) int {
	v, _ := repo.RedisHGet(ctx, fmt.Sprintf(config.KeyDianCoin, typ), uid)
	return 1 - types.ToIntE(v)
}

func SetDianCoin(ctx context.Context, uid int64, typ int) {
	repo.RedisHSet(ctx, fmt.Sprintf(config.KeyDianCoin, typ), uid, "1")
	repo.RedisExpire(ctx, fmt.Sprintf(config.KeyDianCoin, typ), util.DianCoinRefreshTime())
}
