// dao/config_query.go 封装所有配置库（MysqlDefault）的只读查询操作
// 这些表在服务启动时全量加载到内存，后续不做写操作
package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

// ==================== chapter ====================

func GetAllChapter(ctx context.Context) ([]*table.Chapter, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("chapter")
	dest := []*table.Chapter{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== checkin_reward ====================

func GetAllCheckinReward(ctx context.Context) ([]*table.CheckinReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("checkin_reward")
	dest := []*table.CheckinReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== combination ====================

func GetAllCombination(ctx context.Context) ([]*table.Combination, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("combination")
	dest := []*table.Combination{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== position ====================

func GetAllPosition(ctx context.Context) ([]*table.Position, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("position")
	dest := []*table.Position{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== role_title ====================

func GetAllRoleTitle(ctx context.Context) ([]*table.RoleTitle, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("role_title")
	dest := []*table.RoleTitle{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== monster_attr ====================

func GetAllMonsterAttr(ctx context.Context) ([]*table.MonsterAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("monster_attr")
	dest := []*table.MonsterAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== lv_info ====================

func GetAllLvInfo(ctx context.Context) ([]*table.LvInfo, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("lv_info")
	dest := []*table.LvInfo{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== expedition_hero ====================

func GetAllExpeditionHero(ctx context.Context) ([]*table.ExpeditionHero, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("expedition_hero")
	dest := []*table.ExpeditionHero{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// 获取所有 VIP 配置
func GetAllVipConfig(ctx context.Context) ([]*table.VipConfig, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("vip_config")
	dest := []*table.VipConfig{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetAllLibaoRewards(ctx context.Context) ([]*table.LibaoRewards, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("libao_rewards")
	dest := []*table.LibaoRewards{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
