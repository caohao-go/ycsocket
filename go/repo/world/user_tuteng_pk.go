package world

import (
	"context"
	"fmt"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/cache"
	"server_golang/repo/table"
)

// ==================== user_tuteng_pk ====================

// GetUserTutengPkByUserId 根据用户 ID 获取图腾 PK 信息（带缓存，与 PHP getUserTutengPk 一致）
func GetUserTutengPkByUserId(ctx context.Context, userID int64) *table.UserTutengPk {
	redisKey := fmt.Sprintf(config.CacheUserTutengPk, userID)
	val, _ := cache.Get(redisKey)
	if val != "" {
		if val == config.EmptyString {
			return nil
		}
		var result table.UserTutengPk
		if json.Unmarshal(val, &result) == nil {
			return &result
		}
	}

	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_tuteng_pk").AndEqual("user_id", userID).Limit(1)
	pk := table.UserTutengPk{}
	err := client.FindOne(ctx, stmt, &pk)
	if err != nil {
		return nil
	}

	if pk.Id == 0 {
		cache.SetWithTTL(redisKey, config.EmptyString, 600)
		return nil
	}
	cache.SetWithTTL(redisKey, pk, 600)
	return &pk
}

func GetPkList(ctx context.Context, userID int64) ([]*table.UserTutengPk, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_tuteng_pk").AndNotEqual("user_id", userID).Limit(10)
	dest := []*table.UserTutengPk{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// InsertUserTutengPk 插入图腾 PK 信息并清缓存（与 PHP insertUserTutengPk 一致）
func InsertUserTutengPk(ctx context.Context, userID int64, data *table.UserTutengPk) (int64, error) {
	cache.Del(fmt.Sprintf(config.CacheUserTutengPk, userID))
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_tuteng_pk").InsertStruct(data)
	return client.Insert(ctx, stmt)
}

// UpdateUserTutengPkByUserId 更新图腾 PK 信息并清缓存（与 PHP updateUserTutengPk 一致）
func UpdateUserTutengPkByUserId(ctx context.Context, userID int64, data types.Map) error {
	cache.Del(fmt.Sprintf(config.CacheUserTutengPk, userID))
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_tuteng_pk").UpdateMap(extorm.SetMap(data)).AndEqual("user_id", userID)
	_, err := client.Update(ctx, stmt)
	return err
}

// ==================== user_tuteng_pk_detail ====================

// GetUserTutengPkDetail 获取用户PK详情记录（与 PHP getUserTutengPkDetail 一致：order by id desc limit 10）
func GetUserTutengPkDetail(ctx context.Context, userID int64) ([]*table.UserTutengPkDetail, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_tuteng_pk_detail").
		AndSQL("(user_id=? OR opp_user_id=?)", userID, userID).
		Order("id", true).Limit(10)
	dest := []*table.UserTutengPkDetail{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func InsertUserTutengPkDetail(ctx context.Context, data *table.UserTutengPkDetail) (int64, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_tuteng_pk_detail").InsertStruct(data)
	return client.Insert(ctx, stmt)
}
