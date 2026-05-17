package model

import (
	"context"
	"fmt"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

// ---- 好友系统 ----

// GetUserFriendPair 获取好友关系（双向查找，与 PHP getUserFriendPair 一致）
func GetUserFriendPair(ctx context.Context, user1, user2 int64) *table.UsersFriends {
	return world.GetUserFriendPairOne(ctx, user1, user2)
}

// InsertUserFriendPair 插入好友关系
func InsertUserFriendPair(ctx context.Context, user1, user2 int64, status int) (int64, error) {
	return world.InsertUserFriend(ctx, &table.UsersFriends{User1: user1, User2: user2, Status: status})
}

// UpdateUserFriendPair 更新好友关系（双向更新，与 PHP updateUserFriendPair 一致）
func UpdateUserFriendPair(ctx context.Context, user1, user2 int64, status int) error {
	return world.UpdateUserFriendPair(ctx, user1, user2, types.Map{"status": status})
}

// DelUserFriendPair 删除好友关系（双向删除，与 PHP delUserFriendPair 一致）
func DelUserFriendPair(ctx context.Context, user1, user2 int64) error {
	return world.DeleteUserFriendPair(ctx, user1, user2)
}

// GetFriendsList 获取好友UID列表（与 PHP getFriendsList 返回值一致，返回 uid 数组）
// PHP 版本返回 $ret[] 即用户ID数组，status 默认为 0（已确认好友）
func GetFriendsList(ctx context.Context, userID int64, status int) []int64 {
	return world.GetFriendsUIDList(ctx, userID, status)
}

// GetApplyFriendsList 获取好友申请列表（status=1 待处理，与 PHP getApplyFriendsList 一致）
func GetApplyFriendsList(ctx context.Context, userID int64) ([]*table.UsersFriends, error) {
	return world.GetApplyFriendsList(ctx, userID)
}

// ========================= 社交/好友 =========================

// GetLoversList 获取今日送心列表
func GetLoversList(ctx context.Context, uid int64) types.Map {
	v, _ := repo.RedisHGetAll(ctx, fmt.Sprintf(config.KeyLovers, util.DateYmd(), uid))
	return v
}

// GetLoversReceiveList 获取收到的爱心列表
func GetLoversReceiveList(ctx context.Context, uid int64) types.Map {
	v, _ := repo.RedisHGetAll(ctx, fmt.Sprintf(config.KeyReceiveLovers, util.DateYmd(), uid))
	return v
}

// SetLovers 赠送爱心
func SetLovers(ctx context.Context, uid, friendUID int64) {
	k1 := fmt.Sprintf(config.KeyLovers, util.DateYmd(), uid)
	repo.RedisHSet(ctx, k1, friendUID, "1")
	repo.RedisExpire(ctx, k1, 86400)
	k2 := fmt.Sprintf(config.KeyReceiveLovers, util.DateYmd(), friendUID)
	repo.RedisHSet(ctx, k2, uid, "1")
	repo.RedisExpire(ctx, k2, 86400)
}

// LingquLovers 领取爱心
func LingquLovers(ctx context.Context, uid, friendUID int64) {
	repo.RedisHDel(ctx, fmt.Sprintf(config.KeyReceiveLovers, util.DateYmd(), uid), friendUID)
}
