package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

// GetAllItems 获取所有道具配置
func GetAllItems(ctx context.Context) ([]*table.Items, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("items")

	dest := []*table.Items{}
	err := client.FindAll(ctx, stmt, &dest)
	if err != nil {
		return nil, err
	}
	return dest, nil
}

func GetAllItemsCollection(ctx context.Context) ([]*table.ItemsCollection, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("items_collection")
	dest := []*table.ItemsCollection{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
