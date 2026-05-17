package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

func GetAllLuckRewardConfig(ctx context.Context) ([]*table.LuckRewardConfig, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("luck_reward_config")
	dest := []*table.LuckRewardConfig{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetAllLuckIntegralReward(ctx context.Context) ([]*table.LuckIntegralReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("luck_integral_reward")
	dest := []*table.LuckIntegralReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetAllTaobaoRand(ctx context.Context) ([]*table.TaobaoRand, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("taobao_rand")
	dest := []*table.TaobaoRand{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
