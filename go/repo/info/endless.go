package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

// ==================== endless_hero ====================

func GetAllEndlessHero(ctx context.Context) ([]*table.EndlessHero, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("endless_hero")
	dest := []*table.EndlessHero{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== endless_buff ====================

func GetAllEndlessBuff(ctx context.Context) ([]*table.EndlessBuff, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("endless_buff")
	dest := []*table.EndlessBuff{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== endless_layer_reward ====================

func GetAllEndlessLayerReward(ctx context.Context) ([]*table.EndlessLayerReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("endless_layer_reward")
	dest := []*table.EndlessLayerReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== endless_monster_attr ====================

func GetAllEndlessMonsterAttr(ctx context.Context) ([]*table.MonsterAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("endless_monster_attr")
	dest := []*table.MonsterAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
