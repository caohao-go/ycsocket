package user

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

// 插入一条 sequence 记录，返回自增ID
func InsertSequence(ctx context.Context, data *table.Sequence) (int64, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("sequence").InsertStruct(data)
	return client.Insert(ctx, stmt)
}
