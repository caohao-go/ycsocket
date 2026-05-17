package model

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/repo"
)

// 符文重铸交换

// SetRuneExchange 保存符文重铸交换数据
func SetRuneExchange(ctx context.Context, uid int64, exchangeID string, data map[string]*util.Fu) {
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyRuneExchange, uid, exchangeID), data, 86400)
}

// GetRuneExchange 获取符文重铸交换数据
func GetRuneExchange(ctx context.Context, uid int64, exchangeID string) map[string]*util.Fu {
	k := fmt.Sprintf(config.KeyRuneExchange, uid, exchangeID)
	v, _ := repo.RedisGet(ctx, k)
	repo.RedisDel(ctx, k)
	if v == "" {
		return nil
	}
	var ret = map[string]*util.Fu{}
	json.Unmarshal(v, &ret)
	return ret
}

// DelRuneExchange 删除符文重铸交换数据
func DelRuneExchange(ctx context.Context, uid int64, exchangeID string) {
	repo.RedisDel(ctx, fmt.Sprintf(config.KeyRuneExchange, uid, exchangeID))
}
