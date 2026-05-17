package user

import (
	"context"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
	"server_golang/common/types"
	"server_golang/repo"
	"server_golang/repo/table"
)

func ReplaceUserLoginZones(ctx context.Context, data types.Map) error {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_login_zones").ReplaceMap(extorm.SetMap(data))
	_, err := client.Replace(ctx, stmt)
	return err
}

// GetLoginZone 查询用户登录过的区服记录
// 对应 PHP: UserinfoModel::getLoginZone
func GetLoginZone(ctx context.Context, userID int64) ([]types.Map, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_login_zones").AndEqual("user_id", userID).Order("utime", true)
	rows, err := client.FindAllMaps(ctx, stmt)
	if err != nil {
		return nil, err
	}
	result := make([]types.Map, 0, len(rows))
	for _, row := range rows {
		rowMap, _ := types.ToMap(row, "", 0)
		result = append(result, rowMap)
	}
	return result, nil
}

func GetActiveZoneInfo(ctx context.Context) ([]*table.ZoneInfoTable, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("zone_info").AndEqual("status", 1)
	dest := []*table.ZoneInfoTable{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// GetUserLibaoCode 查询礼包码信息
// 对应 PHP: UserinfoModel::getUserLibaoCode
func GetUserLibaoCode(ctx context.Context, code string) (types.Map, error) {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_libao_code").AndEqual("code", code).Limit(1)
	row, err := client.FindOneMap(ctx, stmt)
	if err != nil {
		if extorm.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}
	rowMap, _ := types.ToMap(row, "", 0)
	return rowMap, nil
}

// UpdateUserLibaoCode 更新礼包码使用状态
// 对应 PHP: UserinfoModel::useLibaoCode 中的 update 操作
func UpdateUserLibaoCode(ctx context.Context, code string, updateData types.Map) error {
	client := repo.UserDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_libao_code").
		UpdateMap(extorm.SetMap(updateData)).
		AndEqual("code", code)
	_, err := client.Update(ctx, stmt)
	return err
}
