package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

func GetAllMergeConfig(ctx context.Context) ([]*table.MergeConfig, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("merge_config")
	dest := []*table.MergeConfig{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetAllRuneMerge(ctx context.Context) ([]*table.RuneMerge, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("rune_merge")
	dest := []*table.RuneMerge{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== rune_prop ====================

func GetAllRuneProp(ctx context.Context) ([]*table.RuneProp, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("rune_prop")
	dest := []*table.RuneProp{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== rune_consume ====================

func GetAllRuneConsume(ctx context.Context) ([]*table.RuneConsume, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("rune_consume")
	dest := []*table.RuneConsume{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
