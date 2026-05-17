package world

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/cache"
	"server_golang/repo/table"
)

// GetUserPosition 获取用户阵位
func GetUserPosition(ctx context.Context, userID int64, posType int) (types.Map, error) {
	cacheKey := fmt.Sprintf(config.CacheUserPosition, userID, posType)
	val, ok := cache.Get(cacheKey)
	if ok {
		if val == config.EmptyString {
			return types.Map{}, nil
		}

		ret := types.Map{}
		_ = json.Unmarshal(val, &ret)
		return ret, nil
	}

	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_position").AndEqual("user_id", userID).AndEqual("pos_type", posType).Limit(1)

	dest := table.UserPosition{}
	err := client.FindOne(ctx, stmt, &dest)
	if err != nil {
		return types.Map{}, err
	}

	if dest.Id == 0 {
		cache.SetWithTTL(cacheKey, config.EmptyString, 1800)
	} else {
		cache.SetWithTTL(cacheKey, dest, 1800)
	}
	return types.ObjectToMap(dest), nil
}

func GetUserPositionByUserId(ctx context.Context, userId int64) ([]*table.UserPosition, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_position").AndEqual("user_id", userId)
	dest := []*table.UserPosition{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

func ReplaceUserPosition(ctx context.Context, data *table.UserPosition) error {
	cacheKey := fmt.Sprintf(config.CacheUserPosition, data.UserId, data.PosType)
	cache.SetWithTTL(cacheKey, data, 1800)

	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_position").ReplaceStruct(data)
	_, err := client.Replace(ctx, stmt)
	return err
}
