// dao/guild_query.go 公会相关配置表查询
package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

// ==================== guild_member_limit ====================
func GetAllGuildMemberLimit(ctx context.Context) ([]*table.GuildMemberLimit, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_member_limit")
	dest := []*table.GuildMemberLimit{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== guild_skill_attr ====================
func GetAllGuildSkillAttr(ctx context.Context) ([]*table.GuildSkillAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_skill_attr")
	dest := []*table.GuildSkillAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== guildskill_consume ====================
func GetAllGuildskillConsume(ctx context.Context) ([]*table.GuildskillConsume, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guildskill_consume")
	dest := []*table.GuildskillConsume{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== guild_task ====================
func GetAllGuildTask(ctx context.Context) ([]*table.GuildTask, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_task")
	dest := []*table.GuildTask{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== guild_active_attr ====================
func GetAllGuildActiveAttr(ctx context.Context) ([]*table.GuildActiveAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_active_attr")
	dest := []*table.GuildActiveAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== guild_copy_reward ====================
func GetAllGuildCopyReward(ctx context.Context) ([]*table.GuildCopyReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_copy_reward")
	dest := []*table.GuildCopyReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetGuildCopyRewardByRewardType(ctx context.Context, rewardType int) ([]*table.GuildCopyReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_copy_reward").AndEqual("reward_type", rewardType)
	dest := []*table.GuildCopyReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== guild_boss_config ====================
func GetAllGuildBossConfig(ctx context.Context) ([]*table.GuildBossConfig, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("guild_boss_config")
	dest := []*table.GuildBossConfig{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== record_calculate ====================
func GetAllRecordCalculate(ctx context.Context) ([]*table.RecordCalculate, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("record_calculate")
	dest := []*table.RecordCalculate{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== record_reward ====================
func GetAllRecordReward(ctx context.Context) ([]*table.RecordReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("record_reward")
	dest := []*table.RecordReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
