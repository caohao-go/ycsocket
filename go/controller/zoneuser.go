package controller

import (
	"context"
	"strings"

	"server_golang/common/types"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
	"server_golang/repo/mem/userhero"
)

// 用户信息

func (c *ShinelightController) UpdateUserinfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	nickname := strings.TrimSpace(c.Params.GetStringE("nickname"))
	avatarURL := c.Params.GetStringE("avatar_url")
	gender := c.Params.GetStringE("gender")

	var iGender = 2
	if gender == "" || gender == "1" || gender == "男" {
		iGender = 1
	}

	model.SetUserInfo(userID, nickname, avatarURL, iGender)
	model.ReplaceUserNicknameRecord(ctx, userID, nickname)

	// 记录昵称去重
	if nickname != "" {
		model.SetNicknameSame(ctx, nickname)
	}

	return c.ResponseSuccessToMe(types.Map{})
}

func (c *ShinelightController) NicknameSameAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	nickname := c.Params.GetStringE("nickname")

	// 检查昵称是否已被使用（对应 PHP: RedisPool::instance("nickinfo")->hget('pre_nickname_same', md5($nickname))）
	same := 0
	if model.GetNicknameSame(ctx, nickname) {
		same = 1
	}

	// 微信内容安全检查（对应 PHP: msg_sec_check，errcode == 87014 表示违规）
	if model.MsgSecCheck(ctx, config.MsgSecCheckAppID, config.MsgSecCheckSecret, nickname) {
		return c.ResponseSuccessToMe(types.Map{"same": 1})
	}

	// 敏感词过滤（对应 PHP: str_replace($sensitive_words, "#", $nickname) + strpos($newnickname, "#") !== false）
	if logic.ContainsSensitiveWords(nickname) {
		return c.ResponseSuccessToMe(types.Map{"same": 1})
	}

	return c.ResponseSuccessToMe(types.Map{"same": same})
}

func (c *ShinelightController) UserInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetOrInitUserGradeWithInit(ctx, userID)
	if userGrade == nil {
		return c.ResponseError(99900032, "初始化用户失败")
	}

	// 是否在公会中
	isInGuild := false
	guildInfo := model.GetUsersGuildInfo(ctx, userID)
	if guildInfo != nil && guildInfo.Id != 0 {
		isInGuild = true
	}

	// 初始化标记：新用户插入昵称表
	userInfo := model.GetUserInfoByZoneUserID(ctx, userID)
	if userGrade.GetIntE("init_flag") == 0 && userInfo != nil {
		model.ReplaceUserNicknameRecord(ctx, userID, userInfo.Nickname)
		model.SetUserInfo(userID, userInfo.Nickname, userInfo.AvatarUrl, userInfo.Gender)
		item.AddCoin(userID, 1000)
		item.AddZuan(userID, 100)
	}

	model.SetInitFlag(userID)

	// 用户基础信息
	if userInfo != nil {
		userGrade["nickname"] = userInfo.Nickname
		userGrade["avatar_url"] = userInfo.AvatarUrl
		userGrade["gender"] = userInfo.Gender
	}

	// 记录登录区服信息
	model.ReplaceLoginZone(userID, userGrade.GetIntE("lv"))

	// 加载用户英雄到内存
	userhero.InitUserHeros(userID)

	// 成就任务：战斗力
	fightPoint := model.GetUserFightPoint(ctx, userID, 1)

	// 补充字段
	copyLv := userGrade.GetIntE("copy")
	lv := userGrade.GetIntE("lv")
	userGrade["exp"] = item.Exp(userID)
	userGrade["coin"] = item.Coin(userID)
	userGrade["zuan"] = item.Zuan(userID)
	userGrade["hero_exp"] = item.Total(userID, 7)
	userGrade["fuwen_jinghua"] = item.Total(userID, 20701)
	userGrade["fight_point"] = fightPoint
	userGrade["open_lv"] = logic.GetOpenLv(copyLv)
	userGrade["nextlv_need_exp"] = logic.GetLvUpdateExp(lv)
	userGrade["is_in_guild"] = isInGuild

	// 日常任务：每日登录
	model.SetDailyTaskFinish(ctx, userID, 10001, 1)

	model.AchieveTaskHandle(ctx, userID, 19, fightPoint, 14001, 14011)

	return c.ResponseSuccessToMe(userGrade)
}

func (c *ShinelightController) UserSpaceAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)
	userDetail := types.Map{
		"nickname":   userGrade.GetStringE("nickname"),
		"avatar_url": userGrade.GetStringE("avatar_url"),
		"gender":     userGrade.GetIntE("gender"),
		"lv":         userGrade.GetIntE("lv"),
	}

	userDetail["fight_point"] = model.GetUserFightPoint(ctx, userID, 1)

	// 公会名称
	guildDetail := model.GetUsersGuildDetail(ctx, userID)
	guildName := ""
	if guildDetail != nil {
		guildName = guildDetail.GuildName
	}

	userDetail["guild"] = guildName

	// 阵位英雄展示 (hero_shows)
	heroShows := make([]types.Map, 0)
	position := model.GetUserPositionByID(ctx, userID, 1)
	if position != nil {
		heroPos := position.HeroPos
		if len(heroPos) > 0 {
			ids := make([]int, 0, len(heroPos))
			for id := range heroPos {
				ids = append(ids, id)
			}
			positionHeros := model.GetUserHeroAttrByIDs(ctx, ids, userID, true)
			for _, id := range ids {
				if v, ok := positionHeros[id]; ok {
					heroShows = append(heroShows, types.Map{
						"id":      v.Id,
						"hero_id": v.HeroInfo,
						"star":    v.Star,
						"stage":   v.Stage,
						"lv":      v.Lv,
					})
				}
			}
		}
	}

	// 称号
	userDetail["used_role_title"] = userGrade.GetIntE("role_title")
	userDetail["role_titles"] = model.GetUserRoleTitles(ctx, userID)
	userDetail["hero_shows"] = heroShows

	return c.ResponseSuccessToMe(userDetail)
}

func (c *ShinelightController) OtherUserSpaceAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	otherUID := c.Params.GetInt64E("other_userid")
	posType := c.Params.GetIntE("position")
	if posType != 1 {
		posType = 2
	}

	userGrade := model.GetUserAttr(otherUID)
	if userGrade == nil {
		userGrade = types.Map{}
	}

	delete(userGrade, "id")
	delete(userGrade, "updatetime")

	// 用户基础信息（对应 PHP: UserinfoModel::getUserinfoByZoneUserId($other_userid)，从 nickinfo 读取）
	userInfo := model.GetUserAttr(otherUID)
	if userInfo != nil {
		userGrade["nickname"] = userInfo.GetStringE("nickname")
		userGrade["avatar_url"] = userInfo.GetStringE("avatar_url")
		gender := userInfo.GetIntE("gender")
		if gender != 2 {
			gender = 1
		}
		userGrade["gender"] = gender
	}
	userGrade["fight_point"] = model.GetUserFightPoint(ctx, otherUID, posType)

	// 公会名称
	guildDetail := model.GetUsersGuildDetail(ctx, otherUID)
	guildName := ""
	if guildDetail != nil {
		guildName = guildDetail.GuildName
	}
	userGrade["guild"] = guildName

	// 阵位英雄展示 (hero_shows) — 对齐 PHP: 先查指定阵位，为空则查另一个
	heroShows := make([]types.Map, 0)
	position := model.GetUserPositionByID(ctx, otherUID, posType)
	posHero := position.HeroPos
	if len(posHero) == 0 {
		// 切换阵位
		newPosType := 1
		if posType == 1 {
			newPosType = 2
		}
		position = model.GetUserPositionByID(ctx, otherUID, newPosType)
		posHero = position.HeroPos
	}

	fightPoint := 0
	if len(posHero) > 0 {
		ids := make([]int, 0, len(posHero))
		for id := range posHero {
			ids = append(ids, id)
		}
		positionHeros := model.GetUserHeroAttrByIDs(ctx, ids, otherUID, true)
		for _, id := range ids {
			if v, ok := positionHeros[id]; ok {
				fightPoint += v.FightPoint
				heroID := v.HeroInfo
				star := v.Star
				stage := v.Stage
				skillInfo := logic.GetSkillBaseInfo(heroID, star, stage)
				heroShow := types.Map{
					"id":      v.Id,
					"hero_id": heroID,
					"star":    star,
					"stage":   v.Stage,
					"lv":      v.Lv,
				}
				if skillInfo != nil {
					heroShow["skills"] = skillInfo.Skills
					heroShow["base_skill"] = skillInfo.BaseSkill
				}
				heroShows = append(heroShows, heroShow)
			}
		}
	}

	userGrade["hero_shows"] = heroShows
	userGrade["fight_point"] = fightPoint

	return c.ResponseSuccessToMe(userGrade)
}

// 称号系统

func (c *ShinelightController) ChooseRoleTitleAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	roleTitleID := c.Params.GetIntE("role_title_id")

	// 验证称号是否已激活（与 PHP chooseRoleTitleAction 一致）
	if !model.IsUserRoleExist(ctx, userID, roleTitleID) {
		return c.ResponseError(583324, "称号没有激活")
	}

	model.ChooseRoleTitle(userID, roleTitleID)
	return c.ResponseSuccessToMe(types.Map{})
}
