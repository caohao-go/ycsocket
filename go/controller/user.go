package controller

import (
	"context"
	"fmt"
	"sort"
	"time"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/crypto"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/model"
)

// UserController 处理用户相关请求
type UserController struct {
	*BaseController
}

func dispatchUser(ctx context.Context, base *BaseController, action string) *Result {
	c := &UserController{BaseController: base}
	switch action {
	case "register":
		return c.Register(ctx)
	case "accountLogin":
		return c.AccountLogin(ctx)
	case "login":
		return c.Login(ctx)
	case "appleLogin":
		return c.AppleLogin(ctx)
	case "delUser":
		return c.DelUser(ctx)
	case "modifyInfo":
		return c.ModifyInfo(ctx)
	case "getInfo":
		return c.GetInfo(ctx)
	case "getOtherInfo":
		return c.GetOtherInfo(ctx)
	case "zoneInfo":
		return c.ZoneInfo(ctx)
	case "getZoneUser":
		return c.GetZoneUser(ctx)
	case "location":
		return c.Location(ctx)
	case "gameConf":
		return c.GameConf(ctx)
	case "getOpenGid":
		return c.GetOpenGid(ctx)
	default:
		return base.ResponseError(3, fmt.Sprintf("route error: user/%s", action))
	}
}

func (c *UserController) Register(ctx context.Context) *Result {
	appid := c.Params.GetStringE("appid")
	account := c.Params.GetStringE("account")
	password := c.Params.GetStringE("password")
	nickname := c.Params.GetStringE("nickname")
	city := c.Params.GetStringE("city")

	if appid == "" {
		return c.ResponseError(111, "appid can`t be empty")
	}
	if account == "" {
		return c.ResponseError(112, "account can`t be empty")
	}
	if nickname == "" {
		return c.ResponseError(113, "nickname can`t be empty")
	}
	if len(password) < 6 {
		return c.ResponseError(114, "password must more than 6 word")
	}

	existUser, _ := model.GetUserInfoByOpenId(ctx, account)
	if existUser != nil {
		return c.ResponseError(115, "account exists")
	}

	userID, token, _ := model.RegisterUser(ctx, appid, account, crypto.MD5Str(password), nickname, city)
	if token == "" {
		return c.ResponseError(116, "register failed, please try later")
	}

	return c.ResponseSuccessToMe(types.Map{
		"userid": userID,
		"token":  token,
	})
}

func (c *UserController) AccountLogin(ctx context.Context) *Result {
	appid := c.Params.GetStringE("appid")
	account := c.Params.GetStringE("account")
	password := c.Params.GetStringE("password")

	if appid == "" {
		return c.ResponseError(111, "appid can`t be empty")
	}
	if account == "" {
		return c.ResponseError(112, "account can`t be empty")
	}

	userInfo, _ := model.GetUserInfoByOpenId(ctx, account)
	if userInfo == nil {
		return c.ResponseError(121, "not find user "+account)
	}

	if crypto.MD5Str(password) != string(userInfo.SessionKey) {
		return c.ResponseError(122, "wrong password")
	}

	userID := int64(userInfo.UserId)
	loginTimes := userInfo.LoginTimes
	token, err := model.LoginUser(ctx, userID, "", loginTimes)
	if token == "" || err != nil {
		return c.ResponseError(123, "login failed, please try later")
	}

	return c.ResponseSuccessToMe(types.Map{
		"userid": userID,
		"token":  token,
		"is_new": 0,
	})
}

func (c *UserController) Login(ctx context.Context) *Result {
	appid := c.Params.GetStringE("appid")
	code := c.Params.GetStringE("code")

	wxCfg, ok := config.Cfg.WxApp[appid]
	if !ok || wxCfg.AppID == "" {
		return c.ResponseError(10000002, "access token config not find")
	}

	loginInfo, errMsg := model.GetOpenID(wxCfg.AppID, wxCfg.Secret, code)
	if loginInfo == nil {
		return c.ResponseError(10001006, errMsg)
	}

	openid := loginInfo.GetStringE("openid")
	if openid == "" {
		return c.ResponseError(10001006, "openid is empty")
	}

	result := types.Map{"openid": openid}

	user, _ := model.GetUserInfoByOpenId(ctx, openid)
	if user == nil {
		userID, token, _ := model.RegisterUser(ctx, appid, openid, loginInfo.GetStringE("session_key"), "", "")
		result["userid"] = userID
		result["token"] = token
		result["is_new"] = 1
	} else {
		token, _ := model.LoginUser(ctx, user.UserId, loginInfo.GetStringE("session_key"), user.LoginTimes)
		result["userid"] = user.UserId
		result["token"] = token
		result["is_new"] = 0
	}

	if result.GetStringE("token") == "" {
		return c.ResponseError(10000009, "登陆失败, 请重试")
	}

	return c.ResponseSuccessToMe(result)
}

func (c *UserController) AppleLogin(ctx context.Context) *Result {
	appid := c.Params.GetStringE("appid")
	idToken := c.Params.GetStringE("id_token")

	if appid == "" {
		return c.ResponseError(111, "appid can`t be empty")
	}

	// Apple 登录验证 — 通过 JWT 验证获取 openid
	// 对应 PHP: $this->apple_model->Login($appid, $id_token, $name)
	openid, name := model.AppleLogin(appid, idToken)
	if openid == "" {
		return c.ResponseError(10021032, "verify id token error")
	}

	result := types.Map{"openid": openid}

	user, _ := model.GetUserInfoByOpenId(ctx, openid)
	if user == nil {
		// 新用户注册
		userID, token, _ := model.RegisterUser(ctx, appid, openid, "apple_login", name, "")
		result["userid"] = userID
		result["token"] = token
		result["is_new"] = 1
	} else {
		// 老用户登录
		token, _ := model.LoginUser(ctx, user.UserId, "apple_login", user.LoginTimes)
		result["userid"] = user.UserId
		result["token"] = token
		result["is_new"] = 0
	}

	if result.GetStringE("token") == "" {
		return c.ResponseError(10000009, "登陆失败, 请重试")
	}

	return c.ResponseSuccessToMe(result)
}

func (c *UserController) DelUser(ctx context.Context) *Result {
	userID := c.Params.GetInt64E("userid")
	if userID == 0 {
		return c.ResponseError(10000021, "user_id is empty")
	}
	model.DelUser(ctx, userID)
	return c.ResponseSuccessToMe(types.Map{})
}

func (c *UserController) ModifyInfo(ctx context.Context) *Result {
	userID := c.Params.GetInt64E("userid")
	token := c.Params.GetStringE("token")
	rawData := c.Params.GetStringE("raw_data")

	if userID == 0 {
		return c.ResponseError(10000021, "user_id is empty")
	}
	if rawData == "" {
		return c.ResponseError(10000022, "rawData is empty")
	}

	userInfo := model.GetUserInfoByUserID(ctx, userID)
	if userInfo == nil {
		return c.ResponseError(10000023, "未找到该用户")
	}

	if token == "" || token != userInfo.Token {
		return c.ResponseError(10000024, "token 校验失败")
	}

	// 对应 PHP: $raw_data_array = json_decode($rawData, true)
	rawDataMap := json.ToMap(rawData)
	updateData := types.Map{
		"nickname":   rawDataMap.GetStringE("nick_name"),
		"gender":     rawDataMap.GetIntE("gender"),
		"city":       rawDataMap.GetStringE("city"),
		"avatar_url": rawDataMap.GetStringE("avatar_url"),
	}

	model.UpdateUser(ctx, userID, updateData)
	return c.ResponseSuccessToMe(types.Map{})
}

func (c *UserController) GetInfo(ctx context.Context) *Result {
	userID := c.Params.GetInt64E("userid")
	token := c.Params.GetStringE("token")

	if userID == 0 {
		return c.ResponseError(10000017, "user_id is empty")
	}
	if token == "" {
		return c.ResponseError(10000016, "token is empty")
	}

	userInfo := model.GetUserInfoByUserID(ctx, userID)
	if userInfo == nil {
		return c.ResponseError(10000023, "未找到该用户")
	}
	if token != userInfo.Token {
		return c.ResponseError(10000024, "token 校验失败")
	}

	// 对应 PHP: getInfoAction — 返回字段名与 PHP 保持一致（snake_case）
	return c.ResponseSuccessToMe(types.Map{
		"gender":     userInfo.Gender,
		"avatar_url": userInfo.AvatarUrl,
		"nickname":   userInfo.Nickname,
		"form_id":    userInfo.FormId,
		"user_id":    userID,
	})
}

func (c *UserController) GetOtherInfo(ctx context.Context) *Result {
	userID := c.Params.GetInt64E("userid")
	token := c.Params.GetStringE("token")
	// 对应 PHP: $toUserId = $this->params['to_user_id']
	toUserID := c.Params.GetInt64E("to_user_id")

	if userID == 0 {
		return c.ResponseError(10000017, "user_id is empty")
	}
	if toUserID == 0 {
		return c.ResponseError(10000059, "toUserId is empty")
	}
	if token == "" {
		return c.ResponseError(10000016, "token is empty")
	}

	myUserInfo := model.GetUserInfoByUserID(ctx, userID)
	if myUserInfo == nil {
		return c.ResponseError(10000023, "未找到用户信息")
	}
	if token != myUserInfo.Token {
		return c.ResponseError(10000024, "token 校验失败")
	}

	userInfo := model.GetUserInfoByUserID(ctx, toUserID)
	if userInfo == nil {
		return c.ResponseError(10000023, "未找到该用户")
	}

	// 对应 PHP: getOtherInfoAction — 返回字段名与 PHP 保持一致（snake_case）
	return c.ResponseSuccessToMe(types.Map{
		"gender":     userInfo.Gender,
		"avatar_url": userInfo.AvatarUrl,
		"nickname":   userInfo.Nickname,
		"form_id":    userInfo.FormId,
		"user_id":    userInfo.UserId,
	})
}

func (c *UserController) ZoneInfo(ctx context.Context) *Result {
	userID := c.Params.GetInt64E("userid")
	if userID == 0 {
		return c.ResponseError(10000021, "userid is empty")
	}

	zoneList := make(map[string][]types.Map)
	tmp := config.Cfg
	for _, zi := range tmp.Zone.ZoneInfo {
		zoneID := zi.ZoneID
		if zi.Hequ != 0 {
			zoneID = zi.Hequ
		}
		key := fmt.Sprintf("A%d", zoneID)
		openTimestamp := int64(0)
		if zi.Time != "" {
			if t, err := time.Parse("2006-01-02 15:04:05", zi.Time); err == nil {
				openTimestamp = t.Unix()
			}
		}
		zoneList[key] = append(zoneList[key], types.Map{
			"zone_id":        zi.ZoneID,
			"zone_name":      zi.ZoneName,
			"hequ":           zi.Hequ,
			"hot":            zi.Hot,
			"status":         zi.Status,
			"socket":         zi.Socket,
			"port":           zi.Port,
			"time":           zi.Time,
			"open_timestamp": openTimestamp,
		})
	}

	// 对应 PHP: zoneInfoAction 中 login_zone 逻辑
	// 获取用户登录过的区服记录，并附加 nickinfo（nickname/avatar_url/gender）
	loginZone := model.GetLoginZone(ctx, userID)
	for k := range loginZone {
		nickInfo := model.GetUserInfoByUserID(ctx, userID)
		loginZone[k]["nickname"] = nickInfo.Nickname
		loginZone[k]["avatar_url"] = nickInfo.AvatarUrl
		loginZone[k]["gender"] = nickInfo.Gender
	}

	// 对应 PHP: krsort($result) — 按 key 逆序排序 zone_list
	// 返回 map 格式（与 PHP 保持一致），客户端按 zone_list[key] 方式访问
	keys := make([]string, 0, len(zoneList))
	for k := range zoneList {
		keys = append(keys, k)
	}
	sort.Sort(sort.Reverse(sort.StringSlice(keys)))
	sortedZoneList := types.Map{}
	for _, k := range keys {
		sortedZoneList[k] = zoneList[k]
	}

	return c.ResponseSuccessToMe(types.Map{
		"current_time":   time.Now().Unix(),
		"current_date":   time.Now().Format("2006-01-02 15:04:05"),
		"zone_list":      sortedZoneList,
		"login_zone":     loginZone,
		"recommend_zone": config.Cfg.Zone.RecommendZone,
	})
}

func (c *UserController) GetZoneUser(ctx context.Context) *Result {
	userID := c.Params.GetInt64E("userid")
	token := c.Params.GetStringE("token")
	zoneIDParam := c.Params.GetIntE("zone_id")

	if userID == 0 {
		return c.ResponseError(10000021, "userid is empty")
	}
	if zoneIDParam == 0 {
		return c.ResponseError(10000021, "zone_id is empty")
	}

	userInfo := model.GetUserInfoByUserID(ctx, userID)
	if userInfo == nil {
		return c.ResponseError(10000023, "未找到该用户")
	}
	if token == "" || token != userInfo.Token {
		return c.ResponseError(10000024, "token 校验失败")
	}

	zoneUserID := model.GetUserZoneUID(userID, zoneIDParam)

	// 对应 PHP: $result['hequ'] = Zoneinfo::hequ($zone_id)
	return c.ResponseSuccessToMe(types.Map{
		"zone_user_id": zoneUserID,
		"zone_id":      zoneIDParam,
		"hequ":         logic.Hequ(zoneIDParam),
	})
}

func (c *UserController) Location(ctx context.Context) *Result {
	// 对应 PHP: locationAction — 完整的 filter 过滤逻辑
	appid := c.Params.GetStringE("appid")
	userid := c.Params.GetStringE("userid")
	version := c.Params.GetStringE("version")

	result := types.Map{
		"flag": 0,
		"type": "none",
	}

	filterCfg := config.Cfg.Filter

	// 获取 appid+version 对应的过滤配置
	var versionCfg *config.FilterVersionConfig
	if appid != "" {
		if appVersions, ok := filterCfg.AppVersions[appid]; ok {
			if vc, ok2 := appVersions[version]; ok2 {
				versionCfg = &vc
			}
		}
	}

	// 对应 PHP: if (empty($appid) || empty($config))
	if appid == "" || versionCfg == nil {
		// 默认排除几个地域
		if len(filterCfg.FilterArea) > 0 {
			area := getIPArea(c.IP)
			if area != "" && containsString(filterCfg.FilterArea, area) {
				result["flag"] = 1
				result["type"] = "area"
				return c.ResponseSuccessToMe(result)
			}
		}
		return c.ResponseSuccessToMe(result)
	}

	// 反向用户逻辑
	fanxiang := false
	if versionCfg.Fanxiang {
		fanxiang = true
	} else if len(versionCfg.IncludeUsers) > 0 && userid != "" {
		if containsString(versionCfg.IncludeUsers, userid) {
			fanxiang = true
		}
	}

	// 排除时间
	if len(versionCfg.FilterTimes) > 0 {
		nowTime := time.Now().Format("15:04")
		for _, ft := range versionCfg.FilterTimes {
			if nowTime >= ft[0] && nowTime <= ft[1] {
				if fanxiang {
					result["flag"] = 0
				} else {
					result["flag"] = 1
					result["type"] = "time"
				}
				return c.ResponseSuccessToMe(result)
			}
		}
	}

	// 排除用户
	if len(versionCfg.FilterUsers) > 0 && userid != "" {
		if containsString(versionCfg.FilterUsers, userid) {
			if fanxiang {
				result["flag"] = 0
			} else {
				result["flag"] = 1
				result["type"] = "user"
			}
			return c.ResponseSuccessToMe(result)
		}
	}

	// 排除地域
	if len(versionCfg.FilterArea) > 0 {
		area := getIPArea(c.IP)
		if area != "" && containsString(versionCfg.FilterArea, area) {
			if fanxiang {
				result["flag"] = 0
			} else {
				result["flag"] = 1
				result["type"] = "area"
			}
			return c.ResponseSuccessToMe(result)
		}
	}

	// 最终判断
	if fanxiang {
		result["flag"] = 1
		result["type"] = "fanxiang"
	} else {
		result["flag"] = 0
	}
	return c.ResponseSuccessToMe(result)
}

// getIPArea 根据 IP 获取地域信息
// 对应 PHP: Ip_Location::find($this->ip) 返回 $location[1]（省份/地区）
// TODO: 接入 IP 地理位置库后替换此实现
func getIPArea(ip string) string {
	// 目前项目中无 IP 地理位置库，返回空字符串
	// 后续可接入第三方 IP 定位库（如 ip2region 等）
	log.Infof(context.Background(), "getIPArea called with ip=%s, ip location not implemented yet", ip)
	return ""
}

// containsString 检查 slice 中是否包含指定字符串
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func (c *UserController) GameConf(ctx context.Context) *Result {
	// 对应 PHP: gameConfAction — 按 appid+version 返回游戏配置
	appid := c.Params.GetStringE("appid")
	version := c.Params.GetStringE("version")

	gameConfCfg := config.Cfg.GameConf
	if gameConfCfg == nil {
		return c.ResponseSuccessToMe(types.Map{})
	}

	appConf, ok := gameConfCfg[appid]
	if !ok {
		return c.ResponseSuccessToMe(types.Map{})
	}

	versionConf, ok := appConf[version]
	if !ok {
		return c.ResponseSuccessToMe(types.Map{})
	}

	result := types.Map{}
	for k, v := range versionConf {
		result[k] = v
	}
	return c.ResponseSuccessToMe(result)
}

// GetOpenGid 解密微信群数据获取 open_g_id
// 对应 PHP: getOpenGidAction — 使用 WXBizDataCrypt 解密微信加密数据
func (c *UserController) GetOpenGid(ctx context.Context) *Result {
	userID := c.Params.GetInt64E("userid")
	token := c.Params.GetStringE("token")
	encryptedData := c.Params.GetStringE("encrypted_data")
	iv := c.Params.GetStringE("iv")

	if userID == 0 {
		return c.ResponseError(10000021, "userid is empty")
	}

	userInfo := model.GetUserInfoByUserID(ctx, userID)
	if userInfo == nil {
		return c.ResponseError(10000023, "未找到该用户")
	}

	if token == "" || token != userInfo.Token {
		return c.ResponseError(10000024, "token 校验失败")
	}

	// 解密敏感数据
	sessionKey := userInfo.SessionKey
	appid := userInfo.Appid
	_ = appid // appid 在 PHP 中用于 WXBizDataCrypt 构造，Go 中 AES 解密不需要

	decryptedData, err := crypto.WXBizDataDecrypt(sessionKey, encryptedData, iv)
	if err != nil {
		log.Errorf(ctx, 0, "getOpenGid decryptData error: %v", err)
		return c.ResponseError(10003212, fmt.Sprintf("解密失败 - %v", err))
	}

	dataMap := json.ToMap(decryptedData)
	openGID := dataMap.GetStringE("openGId")
	if openGID == "" {
		// 兼容 PHP 中的 open_g_id 字段名
		openGID = dataMap.GetStringE("open_g_id")
	}

	return c.ResponseSuccessToMe(types.Map{
		"open_g_id": openGID,
	})
}
