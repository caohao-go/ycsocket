package controller

import (
	"context"
	"fmt"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/table"
)

// 分享/邀请/视频

func (c *ShinelightController) ShareinfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 已分享多少次
	times := model.GetShareTimes(ctx, userID)
	// 获取对应奖励配置
	rewards, ok := logic.ShareRewards[times]
	if !ok {
		rewards = []util.TypeNum{}
	}
	// 3分钟后分享的时间戳
	shareTime := model.GetShareTime(ctx, userID)

	// rewards 转为 []types.Map
	rewardMaps := make([]types.Map, 0, len(rewards))
	for _, r := range rewards {
		rewardMaps = append(rewardMaps, types.Map{"type": r.Type, "num": r.Num})
	}

	data := types.Map{
		"times":   times + 1,
		"rewards": rewardMaps,
		"time":    shareTime,
	}
	return c.ResponseSuccessToMe(data)
}

func (c *ShinelightController) SharelingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 加锁防并发
	lockKey := fmt.Sprintf("sharelingqu%d", userID)
	if lockErr := lock.Lock(lockKey, 5); lockErr != nil {
		return c.ResponseError(35671, lockErr.Error())
	}
	defer lock.Unlock(lockKey)

	// 已分享多少次
	times := model.GetShareTimes(ctx, userID)
	if times >= 2 {
		return c.ResponseError(35671, "今天已经分享2次了！")
	}

	// 获取奖励配置
	rewards, ok := logic.ShareRewards[times]
	if !ok || len(rewards) == 0 {
		return c.ResponseError(35671, "奖励档位错误！")
	}

	// 给予奖励

	model.GiveReward(userID, rewards...)

	// 增加次数
	model.IncrShareTimes(ctx, userID)

	// 如果是第一次分享，设置3分钟后分享
	if times == 0 {
		model.SetShareTime(ctx, userID)
	}

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 48, 1)

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

// InviteinfoAction 邀请界面
// 对应 PHP: Shinelight::inviteinfoAction
func (c *ShinelightController) InviteinfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 获取邀请领取状态
	times := model.GetInviteInfo(ctx, userID)

	data := types.Map{}
	data["times"] = []types.Map{
		{"num": 1, "status": times.GetIntE("1")},
		{"num": 3, "status": times.GetIntE("3")},
		{"num": 5, "status": times.GetIntE("5")},
	}

	// 获取所有我邀请的玩家数量
	players, _ := model.GetMyInvitePlayers(ctx, userID)
	data["count"] = len(players)

	return c.ResponseSuccessToMe(data)
}

// InviteAction 被邀请玩家调用
// 对应 PHP: Shinelight::inviteAction
func (c *ShinelightController) InviteAction(ctx context.Context) *Result {
	userID := c.Params.GetInt64E("userid")
	inviteZoneUID := c.Params.GetInt64E("invite_zone_uid")

	// 获取用户信息，检查登录次数
	userInfo := model.GetUserInfoByZoneUserID(ctx, userID)

	// 玩家登陆次数大于2次，则认为玩家已经存在
	if userInfo.LoginTimes > 2 {
		return c.ResponseSuccessToMe(types.Map{"success": 0})
	}

	// 玩家是否已经被邀请过一次
	existInvite, _ := model.GetInviteByUserID(ctx, userID)
	if existInvite != nil {
		return c.ResponseSuccessToMe(types.Map{"success": 0})
	}

	// 插入邀请记录
	model.InsertInviteInfo(ctx, &table.InviteInfo{
		UserId:        userID,
		InviteZoneUid: inviteZoneUID,
	})

	return c.ResponseSuccessToMe(types.Map{"success": 1})
}

// InvitelingquAction 邀请领取
// 对应 PHP: Shinelight::invitelingquAction
func (c *ShinelightController) InvitelingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	num := c.Params.GetIntE("num")

	// 加锁防并发
	lockKey := fmt.Sprintf("invitelingqu%d", userID)
	if lockErr := lock.Lock(lockKey, 5); lockErr != nil {
		return c.ResponseError(356712, lockErr.Error())
	}
	defer lock.Unlock(lockKey)

	// 获取所有我邀请的玩家
	players, _ := model.GetMyInvitePlayers(ctx, userID)
	countNum := len(players)
	if num > countNum {
		return c.ResponseError(356712, "邀请的人数不够哦！")
	}

	// 获取邀请领取状态

	times := model.GetInviteInfo(ctx, userID)
	if times.GetIntE(num) > 0 {
		return c.ResponseError(3567121, "奖励已经领取！")
	}

	// 获取奖励配置
	rewards, ok := logic.InviteRewards[num]
	if !ok || len(rewards) == 0 {
		return c.ResponseError(356712, "奖励档位错误！")
	}

	// 给予奖励
	model.GiveReward(userID, rewards...)

	// 设置领取状态
	model.SetInvitePlayerStatus(ctx, userID, num)

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

func (c *ShinelightController) VedioinfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	data := model.GetVedioInfoFull(ctx, userID)

	// 对齐 PHP: `$data['status'] = array_values($data['status'])`，把 status 由 {1,3,5,7:object}
	// 转为 [num=1, num=3, num=5, num=7] 的数组，供前端按序访问。
	statusList := make([]types.Map, 0, 4)
	if statusMap, err := types.ToMap(data["status"], ""); err == nil && statusMap != nil {
		for _, k := range []string{"1", "3", "5", "7"} {
			if entry, err := types.ToMap(statusMap[k], ""); err == nil && entry != nil {
				statusList = append(statusList, entry)
			}
		}
	}
	data["status"] = statusList

	return c.ResponseSuccessToMe(data)
}

// VedioAction 看视频
// 对应 PHP: Shinelight::vedioAction
func (c *ShinelightController) VedioAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 获取看视频状态
	data := model.GetVedioInfoFull(ctx, userID)
	if data.GetIntE("times") > 7 {
		return c.ResponseError(356711, "今天已经看了7次视频了！")
	}

	// 设置状态
	model.IncrVedioTimes(ctx, userID)
	return c.ResponseSuccessToMe(types.Map{"success": 1})
}

// VediolingquAction 看视频领取奖励
// 对应 PHP: Shinelight::vediolingquAction
func (c *ShinelightController) VediolingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	num := c.Params.GetIntE("num") // 1,3,5,7

	// 加锁防并发
	lockKey := fmt.Sprintf("vediolingqu%d", userID)
	if lockErr := lock.Lock(lockKey, 5); lockErr != nil {
		return c.ResponseError(356712, lockErr.Error())
	}
	defer lock.Unlock(lockKey)

	// 获取奖励配置
	rewards, ok := logic.VedioRewards[num]
	if !ok || len(rewards) == 0 {
		return c.ResponseError(356712, "奖励档位错误！")
	}

	// 获取看视频状态

	dates := model.GetVedioInfoFull(ctx, userID)

	// 检查看视频次数是否足够
	if dates.GetIntE("times") < num {
		return c.ResponseError(356712, "看视频的次数不够哦！")
	}

	// 检查奖励是否已领取
	statusMap, _ := types.ToMap(dates["status"], "")

	if statusEntry := statusMap.GetMapE(num); len(statusEntry) > 0 {
		if statusEntry.GetIntE("status") == 1 {
			return c.ResponseError(3567121, "奖励已经领取！")
		}
	}

	// 设置领取状态
	model.SetVedioStatus(ctx, userID, num)

	// 给予奖励
	model.GiveReward(userID, rewards...)

	return c.ResponseSuccessToMe(types.Map{"rewards": rewards})
}

// ======================== 礼包码/十连抽/引导 ========================
