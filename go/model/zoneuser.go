package model

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/table"
	"server_golang/repo/user"
	"server_golang/repo/world"
)

// GetOrInitUserGradeWithInit 获取/初始化用户等级信息（带 ShinelightModel 的初始化逻辑）
func GetOrInitUserGradeWithInit(ctx context.Context, userID int64) types.Map {
	userInfo := GetUserAttr(userID)
	if userInfo == nil {
		return nil
	}
	// 如果是新用户，执行 ShinelightModel 特有的初始化
	if userInfo.GetIntE(config.AttrInitFlag) == 0 {
		InitContents(ctx, userID)
		InitContentsInt(ctx, userID)
	}
	return userInfo
}

// GetUserInfoByZoneUserID 获取用户信息（带 ristretto 本地缓存）
// 对应 PHP: UserinfoModel::getUserinfoByUserid
// 包含从 nickinfo 覆盖 nickname/avatar_url/gender 的逻辑
func GetUserInfoByZoneUserID(ctx context.Context, zoneUserID int64) *table.UserInfo {
	userId := GetUIDByZoneUserID(zoneUserID)
	return GetUserInfoByUserID(ctx, userId)
}

// GetLoginZone 查询用户登录过的区服记录
// 对应 PHP: UserinfoModel::getLoginZone
func GetLoginZone(ctx context.Context, userID int64) []types.Map {
	zones, err := user.GetLoginZone(ctx, userID)
	if err != nil || zones == nil {
		return []types.Map{}
	}
	return zones
}

// ReplaceLoginZone 记录用户登录区服信息
func ReplaceLoginZone(zoneUserID int64, lv int) {
	userID := GetUIDByZoneUserID(zoneUserID)
	zoneID := GetUserZoneID(zoneUserID)
	go func() {
		user.ReplaceUserLoginZones(context.Background(), types.Map{
			"user_id": userID,
			"zone":    zoneID,
			"lv":      lv,
		})
	}()
}

// GetUsersWithDetail 批量获取用户详情 + user_grade 表额外字段
// 对齐 PHP CoreModel::getMuiltiUsersDetail($uids, $users_grade_key, $pos_type) 的完整逻辑
// usersGradeKey: 需要额外获取的 user_grade 字段（如 lv, fight_point, vip_level, off_time）
// posType: 取战斗力时的阵位类型，1=剧情 2=竞技场
func GetUsersWithDetail(ctx context.Context, zoneUserIds []int64, posType int, usersGradeKey ...string) map[int64]types.Map {
	if len(zoneUserIds) == 0 {
		return make(map[int64]types.Map)
	}

	// 1. 获取基础用户信息（nickname, avatar_url）
	userGrades := GetMultiUserAttr(zoneUserIds)

	result := make(map[int64]types.Map, len(userGrades))
	for zoneUserId, v := range userGrades {
		result[zoneUserId] = types.Map{
			"user_id":    zoneUserId,
			"nickname":   v.GetStringE("nickname"),
			"avatar_url": v.GetStringE("avatar_url"),
			"gender":     v.GetIntE("gender"),
		}
	}

	// 2. 如果需要额外字段，从 user_grade 表批量查询
	if len(usersGradeKey) > 0 {
		for _, uid := range zoneUserIds {
			info, ok := result[uid]
			if !ok {
				continue
			}

			userGrade, ok2 := userGrades[uid]
			if !ok2 {
				continue
			}

			for _, key := range usersGradeKey {
				if key == "fight_point" {
					if posType <= 0 {
						posType = 1
					}
					info[key] = GetUserFightPoint(ctx, uid, posType)
				} else {
					info[key] = userGrade.GetIntE(key)
				}
			}
			result[uid] = info
		}
	}

	return result
}

// GetUserByName 根据昵称模糊搜索用户（与 PHP getUserByName 一致，LIKE 模糊搜索 user_nickname 表）
func GetUserByName(ctx context.Context, nickname string) []*table.UserNickname {
	result, err := world.GetUsersByNicknameLike(ctx, nickname)
	if err != nil {
		return nil
	}
	return result
}

// ReplaceUserNicknameRecord 插入用户昵称记录（写入 user_nickname 表）
func ReplaceUserNicknameRecord(ctx context.Context, userID int64, nickname string) (int64, error) {
	return world.ReplaceUserNickname(ctx, &table.UserNickname{
		UserId:   userID,
		Nickname: nickname,
	})
}

// SetNicknameSame 记录昵称去重
func SetNicknameSame(ctx context.Context, nickname string) {
	h := md5.Sum([]byte(strings.TrimSpace(nickname)))
	key := hex.EncodeToString(h[:])
	repo.RedisSet(ctx, fmt.Sprintf("nickname_same:%s", key), 1, 0)
}

// GetNicknameSame 检查昵称是否已被使用
func GetNicknameSame(ctx context.Context, nickname string) bool {
	h := md5.Sum([]byte(strings.TrimSpace(nickname)))
	key := hex.EncodeToString(h[:])
	ok, _ := repo.RedisGet(ctx, fmt.Sprintf("nickname_same:%s", key))
	return ok == "1"
}

// GetUIDByZoneUserID 从区服用户ID中提取原始用户ID
func GetUIDByZoneUserID(zoneUserID int64) int64 {
	if zoneUserID <= 10000 {
		return zoneUserID
	}
	return zoneUserID / 1000
}

// GetUserZoneID 从区服用户ID中提取区服ID
func GetUserZoneID(zoneUserID int64) int {
	if zoneUserID <= 10000 {
		return 1
	}
	return int(zoneUserID % 1000)
}

// GetUserZoneUID 生成区服用户ID: userID * 1000 + zoneID
func GetUserZoneUID(userID int64, zoneID int) int64 {
	if zoneID <= 0 {
		zoneID = 1
	}
	if userID <= 10000 {
		return userID
	}
	return userID*1000 + int64(zoneID)
}

// ---- 称号系统 ----

// GetUserRoleTitles 获取用户称号列表
func GetUserRoleTitles(ctx context.Context, userID int64) []*table.UserRoleTitle {
	return world.GetUserRoleTitlesByUserId(ctx, userID)
}

// IsUserRoleExist 判断用户称号是否存在
func IsUserRoleExist(ctx context.Context, userID int64, roleID int) bool {
	return world.IsUserRoleTitleExist(ctx, userID, roleID)
}
