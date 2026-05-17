package model

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/mem/daily"
)

// ========================= 挂机/快速挂机 =========================

// GetFastUsedCount 获取快速战斗已使用次数
func GetFastUsedCount(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyFastUsedCnt)
	return types.ToIntE(v)
}

// IncrFastUsedCount 增加快速战斗已使用次数
func IncrFastUsedCount(ctx context.Context, uid int64) {
	daily.Incr(uid, config.DailyFastUsedCnt, 1)
}

// GetOnHookTime 获取上次挂机时间，返回分钟数和小时数
func GetOnHookTime(ctx context.Context, uid int64, timeUp int, set bool) (minVal int, hourVal int) {
	rkMin := fmt.Sprintf(config.KeyOnHookTimeM, uid)
	rkHour := fmt.Sprintf(config.KeyOnHookTimeH, uid)
	rkNew := fmt.Sprintf(config.KeyNewOnHook, uid)
	minLast, _ := repo.RedisGet(ctx, rkMin)
	minLastT := types.ToIntE(minLast)
	hourLast, _ := repo.RedisGet(ctx, rkHour)
	hourLastT := types.ToIntE(hourLast)
	now := int(time.Now().Unix())
	minDur := now - minLastT
	newHook := false
	newOnHookVal, _ := repo.RedisGet(ctx, rkNew)
	if minLastT == 0 && types.ToIntE(newOnHookVal) == 0 {
		repo.RedisSet(ctx, rkMin, now-120, 12*3600)
		repo.RedisSet(ctx, rkNew, "1")
		minDur = 120
		newHook = true
	} else if minLastT == 0 {
		minDur = (12 + timeUp) * 3600
	}
	hourDur := now - hourLastT
	if newHook {
		repo.RedisSet(ctx, rkHour, now-120, 12*3600)
		repo.RedisSet(ctx, rkNew, "1")
		hourDur = 120
	} else if hourLastT == 0 {
		hourDur = (12 + timeUp) * 3600
	}
	minVal = minDur / 60
	hourVal = hourDur / 3600
	if set {
		repo.RedisSet(ctx, rkMin, now-minDur%60, 12*3600)
		repo.RedisSet(ctx, rkHour, now-hourDur%3600, 12*3600)
	}
	return
}
