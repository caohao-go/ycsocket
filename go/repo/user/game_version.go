package user

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

func GetAllGameVersion(ctx context.Context) ([]*table.GameVersion, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("game_version")
	dest := []*table.GameVersion{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetGameVersionLatest(ctx context.Context) (*table.GameVersion, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("game_version").Order("id", true).Limit(1)
	dest := &table.GameVersion{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil {
		return nil, err
	}
	return dest, nil
}
