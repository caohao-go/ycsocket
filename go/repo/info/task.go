package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

// ==================== task_weekly_config ====================

func GetAllTaskWeeklyConfig(ctx context.Context) ([]*table.TaskWeeklyConfig, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("task_weekly_config")
	dest := []*table.TaskWeeklyConfig{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== task_config ====================

func GetAllTaskConfig(ctx context.Context) ([]*table.TaskConfig, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("task_config")
	dest := []*table.TaskConfig{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== task_active_config ====================

func GetAllTaskActiveConfig(ctx context.Context) ([]*table.TaskActiveConfig, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("task_active_config")
	dest := []*table.TaskActiveConfig{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== task_achieve ====================

func GetAllTaskAchieve(ctx context.Context) ([]*table.TaskAchieve, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("task_achieve")
	dest := []*table.TaskAchieve{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== task_guide ====================

func GetAllTaskGuide(ctx context.Context) ([]*table.TaskGuide, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("task_guide")
	dest := []*table.TaskGuide{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
