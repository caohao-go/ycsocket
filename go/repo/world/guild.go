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

// ==================== guild ====================

// GetGuildById 根据 ID 获取公会信息（带缓存）
func GetGuildById(ctx context.Context, id int) *table.Guild {
	redisKey := fmt.Sprintf(config.CacheGuildInfo, id)
	val, _ := cache.Get(redisKey)
	if val != "" {
		if val == config.EmptyString {
			return nil
		}
		var result table.Guild
		if json.Unmarshal(val, &result) == nil {
			return &result
		}
	}
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild").AndEqual("id", id).Limit(1)
	dest := &table.Guild{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil {
		if extorm.IsNil(err) {
			cache.SetWithTTL(redisKey, config.EmptyString, 600)
			return nil
		}
		return nil
	}
	cache.SetWithTTL(redisKey, dest, 600)
	return dest
}

func GetGuildListByName(ctx context.Context, name string) ([]*table.Guild, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild").AndLike("guild_name", "%"+name+"%", false).Limit(20)
	dest := []*table.Guild{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetGuildListTop50(ctx context.Context) ([]*table.Guild, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild").Order("people_num", true).Limit(50)
	dest := []*table.Guild{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetGuildAll(ctx context.Context) ([]*table.Guild, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild")
	dest := []*table.Guild{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetGuildRank(ctx context.Context) ([]*table.Guild, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild").Orders("guild_lv DESC", "exp DESC").Limit(50)
	dest := []*table.Guild{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetGuildsForFight(ctx context.Context) ([]*table.Guild, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild").AndGte("people_num", 5).AndGte("guild_lv", 3)
	dest := []*table.Guild{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func IncrGuildPeopleNum(ctx context.Context, id int) error {
	client := repo.WorldDB()
	_, err := client.Exec(ctx, "UPDATE guild SET people_num=people_num+1 WHERE id=?", id)
	return err
}

func DecrGuildPeopleNum(ctx context.Context, id int) error {
	client := repo.WorldDB()
	_, err := client.Exec(ctx, "UPDATE guild SET people_num=people_num-1 WHERE id=?", id)
	return err
}

func InsertGuild(ctx context.Context, data *table.Guild) (int64, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild").InsertStruct(data)
	return client.Insert(ctx, stmt)
}

func UpdateGuildByIdWithCache(ctx context.Context, id int, data types.Map) error {
	cache.Del(fmt.Sprintf(config.CacheGuildInfo, id))
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild").UpdateMap(extorm.SetMap(data)).AndEqual("id", id)
	_, err := client.Update(ctx, stmt)
	return err
}

// ==================== users_guild ====================

// GetUsersGuildByUserId 根据用户 ID 获取公会关系
func GetUsersGuildByUserId(ctx context.Context, userID int64) *table.UsersGuild {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("users_guild").AndEqual("user_id", userID).Limit(1)
	dest := &table.UsersGuild{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil {
		return nil
	}
	return dest
}

// GetGuildFuhuizhangCount 获取公会副会长数量（按 zhiwei=2 查询，与 PHP 一致）
func GetGuildFuhuizhangCount(ctx context.Context, guildID int) (int, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("users_guild").AndEqual("guild_id", guildID).AndEqual("zhiwei", 2)
	cnt, err := client.Count(ctx, stmt)
	return int(cnt), err
}

func GetGuildsUserByGuildId(ctx context.Context, guildID int, onlyUserId bool) ([]*table.UsersGuild, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("users_guild").AndEqual("guild_id", guildID)
	if onlyUserId {
		stmt.Select("user_id")
	}
	dest := []*table.UsersGuild{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetGuildsUsersByGuildIds(ctx context.Context, guildIDs []int) ([]*table.UsersGuild, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("users_guild").Select("user_id", "guild_id").AndIn("guild_id", guildIDs)
	dest := []*table.UsersGuild{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func InsertUsersGuild(ctx context.Context, data *table.UsersGuild) (int64, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("users_guild").InsertStruct(data)
	return client.Insert(ctx, stmt)
}

func UpdateUsersGuildByUserId(ctx context.Context, userID int64, data types.Map) error {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("users_guild").UpdateMap(extorm.SetMap(data)).AndEqual("user_id", userID)
	_, err := client.Update(ctx, stmt)
	return err
}

// ==================== guild_apply ====================

func GetGuildsApplyUser(ctx context.Context, guildID int) ([]*table.GuildApply, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_apply").AndEqual("guild_id", guildID)
	dest := []*table.GuildApply{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func InsertGuildApply(ctx context.Context, data *table.GuildApply) (int64, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_apply").InsertStruct(data)
	return client.Insert(ctx, stmt)
}

// DeleteUsersGuildByUserId 删除用户公会关系
func DeleteUsersGuildByUserId(ctx context.Context, userID int64) error {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("users_guild").AndEqual("user_id", userID)
	_, err := client.Delete(ctx, stmt)
	return err
}

// DeleteGuildApply 删除公会申请
func DeleteGuildApply(ctx context.Context, guildID, userID int64) error {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_apply").AndEqual("guild_id", guildID).AndEqual("user_id", userID)
	_, err := client.Delete(ctx, stmt)
	return err
}

// ==================== guild_chapter_blood ====================

// GetGuildChapterBlood 获取公会副本章节血量
func GetGuildChapterBlood(ctx context.Context, guildID int) *table.GuildChapterBlood {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_chapter_blood").AndEqual("id", guildID).Limit(1)
	dest := &table.GuildChapterBlood{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil {
		return nil
	}
	return dest
}

// InsertGuildChapterBlood 插入公会副本章节血量
func InsertGuildChapterBlood(ctx context.Context, data *table.GuildChapterBlood) (int64, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_chapter_blood").InsertStruct(data)
	return client.Insert(ctx, stmt)
}

// UpdateGuildChapterBlood 更新公会副本章节血量
func UpdateGuildChapterBlood(ctx context.Context, guildID int, chapter, chapterBlood int) error {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_chapter_blood").
		UpdateMap(extorm.SetMap(types.Map{"chapter": chapter, "chapter_blood": chapterBlood})).
		AndEqual("id", guildID)
	_, err := client.Update(ctx, stmt)
	return err
}
