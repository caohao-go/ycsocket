package world

import (
	"context"

	"server_golang/common/types"
	"server_golang/repo"
	"server_golang/repo/table"
)

// 获取最近活跃的100个用户ID
func GetFriendsRecommendUserIds(ctx context.Context) ([]int64, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_grade").Order("off_time", true).Limit(100).Select("user_id")
	dest := []int64{}
	err := client.FindAllSingleCol(ctx, stmt, &dest)
	return dest, err
}

// ==================== users_friends ====================

// GetUserFriendPairOne 获取好友关系（双向查找，与 PHP getUserFriendPair 一致）
func GetUserFriendPairOne(ctx context.Context, user1, user2 int64) *table.UsersFriends {
	client := repo.WorldDB()

	// 先查 user1=user1, user2=user2
	stmt := repo.NewStmt()
	stmt.SetTableName("users_friends").AndEqual("user1", user1).AndEqual("user2", user2).Limit(1)
	dest := &table.UsersFriends{}
	err := client.FindOne(ctx, stmt, dest)
	if err == nil {
		return dest
	}

	// 再查 user1=user2, user2=user1（与 PHP 保持一致的双向查找）
	stmt2 := repo.NewStmt()
	stmt2.SetTableName("users_friends").AndEqual("user1", user2).AndEqual("user2", user1).Limit(1)
	dest2 := &table.UsersFriends{}
	err = client.FindOne(ctx, stmt2, dest2)
	if err == nil {
		return dest2
	}

	return nil
}

// GetFriendsList 获取好友列表（查 user1 或 user2 为当前用户的记录）
func GetFriendsList(ctx context.Context, userID int64, status int) ([]*table.UsersFriends, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("users_friends").AndSQL("(user1=? OR user2=?)", userID, userID).AndEqual("status", status)
	dest := []*table.UsersFriends{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// GetFriendsUIDList 获取好友UID列表（与 PHP getFriendsList 返回值一致，返回 uid 数组）
func GetFriendsUIDList(ctx context.Context, userID int64, status int) []int64 {
	friends, err := GetFriendsList(ctx, userID, status)
	if err != nil || len(friends) == 0 {
		return nil
	}
	ret := make([]int64, 0, len(friends))
	for _, f := range friends {
		if f.User1 == userID {
			ret = append(ret, f.User2)
		} else {
			ret = append(ret, f.User1)
		}
	}
	return ret
}

// GetApplyFriendsList 获取好友申请列表（status=1 待处理，与 PHP getApplyFriendsList 一致）
func GetApplyFriendsList(ctx context.Context, userID int64) ([]*table.UsersFriends, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("users_friends").AndEqual("user2", userID).AndEqual("status", 1)
	dest := []*table.UsersFriends{}
	err := client.FindAll(ctx, stmt, &dest)
	return dest, err
}

// InsertUserFriend 插入好友关系
func InsertUserFriend(ctx context.Context, data *table.UsersFriends) (int64, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("users_friends").InsertStruct(data)
	return client.Insert(ctx, stmt)
}

// UpdateUserFriendPair 更新好友关系（双向更新，与 PHP updateUserFriendPair 一致）
func UpdateUserFriendPair(ctx context.Context, user1, user2 int64, data types.Map) error {
	client := repo.WorldDB()
	// PHP: update users_friends set status={$status} where (user1={$user1} and user2={$user2}) or (user1={$user2} and user2={$user1})
	sql := "UPDATE users_friends SET status=? WHERE (user1=? AND user2=?) OR (user1=? AND user2=?)"
	_, err := client.Exec(ctx, sql, data["status"], user1, user2, user2, user1)
	return err
}

// DeleteUserFriendPair 删除好友关系（双向删除，与 PHP delUserFriendPair 一致）
func DeleteUserFriendPair(ctx context.Context, user1, user2 int64) error {
	client := repo.WorldDB()
	// PHP: delete from users_friends where (user1={$user1} and user2={$user2}) or (user1={$user2} and user2={$user1})
	sql := "DELETE FROM users_friends WHERE (user1=? AND user2=?) OR (user1=? AND user2=?)"
	_, err := client.Exec(ctx, sql, user1, user2, user2, user1)
	return err
}

// GetUsersByNicknameLike 根据昵称模糊搜索用户（与 PHP getUserByName 一致，查 user_nickname 表）
func GetUsersByNicknameLike(ctx context.Context, nickname string) ([]*table.UserNickname, error) {
	client := repo.WorldDB()
	sql := "SELECT user_id, nickname FROM user_nickname WHERE nickname LIKE ?"

	ret := []*table.UserNickname{}
	err := client.QueryToStructs(ctx, &ret, sql, "%"+nickname+"%")
	if err != nil {
		return nil, err
	}
	return ret, nil
}

func ReplaceUserNickname(ctx context.Context, data *table.UserNickname) (int64, error) {
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_nickname").ReplaceStruct(data)
	return client.Replace(ctx, stmt)
}
