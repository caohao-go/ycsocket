package controller

import (
	"context"

	"server_golang/common/types"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/table"
)

// 排行榜

func (c *ShinelightController) RankBestAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := make(types.Map)

	// 剧情进度 Top1
	copyRank := model.GetRankList(ctx, config.RankCopy, true, 0, 1)
	if len(copyRank) > 0 {
		data["copy"] = copyRank[0]
	}
	// 爬塔 Top1
	towerRank := model.GetRankList(ctx, config.RankClimbtower, true, 0, 1)
	if len(towerRank) > 0 {
		data["climbtower_layer"] = towerRank[0]
	}
	// 公会 Top1
	guildRank, _ := model.GetGuildRank(ctx)
	if len(guildRank) > 0 {
		data["guild_fight"] = types.Map{
			"user_id":    0,
			"nickname":   guildRank[0].GuildName,
			"avatar_url": guildRank[0].OwnAvatarURL,
			"score":      guildRank[0].FightPoint,
		}
	}
	// 竞技场 Top1
	pkRank := model.GetTutengRankList(ctx, true, 0, 99)
	if len(pkRank) > 0 {
		data["pk"] = pkRank[0]
	}
	// 战力 Top1
	fpRank := model.GetRankList(ctx, config.RankFightPoint, true, 0, 1)
	if len(fpRank) > 0 {
		data["fight_point"] = fpRank[0]
	}

	// 最强英雄 — 对齐 PHP: 从 hero_fight_point 排行榜取 Top1
	data["best_hero"] = nil
	scoreRanks := model.GetTopFightpointHero(ctx)
	if len(scoreRanks) > 0 {
		heroIDs := make([]int, 0, len(scoreRanks))
		for _, sm := range scoreRanks {
			heroIDs = append(heroIDs, types.ToIntE(sm.Member))
		}
		userHeros := model.GetUserHeroByIDs(ctx, heroIDs)
		if len(userHeros) > 0 {
			// 构建英雄映射和收集用户ID
			userheroMap := make(map[int64]*table.UserHero)
			userIDs := make([]int64, 0)
			for _, uh := range userHeros {
				userheroMap[int64(uh.Id)] = uh
				userIDs = append(userIDs, uh.UserId)
			}

			// 批量获取用户信息（带 user_grade 的 lv / vip_level，对齐 PHP）
			userInfos := model.GetUsersWithDetail(ctx, userIDs, 1, config.AttrLv, config.AttrVipLevel)

			// 取第一条（最高分）
			top := scoreRanks[0]
			if uh, ok := userheroMap[types.ToInt64E(top.Member)]; ok {
				userID := uh.UserId
				tmp := types.Map{
					"id":          types.ToInt64E(top.Member),
					"fight_point": int(top.Score),
					"hero_lv":     uh.Lv,
					"hero_id":     uh.HeroId,
					"star":        uh.Star,
					"property":    logic.HeroProperty(uh.HeroId),
					"userid":      userID,
				}
				if info, ok := userInfos[userID]; ok {
					tmp["nickname"] = info["nickname"]
					tmp["avatar_url"] = info["avatar_url"]
					tmp["user_lv"] = info.GetIntE("lv")
					tmp["vip_level"] = info.GetIntE("vip_level")
				}
				data["best_hero"] = tmp
			}
		}
	}

	return c.ResponseSuccessToMe(data)
}

// 排行榜列表
func (c *ShinelightController) RankListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	rankType := c.Params.GetIntE("type")

	data := make(types.Map)

	realType := config.RankCopy
	switch rankType {
	case 1:
		realType = config.RankCopy
	case 2:
		realType = config.RankClimbtower
	case 3:
		realType = config.RankGuildFight
	case 4:
		realType = "pk"
	case 5:
		realType = config.RankFightPoint
	}

	if rankType == 4 {
		data["my_rank"] = model.GetMyTutengRank(ctx, userID)
		data["my_score"] = model.GetMyTutengScore(ctx, userID)
		data["rank_list"] = model.GetTutengRankList(ctx, true, 0, 99)
	} else {
		data["my_rank"] = model.GetMyRank(ctx, realType, userID)
		data["my_score"] = model.GetMyRankScore(ctx, realType, userID)
		data["rank_list"] = model.GetRankList(ctx, realType, true, 0, 99, "lv", "vip_level")
	}

	return c.ResponseSuccessToMe(data)
}

// 英雄战力排行榜 — 对齐 PHP heroRankListAction: 从 Redis ZSet 读取英雄排行，并丰化英雄+用户信息
func (c *ShinelightController) HeroRankListAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	prop := c.Params.GetIntE("prop")

	scoreRanks := model.GetFightpointRank(ctx, prop)

	rankList := make([]types.Map, 0)
	if len(scoreRanks) > 0 {
		// 收集英雄记录ID（保持 Redis 返回的排序）
		heroIDs := make([]int, 0, len(scoreRanks))
		for _, sm := range scoreRanks {
			heroIDs = append(heroIDs, types.ToIntE(sm.Member))
		}

		// 批量查询英雄信息
		userHeros := model.GetUserHeroByIDs(ctx, heroIDs)
		userheroMap := make(map[int64]*table.UserHero)
		userIDs := make([]int64, 0)
		for _, uh := range userHeros {
			userheroMap[int64(uh.Id)] = uh
			userIDs = append(userIDs, uh.UserId)
		}

		// 批量获取用户信息（带 user_grade 的 lv / vip_level，对齐 PHP ['lv','vip_level']）
		userInfos := model.GetUsersWithDetail(ctx, userIDs, 1, config.AttrLv, config.AttrVipLevel)

		n := 0
		for _, sm := range scoreRanks {
			var heroRecordID = types.ToInt64E(sm.Member)
			uh, ok := userheroMap[heroRecordID]
			if !ok {
				continue
			}
			userID := uh.UserId
			info, infoOK := userInfos[userID]
			if !infoOK {
				continue
			}
			n++
			if n > 100 {
				break
			}
			tmp := types.Map{
				"id":          heroRecordID,
				"fight_point": int(sm.Score),
				"hero_lv":     uh.Lv,
				"hero_id":     uh.HeroId,
				"star":        uh.Star,
				"property":    logic.HeroProperty(uh.HeroId),
				"userid":      userID,
				"nickname":    info["nickname"],
				"avatar_url":  info["avatar_url"],
				"user_lv":     info.GetIntE("lv"),
				"vip_level":   info.GetIntE("vip_level"),
			}
			rankList = append(rankList, tmp)
		}
	}

	return c.ResponseSuccessToMe(types.Map{"rank_list": rankList})
}
