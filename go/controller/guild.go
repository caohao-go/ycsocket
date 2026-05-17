package controller

import (
	"context"
	"fmt"
	"sort"
	"time"

	"server_golang/common/lock"
	"server_golang/common/util"
	"server_golang/repo/mem/item"

	"server_golang/common/types"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/table"
)

// 公会系统

// CreateGuildAction 创建公会（与 PHP createGuildAction 一致）
func (c *ShinelightController) CreateGuildAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	guildName := c.Params.GetStringE("guild_name")
	if guildName == "" {
		return c.ResponseError(55001, "公会名不能为空")
	}
	lvLimit := c.Params.GetIntE("lv_limit")
	needCheck := c.Params.GetIntE("need_check")
	declar := c.Params.GetStringE("declar")

	lock.Lock(fmt.Sprintf("create_guild%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("create_guild%d", userID))

	// VIP 检查（与 PHP 一致：VIP2 才能创建）
	userGrade := model.GetUserAttr(userID)
	vipLevel := userGrade.GetIntE("vip_level")
	if vipLevel < 2 {
		return c.ResponseError(9923, "创建行会需要VIP2")
	}

	myGuild := model.GetUsersGuildID(ctx, userID)
	if myGuild > 0 {
		return c.ResponseError(3245, "请先退出当前行会")
	}

	// 消耗钻石 100（与 PHP 一致）
	if item.NotEnough(userID, 2, 100) {
		return c.ResponseError(666666, "钻石不够")
	}
	item.Sub(userID, 2, 100)

	guildID, err2 := model.CreateGuild(ctx, &table.Guild{
		GuildName: guildName,
		GuildLv:   1,
		PeopleNum: 1,
		Creator:   userID,
		OwnUser:   userID,
		LvLimit:   lvLimit,
		NeedCheck: needCheck,
		Declar:    declar,
	})
	if err2 != nil {
		return c.ResponseError(55003, "创建失败")
	}

	// 初始化公会boss数据（与 PHP 一致）
	model.InsertGuildChapterBlood(ctx, &table.GuildChapterBlood{
		Id:           int(guildID),
		Chapter:      1,
		ChapterBlood: logic.GetCopyChapterHP(1),
	})

	// 加入公会
	ret, _ := model.InsertUsersGuild(ctx, &table.UsersGuild{
		GuildId: guildID, Zhiwei: 1, UserId: userID,
	}, userID)
	if !ret {
		return c.ResponseError(9129, "加入失败")
	}

	model.GuideTaskHandle(ctx, userID, 93, 1)

	return c.ResponseSuccessToMe(types.Map{"guild_id": guildID})
}

// GuildInfoAction 公会信息（与 PHP guildInfoAction 一致）
func (c *ShinelightController) GuildInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := types.Map{}

	// 等级检查（与 PHP 一致：lv < 15 返回 status=1）
	userGrade := model.GetUserAttr(userID)
	lv := userGrade.GetIntE("lv")
	if lv < 15 {
		data["status"] = 1
		return c.ResponseSuccessToMe(data)
	}

	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		data["status"] = 2
		return c.ResponseSuccessToMe(data)
	}

	data["status"] = 0
	tmp := model.GetGuildInfoByID(ctx, guildID)
	if tmp == nil || tmp.Id == 0 {
		return c.ResponseError(99, "行会信息不存在")
	}

	guidInfo := types.ObjectToMap(tmp)

	// 会长信息（与 PHP 一致）
	ownUserInfo := model.GetUserAttr(tmp.OwnUser)
	if ownUserInfo != nil {
		guidInfo["own_nickname"] = ownUserInfo.GetStringE("nickname")
		guidInfo["own_avatar_url"] = ownUserInfo.GetStringE("avatar_url")
	}

	// 成员上限和升级经验（与 PHP 一致）
	if lvData, ok := logic.GuildLvDatas[tmp.GuildLv]; ok {
		guidInfo["member_limit"] = lvData.MemberNum
		guidInfo["up_lv_need_exp"] = lvData.Exp
	}

	data["guid_info"] = guidInfo

	// 弹劾会长状态（与 PHP 一致）
	usersGuildInfo := model.GetUsersGuildInfo(ctx, userID)
	data["tanhe_stutas"] = 0
	if usersGuildInfo != nil && usersGuildInfo.Zhiwei != 1 {
		ownUserGrade := model.GetUserAttr(tmp.OwnUser)
		offTimeStr := ownUserGrade.GetStringE("off_time")
		if offTimeStr != "" && offTimeStr != "0" {
			offTime, _ := time.Parse("2006-01-02 15:04:05", offTimeStr)
			offDays := int(time.Since(offTime).Hours() / 24)
			if offDays >= 7 {
				data["tanhe_stutas"] = 1
			}
		}
	}

	return c.ResponseSuccessToMe(data)
}

// GuildListAction 公会列表
func (c *ShinelightController) GuildListAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	list, _ := model.GetGuildList(ctx)
	return c.ResponseSuccessToMe(types.Map{"list": list})
}

// GetGuildUsersAction 公会成员列表（与 PHP getGuildUsersAction 一致）
func (c *ShinelightController) GetGuildUsersAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := types.Map{}
	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		data["status"] = 2
		return c.ResponseSuccessToMe(data)
	}

	data["status"] = 0
	guidInfo := model.GetGuildInfoByID(ctx, guildID)
	if guidInfo == nil {
		return c.ResponseError(99, "行会信息不存在")
	}

	guildLv := guidInfo.GuildLv
	if lvData, ok := logic.GuildLvDatas[guildLv]; ok {
		data["member_limit"] = lvData.MemberNum
	}

	members := model.GetGuildsUser(ctx, guildID, false, userID)
	data["guild_users"] = members
	data["member_num"] = len(members)

	return c.ResponseSuccessToMe(data)
}

// ApplyGuildAction 加入/申请公会（与 PHP applyGuildAction 一致）
func (c *ShinelightController) ApplyGuildAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	guildID := c.Params.GetIntE("guild_id")

	guildInfo := model.GetGuildInfoByID(ctx, guildID)
	if guildInfo == nil {
		return c.ResponseError(99, "行会信息不存在")
	}

	myGuild := model.GetUsersGuildID(ctx, userID)
	if myGuild > 0 {
		return c.ResponseError(3245, "请先退出当前行会")
	}

	// 退出等待时间检查（与 PHP 一致）
	waitTime := model.GetQuitWaitTime(ctx, userID)
	if waitTime > 0 {
		return c.ResponseError(93298, fmt.Sprintf("需要等待%d小时", waitTime/3600))
	}

	// 等级限制检查（与 PHP 一致）
	lvLimit := guildInfo.LvLimit
	if lvLimit != 0 {
		userGrade := model.GetUserAttr(userID)
		if userGrade.GetIntE("lv") < lvLimit {
			return c.ResponseError(9329, "等级限制")
		}
	}

	needCheck := guildInfo.NeedCheck
	apply := 0
	if needCheck != 0 {
		// 需要审核
		model.AddUsersGuildApply(ctx, guildID, userID)
		apply = 1
	} else {
		// 直接加入
		ret, _ := model.InsertUsersGuild(ctx, &table.UsersGuild{
			GuildId: guildID, UserId: userID,
		}, userID)
		if !ret {
			return c.ResponseError(9129, "人数已满")
		}

		model.GuideTaskHandle(ctx, userID, 93, 1)

		apply = 2
		// 从申请列表里面删除
		model.DelUsersGuildApply(ctx, guildID, userID)
	}

	return c.ResponseSuccessToMe(types.Map{"need_check": needCheck, "status": apply})
}

// QuitGuildAction 退出公会（与 PHP quitGuildAction 一致）
func (c *ShinelightController) QuitGuildAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	usersGuildInfo := model.GetUsersGuildInfo(ctx, userID)
	if usersGuildInfo == nil {
		return c.ResponseSuccessToMe(types.Map{"status": 2})
	}

	// 会长不能直接退出（与 PHP 一致：zhiwei==1）
	if usersGuildInfo.Zhiwei == 1 {
		return c.ResponseError(996, "请先指定会长")
	}

	guildID := usersGuildInfo.GuildId
	model.UserQuitGuild(ctx, userID)

	// 清空个人活跃（与 PHP 一致）
	model.UpdateUsersContentInt(ctx, userID, "guild_active", 0)

	// 清空个人贡献（与 PHP 一致）
	gongxianData := model.GetGuildGongxian(ctx, guildID)
	if len(gongxianData) > 0 {
		model.HdelGuildGongxian(ctx, guildID, userID)
	}

	// 退出等待时间（与 PHP 一致：第一次退出不需要等12小时）
	model.QuitWaitTime(ctx, userID)

	return c.ResponseSuccessToMe(types.Map{"success": 1})
}

// GuildApplyListAction 公会申请列表（与 PHP guildApplyListAction 一致）
func (c *ShinelightController) GuildApplyListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	usersGuildInfo := model.GetUsersGuildInfo(ctx, userID)
	if usersGuildInfo == nil {
		return c.ResponseSuccessToMe(types.Map{"list": []interface{}{}})
	}

	// 只有会长和副会长能看申请列表（与 PHP 一致：zhiwei==1 or zhiwei==2）
	zhiwei := usersGuildInfo.Zhiwei
	if zhiwei != 1 && zhiwei != 2 {
		return c.ResponseSuccessToMe(types.Map{"list": []interface{}{}})
	}

	guildID := usersGuildInfo.GuildId
	list := model.GetGuildsApplyUser(ctx, guildID)
	return c.ResponseSuccessToMe(types.Map{"list": list})
}

// GuildApplyHandleAction 处理公会申请（与 PHP guildApplyHandleAction 一致）
func (c *ShinelightController) GuildApplyHandleAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	// 参数名与 PHP 保持一致：apply_userid、op（0-拒绝 1-接受）
	applyUserID := c.Params.GetInt64E("apply_userid")
	op := c.Params.GetIntE("op")

	// 权限校验（与 PHP 一致：zhiwei < 1 不可操作）
	guildInfo := model.GetUsersGuildInfo(ctx, userID)
	if guildInfo == nil || guildInfo.Zhiwei < 1 {
		return c.ResponseError(2949, "权限不够")
	}
	guildID := guildInfo.GuildId

	// 检查申请者是否已加入公会（与 PHP 一致）
	applyGuildInfo := model.GetUsersGuildInfo(ctx, applyUserID)
	if applyGuildInfo != nil && applyGuildInfo.GuildId == guildID {
		return c.ResponseError(29449, "已经加入")
	}
	if applyGuildInfo != nil {
		return c.ResponseError(29449, "已经加入其它工会")
	}

	// 检查申请记录是否存在（与 PHP 一致）
	applyScore := model.GetUserGuildApplyByUID(ctx, guildID, applyUserID)
	if applyScore == 0 {
		return c.ResponseError(2999, "操作错误")
	}

	if op == 1 {
		ret, _ := model.InsertUsersGuild(ctx, &table.UsersGuild{
			GuildId: guildID, UserId: applyUserID,
		}, applyUserID)
		if !ret {
			return c.ResponseError(91291, "人数已满")
		}

		model.GuideTaskHandle(ctx, applyUserID, 93, 1)
	}
	model.DelUsersGuildApply(ctx, guildID, applyUserID)
	return c.ResponseSuccessToMe(types.Map{})
}

// GuildEditAction 编辑公会信息（合并 PHP setGuildLimitAction + setGuildDeclarAction 逻辑）
// 只有会长可以修改；支持 declar、lv_limit、need_check 字段
func (c *ShinelightController) GuildEditAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	usersGuild := model.GetUsersGuildInfo(ctx, userID)
	if usersGuild == nil {
		return c.ResponseError(55010, "未加入公会")
	}
	guildID := usersGuild.GuildId

	// 行会信息不存在校验（与 PHP 一致）
	guildInfo := model.GetGuildInfoByID(ctx, guildID)
	if guildInfo == nil {
		return c.ResponseError(99, "行会信息不存在")
	}

	// 只有会长可以修改（与 PHP 一致）
	if guildInfo.OwnUser != userID {
		return c.ResponseError(545799, "只有会长可以修改")
	}

	// 参数名与 PHP 保持一致：declar / lv_limit / need_check
	data := types.Map{}
	if declar, ok := c.Params["declar"]; ok {
		data["declar"] = types.ToString(declar)
	}
	if _, ok := c.Params["lv_limit"]; ok {
		data["lv_limit"] = c.Params.GetIntE("lv_limit")
	}
	if _, ok := c.Params["need_check"]; ok {
		data["need_check"] = c.Params.GetIntE("need_check")
	}
	if len(data) > 0 {
		model.UpdateGuildInfoByID(ctx, guildID, data)
	}
	return c.ResponseSuccessToMe(types.Map{})
}

// GuildRankListAction 公会排行榜（与 PHP guildRankListAction 一致）
func (c *ShinelightController) GuildRankListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := types.Map{}
	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		data["status"] = 2
	} else {
		data["status"] = 0
	}

	list, _ := model.GetGuildRank(ctx)
	data["list"] = list

	if guildID > 0 {
		for k, v := range list {
			if v.Id == guildID {
				data["myguild_index"] = k
				break
			}
		}
	}

	return c.ResponseSuccessToMe(data)
}

// GuildTanheAction 弹劾会长（与 PHP guildTanheAction 一致）
func (c *ShinelightController) GuildTanheAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	usersGuildInfo := model.GetUsersGuildInfo(ctx, userID)
	if usersGuildInfo == nil {
		return c.ResponseError(9968, "数据错误")
	}

	// 会长不能弹劾自己（与 PHP 一致：zhiwei==1）
	if usersGuildInfo.Zhiwei == 1 {
		return c.ResponseError(9967, "您已经是会长了")
	}

	guildID := usersGuildInfo.GuildId
	guildInfo := model.GetGuildInfoByID(ctx, guildID)
	if guildInfo == nil {
		return c.ResponseError(99, "system error")
	}

	// 检查会长离线时间（与 PHP 一致：超过7天才可以弹劾）
	ownUserGrade := model.GetUserAttr(guildInfo.OwnUser)
	offTimeStr := ownUserGrade.GetStringE("off_time")
	if offTimeStr != "" && offTimeStr != "0" {
		offTime, _ := time.Parse("2006-01-02 15:04:05", offTimeStr)
		offDays := int(time.Since(offTime).Seconds() / 86400)
		if offDays < 7 {
			return c.ResponseError(9969, "会长离线未超过7天")
		}
	} else {
		return c.ResponseError(9969, "会长离线未超过7天")
	}

	// 消耗钻石 300（与 PHP 一致）
	if item.NotEnough(userID, 2, 300) {
		return c.ResponseError(99691, "弹劾会长需要300砖石")
	}
	item.Sub(userID, 2, 300)

	// 更改帮会状态（与 PHP 一致：转让会长给弹劾者）
	model.GuildChangeOwner(ctx, guildID, guildInfo.OwnUser, userID)

	return c.ResponseSuccessToMe(types.Map{})
}

// GuildOpAction 公会操作（与 PHP guildOpAction 一致）
// op: 1-转让会长 2-任命副会长 3-移除副会长 4-移除会员
func (c *ShinelightController) GuildOpAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	opUserID := c.Params.GetInt64E("op_userid")
	op := c.Params.GetIntE("op")

	usersGuild := model.GetUsersGuildInfo(ctx, userID)
	if usersGuild == nil {
		return c.ResponseError(55010, "未加入公会")
	}
	guildID := usersGuild.GuildId
	zhiwei := usersGuild.Zhiwei

	opUsersGuild := model.GetUsersGuildInfo(ctx, opUserID)

	// 不能对自己操作（与 PHP 一致）
	if userID == opUserID {
		return c.ResponseError(34215, "参数错误")
	}

	// 检查操作对象是否在同一公会（与 PHP 一致）
	if model.GetUsersGuildID(ctx, opUserID) != guildID {
		return c.ResponseError(3215, "玩家已退出工会")
	}

	guildInfo := model.GetGuildInfoByID(ctx, guildID)
	if guildInfo == nil {
		return c.ResponseError(99, "行会信息不存在")
	}

	switch op {
	case 1:
		// 转让会长（与 PHP 一致：只有会长才能操作）
		if zhiwei != 1 {
			return c.ResponseError(9763, "会长才可以操作")
		}
		model.GuildChangeOwner(ctx, guildID, userID, opUserID)

	case 2:
		// 任命副会长（与 PHP 一致：只有会长才能操作，最多3个副会长）
		if zhiwei != 1 {
			return c.ResponseError(94954, "会长才可以操作")
		}
		cnt := model.GetGuildFuhuizhangCount(ctx, guildID)
		if cnt >= 3 {
			return c.ResponseError(9453, "副会长最多任命3个")
		}
		model.UpdateUsersGuild(ctx, opUserID, types.Map{"zhiwei": 2})

	case 3:
		// 移除副会长（与 PHP 一致：只有会长才能操作）
		if zhiwei != 1 {
			return c.ResponseError(94954, "会长才可以操作")
		}
		model.UpdateUsersGuild(ctx, opUserID, types.Map{"zhiwei": 0})

	default:
		// 移除会员（与 PHP 一致：会长或副会长才能操作，副会长不能踢副会长）
		if zhiwei != 1 && zhiwei != 2 {
			return c.ResponseError(459854, "会长或者副会长才可以操作")
		}
		if zhiwei == 2 && opUsersGuild != nil && opUsersGuild.Zhiwei == 2 {
			return c.ResponseError(459854, "会长才可以操作")
		}
		model.UserQuitGuild(ctx, opUserID)
	}

	return c.ResponseSuccessToMe(types.Map{})
}

// GetContributionAction 获取贡献信息（与 PHP getContributionAction 一致）
func (c *ShinelightController) GetContributionAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := types.Map{}
	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		data["status"] = 2
		return c.ResponseSuccessToMe(data)
	}

	data["status"] = 0
	guildInfo := model.GetGuildInfoByID(ctx, guildID)
	if guildInfo == nil {
		return c.ResponseError(99, "行会信息不存在")
	}

	data["guild_lv"] = guildInfo.GuildLv

	// 下一级成员上限（与 PHP 一致：guild_lv + 1 对应的 member_limit）
	nextLv := guildInfo.GuildLv + 1
	if nextLvData, ok := logic.GuildLvDatas[nextLv]; ok {
		data["next_member_limit"] = nextLvData.MemberNum
	} else {
		data["next_member_limit"] = 0
	}

	// 今日帮会捐献活跃值（与 PHP 一致）
	data["contribution_active"] = model.GetGuildContributeActive(ctx, guildID)

	// 捐献活跃领取状态（与 PHP 一致）
	data["active_rewards"] = model.GetActiveLingquStatus(ctx, userID)

	// 捐献的获取奖励与捐献状态（与 PHP 一致：硬编码3种捐献档次）
	contributions := []types.Map{
		{
			"id":         1,
			"status":     model.GetGuildContribute(ctx, userID, 1),
			"contribute": types.Map{"type": 1, "num": 20000},
			"reward":     []types.Map{{"type": 8, "num": 100}, {"type": 9, "num": 50}},
		},
		{
			"id":         2,
			"status":     model.GetGuildContribute(ctx, userID, 2),
			"contribute": types.Map{"type": 2, "num": 100},
			"reward":     []types.Map{{"type": 8, "num": 250}, {"type": 9, "num": 100}},
		},
		{
			"id":         3,
			"status":     model.GetGuildContribute(ctx, userID, 3),
			"contribute": types.Map{"type": 2, "num": 300},
			"reward":     []types.Map{{"type": 8, "num": 800}, {"type": 9, "num": 200}},
		},
	}
	data["contribution"] = contributions

	return c.ResponseSuccessToMe(data)
}

// ContributeAction 捐献（与 PHP contributeAction 一致）
func (c *ShinelightController) ContributeAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")
	if id != 1 && id != 2 && id != 3 {
		return c.ResponseError(99, "id error")
	}

	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		return c.ResponseError(99, "未加入行会")
	}

	lock.Lock(fmt.Sprintf("contribute%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("contribute%d", userID))

	guildInfo := model.GetGuildInfoByID(ctx, guildID)
	if guildInfo == nil {
		return c.ResponseError(99, "行会信息不存在")
	}

	// 捐献配置（与 PHP 一致：硬编码3种）
	type contributeConfig struct {
		ContributeType int
		ContributeNum  int
		Reward         []util.TypeNum
	}
	contributions := map[int]contributeConfig{
		1: {ContributeType: 1, ContributeNum: 20000, Reward: []util.TypeNum{{Type: 8, Num: 100}, {Type: 9, Num: 50}}},
		2: {ContributeType: 2, ContributeNum: 100, Reward: []util.TypeNum{{Type: 8, Num: 250}, {Type: 9, Num: 100}}},
		3: {ContributeType: 2, ContributeNum: 300, Reward: []util.TypeNum{{Type: 8, Num: 800}, {Type: 9, Num: 200}}},
	}
	contribution := contributions[id]

	// 检查今天是否已经捐献（与 PHP 一致）
	if model.GetGuildContribute(ctx, userID, id) != 0 {
		return c.ResponseError(2149, "今天已经领取")
	}

	// 检查货币是否足够（与 PHP 一致）
	if item.NotEnough(userID, contribution.ContributeType, contribution.ContributeNum) {
		return c.ResponseError(666666, "货币不够")
	}

	// 设置捐献状态（与 PHP 一致）
	model.SetGuildContribute(ctx, userID, id)

	// 删除捐献货币（与 PHP 一致）
	item.Sub(userID, contribution.ContributeType, contribution.ContributeNum)

	// 增加当日捐献活跃值（与 PHP 一致）
	model.IncrGuildContributeActive(ctx, guildID, id)

	// 增加帮会经验和等级（与 PHP 一致）
	beforeExp := guildInfo.Exp
	rewardNum9 := 0
	if len(contribution.Reward) >= 2 {
		rewardNum9 = contribution.Reward[1].Num
	}
	afterExp := beforeExp + rewardNum9

	afterLv := guildInfo.GuildLv
	// 按等级顺序遍历公会等级数据（修复 map 遍历顺序不确定问题）
	lvKeys := make([]int, 0, len(logic.GuildLvDatas))
	for k := range logic.GuildLvDatas {
		lvKeys = append(lvKeys, k)
	}
	sort.Ints(lvKeys)
	for _, k := range lvKeys {
		data := logic.GuildLvDatas[k]
		if afterExp <= data.Exp {
			afterLv = k
			break
		}
	}

	model.UpdateGuildInfoByID(ctx, guildID, types.Map{"guild_lv": afterLv, "exp": afterExp})

	// 给予捐献奖励（与 PHP 一致）
	model.GiveReward(userID, contribution.Reward...)

	// 日常任务（与 PHP 一致：捐献1次，任务 10008）
	model.SetDailyTaskFinish(ctx, userID, 10008, 1)

	// 公会活跃（与 PHP 一致）
	tasks := model.GetGuildTask(ctx, userID)
	for _, v := range tasks {
		taskID := v.TaskId
		finishCount := v.FinishCount
		taskCountLimit := v.TaskCountLimit

		if taskID == 5 && contribution.ContributeType == 1 {
			// 公会捐献金币（与 PHP 一致）
			if finishCount < taskCountLimit && finishCount >= 0 {
				model.IncrTaskFinishNumStr(ctx, userID, "guild_5", 1)
				model.IncrUsersContentInt(ctx, userID, "guild_active", 5)
			}
		}
		if taskID == 6 && contribution.ContributeType == 2 && contribution.ContributeNum == 100 {
			// 公会捐献钻石100（与 PHP 一致）
			if finishCount < taskCountLimit && finishCount >= 0 {
				model.IncrTaskFinishNumStr(ctx, userID, "guild_6", 1)
				model.IncrUsersContentInt(ctx, userID, "guild_active", 10)
			}
		}
		if taskID == 7 && contribution.ContributeType == 2 && contribution.ContributeNum == 300 {
			// 公会捐献钻石300（与 PHP 一致）
			if finishCount < taskCountLimit && finishCount >= 0 {
				model.IncrTaskFinishNumStr(ctx, userID, "guild_7", 1)
				model.IncrUsersContentInt(ctx, userID, "guild_active", 20)
			}
		}
	}

	model.GuideTaskHandle(ctx, userID, 99, 1)

	return c.ResponseSuccessToMe(types.Map{})
}

// ContributeLingquAction 贡献领取（与 PHP contributeLingquAction 一致）
func (c *ShinelightController) ContributeLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")
	if id != 1 && id != 2 && id != 3 {
		return c.ResponseError(99, "id error")
	}

	// 活跃阈值（与 PHP 一致：id=3→3000, id=2→2000, id=1→1000）
	key := 1000
	if id == 3 {
		key = 3000
	} else if id == 2 {
		key = 2000
	}

	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		return c.ResponseSuccessToMe(types.Map{"status": 2})
	}

	lock.Lock(fmt.Sprintf("contribute_lingqu%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("contribute_lingqu%d", userID))

	// 获取当日捐献活跃值（与 PHP 一致）
	juanxianNum := model.GetGuildContributeActive(ctx, guildID)
	if juanxianNum < key {
		return c.ResponseError(3994, "活跃值不够")
	}

	// 捐献活跃领取状态（与 PHP 一致）
	activeRewards := model.GetActiveLingquStatus(ctx, userID)
	if activeRewards.GetIntE(key) == 1 {
		return c.ResponseError(3995, "今日已经领取")
	}

	// 设置领取状态（与 PHP 一致）
	model.SetGuildContributeActiveLingqu(ctx, userID, id)

	// 获取活跃奖励配置（与 PHP 一致：从 GuildActiveAttrDatas 读取对应 key 的奖励）
	rewards := logic.GetActiveRewards()
	var reward = []util.TypeNum{}
	for _, r := range rewards {
		// 根据活跃等级匹配奖励阈值
		lv := r.GetIntE("lv")
		if lvData, ok := logic.GuildActiveAttrDatas[lv]; ok {
			if lvData.TotalActiveNum == key {
				reward = r["reward"].([]util.TypeNum)
				break
			}
		}
	}

	// 如果无法精确匹配，按默认方式给奖励
	if len(reward) == 0 {
		// 按 id 给固定奖励（兜底）
		switch id {
		case 1:
			reward = []util.TypeNum{{Type: 8, Num: 50}, {Type: 1, Num: 10000}}
		case 2:
			reward = []util.TypeNum{{Type: 8, Num: 100}, {Type: 1, Num: 20000}}
		case 3:
			reward = []util.TypeNum{{Type: 8, Num: 200}, {Type: 1, Num: 50000}}
		}
	}

	// 给予奖励（与 PHP 一致）
	model.GiveReward(userID, reward...)
	return c.ResponseSuccessToMe(types.Map{"rewards": reward})
}

// GetGuildSkillAction 公会技能列表（与 PHP getGuildSkillAction 一致）
func (c *ShinelightController) GetGuildSkillAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	skills := model.GetGuildSkills(ctx, userID)
	// 获得玩家的帮会贡献（经验）（与 PHP 一致）
	gongxian := item.Total(userID, 8)
	return c.ResponseSuccessToMe(types.Map{"guild_skill": skills, "gongxian": gongxian})
}

// ActiveGuildSkillAction 激活公会技能（与 PHP activeGuildSkillAction 一致）
func (c *ShinelightController) ActiveGuildSkillAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	prop := c.Params.GetIntE("skill_id")
	if prop <= 0 {
		prop = c.Params.GetIntE("prop")
	}

	lock.Lock(fmt.Sprintf("active_guild_skill%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("active_guild_skill%d", userID))

	oldFP := model.GetUserFightPoint(ctx, userID, 1)

	// 获取当前技能数据计算消耗（与 PHP 一致：从配置表读取）
	key := prop - 1
	datas := model.GetUsersContent(ctx, userID, "guild_skills")

	dataMap := datas.GetMapE(key)

	lv := 1
	attrKey := 0
	if len(dataMap) > 0 {
		lv = dataMap.GetIntE("lv")
		attrLvRaw := dataMap["attr_lv"]
		// 找出下一个要升级的 attr_key
		attrLvs, _ := types.ToIntArray(attrLvRaw)

		for k, alv := range attrLvs {
			if alv < lv {
				attrKey = k
				break
			}
		}
	}

	// 从配置表获取消耗（与 PHP 一致）
	needNum8, needNum1 := 0, 0
	if consumeData, ok := logic.GuildSkillConsumeDatas[prop]; ok {
		if lvData, ok := consumeData[lv]; ok {
			if consume, ok := lvData[attrKey]; ok {
				for _, c := range consume {
					if types.ToIntE(c.Type) == 8 {
						needNum8 = c.Num
					}
					if types.ToIntE(c.Type) == 1 {
						needNum1 = c.Num
					}
				}
			}
		}
	}

	if needNum8 > 0 && item.NotEnough(userID, 8, needNum8) {
		return c.ResponseError(7777, "帮贡不够")
	}
	if needNum1 > 0 && item.NotEnough(userID, 1, needNum1) {
		return c.ResponseError(666666, "金币不够")
	}

	if needNum8 > 0 {
		item.Sub(userID, 8, needNum8)
	}
	if needNum1 > 0 {
		item.Sub(userID, 1, needNum1)
	}

	// 激活技能
	model.ActiveGuildSkill(ctx, userID, prop)

	newFP := model.GetUserFightPoint(ctx, userID, 1)

	// 公会活跃（与 PHP 一致：task_id=2 公会技能）
	tasks := model.GetGuildTask(ctx, userID)
	for _, v := range tasks {
		if v.TaskId == 2 {
			finishCount := v.FinishCount
			taskCountLimit := v.TaskCountLimit
			if finishCount < taskCountLimit && finishCount >= 0 {
				model.IncrTaskFinishNumStr(ctx, userID, "guild_2", 1)
				model.IncrUsersContentInt(ctx, userID, "guild_active", 10)
			}
			break
		}
	}

	model.GuideTaskHandle(ctx, userID, 101, 1)

	return c.ResponseSuccessToMe(types.Map{
		"old_fight_point": oldFP,
		"new_fight_point": newFP,
	})
}
