package model

import (
	"context"
	"fmt"

	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
)

// SetHeroExchange 设置英雄转换数据
func SetHeroExchange(ctx context.Context, uid int64, exchangeID string, oldID int, newHeroID int) {
	data := types.Map{"old": oldID, "new": newHeroID}
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyHeroExchange, uid, exchangeID), data, 86400)
}

// GetHeroExchange 获取英雄转换数据
func GetHeroExchange(ctx context.Context, uid int64, exchangeID string) (int, int) {
	k := fmt.Sprintf(config.KeyHeroExchange, uid, exchangeID)
	v, _ := repo.RedisGet(ctx, k)
	repo.RedisDel(ctx, k)
	if v == "" {
		return 0, 0
	}
	ret := types.ToMapE(v)
	if ret == nil {
		return 0, 0
	}
	oldID := ret.GetIntE("old")
	newID := ret.GetIntE("new")
	return oldID, newID
}

// DelHeroExchange 删除英雄转换数据
func DelHeroExchange(ctx context.Context, uid int64, exchangeID string) {
	repo.RedisDel(ctx, fmt.Sprintf(config.KeyHeroExchange, uid, exchangeID))
}
