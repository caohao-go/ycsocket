package world

import (
	"context"
	"fmt"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
	"server_golang/common/json"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/cache"
	"server_golang/repo/table"
)

// GetUserClimbtowerRecordByLayer 根据层数获取爬塔记录（带缓存，600s TTL）
func GetUserClimbtowerRecordByLayer(ctx context.Context, layer int) *table.UserClimbtowerRecord {
	redisKey := fmt.Sprintf(config.CacheUserClimbtowerRecord, layer)
	val, _ := cache.Get(redisKey)
	if val != "" {
		if val == config.EmptyString {
			return nil
		}
		var result table.UserClimbtowerRecord
		if json.Unmarshal(val, &result) == nil {
			return &result
		}
	}
	dest := getUserClimbtowerRecordByLayer(ctx, layer)
	if dest == nil {
		cache.SetWithTTL(redisKey, config.EmptyString, 600)
		return nil
	}
	cache.SetWithTTL(redisKey, dest, 600)
	return dest
}

// ReplaceUserClimbtowerRecord 替换爬塔记录并清缓存
func ReplaceUserClimbtowerRecord(ctx context.Context, data *table.UserClimbtowerRecord) error {
	cache.Del(fmt.Sprintf(config.CacheUserClimbtowerRecord, data.Layer))
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_climbtower_record").ReplaceStruct(data)
	_, err := client.Replace(ctx, stmt)
	return err
}

// 根据层数获取爬塔记录
func getUserClimbtowerRecordByLayer(ctx context.Context, layer int) *table.UserClimbtowerRecord {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_climbtower_record").AndEqual("layer", layer).Limit(1)
	dest := &table.UserClimbtowerRecord{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil {
		if extorm.IsNil(err) {
			return nil
		}
		return nil
	}
	return dest
}
