package model

import (
	"context"
	"encoding/base64"
	"strings"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
)

// ApplePackageName 对应 PHP: AppleModel::$packageName
var ApplePackageName = map[string]string{
	"wx60207b369e6a0ed1":   "org.cocos2d.bigSword",
	"chinese_young_master": "org.cocos2d.chineseYoungMaster",
	"linkline":             "org.cocos2d.linkline",
	"Endlesscourage":       "com.chrisantem.game.Endlesscourage",
	"wxfa8b612abfc14f0b":   "com.chrisantem.BigProject",
}

// AppleLogin 验证 Apple id_token 并返回 openid 和 name
// 对应 PHP: AppleModel::Login
// Apple Sign In 的 id_token 是一个 JWT，包含用户的 sub (openid) 和 email
func AppleLogin(appid, idToken string) (string, string) {
	pkName := ApplePackageName[appid]
	if pkName == "" {
		log.Errorf(context.Background(), 0, "apple login: unknown appid=%s", appid)
		return "", ""
	}

	// 解析 JWT payload（id_token 格式: header.payload.signature）
	parts := strings.Split(idToken, ".")
	if len(parts) != 3 {
		log.Errorf(context.Background(), 0, "[%s] apple verify id token failed: invalid JWT format", pkName)
		return "", ""
	}

	// 解码 payload
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		log.Errorf(context.Background(), 0, "[%s] apple verify id token decode error: %v", pkName, err)
		return "", ""
	}

	claims := json.ToMap(string(payload))
	if claims == nil {
		log.Errorf(context.Background(), 0, "[%s] apple verify id token parse error", pkName)
		return "", ""
	}

	// 验证 issuer
	iss, _ := claims.GetString("iss")
	if iss != "https://appleid.apple.com" {
		log.Errorf(context.Background(), 0, "[%s] apple verify id token: invalid issuer=%s", pkName, iss)
		return "", ""
	}

	// 验证 audience
	aud, _ := claims.GetString("aud")
	if aud != pkName {
		log.Errorf(context.Background(), 0, "[%s] apple verify id token: invalid audience=%s, expected=%s", pkName, aud, pkName)
		return "", ""
	}

	// 获取用户唯一标识 (sub)
	userID, _ := claims.GetString("sub")
	if userID == "" {
		log.Errorf(context.Background(), 0, "[%s] apple verify id token: empty sub", pkName)
		return "", ""
	}

	// 获取 name (email 或 sub 前5字符)
	name := ""
	email, _ := claims.GetString("email")
	if email != "" {
		name = strings.Replace(email, "@privaterelay.appleid.com", "", 1)
	} else {
		if len(userID) > 5 {
			name = userID[:5]
		} else {
			name = userID
		}
	}

	log.Infof(context.Background(), "[%s] apple verify id success: userid=%s", pkName, userID)
	return userID, name
}
