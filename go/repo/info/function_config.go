package info

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

func GetAllFunctionConfig(ctx context.Context) ([]*table.FunctionConfig, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("function_config")
	dest := []*table.FunctionConfig{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetCheckpointDaliyRewardByCopyId(ctx context.Context, copyId int) ([]*table.CheckpointDaliyReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("checkpoint_daliy_reward").AndEqual("copy_id", copyId)
	dest := []*table.CheckpointDaliyReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func GetCheckpointDaliyRewardMap(ctx context.Context) (map[int]*table.CheckpointDaliyReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("checkpoint_daliy_reward")
	dest := []*table.CheckpointDaliyReward{}
	err := client.FindAll(ctx, stmt, &dest)
	if err != nil {
		return nil, err
	}

	result := make(map[int]*table.CheckpointDaliyReward)
	for _, v := range dest {
		result[v.CopyId] = v
	}
	return result, nil
}

// ==================== checkpoint_reward ====================

func GetAllCheckpointReward(ctx context.Context) ([]*table.CheckpointReward, error) {
	client := repo.InfoDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("checkpoint_reward")
	dest := []*table.CheckpointReward{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}
