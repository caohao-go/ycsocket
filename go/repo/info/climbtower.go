package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

func GetAllClimbtowerHero(ctx context.Context) ([]*table.ClimbtowerHero, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("climbtower_hero")
	dest := []*table.ClimbtowerHero{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetAllClimbtowerHeroOrderByLayer(ctx context.Context) ([]*table.ClimbtowerHero, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("climbtower_hero").Order("layer")
	dest := []*table.ClimbtowerHero{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// ==================== climbtower_monster_attr ====================

func GetAllClimbtowerMonsterAttr(ctx context.Context) ([]*table.MonsterAttr, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("climbtower_monster_attr")
	dest := []*table.MonsterAttr{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
