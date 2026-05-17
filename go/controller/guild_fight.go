package controller

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/cache"
)

// GetGuildFightAction 公会战信息（与 PHP getGuildFightAction 一致）
func (c *ShinelightController) GetGuildFightAction(ctx context.Context) *Result {
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

	data = model.GetGuildFightInfo(ctx, userID, guildID)

	// 排行榜
	changci := logic.GetGuildFightChangci()
	rankKey := fmt.Sprintf("%s_%d_%d", config.RankGuildFight, guildID, changci)
	data["rank_list"] = model.GetRankList(ctx, rankKey, true, 0, 2)

	return c.ResponseSuccessToMe(data)
}

// FightGuildAction 公会战战斗（与 PHP fightGuildAction 一致）
func (c *ShinelightController) FightGuildAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	star := c.Params.GetIntE("star")
	pos := c.Params.GetIntE("pos")
	position := c.Params.GetIntE("position")
	_, fightHeros := util.ToPosHeros(c.Params.GetStringE("fight_heros"))

	if time.Now().Hour() >= 21 {
		return c.ResponseError(8822, "行会战已结束")
	}

	if star < 0 || star > 3 {
		return c.ResponseError(399, "星数不对")
	}

	if len(fightHeros) == 0 {
		return c.ResponseError(6742, "请选择出战英雄")
	}

	guildID := model.GetUsersGuildID(ctx, userID)
	if guildID == 0 {
		return c.ResponseError(3288, "已经退出行会")
	}

	data := model.GetGuildFightInfo(ctx, userID, guildID)
	status := data.GetIntE("status")
	if status == 2 {
		return c.ResponseError(332, "战斗分配中")
	}
	if status != 1 {
		return c.ResponseError(432, "战斗已结束")
	}

	myPos := data.GetIntE("my_pos")
	if myPos == -1 {
		return c.ResponseError(499, "未参与本次行会战")
	}

	myGuild, _ := types.ToMap(data["my_guild"], "")
	opGuild, _ := types.ToMap(data["op_guild"], "")

	myUsers, _ := types.ToMapArray(myGuild["users"], "")
	opUsers, _ := types.ToMapArray(opGuild["users"], "")

	if myPos >= len(myUsers) {
		return c.ResponseError(499, "位置错误")
	}

	if types.Map(myUsers[myPos]).GetIntE("fight_count") >= 2 {
		return c.ResponseError(569, "次数用完")
	}

	if pos < 0 || pos >= len(opUsers) {
		return c.ResponseError(399, "pos is invalid")
	}

	if types.Map(opUsers[pos]).GetIntE("be_fight_count") >= 5 {
		return c.ResponseError(549, "最多只能被打5次")
	}

	if types.Map(opUsers[pos]).GetIntE("be_star")+star > 3 {
		return c.ResponseError(577, "该玩家正在被挑战，请稍后进入行会战查看战果。")
	}

	myGuildID := myGuild.GetIntE("guild_id")
	opGuildID := opGuild.GetIntE("guild_id")

	myGuildIDLock, _ := cache.Get(fmt.Sprintf(config.CacheLockGuildFight, myGuildID))
	opGuildIDLock, _ := cache.Get(fmt.Sprintf(config.CacheLockGuildFight, opGuildID))
	if myGuildIDLock != "" || opGuildIDLock != "" {
		return c.ResponseError(321, "有战斗正在激烈进行，请3秒后再试")
	}

	// 锁住双方
	cache.SetWithTTL(fmt.Sprintf(config.CacheLockGuildFight, myGuildID), "1", 5)
	cache.SetWithTTL(fmt.Sprintf(config.CacheLockGuildFight, opGuildID), "1", 5)

	// 获取我方英雄
	fightHeroDetails := c.getFightHeroByPosMap(ctx, userID, fightHeros)

	// 获取对手英雄
	opHeroData := types.ToMapArrayE(opUsers[pos]["heros"])
	opHerosTyped := make([]*logic.HeroBaseInfo, 0, len(opHeroData))
	for _, hm := range opHeroData {
		opHerosTyped = append(opHerosTyped, &logic.HeroBaseInfo{
			Id:        hm.GetIntE("id"),
			UserId:    hm.GetInt64E("user_id"),
			HeroId:    hm.GetIntE("hero_id"),
			Star:      hm.GetIntE("star"),
			Stage:     hm.GetIntE("stage"),
			Lv:        hm.GetIntE("lv"),
			Pos:       hm.GetIntE("pos"),
			Hp:        hm.GetIntE("hp"),
			CurrentHp: hm.GetIntE("current_hp"),
		})
	}
	opHeroDetail := model.GetFightHeroAttrWithSkill(ctx, opHerosTyped)

	fight := logic.NewFight(fightHeroDetails, opHeroDetail)
	winner, fightResult := fight.FightExec(15)

	model.SetFightHeros(ctx, userID, "guild_fight", fightHeros, position)

	// 判断胜负
	success := 0
	if winner == "P1" {
		success = 1
	}

	// 战斗日志
	changci := logic.GetGuildFightChangci()
	fightLog := types.Map{
		"success":        success,
		"time":           time.Now().Unix(),
		"star":           star,
		"add_zhanji":     0,
		"atk_nickname":   myUsers[myPos].GetStringE("nickname"),
		"def_nickname":   opUsers[pos].GetStringE("nickname"),
		"atk_guild_name": myGuild.GetStringE("guild_name"),
		"def_guild_name": opGuild.GetStringE("guild_name"),
		"atk_guild_star": myGuild.GetIntE("total_star"),
	}

	if success == 1 {
		fightLog["atk_guild_star"] = fightLog.GetIntE("atk_guild_star") + star

		if star == 0 {
			// 与 PHP 一致：star=0 的无伤胜利，累加 increate 属性加成
			increateVal := 2
			if pos < 5 {
				increateVal = 4
			}
			increate, _ := types.ToMap(myGuild["increate"], "")
			if increate == nil {
				increate = types.Map{"atk": 0, "def": 0, "hp": 0, "speed": 0}
			}
			increate["atk"] = increate.GetIntE("atk") + increateVal
			increate["def"] = increate.GetIntE("def") + increateVal
			increate["hp"] = increate.GetIntE("hp") + increateVal
			increate["speed"] = increate.GetIntE("speed") + increateVal
			myGuild["increate"] = increate

			model.AddZhanji(ctx, myGuildID, userID, 5)
			fightLog["add_zhanji"] = 5
		} else {
			// 更新星数
			myUsers[myPos]["star"] = types.Map(myUsers[myPos]).GetIntE("star") + star
			opUsers[pos]["be_star"] = types.Map(opUsers[pos]).GetIntE("be_star") + star

			// 计算战绩（与 PHP 一致：直接取 zhanji[star-1]['zhanji']，不做兜底）
			addZhanji := 0
			opZhanji := opUsers[pos]["zhanji"]
			if starIdx := star - 1; starIdx >= 0 {
				if types.IsArray(opZhanji) {
					zhanjiArr, err := types.ToArray(opZhanji)
					if err == nil && len(zhanjiArr) > starIdx {
						addZhanji = types.ToMapE(zhanjiArr[starIdx]).GetIntE("zhanji")
					}
				}
			}
			model.AddZhanji(ctx, myGuildID, userID, addZhanji)
			fightLog["add_zhanji"] = addZhanji
		}
	} else {
		model.AddZhanji(ctx, myGuildID, userID, 1)
		fightLog["add_zhanji"] = 1
		opUsers[pos]["suc_def"] = types.Map(opUsers[pos]).GetIntE("suc_def") + 1
	}

	// 更新战斗次数
	myUsers[myPos]["fight_count"] = types.Map(myUsers[myPos]).GetIntE("fight_count") + 1
	opUsers[pos]["be_fight_count"] = types.Map(opUsers[pos]).GetIntE("be_fight_count") + 1

	// 保存战斗信息
	myGuild["users"] = myUsers
	opGuild["users"] = opUsers
	model.SetGuildFightInfo(ctx, myGuildID, changci, myGuild)
	model.SetGuildFightInfo(ctx, opGuildID, changci, opGuild)

	// 更新星数记录
	model.SetGuildStars(ctx, myGuildID, changci, userID, myUsers[myPos].GetIntE("star"))

	// 记录日志
	model.AddGuildFightLog(ctx, myGuildID, changci, "1", fightLog)
	model.AddGuildFightLog(ctx, opGuildID, changci, "2", fightLog)
	model.AddMyGuildFightLog(ctx, userID, changci, "1", fightLog)
	opUserID := opUsers[pos].GetInt64E("user_id")
	model.AddMyGuildFightLog(ctx, opUserID, changci, "2", fightLog)

	// 解锁
	cache.Del(fmt.Sprintf(config.CacheLockGuildFight, myGuildID))
	cache.Del(fmt.Sprintf(config.CacheLockGuildFight, opGuildID))

	myHero := logic.GetBaseFromHero(fightHeroDetails)
	oppHero := logic.GetBaseFromHero(opHeroDetail)

	// 记录位置日志（与 PHP 一致：附加 my_hero/opp_hero/fight_result 供回看战斗）
	// Hero 已有 json tag，可直接 JSON 序列化
	posLog := types.CopyMap(fightLog)
	posLog["my_hero"] = fightHeroDetails
	posLog["opp_hero"] = opHeroDetail
	posLog["fight_result"] = fightResult
	model.AddPosGuildLog(ctx, opGuildID, pos, changci, posLog)

	return c.ResponseSuccessToMe(types.Map{
		"success":      success,
		"my_hero":      myHero,
		"opp_hero":     oppHero,
		"fight_result": fightResult,
	})
}

// FightRankAction 公会战排行榜（战绩排行，与 PHP fightRankAction 一致）
func (c *ShinelightController) FightRankAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	guildID := c.Params.GetIntE("guild_id")
	if guildID == 0 {
		guildID = model.GetUsersGuildID(ctx, userID)
	}

	changci := logic.GetGuildFightChangci()
	rankKey := fmt.Sprintf("%s_%d_%d", config.RankGuildFight, guildID, changci)

	// 获取我的排名、战绩分数、星数（对齐 PHP fightRankAction）
	myRank := model.GetMyRank(ctx, rankKey, userID)
	myZhanji := model.GetMyRankScore(ctx, rankKey, userID)
	myStar := model.HgetGuildStars(ctx, guildID, changci, userID)

	// 获取排行榜列表
	rankList := model.GetRankList(ctx, rankKey, true, 0, 99, "lv", "fight_point")

	// 获取全公会星数信息，遍历排行榜为每个用户附加 star 字段
	starInfo := model.GetGuildStars(ctx, guildID, changci)
	for i, item := range rankList {
		uid := item.GetStringE("user_id")
		rankList[i]["star"] = types.ToIntE(starInfo[uid])
	}

	return c.ResponseSuccessToMe(types.Map{
		"my_rank":   myRank,
		"my_zhanji": myZhanji,
		"my_star":   myStar,
		"rank_list": rankList,
	})
}

// GuildPosFightLogAction 公会战位置战斗日志（与 PHP guildPosFightLogAction 一致）
func (c *ShinelightController) GuildPosFightLogAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	guildID := c.Params.GetIntE("guild_id")
	pos := c.Params.GetIntE("pos")
	changci := logic.GetGuildFightChangci()

	data := model.GetPosGuildLog(ctx, guildID, pos, changci)
	return c.ResponseSuccessToMe(types.Map{"logs": data})
}

// GuildFightLogAction 公会战战斗日志（与 PHP guildFightLogAction 一致）
func (c *ShinelightController) GuildFightLogAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	guildID := c.Params.GetIntE("guild_id")
	changci := logic.GetGuildFightChangci()

	myData := model.GetMyGuildFightLog(ctx, userID, changci)
	logsData := model.GetGuildFightLog(ctx, guildID, changci)
	return c.ResponseSuccessToMe(types.Map{"my_log": myData, "logs": logsData})
}

// GuildDuizhanListAction 对战列表（与 PHP guildDuizhanListAction 一致）
func (c *ShinelightController) GuildDuizhanListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	guildID := model.GetUsersGuildID(ctx, userID)

	duizhanRaw := model.GetDuizhanInfo(ctx)
	duizhanMap, _ := types.ToMap(duizhanRaw, "")
	duizhanArr, _ := duizhanMap["duizhan"].([]interface{})

	guilds, _ := model.GetGuildList(ctx)
	guildsMap := make(map[int]*model.GuildInfo)
	for _, g := range guilds {
		guildsMap[g.Id] = g
	}

	resultData := make([]types.Map, 0, len(duizhanArr))
	var myData types.Map
	for _, v := range duizhanArr {
		vm, _ := types.ToMap(v, "")
		p1GuildID := vm.GetIntE("p1")
		p2GuildID := vm.GetIntE("p2")

		entry := types.Map{
			"p1":                p1GuildID,
			"p2":                p2GuildID,
			"p1_name":           guildsMap[p1GuildID].GuildName,
			"p2_name":           guildsMap[p2GuildID].GuildName,
			"p1_own_nickname":   guildsMap[p1GuildID].OwnNickname,
			"p2_own_nickname":   guildsMap[p2GuildID].OwnNickname,
			"p1_own_avatar_url": guildsMap[p1GuildID].OwnAvatarURL,
			"p2_own_avatar_url": guildsMap[p2GuildID].OwnAvatarURL,
		}
		resultData = append(resultData, entry)

		if guildID == p1GuildID || guildID == p2GuildID {
			myData = entry
		}
	}

	return c.ResponseSuccessToMe(types.Map{"data": resultData, "mydata": myData})
}

// GuildLogAction 公会日志（与 PHP guildLogAction 一致 - 返回空数组）
func (c *ShinelightController) GuildLogAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	return c.ResponseSuccessToMe(types.Map{"data": []interface{}{}})
}
