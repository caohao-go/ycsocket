package model

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/mem/content"
	"server_golang/repo/mem/daily"
)

// GetUserVipContents 获取用户VIP内容（解析JSON为Map）
func GetUserVipContents(userID int64) types.Map {
	return content.GetMap(userID, "user_vip")
}

// UpdateUserVipContents 更新用户VIP内容
func UpdateUserVipContents(userID int64, tmp types.Map) {
	content.SetMap(userID, "user_vip", tmp)
}

// ========================= 充值数据 =========================

// GetVipContentsCurWeek 获取VIP当周内容
func GetVipContentsCurWeek(ctx context.Context, uid int64) types.Map {
	k := fmt.Sprintf(config.DWVipContentW, util.DateW(), uid)
	v, _ := repo.RedisGet(ctx, k)
	if v == "" {
		content := types.Map{"xianzhi": types.Map{"limit": 0}}
		repo.RedisSet(ctx, k, content, 7*86400)
		return content
	}
	ret := types.ToMapE(v)
	return ret
}

func SetVipContentsCurWeek(ctx context.Context, uid int64, content interface{}) {
	repo.RedisSet(ctx, fmt.Sprintf(config.DWVipContentW, util.DateW(), uid), content, 7*86400)
}

// GetVipContentsCurMon 获取当月 VIP 内容
func GetVipContentsCurMon(ctx context.Context, uid int64) types.Map {
	k := fmt.Sprintf(config.KeyVipContentM, uid)
	v, _ := repo.RedisGet(ctx, k)
	if v == "" {
		content := types.Map{"kuaisu": types.Map{"limit": 0, "expire": 0}}
		repo.RedisSet(ctx, k, content, 31*86400)
		return content
	}
	ret := types.ToMapE(v)
	return ret
}

func SetVipContentsCurMon(ctx context.Context, uid int64, content interface{}) {
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyVipContentM, uid), content, 31*86400)
}

// GetRongyaoCardExpire 获取月卡到期时间
func GetRongyaoCardExpire(ctx context.Context, uid int64, typ int) int {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyMonCardExpire, typ, uid))
	return types.ToIntE(v)
}

func SetRongyaoCardExpire(ctx context.Context, uid int64, typ int) {
	k := fmt.Sprintf(config.KeyMonCardExpire, typ, uid)
	repo.RedisSet(ctx, k, time.Now().Unix()+30*86400, 30*86400)
}

// 限时充值 + 成就任务处理

// ========================= 限时充值 =========================

// GetVipContentsLimitWeek 获取VIP周限购内容
func GetVipContentsLimitWeek(ctx context.Context, uid int64, zhouqi int) types.Map {
	k := fmt.Sprintf(config.KeyVipContentsLimitW, zhouqi, uid)
	v, _ := repo.RedisGet(ctx, k)
	if v == "" {
		c := types.Map{"week_libao": types.Map{"limit": types.Map{}}, "qianggou": types.Map{"limit": types.Map{}}}
		repo.RedisSet(ctx, k, c, 7*86400)
		return c
	}
	ret := types.ToMapE(v)
	return ret
}

func SetVipContentsLimitWeek(ctx context.Context, uid int64, zhouqi int, content interface{}) {
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyVipContentsLimitW, zhouqi, uid), content, 7*86400)
}

// GetVipContentsLimitMon 获取VIP月限购内容
func GetVipContentsLimitMon(ctx context.Context, uid int64, zhouqi int) types.Map {
	k := fmt.Sprintf(config.KeyVipContentsLimitM, zhouqi, uid)
	v, _ := repo.RedisGet(ctx, k)
	if v == "" {
		c := types.Map{"yuedu_libao": types.Map{"limit": types.Map{}}}
		repo.RedisSet(ctx, k, c, 32*86400)
		return c
	}
	ret := types.ToMapE(v)
	return ret
}

func SetVipContentsLimitMon(ctx context.Context, uid int64, zhouqi int, content interface{}) {
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyVipContentsLimitM, zhouqi, uid), content, 32*86400)
}

// ========================= 充值/VIP =========================

func GetVipContentsCurDay(ctx context.Context, uid int64) types.Map {
	v, _ := daily.Get(uid, config.DailyVipContentsDay)
	if v == nil {
		content := types.Map{
			"leiji_xiaofei":     0,
			"leiji_chong":       0,
			"chong_180_ling":    0,
			"day_libao":         types.Map{"limit": types.Map{}},
			"leiji_chong_libao": types.Map{"lingqu": types.Map{}}}
		daily.Set(uid, config.DailyVipContentsDay, json.Marshal(content))
		return content
	}
	s := types.ToString(v)
	if s == "" {
		content := types.Map{
			"leiji_xiaofei":     0,
			"leiji_chong":       0,
			"chong_180_ling":    0,
			"day_libao":         types.Map{"limit": types.Map{}},
			"leiji_chong_libao": types.Map{"lingqu": types.Map{}}}
		daily.Set(uid, config.DailyVipContentsDay, json.Marshal(content))
		return content
	}
	ret := types.ToMapE(s)
	return ret
}

// SetVipContentsCurDay 设置VIP当天内容
func SetVipContentsCurDay(ctx context.Context, uid int64, content interface{}) {
	daily.Set(uid, config.DailyVipContentsDay, json.Marshal(content))
}

// GetRongyaoCardAmt 获取月卡当日累计充值金额
func GetRongyaoCardAmt(ctx context.Context, uid int64, typ int) int {
	v, _ := daily.GetByPrefix(uid, config.DailyMonCardAmt, typ)
	return types.ToIntE(v)
}

func IncrRongyaoCardAmt(ctx context.Context, uid int64, typ, amt int) int64 {
	return daily.IncrByPrefix(uid, config.DailyMonCardAmt, typ, int64(amt))
}

// GetRongyaoCardLingqu 获取月卡领取状态
func GetRongyaoCardLingqu(ctx context.Context, uid int64, typ int) int {
	v, _ := daily.GetByPrefix(uid, config.DailyMonCardLingqu, typ)
	return types.ToIntE(v)
}

// SetRongyaoCardLingqu 设置月卡已领取
func SetRongyaoCardLingqu(ctx context.Context, uid int64, typ int) {
	daily.SetByPrefix(uid, config.DailyMonCardLingqu, typ, "1")
}

// GetVipZhizunLingqu 获取至尊月卡今日领取状态
func GetVipZhizunLingqu(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyVipZhizunLingqu)
	return types.ToIntE(v)
}

// SetVipZhizunLingqu 设置至尊月卡今日已领取
func SetVipZhizunLingqu(ctx context.Context, uid int64) {
	daily.Set(uid, config.DailyVipZhizunLingqu, "2")
}

func GetVipBuyCount(ctx context.Context, uid int64, vipLv int) int {
	v, _ := daily.GetByPrefix(uid, config.DailyFastClimbCnt, vipLv)
	return types.ToIntE(v)
}

func IncrVipBuyCount(ctx context.Context, uid int64, vipLv int) {
	daily.IncrByPrefix(uid, config.DailyFastClimbCnt, vipLv, 1)
}
