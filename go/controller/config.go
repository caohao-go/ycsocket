package controller

import (
	"context"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/model"
)

// 基础配置信息

func (c *ShinelightController) VerAction(ctx context.Context) *Result {
	return c.ResponseSuccessToAll(types.Map{"ver": "1.0.0"})
}

func (c *ShinelightController) SwitchAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	userGrade := model.GetUserAttr(userID)
	data := types.Map{}
	if userGrade != nil {
		regT := userGrade.GetStringE("reg_t")
		if regT < "2020-07-16 15:00:00" {
			data["guild_task_open"] = 0
		} else {
			data["guild_task_open"] = 1
		}
	}
	return c.ResponseSuccessToAll(data)
}

func (c *ShinelightController) GonggaoAction(ctx context.Context) *Result {
	userID := c.Params.GetInt64E("userid")
	content := c.Params.GetStringE("content")
	if content == "" {
		return c.ResponseError(13342339, "内容不能为空")
	}
	if userID != 99999999 {
		return c.ResponseError(133425539, "无权限")
	}
	return c.ResponseSuccessToAll(types.Map{"content": content})
}

// ======================== 配置和奖励数据 ========================

func (c *ShinelightController) ConfigAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := make(types.Map)

	// pk次数奖励
	if pkCountRewards := logic.GetRewardByCopyID(ctx, logic.RewardCopyIDPKCount); pkCountRewards != nil {
		pkRewards := make([]types.Map, 0)
		for _, v := range pkCountRewards {
			pkRewards = append(pkRewards, types.Map{
				"cnt":    v.RankCount.Max,
				"reward": v.Reward,
			})
		}
		data["pk_rewards"] = pkRewards
	}

	// 每日排名奖励
	if pkDailyRewards := logic.GetRewardByCopyID(ctx, logic.RewardCopyIDPKDaily); pkDailyRewards != nil {
		dailyRewards := make([]types.Map, 0)
		for _, v := range pkDailyRewards {
			dailyRewards = append(dailyRewards, types.Map{
				"min":    v.RankCount.Min,
				"max":    v.RankCount.Max,
				"reward": v.Reward,
			})
		}
		data["pk_daily_rewards"] = dailyRewards
	}

	// 赛季排名奖励
	if pkRankRewards := logic.GetRewardByCopyID(ctx, logic.RewardCopyIDPKRank); pkRankRewards != nil {
		rankRewards := make([]types.Map, 0)
		for _, v := range pkRankRewards {
			rankRewards = append(rankRewards, types.Map{
				"min":    v.RankCount.Min,
				"max":    v.RankCount.Max,
				"reward": v.Reward,
			})
		}
		data["pk_rank_rewards"] = rankRewards
	}

	// 无尽试炼奖励
	if endlessRewards := logic.GetRewardByCopyID(ctx, logic.RewardCopyIDEndless); endlessRewards != nil {
		endlessRewardsList := make([]types.Map, 0)
		for _, v := range endlessRewards {
			endlessRewardsList = append(endlessRewardsList, types.Map{
				"min":    v.RankCount.Min,
				"max":    v.RankCount.Max,
				"reward": v.Reward,
			})
		}
		data["endless_reward"] = endlessRewardsList
	}

	// 英雄远征总奖励 — 对齐 PHP Expedition::getTotalRewards()
	// PHP 返回: [{copy_id, name, open_type, rewards}, ...]
	expeditionRewardsList := make([]types.Map, 0)
	copyIDs := []int{20201, 20202, 20203}
	names := []string{"普通模式", "困难模式", "地狱模式"}
	// PHP: 普通模式 open_type=[], 困难模式 open_type={type:2,num:1000000}, 地狱模式 open_type={type:2,num:1800000}
	openTypes := []interface{}{
		[]types.Map{},
		types.Map{"type": 2, "num": 1000000},
		types.Map{"type": 2, "num": 1800000},
	}
	for idx, copyID := range copyIDs {
		rewards := make([]util.TypeNum, 0)
		if totalRewards, ok := logic.ExpeditionTotalRewards[copyID]; ok && len(totalRewards) > 0 {
			rewards = totalRewards
		}
		expeditionRewardsList = append(expeditionRewardsList, types.Map{
			"copy_id":   copyID,
			"name":      names[idx],
			"open_type": openTypes[idx],
			"rewards":   rewards,
		})
	}
	data["expedition_rewards"] = expeditionRewardsList

	// 探宝数据 — 对齐 PHP: $data['tanbao_10'] = Tanbao::$luck_datas[1]
	// PHP 返回 integral => [{'type': ..., 'num': ...}] 的 map 结构
	if luckData1, ok := logic.TanbaoLuckData[1]; ok {
		tanbao10 := make(types.Map)
		for integral, rewards := range luckData1 {
			tanbao10[types.ToString(integral)] = rewards
		}
		data["tanbao_10"] = tanbao10
	}
	if luckData2, ok := logic.TanbaoLuckData[2]; ok {
		tanbao12 := make(types.Map)
		for integral, rewards := range luckData2 {
			tanbao12[types.ToString(integral)] = rewards
		}
		data["tanbao_12"] = tanbao12
	}

	// 公会活跃奖励
	activeRewards := logic.GetActiveRewards()
	if len(activeRewards) == 0 {
		activeRewards = []types.Map{}
	}
	data["active_rewards"] = activeRewards

	// 公会战排名奖励
	recordRewards := make([]types.Map, 0)
	for _, v := range logic.GuildRecordRewardDatas {
		recordRewards = append(recordRewards, types.Map{
			"id":     v.ID,
			"rank":   v.Rank,
			"reward": v.Reward,
		})
	}
	if len(recordRewards) == 0 {
		recordRewards = []types.Map{}
	}
	data["guild_fight_rank_reward"] = recordRewards

	return c.ResponseSuccessToMe(data)
}

func (c *ShinelightController) RedPointInitAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 获取今日红点状态（对应 PHP RedPoint::getTodayPoint($userId)）
	data := model.GetTodayPoint(ctx, userID)
	// 追加排行榜红点（对应 PHP $data[] = array('type' => 1, 'id' => 4, 'pk' => 0, 'num' => 1)）
	data = append(data, types.Map{"type": 1, "id": 4, "pk": 0, "num": 1})

	return c.ResponseSuccessToMe(types.Map{"points": data})
}
