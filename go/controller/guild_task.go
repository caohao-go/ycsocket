package controller

import (
	"context"
	"fmt"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/cache"
	"server_golang/repo/mem/item"
)

// GuildCopyAction 公会副本信息（与 PHP guildCopyAction 一致）
func (c *ShinelightController) GuildCopyAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		return c.ResponseSuccessToMe(types.Map{"status": 2})
	}

	guildChapter := model.GetGuildChapterBlood(ctx, guildID)
	if guildChapter == nil {
		return c.ResponseError(99, "公会副本数据不存在")
	}

	data := types.Map{"status": 0}
	copyInfo := model.GetCopyInfo(ctx, userID, guildID, guildChapter.Chapter, guildChapter.ChapterBlood)
	for k, v := range copyInfo {
		data[k] = v
	}

	// 排行榜
	rankKey := fmt.Sprintf("%s_%d_%d", config.RankGuildHarm, guildID, guildChapter.Chapter)
	data["rank_list"] = model.GetRankList(ctx, rankKey, true, 0, 2)

	return c.ResponseSuccessToMe(data)
}

// UpGuildCopyAction 公会副本战斗（与 PHP upGuildCopyAction 一致）
func (c *ShinelightController) UpGuildCopyAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	position := c.Params.GetIntE("position")
	_, fightHeros := util.ToPosHeros(c.Params.GetStringE("fight_heros"))

	if len(fightHeros) == 0 {
		return c.ResponseError(6742, "请选择出战英雄")
	}

	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		return c.ResponseError(9382, "未加入行会")
	}

	guildChapter := model.GetGuildChapterBlood(ctx, guildID)
	if guildChapter == nil {
		return c.ResponseError(99, "system error")
	}

	copyInfo := model.GetCopyInfo(ctx, userID, guildID, guildChapter.Chapter, guildChapter.ChapterBlood)
	if copyInfo == nil {
		return c.ResponseError(99, "system error")
	}

	freeTimes := copyInfo.GetIntE("free_times")
	vipTimes := copyInfo.GetIntE("vip_times")
	if freeTimes <= 0 && vipTimes <= 0 {
		return c.ResponseError(88888, "今日挑战次数用完了")
	}

	isCopyIng, _ := cache.Get(fmt.Sprintf(config.CacheLockGuildCopy, guildID))
	if isCopyIng != "" {
		return c.ResponseError(9021, "有战斗正在激烈进行，请3秒后再试")
	}

	// 免费次数用完，需要消耗钻石
	if freeTimes <= 0 {
		costType := copyInfo.GetIntE("cost_type")
		basicCost := copyInfo.GetIntE("basic_cost")
		if item.NotEnough(userID, costType, basicCost) {
			return c.ResponseError(666666, "钻石不够")
		}
		item.Sub(userID, costType, basicCost)
	}

	// 锁住副本
	cache.SetWithTTL(fmt.Sprintf(config.CacheLockGuildCopy, guildID), "1", 5)

	// 获取我方英雄
	fightHeroDetails := c.getFightHeroByPosMap(ctx, userID, fightHeros)

	// 获取Boss
	chapter := copyInfo.GetIntE("chapter")
	bossConfig := logic.GuildBossConfigDatas[chapter]
	bossID := bossConfig.BossId
	opHeroDetail := logic.GetMonstersByCopyBoss(bossID)

	// 去掉BOSS强化加血技能（与 PHP 一致）
	for k := range opHeroDetail {
		filtered := make([]*logic.Skill, 0)
		for _, s := range opHeroDetail[k].Skills {
			if s.Type != 3 {
				filtered = append(filtered, s)
			}
		}
		opHeroDetail[k].Skills = filtered
	}

	// 设置Boss当前血量
	chapterBlood := copyInfo.GetIntE("chapter_blood")
	chapterCurrentBlood := copyInfo.GetIntE("chapter_current_blood")
	if len(opHeroDetail) > 0 {
		opHeroDetail[0].CurrentHP = chapterCurrentBlood
		opHeroDetail[0].Hp = chapterBlood
	}

	// 真实战斗（5回合上限，与 PHP 一致）
	fight := logic.NewFight(fightHeroDetails, opHeroDetail)
	_, fightResult := fight.FightExec(5)

	// 获取boss剩余血量
	leftBlood := 0
	for _, ha := range fight.HerosAttr["P2"] {
		hp := ha.CurrentHP
		if hp < 0 {
			hp = 0
		}
		leftBlood = hp
		break
	}

	// 计算伤害
	harmBlood := guildChapter.ChapterBlood - leftBlood
	if harmBlood <= 0 {
		harmBlood = 1
	}

	// 增加伤害排行
	rankKey := fmt.Sprintf("%s_%d_%d", config.RankGuildHarm, guildID, chapter)
	model.IncrRankScore(ctx, rankKey, userID, float64(harmBlood), 259200)

	// 重新获取最新章节数据（并发安全）
	latestChapter := model.GetGuildChapterBlood(ctx, guildID)
	if latestChapter != nil {
		if harmBlood < latestChapter.ChapterBlood {
			model.UpdateGuildChapterBlood(ctx, guildID, latestChapter.Chapter, latestChapter.ChapterBlood-harmBlood)
		} else {
			// 击杀Boss，进入下一章
			newChapter := latestChapter.Chapter + 1
			model.UpdateGuildChapterBlood(ctx, guildID, newChapter, logic.GetCopyChapterHP(newChapter))

			// 邮件发送击杀奖励+排行奖励（与 PHP upGuildCopyAction 一致）
			model.SendGuildBossKillMail(ctx, userID, guildID, latestChapter.Chapter, copyInfo["hit_reward"])

			// 成就任务 - 击杀公会boss
			model.IncrKillGuildbossNum(userID)
			killNum := model.GetKillGuildbossNum(userID)
			model.AchieveTaskHandle(ctx, userID, 16, killNum, 11101, 11107)
		}
	}

	model.SetLastCopyHarmBlood(ctx, userID, guildID, harmBlood)
	model.SetFightHeros(ctx, userID, "guild_copy", fightHeros, position)

	// 解锁战斗
	cache.Del(fmt.Sprintf(config.CacheLockGuildCopy, guildID))

	// 发放伤害奖励
	harmReward, _ := copyInfo["harm_reward"].([]util.TypeNum)
	if len(harmReward) > 0 {
		model.GiveReward(userID, harmReward...)
	}

	myHero := logic.GetBaseFromHero(fightHeroDetails)
	oppHero := logic.GetBaseFromHero(opHeroDetail)

	model.IncrGuildCopyCount(ctx, userID)

	// 公会活跃（与 PHP 一致：task_id=3 公会副本）
	tasks := model.GetGuildTask(ctx, userID)
	for _, v := range tasks {
		if v.TaskId == 3 {
			finishCount := v.FinishCount
			taskCountLimit := v.TaskCountLimit
			if finishCount < taskCountLimit && finishCount >= 0 {
				model.IncrTaskFinishNumStr(ctx, userID, "guild_3", 1)
				model.IncrUsersContentInt(ctx, userID, "guild_active", 5)
			}
			break
		}
	}

	// 日常任务（与 PHP 一致：10007）
	model.SetDailyTaskFinish(ctx, userID, 10007, 1)

	// 成就任务 - 挑战公会boss
	model.IncrGuildbossNum(userID)
	guildBossNum := model.GetGuildbossNum(userID)
	model.AchieveTaskHandle(ctx, userID, 15, guildBossNum, 11001, 11004)

	// 判断胜负
	success := 0
	if len(fightResult) > 0 {
		lastRound := fightResult[len(fightResult)-1]
		if ops, err := types.ToMapArray(lastRound["ops"], ""); err == nil && len(ops) > 0 {
			lastOp := ops[len(ops)-1]
			if ret := types.ToString(lastOp["ret"]); ret == "P1" {
				success = 2
			} else if ret == "P" {
				success = 1
			}
		}
	}

	retData := types.Map{
		"success":      success,
		"harm_blood":   harmBlood,
		"rewards":      copyInfo["harm_reward"],
		"my_hero":      myHero,
		"opp_hero":     oppHero,
		"fight_result": fightResult,
	}

	// 引导任务
	model.GuideTaskHandle(ctx, userID, 97, 1)

	return c.ResponseSuccessToMe(retData)
}

// SaodangGuildCopyAction 公会副本扫荡（与 PHP saodangGuildCopyAction 一致）
func (c *ShinelightController) SaodangGuildCopyAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		return c.ResponseError(9382, "未加入行会")
	}

	guildChapter := model.GetGuildChapterBlood(ctx, guildID)
	if guildChapter == nil {
		return c.ResponseError(99, "system error")
	}

	copyInfo := model.GetCopyInfo(ctx, userID, guildID, guildChapter.Chapter, guildChapter.ChapterBlood)
	if copyInfo == nil {
		return c.ResponseError(99, "system error")
	}

	// 上次伤害（与 PHP 一致：没打过不能扫荡）
	harmBlood := model.GetLastCopyHarmBlood(ctx, userID, guildID)
	if harmBlood == 0 {
		return c.ResponseError(3299, "还未打BOSS")
	}

	freeTimes := copyInfo.GetIntE("free_times")
	vipTimes := copyInfo.GetIntE("vip_times")
	if freeTimes <= 0 && vipTimes <= 0 {
		return c.ResponseError(88888, "今日挑战次数用完了")
	}

	if freeTimes <= 0 {
		costType := copyInfo.GetIntE("cost_type")
		basicCost := copyInfo.GetIntE("basic_cost")
		if item.NotEnough(userID, costType, basicCost) {
			return c.ResponseError(666666, "钻石不够")
		}
		item.Sub(userID, costType, basicCost)
	}

	model.IncrGuildCopyCount(ctx, userID)

	// 公会活跃（与 PHP 一致）
	tasks := model.GetGuildTask(ctx, userID)
	for _, v := range tasks {
		if v.TaskId == 3 {
			finishCount := v.FinishCount
			taskCountLimit := v.TaskCountLimit
			if finishCount < taskCountLimit && finishCount >= 0 {
				model.IncrTaskFinishNumStr(ctx, userID, "guild_3", 1)
				model.IncrUsersContentInt(ctx, userID, "guild_active", 5)
			}
			break
		}
	}

	// 增加伤害排行
	chapter := guildChapter.Chapter
	rankKey := fmt.Sprintf("%s_%d_%d", config.RankGuildHarm, guildID, chapter)
	model.IncrRankScore(ctx, rankKey, userID, float64(harmBlood), 259200)

	// 更新章节血量
	latestChapter := model.GetGuildChapterBlood(ctx, guildID)
	if latestChapter != nil {
		if harmBlood < latestChapter.ChapterBlood {
			model.UpdateGuildChapterBlood(ctx, guildID, latestChapter.Chapter, latestChapter.ChapterBlood-harmBlood)
		} else {
			newChapter := latestChapter.Chapter + 1
			model.UpdateGuildChapterBlood(ctx, guildID, newChapter, logic.GetCopyChapterHP(newChapter))

			// 邮件发送击杀奖励+排行奖励（与 PHP saodangGuildCopyAction 一致）
			model.SendGuildBossKillMail(ctx, userID, guildID, latestChapter.Chapter, copyInfo["hit_reward"])

			model.IncrKillGuildbossNum(userID)
			killNum := model.GetKillGuildbossNum(userID)
			model.AchieveTaskHandle(ctx, userID, 16, killNum, 11101, 11107)
		}
	}

	// 日常任务（与 PHP 一致：10007）
	model.SetDailyTaskFinish(ctx, userID, 10007, 1)

	return c.ResponseSuccessToMe(types.Map{"harm_blood": harmBlood, "rewards": copyInfo["harm_reward"]})
}

// GuildCopyRankAction 公会副本排行（与 PHP guildCopyRankAction 一致）
func (c *ShinelightController) GuildCopyRankAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	chapter := c.Params.GetIntE("chapter")

	data := types.Map{}
	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		data["status"] = 2
		return c.ResponseSuccessToMe(data)
	}

	data["status"] = 0
	if chapter <= 1 {
		guildChapter := model.GetGuildChapterBlood(ctx, guildID)
		if guildChapter != nil {
			chapter = guildChapter.Chapter
		}
	}

	rankKey := fmt.Sprintf("%s_%d_%d", config.RankGuildHarm, guildID, chapter)
	data["my_rank"] = model.GetMyRank(ctx, rankKey, userID)
	data["my_harm"] = model.GetMyRankScore(ctx, rankKey, userID)
	data["rank_list"] = model.GetRankList(ctx, rankKey, true, 0, 99)

	return c.ResponseSuccessToMe(data)
}

// GuildCopyAddAtkAction 公会副本加成（与 PHP guildCopyAddAtkAction 一致）
func (c *ShinelightController) GuildCopyAddAtkAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		return c.ResponseError(9382, "未加入行会")
	}

	// 消耗钻石 20（与 PHP 一致，不是50）
	if item.NotEnough(userID, 2, 20) {
		return c.ResponseError(666666, "钻石不够")
	}
	item.Sub(userID, 2, 20)

	// 增加公会副本攻击加成（与 PHP 一致）
	model.AddGuildCopyAtkAdd(ctx, guildID)

	// 公会活跃（与 PHP 一致：task_id=1 公会副本加成）
	tasks := model.GetGuildTask(ctx, userID)
	for _, v := range tasks {
		if v.TaskId == 1 {
			finishCount := v.FinishCount
			taskCountLimit := v.TaskCountLimit
			if finishCount < taskCountLimit && finishCount >= 0 {
				model.IncrTaskFinishNumStr(ctx, userID, "guild_1", 7)
				model.IncrUsersContentInt(ctx, userID, "guild_active", 50)
			}
			break
		}
	}

	model.GuideTaskHandle(ctx, userID, 95, 1)

	return c.ResponseSuccessToMe(types.Map{})
}
