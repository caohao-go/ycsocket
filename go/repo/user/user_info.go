package user

import (
	"context"
	"fmt"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/cache"
	"server_golang/repo/table"
)

// 根据 open_id 获取用户信息
func GetUserInfoByOpenId(ctx context.Context, openId string) (*table.UserInfo, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_info").AndEqual("open_id", openId).Limit(1)

	dest := &table.UserInfo{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil {
		if extorm.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}
	return dest, nil
}

// 根据 user_id 获取用户信息（返回 Map，兼容旧代码过渡用）
func GetUserInfoByUserId(ctx context.Context, userId int64) (*table.UserInfo, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_info").AndEqual("user_id", userId).Limit(1)

	dest := table.UserInfo{}
	err := client.FindOne(ctx, stmt, &dest)
	if err != nil {
		return nil, err
	}

	return &dest, nil
}

// 插入用户信息
func InsertUserInfo(ctx context.Context, data *table.UserInfo) (int64, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_info").InsertStruct(data)
	return client.Insert(ctx, stmt)
}

// 根据 user_id 更新用户信息
func UpdateUserInfoByUserId(ctx context.Context, userId int64, updateData types.Map) error {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_info").
		UpdateMap(extorm.SetMap(updateData)).
		AndEqual("user_id", userId)
	_, err := client.Update(ctx, stmt)
	cache.Del(fmt.Sprintf(config.CacheRedisUserInfo, userId))
	return err
}

// 根据 user_id 删除用户信息
func DeleteUserInfoByUserId(ctx context.Context, userId int64) error {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_info").AndEqual("user_id", userId)
	_, err := client.Delete(ctx, stmt)
	cache.Del(fmt.Sprintf(config.CacheRedisUserInfo, userId))
	return err
}
