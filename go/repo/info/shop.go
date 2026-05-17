package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

func GetAllShopSell(ctx context.Context) ([]*table.ShopSell, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("shop_sell")
	dest := []*table.ShopSell{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetAllShopFunctionConfig(ctx context.Context) ([]*table.ShopFunctionConfig, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("shop_function_config")
	dest := []*table.ShopFunctionConfig{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
