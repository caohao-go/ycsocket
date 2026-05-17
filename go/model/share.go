package model

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/mem/content"
	"server_golang/repo/mem/daily"
	"server_golang/repo/table"
	"server_golang/repo/user"
)

// ---- 邀请/视频 ----

// GetInviteInfo 获取邀请信息
// 对应 PHP ShinelightModel::getInviteInfo — 含空值保护：若数据为空则初始化并插入
func GetInviteInfo(ctx context.Context, userID int64) types.Map {
	times := content.GetMap(userID, "invite_player")
	if len(times) == 0 {
		times = types.Map{"1": 0, "3": 0, "5": 0}
		content.SetMap(userID, "invite_player", times)
	}
	return times
}

// SetInvitePlayerStatus 设置邀请状态
func SetInvitePlayerStatus(ctx context.Context, userID int64, num int) error {
	inviteInfo := GetInviteInfo(ctx, userID)
	inviteInfo[types.ToString(num)] = 1
	content.SetMap(userID, "invite_player", inviteInfo)
	return nil
}

// ========================= 视频/分享 =========================

func GetVedioTimes(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyVedioTime)
	return types.ToIntE(v)
}

func IncrVedioTimes(ctx context.Context, uid int64) {
	daily.Incr(uid, config.DailyVedioTime, 1)
}

func GetShareTimes(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyShareTime)
	return types.ToIntE(v)
}

func IncrShareTimes(ctx context.Context, uid int64) {
	daily.Incr(uid, config.DailyShareTime, 1)
}

// GetVedioInfoFull 获取完整的视频信息
// 对应 PHP: ShinelightModel::getVedioInfo
func GetVedioInfoFull(ctx context.Context, uid int64) types.Map {
	times := GetVedioTimes(ctx, uid)
	status := GetVedioStatus(ctx, uid)

	statusMap := types.Map{
		"1": types.Map{"num": 1, "status": types.ToIntE(status["1"])},
		"3": types.Map{"num": 3, "status": types.ToIntE(status["3"])},
		"5": types.Map{"num": 5, "status": types.ToIntE(status["5"])},
		"7": types.Map{"num": 7, "status": types.ToIntE(status["7"])},
	}

	return types.Map{
		"times":  times,
		"status": statusMap,
	}
}

// ========================= 邀请记录 =========================

// GetMyInvitePlayers 查询我邀请的所有玩家
func GetMyInvitePlayers(ctx context.Context, inviteZoneUID int64) ([]*table.InviteInfo, error) {
	return user.GetMyInvitePlayers(ctx, inviteZoneUID)
}

// GetInviteByUserID 查询是否已被邀请过
func GetInviteByUserID(ctx context.Context, userID int64) (*table.InviteInfo, error) {
	return user.GetInviteByUserID(ctx, userID)
}

// InsertInviteInfo 插入邀请记录
func InsertInviteInfo(ctx context.Context, data *table.InviteInfo) (int64, error) {
	return user.InsertInviteInfo(ctx, data)
}

// ========================= 视频/分享 =========================

func GetVedioStatus(ctx context.Context, uid int64) types.Map {
	v, _ := repo.RedisHGetAll(ctx, fmt.Sprintf(config.KeyVedioStatus, uid, util.DateYmd()))
	return v
}

// SetVedioStatus 设置视频领取状态
func SetVedioStatus(ctx context.Context, uid int64, num int) {
	k := fmt.Sprintf(config.KeyVedioStatus, uid, util.DateYmd())
	repo.RedisHSet(ctx, k, num, "1")
	repo.RedisExpire(ctx, k, 86400)
}

// GetShareTime 获取分享时间戳
func GetShareTime(ctx context.Context, uid int64) int {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyShareTimeRefresh, uid, util.DateYmd()))
	return types.ToIntE(v)
}

// SetShareTime 设置分享时间（3分钟后）
func SetShareTime(ctx context.Context, uid int64) {
	k := fmt.Sprintf(config.KeyShareTimeRefresh, uid, util.DateYmd())
	repo.RedisSet(ctx, k, time.Now().Unix()+180, 180)
}
