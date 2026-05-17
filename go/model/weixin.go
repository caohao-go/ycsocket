// 微信相关接口（对应 PHP WeixinModel）
package model

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/helper"
	"server_golang/common/json"
)

// accessTokenCache 缓存 access_token（对应 PHP RedisPool::instance("userinfo")->get("wxfy_access_token".$appId)）
var accessTokenCache sync.Map

type accessTokenEntry struct {
	Token    string
	ExpireAt time.Time
}

// GetAccessToken 获取微信 access_token（对应 PHP WeixinModel::getAccessToken）
// 带内存缓存，过期后重新获取
func GetAccessToken(ctx context.Context, appID, secret string) (string, error) {
	cacheKey := "wxfy_access_token" + appID

	// 先从缓存取
	if v, ok := accessTokenCache.Load(cacheKey); ok {
		entry := v.(*accessTokenEntry)
		if time.Now().Before(entry.ExpireAt) {
			return entry.Token, nil
		}
	}

	// 请求微信接口
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, secret)
	result, err := helper.HttpsRequest(url)
	if err != nil {
		log.Errorf(ctx, 0, "GetAccessToken request error: %v", err)
		return "", err
	}

	res := json.ToMap(result)
	accessToken := res.GetStringE("access_token")
	if accessToken == "" {
		log.Errorf(ctx, 0, "get access token error, appid=%s, result=%s", appID, result)
		return "", fmt.Errorf("get access token error, appid=%s", appID)
	}

	// 缓存 1 小时（与 PHP expire 3600 一致）
	accessTokenCache.Store(cacheKey, &accessTokenEntry{
		Token:    accessToken,
		ExpireAt: time.Now().Add(3600 * time.Second),
	})

	return accessToken, nil
}

// MsgSecCheck 微信内容安全检查（对应 PHP nicknameSameAction 中的 msg_sec_check 调用）
// 返回 true 表示内容违规（对应 PHP errcode == 87014）
func MsgSecCheck(ctx context.Context, appID, secret, content string) bool {
	accessToken, err := GetAccessToken(ctx, appID, secret)
	if err != nil {
		// access_token 获取失败，不拦截（与 PHP 行为一致：获取失败时继续后续逻辑）
		return false
	}

	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/msg_sec_check?access_token=%s", accessToken)
	body := json.Marshal(map[string]string{"content": content})
	result, err := helper.HttpsPostJSON(url, body)
	if err != nil {
		return false
	}

	res := json.ToMap(result)
	errcode := res.GetIntE("errcode")
	return errcode == 87014
}
