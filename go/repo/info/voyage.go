package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

func GetAllVoyageProbability(ctx context.Context) ([]*table.VoyageProbability, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("voyage_probability")
	dest := []*table.VoyageProbability{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetAllVoyageFunction(ctx context.Context) ([]*table.VoyageFunction, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("voyage_function")
	dest := []*table.VoyageFunction{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
