package user

import (
	"context"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
	"server_golang/repo"
	"server_golang/repo/table"
)

// GetInviteByUserID 查询是否已被邀请过
// 对应 PHP: UserinfoModel::getInviteByUserid
func GetInviteByUserID(ctx context.Context, userID int64) (*table.InviteInfo, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("invite_info").AndEqual("user_id", userID).Limit(1)

	dest := &table.InviteInfo{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil {
		if extorm.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}
	return dest, nil
}

// GetMyInvitePlayers 查询我邀请的所有玩家
// 对应 PHP: UserinfoModel::getMyInvitePlayer
func GetMyInvitePlayers(ctx context.Context, inviteZoneUID int64) ([]*table.InviteInfo, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("invite_info").AndEqual("invite_zone_uid", inviteZoneUID)

	dest := []*table.InviteInfo{}
	err := client.FindAll(ctx, stmt, &dest)
	if err != nil {
		if extorm.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}
	return dest, nil
}

// InsertInviteInfo 插入邀请记录
// 对应 PHP: UserinfoModel::insertInviteInfo
func InsertInviteInfo(ctx context.Context, data *table.InviteInfo) (int64, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("invite_info").InsertStruct(data)
	return client.Insert(ctx, stmt)
}
