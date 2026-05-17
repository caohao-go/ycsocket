package world

import (
	"context"

	"server_golang/repo"
	"server_golang/repo/table"
)

// IsUserRoleTitleExist 判断用户称号是否存在
func IsUserRoleTitleExist(ctx context.Context, userID int64, roleID int) bool {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_role_title").AndEqual("user_id", userID).AndEqual("role_id", roleID)
	cnt, err := client.Count(ctx, stmt)
	if err != nil || cnt == 0 {
		return false
	}
	return true
}

// GetUserRoleTitlesByUserId 获取用户称号列表
func GetUserRoleTitlesByUserId(ctx context.Context, userID int64) []*table.UserRoleTitle {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_role_title").AndEqual("user_id", userID)
	dest := []*table.UserRoleTitle{}
	err := client.FindAll(ctx, stmt, &dest)
	if err != nil {
		return nil
	}
	return dest
}
