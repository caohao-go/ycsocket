package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

// ==================== copy ====================

func GetAllCopy(ctx context.Context) ([]*table.Copy, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("copy")
	dest := []*table.Copy{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetAllCopyReward(ctx context.Context) ([]*table.CopyReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("copy_reward")
	dest := []*table.CopyReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== copy_boss_reward ====================

func GetAllCopyBossReward(ctx context.Context) ([]*table.CopyBossReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("copy_boss_reward")
	dest := []*table.CopyBossReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== onhook_reward ====================

func GetAllOnhookReward(ctx context.Context) ([]*table.OnhookReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("onhook_reward")
	dest := []*table.OnhookReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== random_count_config ====================

func GetAllRandomCountConfig(ctx context.Context) ([]*table.RandomCountConfig, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("random_count_config")
	dest := []*table.RandomCountConfig{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
