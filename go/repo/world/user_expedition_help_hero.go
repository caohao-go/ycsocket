package world

import (
	"context"
	"crypto/md5"
	"fmt"
	"sort"

	"server_golang/common/json"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/cache"
	"server_golang/repo/table"
)

// ReplaceUserExpeditionHelpHero 替换远征助战英雄
func ReplaceUserExpeditionHelpHero(ctx context.Context, data *table.UserExpeditionHelpHero) error {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_expedition_help_hero").ReplaceStruct(data)
	_, err := client.Replace(ctx, stmt)
	return err
}

// GetUserExpeditionHelpHeroByUserIds 根据用户 ID 列表获取助战英雄
func GetUserExpeditionHelpHeroByUserIds(ctx context.Context, userIDs []int64) []*table.UserExpeditionHelpHero {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_expedition_help_hero").AndIn("user_id", userIDs)
	dest := []*table.UserExpeditionHelpHero{}
	err := client.FindAll(ctx, stmt, &dest)
	if err != nil {
		return nil
	}
	return dest
}

// GetUserExpeditionHelpHeroByUserIdsWithCache 根据用户 ID 列表获取助战英雄（带缓存，10s TTL）
func GetUserExpeditionHelpHeroByUserIdsWithCache(ctx context.Context, userIDs []int64) []*table.UserExpeditionHelpHero {
	redisKey := fmt.Sprintf(config.CacheUserExpeditionHelpHero, expeditionHelpHeroCacheKey(userIDs))
	val, _ := cache.Get(redisKey)
	if val != "" {
		if val == config.EmptyString {
			return nil
		}
		var result []*table.UserExpeditionHelpHero
		if json.Unmarshal(val, &result) == nil {
			return result
		}
	}
	dest := GetUserExpeditionHelpHeroByUserIds(ctx, userIDs)
	if len(dest) == 0 {
		cache.SetWithTTL(redisKey, config.EmptyString, 10)
		return nil
	}
	cache.SetWithTTL(redisKey, dest, 10)
	return dest
}

// expeditionHelpHeroCacheKey 生成与 PHP md5(serialize($user_ids)) 对应的缓存 key 后缀
func expeditionHelpHeroCacheKey(userIDs []int64) string {
	sorted := make([]int64, len(userIDs))
	copy(sorted, userIDs)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	return fmt.Sprintf("%x", md5.Sum(json.MarshalToBytes(sorted)))
}
