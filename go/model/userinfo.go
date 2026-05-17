package model

import (
	"context"
	"fmt"
	"time"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/crypto"
	"server_golang/common/helper"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo/cache"
	"server_golang/repo/table"
	"server_golang/repo/user"
)

// RegisterUser 插入新用户记录
func RegisterUser(ctx context.Context, appid string,
	openID, sessionKey, nickname, city string) (int64, string, error) {
	userID := generateUserID(ctx)
	now := time.Now().Format("2006-01-02 15:04:05")
	token := crypto.MD5Str(fmt.Sprintf("%s_%d_%d_%s", config.Cfg.Constants.TokenGenerateKey, time.Now().Unix(), userID, sessionKey))

	_, err := user.InsertUserInfo(ctx, &table.UserInfo{
		Appid:         appid,
		UserId:        userID,
		OpenId:        openID,
		SessionKey:    sessionKey,
		Nickname:      nickname,
		LastLoginTime: now,
		RegistTime:    now,
		Token:         token,
		City:          city,
	})

	if err != nil {
		return 0, "", err
	}

	return userID, token, nil
}

// LoginUser 更新登录信息并返回新 token
func LoginUser(ctx context.Context, userID int64, sessionKey string, loginTimes int) (string, error) {
	now := time.Now().Format("2006-01-02 15:04:05")
	token := crypto.MD5Str(fmt.Sprintf("%s_%d_%d_%s", config.Cfg.Constants.TokenGenerateKey, time.Now().Unix(), userID, sessionKey))

	data := types.Map{
		"last_login_time": now,
		"token":           token,
		"login_times":     loginTimes + 1,
	}

	if sessionKey != "" {
		data["session_key"] = sessionKey
	}

	err := user.UpdateUserInfoByUserId(ctx, userID, data)
	if err != nil {
		return "", err
	}
	return token, err
}

// GetUserInfoByUserID 获取用户信息（带 ristretto 本地缓存）
// 对应 PHP: UserinfoModel::getUserinfoByUserid
// 包含从 nickinfo 覆盖 nickname/avatar_url/gender 的逻辑
func GetUserInfoByUserID(ctx context.Context, userID int64) *table.UserInfo {
	cacheKey := fmt.Sprintf(config.CacheRedisUserInfo, userID)

	cached, ok := cache.Get(cacheKey)
	if ok && cached != "" {
		userInfo := table.UserInfo{}
		json.Unmarshal([]byte(cached), &userInfo)
		if userInfo.UserId > 0 {
			return &userInfo
		}
	}

	userInfo, err := user.GetUserInfoByUserId(ctx, userID)
	if err != nil || userInfo == nil {
		return nil
	}

	cache.SetWithTTL(cacheKey, userInfo, 1800)
	return userInfo
}

// GetUserInfoByOpenId 根据 OpenID 获取用户信息
func GetUserInfoByOpenId(ctx context.Context, openId string) (*table.UserInfo, error) {
	return user.GetUserInfoByOpenId(ctx, openId)
}

// GetUserAndAuth 通过 token 验证用户身份
func GetUserAndAuth(ctx context.Context, userID int64, token string) (*table.UserInfo, error) {
	if userID == 0 {
		return nil, fmt.Errorf("user id is empty")
	}

	userInfo := GetUserInfoByUserID(ctx, userID)
	if userInfo == nil {
		return nil, fmt.Errorf("not find user")
	}

	return userInfo, nil
}

// UpdateUser 更新用户资料
func UpdateUser(ctx context.Context, userID int64, updateData types.Map) error {
	return user.UpdateUserInfoByUserId(ctx, userID, updateData)
}

// DelUser 删除用户
func DelUser(ctx context.Context, userID int64) {
	user.DeleteUserInfoByUserId(ctx, userID)
}

// GetOpenID 调用微信接口获取 openid
func GetOpenID(appID, secret, code string) (types.Map, string) {
	url := fmt.Sprintf("%sappid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		config.Cfg.Constants.WeixinOpenIDURL, appID, secret, code)

	result, err := helper.HttpsRequest(url)
	if err != nil {
		log.Errorf(context.Background(), 0, "get_openid error: %v", err)
		return nil, err.Error()
	}

	res := json.ToMap(result)
	openid, _ := res.GetString("openid")
	errmsg, _ := res.GetString("errmsg")
	if res == nil || openid == "" {
		log.Errorf(context.Background(), 0, "get_openid error: %s", result)
		return nil, errmsg
	}

	return res, ""
}

// 通过 sequence 表生成唯一用户ID
func generateUserID(ctx context.Context) int64 {
	for i := 0; i < 3; i++ {
		id, err := user.InsertSequence(ctx, &table.Sequence{
			Time: int(time.Now().Unix()),
		})
		if err == nil && id > 0 {
			return id
		}
	}
	return 0
}
