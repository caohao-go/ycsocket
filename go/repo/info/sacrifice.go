package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

func GetAllSacrificeBase(ctx context.Context) ([]*table.SacrificeBase, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("sacrifice_base")
	dest := []*table.SacrificeBase{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetAllHeroStarUpReturn(ctx context.Context) ([]*table.HeroStarUpReturn, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("hero_star_up_return")
	dest := []*table.HeroStarUpReturn{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
