package world

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/cache"
	"server_golang/repo/table"
)

// DeleteUserTongguanReward 删除通关奖励并清缓存
func DeleteUserTongguanReward(ctx context.Context, userID int64, rewardType, copyID int) error {
	cache.Del(fmt.Sprintf(config.CacheUserTongguanReward, userID, rewardType))
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_tongguan_reward").
		AndEqual("user_id", userID).
		AndEqual("type", rewardType).
		AndEqual("copy", copyID)
	_, err := client.Delete(ctx, stmt)
	return err
}

// ReplaceUserTongguanReward 替换通关奖励并清缓存
func ReplaceUserTongguanReward(ctx context.Context, data *table.UserTongguanReward) error {
	cache.Del(fmt.Sprintf(config.CacheUserTongguanReward, data.UserId, data.Type))
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_tongguan_reward").ReplaceStruct(data)
	_, err := client.Replace(ctx, stmt)
	return err
}

// GetUserTongguanRewardList 获取通关奖励列表（带缓存）
func GetUserTongguanRewardList(ctx context.Context, userID int64, rewardType int) []*table.UserTongguanReward {
	redisKey := fmt.Sprintf(config.CacheUserTongguanReward, userID, rewardType)
	val, _ := cache.Get(redisKey)
	if val != "" {
		if val == config.EmptyString {
			return nil
		}
		var result []*table.UserTongguanReward
		if json.Unmarshal(val, &result) == nil {
			return result
		}
	}
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_tongguan_reward").AndEqual("user_id", userID).AndEqual("type", rewardType)
	dest := []*table.UserTongguanReward{}
	err := client.FindAll(ctx, stmt, &dest)
	if err != nil || len(dest) == 0 {
		cache.SetWithTTL(redisKey, config.EmptyString, 600)
		return nil
	}
	cache.SetWithTTL(redisKey, dest, 600)
	return dest
}
