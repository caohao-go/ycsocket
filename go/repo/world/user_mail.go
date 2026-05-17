package world

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/json"
	"server_golang/repo"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo/cache"
	"server_golang/repo/table"
)

// GetUserMailListWithCache 获取用户邮件列表（带缓存）
// 与 PHP getUserMail 对齐：
//  1. 查询 send_time >= now - 15 天 的邮件
//  2. 按 send_time 降序
//  3. 结果以 []*table.UserMail 返回；上层负责 add_items/send_time 字段的转换
func GetUserMailListWithCache(ctx context.Context, userID int64) []*table.UserMail {
	redisKey := fmt.Sprintf(config.CacheUserMail, userID)
	val, _ := cache.Get(redisKey)
	if val != "" {
		if val == config.EmptyString {
			return nil
		}
		var result []*table.UserMail
		if json.Unmarshal(val, &result) == nil {
			return result
		}
	}
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	// 15 天过滤（与 PHP 一致）：send_time >= date('Y-m-d H:i:s', time()-86400*15)
	cutoff := time.Now().Add(-15 * 24 * time.Hour).Format("2006-01-02 15:04:05")
	stmt.SetTableName("user_mail").
		AndEqual("user_id", userID).
		AndGte("send_time", cutoff).
		Order("send_time", true)
	dest := []*table.UserMail{}
	err := client.FindAll(ctx, stmt, &dest)
	if err != nil || len(dest) == 0 {
		cache.SetWithTTL(redisKey, config.EmptyString, 600)
		return nil
	}
	cache.SetWithTTL(redisKey, dest, 600)
	return dest
}

// GetUserMailByIds 根据 user_id 和 id 列表获取邮件
func GetUserMailByIds(ctx context.Context, userID int64, ids []int64) []*table.UserMail {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_mail").AndEqual("user_id", userID).AndIn("id", ids)
	dest := []*table.UserMail{}
	err := client.FindAll(ctx, stmt, &dest)
	if err != nil {
		return nil
	}
	return dest
}

// UpdateUserMailByIdWithCache 根据 user_id + id 更新邮件并清缓存
// 与 PHP readUserMail 一致：updateTable("user_mail", ["user_id"=>$userid, "id"=>$id], ...)
func UpdateUserMailByIdWithCache(ctx context.Context, userID, id int64, data types.Map) error {
	cache.Del(fmt.Sprintf(config.CacheUserMail, userID))
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_mail").
		UpdateMap(extorm.SetMap(data)).
		AndEqual("user_id", userID).
		AndEqual("id", id)
	_, err := client.Update(ctx, stmt)
	return err
}

// LingquMailByIdsWithCache 批量标记邮件附件为已领取并清缓存
// 与 PHP lingquMail 一致：updateTable("user_mail", [..., "id"=>$ids], ['item_get_flag'=>2], $redis_key)
func LingquMailByIdsWithCache(ctx context.Context, userID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	cache.Del(fmt.Sprintf(config.CacheUserMail, userID))
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_mail").
		UpdateMap(extorm.SetMap(types.Map{"item_get_flag": 2})).
		AndIn("id", ids)
	_, err := client.Update(ctx, stmt)
	return err
}

// DeleteUserMailByIdsWithCache 根据 user_id + ID 列表删除邮件并清缓存
// 与 PHP delUserMail 一致：deleteTable("user_mail", ["user_id"=>$userid, "id"=>$ids], $redis_key)
func DeleteUserMailByIdsWithCache(ctx context.Context, userID int64, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	cache.Del(fmt.Sprintf(config.CacheUserMail, userID))
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_mail").AndEqual("user_id", userID).AndIn("id", ids)
	_, err := client.Delete(ctx, stmt)
	return err
}

// InsertUserMail 插入邮件并清缓存（与 PHP insertUserMail 一致）
func InsertUserMail(ctx context.Context, userID int64, title, content, sendTime string, itemGetFlag int, addItems string) {
	cache.Del(fmt.Sprintf(config.CacheUserMail, userID))
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	mail := &table.UserMail{
		UserId:      userID,
		Title:       title,
		Content:     content,
		SendTime:    sendTime,
		ItemGetFlag: itemGetFlag,
	}
	if addItems != "" {
		mail.AddItems = addItems
	}
	stmt.SetTableName("user_mail").InsertStruct(mail)
	client.Insert(ctx, stmt)
}
