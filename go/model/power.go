// 体力系统模块 - 体力恢复、消耗、加体力
package model

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
)

// 体力类型常量
const (
	PowerType10024             = 10024
	PowerType10040             = 10040
	PowerType10030             = 10030
	PowerTypePK                = 9999
	PowerTypeVoyage            = 3333
	PowerTypeBaseZhaohuanQuan  = 5231
	PowerTypeHighZhaohuanQuan  = 5232
	PowerTypeBaseTanbaoRefresh = 54311
	PowerTypeHighTanbaoRefresh = 66443
)

// 最大体力配置
var ConfMaxPower = map[int]int{
	PowerType10024:             1,
	PowerType10030:             3,
	PowerType10040:             5,
	PowerTypePK:                5,
	PowerTypeVoyage:            22000,
	PowerTypeBaseZhaohuanQuan:  1,
	PowerTypeHighZhaohuanQuan:  1,
	PowerTypeBaseTanbaoRefresh: 1,
	PowerTypeHighTanbaoRefresh: 1,
}

// 体力恢复时间（秒）
var ConfRecoverTime = map[int]int{
	PowerType10024:             10800,
	PowerType10030:             28800,
	PowerType10040:             7200,
	PowerTypePK:                120,
	PowerTypeVoyage:            4,
	PowerTypeBaseZhaohuanQuan:  21600,
	PowerTypeHighZhaohuanQuan:  21600,
	PowerTypeBaseTanbaoRefresh: 10800,
	PowerTypeHighTanbaoRefresh: 10800,
}

// GetUserPower 获取用户体力值
func GetUserPower(ctx context.Context, userID int64, powerType int, needTime *int) int {
	*needTime = 0
	userPowerInfo := GetRedisUserPower(ctx, userID, powerType)
	now := int(time.Now().Unix())
	return resetPower(ctx, userID, powerType, userPowerInfo, &now, needTime)
}

// AddUserPower 给用户加体力
func AddUserPower(ctx context.Context, userID int64, powerType int, needTime *int, addNum int) int {
	*needTime = 0
	userPowerInfo := GetRedisUserPower(ctx, userID, powerType)
	if userPowerInfo != nil {
		maxPower := ConfMaxPower[powerType]
		power := types.ToIntE(userPowerInfo["power"]) + addNum
		if power > maxPower {
			power = maxPower
		}
		userPowerInfo["power"] = power
	}
	now := int(time.Now().Unix())
	return resetPower(ctx, userID, powerType, userPowerInfo, &now, needTime)
}

// SubUserPower 使用体力
func SubUserPower(ctx context.Context, userID int64, powerType int, needTime *int, subNum int) int {
	*needTime = 0
	userPowerInfo := GetRedisUserPower(ctx, userID, powerType)
	now := int(time.Now().Unix())
	currentPower := resetPower(ctx, userID, powerType, userPowerInfo, &now, needTime)
	if currentPower-subNum < 0 {
		return -1
	}
	userPowerInfo["last_incr_time"] = now
	userPowerInfo["power"] = currentPower - subNum
	return resetPower(ctx, userID, powerType, userPowerInfo, &now, needTime)
}

// resetPower 重置体力（内部）
func resetPower(ctx context.Context, userID int64, powerType int, userPowerInfo types.Map, currentTime *int, needTime *int) int {
	maxPower := ConfMaxPower[powerType]
	recoverTime := ConfRecoverTime[powerType]

	data := types.Map{
		"power":          maxPower,
		"last_incr_time": *currentTime,
	}

	if userPowerInfo != nil && types.ToIntE(userPowerInfo["power"]) < maxPower {
		elapsed := *currentTime - types.ToIntE(userPowerInfo["last_incr_time"])
		addPowerNum := elapsed / recoverTime
		currentPower := types.ToIntE(userPowerInfo["power"]) + addPowerNum

		if currentPower < maxPower {
			*needTime = recoverTime - ((*currentTime - types.ToIntE(userPowerInfo["last_incr_time"])) % recoverTime)
			*currentTime = *currentTime - (elapsed % recoverTime)
			data["power"] = currentPower
			data["last_incr_time"] = *currentTime
		}
	}

	expireTime := maxPower * recoverTime * 2
	SetRedisUserPower(ctx, userID, powerType, data, expireTime)
	return types.ToIntE(data["power"])
}

// ========================= 用户体力 =========================

func GetRedisUserPower(ctx context.Context, uid int64, typ int) types.Map {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyPreRedisUserPower, typ, uid))
	if v == "" {
		return types.Map{}
	}
	ret := types.ToMapE(v)
	if ret == nil {
		return types.Map{}
	}
	return ret
}

func SetRedisUserPower(ctx context.Context, uid int64, typ int, data interface{}, expire int) {
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyPreRedisUserPower, typ, uid), data, expire)
}
