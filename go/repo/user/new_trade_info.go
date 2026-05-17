package user

import (
	"context"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
	"server_golang/common/types"
	"server_golang/repo"
	"server_golang/repo/table"
)

// 插入订单记录
func InsertNewTradeInfo(ctx context.Context, data *table.NewTradeInfo) (int64, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("new_trade_info").InsertStruct(data)
	return client.Insert(ctx, stmt)
}

// 根据 id 获取订单
func GetNewTradeInfoById(ctx context.Context, id int64) (*table.NewTradeInfo, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("new_trade_info").AndEqual("id", id).Limit(1)

	dest := &table.NewTradeInfo{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil {
		if extorm.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}
	return dest, nil
}

// 根据 id 更新订单
func UpdateNewTradeInfoById(ctx context.Context, id int64, updateData types.Map) error {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("new_trade_info").
		UpdateMap(extorm.SetMap(updateData)).
		AndEqual("id", id)
	_, err := client.Update(ctx, stmt)
	return err
}
