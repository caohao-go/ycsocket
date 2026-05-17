package controller

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"server_golang/common/lock"
	"server_golang/config"
	"server_golang/repo/mem/item"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/table"
)

// PK竞技系统

func (c *ShinelightController) GetpkOpAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	yindao := c.Params.GetIntE("yindao")

	myRank := model.GetMyTutengRank(ctx, userID)
	if myRank == 0 {
		return c.ResponseError(99, "system error")
	}
	maxRank := model.GetTutengMaxRank(ctx)

	var userIDs []int64
	if yindao > 0 {
		userIDs = []int64{127, 128, 129}
	} else {
		userIDs = model.GetDalilyPkUserID(ctx, userID)
	}

	data := make(types.Map)
	allocateUsers := make([]types.Map, 0)

	if len(userIDs) == 0 {
		// 分配排名
		allocateRank := allocatePKRanks(myRank, maxRank)

		uids := make([]int64, 0)
		for _, rank := range allocateRank {
			rankInfo := model.GetRankUserid(ctx, rank)
			tmp := types.Map{"rank": rank, "score": 0}
			if rankInfo != nil {
				tmp["userid"] = rankInfo["user_id"]
				tmp["score"] = rankInfo["score"]
				uid := rankInfo.GetInt64E("user_id")
				if uid > 0 {
					uids = append(uids, uid)
				}
			}
			allocateUsers = append(allocateUsers, tmp)
		}

		model.SetDalilyPkUserID(ctx, userID, uids)

		if len(uids) > 0 {
			userInfos := model.GetUsersWithDetail(ctx, uids, 2, config.AttrLv, config.AttrFightPoint)
			for k, au := range allocateUsers {
				uid := au.GetInt64E("userid")
				if info, ok := userInfos[uid]; ok {
					allocateUsers[k]["nickname"] = info["nickname"]
					allocateUsers[k]["avatar_url"] = info["avatar_url"]
					allocateUsers[k]["fight_point"] = info["fight_point"]
					allocateUsers[k]["lv"] = info["lv"]
				}
			}
		}
	} else {
		userInfos := model.GetUsersWithDetail(ctx, userIDs, 2, config.AttrLv, config.AttrFightPoint)
		for i, uid := range userIDs {
			if i >= 3 {
				break
			}
			rank := model.GetMyTutengRank(ctx, uid)
			score := model.GetMyTutengScore(ctx, uid)
			tmp := types.Map{"userid": uid, "rank": rank, "score": score}
			if info, ok := userInfos[uid]; ok {
				tmp["nickname"] = info["nickname"]
				tmp["avatar_url"] = info["avatar_url"]
				tmp["fight_point"] = info["fight_point"]
				tmp["lv"] = info["lv"]
			}
			allocateUsers = append(allocateUsers, tmp)
		}
	}

	var needTime int
	data["allocate_users"] = allocateUsers
	data["user_power"] = model.GetUserPower(ctx, userID, model.PowerTypePK, &needTime)
	data["free_times"] = model.GetRedisFreeTutengPkTimes(ctx, userID)
	data["quan_count"] = item.Total(userID, 21201)
	data["times"] = model.GetRedisRewardTimes(ctx, userID, logic.RewardCopyIDPKCount)
	data["rank"] = myRank
	data["score"] = model.GetMyTutengScore(ctx, userID)

	// 成就任务
	model.AchieveTaskHandle(ctx, userID, 8, model.GetMyTutengScore(ctx, userID), 5001, 5002)
	model.AchieveTaskHandle(ctx, userID, 22, int(myRank), 16001, 16004)

	data["reward"] = model.GetAllRewardStatus(ctx, userID, logic.RewardCopyIDPKCount)
	data["left_time"] = util.LeftTimeToWeekend2100()

	return c.ResponseSuccessToMe(data)
}

// 刷新PK对手
func (c *ShinelightController) PkOpAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	myRank := model.GetMyTutengRank(ctx, userID)
	if myRank == 0 {
		return c.ResponseError(99, "system error")
	}
	maxRank := model.GetTutengMaxRank(ctx)

	allocateRank := allocatePKRanks(myRank, maxRank)
	allocateUsers := make([]types.Map, 0)
	uids := make([]int64, 0)
	for _, rank := range allocateRank {
		rankInfo := model.GetRankUserid(ctx, rank)
		tmp := types.Map{"rank": rank, "score": 0}
		if rankInfo != nil {
			tmp["userid"] = rankInfo["user_id"]
			tmp["score"] = rankInfo["score"]
			uid := rankInfo.GetInt64E("user_id")
			if uid > 0 {
				uids = append(uids, uid)
			}
		}
		allocateUsers = append(allocateUsers, tmp)
	}
	model.SetDalilyPkUserID(ctx, userID, uids)

	if len(uids) > 0 {
		userInfos := model.GetUsersWithDetail(ctx, uids, 2, config.AttrLv, config.AttrFightPoint)
		for k, au := range allocateUsers {
			uid := au.GetInt64E("userid")
			if info, ok := userInfos[uid]; ok {
				allocateUsers[k]["nickname"] = info["nickname"]
				allocateUsers[k]["avatar_url"] = info["avatar_url"]
				allocateUsers[k]["fight_point"] = info["fight_point"]
				allocateUsers[k]["lv"] = info["lv"]
			}
		}
	}

	var needTime int
	data := types.Map{
		"allocate_users": allocateUsers,
		"user_power":     model.GetUserPower(ctx, userID, model.PowerTypePK, &needTime),
		"free_times":     model.GetRedisFreeTutengPkTimes(ctx, userID),
		"quan_count":     item.Total(userID, 21201),
		"times":          model.GetRedisRewardTimes(ctx, userID, logic.RewardCopyIDPKCount),
		"rank":           myRank,
		"score":          model.GetMyTutengScore(ctx, userID),
		"reward":         model.GetAllRewardStatus(ctx, userID, logic.RewardCopyIDPKCount),
	}
	return c.ResponseSuccessToMe(data)
}

// PK 战斗结果
func (c *ShinelightController) PkRetAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	oppUserID := c.Params.GetInt64E("opp_userid")

	lock.Lock(fmt.Sprintf("heroStarUp%d", userID), 3)
	defer lock.Unlock(fmt.Sprintf("heroStarUp%d", userID))

	userGrade := model.GetUserAttr(userID)
	if userGrade.GetIntE("lv") < 8 {
		return c.ResponseError(32159, "等级不够")
	}

	quanCount := item.Total(userID, 21201)
	if quanCount <= 0 {
		freeTimes := model.GetRedisFreeTutengPkTimes(ctx, userID)
		var needTime int
		power := model.GetUserPower(ctx, userID, model.PowerTypePK, &needTime)
		if freeTimes <= 0 && power <= 0 {
			return c.ResponseError(9411, "体力不够")
		}
	}

	// 获取己方英雄（阵位1，含属性+技能）
	myHeroAttrs := model.GetUserPositionWithHeroAttrs(ctx, userID, 1)

	// 获取对手英雄（阵位2优先，否则阵位1）
	oppHeroAttrs := model.GetUserPositionWithHeroAttrs(ctx, oppUserID, 2)
	if len(oppHeroAttrs) == 0 {
		oppHeroAttrs = model.GetUserPositionWithHeroAttrs(ctx, oppUserID, 1)
	}

	if len(myHeroAttrs) == 0 || myHeroAttrs[0].HeroInfo == 0 {
		return c.ResponseError(4345, "未配置英雄")
	}
	if len(oppHeroAttrs) == 0 || oppHeroAttrs[0].HeroInfo == 0 {
		return c.ResponseError(4422, "对手已经下线英雄，请选择另外对手")
	}

	// 扣费（校验英雄后再扣，避免无效消耗）
	if model.GetRedisFreeTutengPkTimes(ctx, userID) > 0 {
		model.UseRedisFreeTutengPkTimes(ctx, userID)
	} else {
		if quanCount <= 0 {
			return c.ResponseError(32158, "挑战卷不够")
		}
		item.Sub(userID, 21201, 1)
		var needTime int
		model.SubUserPower(ctx, userID, model.PowerTypePK, &needTime, 1)
	}

	fightHero := logic.GetBaseFromHero(myHeroAttrs)
	opFightHero := logic.GetBaseFromHero(oppHeroAttrs)

	fight := logic.NewFight(myHeroAttrs, oppHeroAttrs)
	winnerStr, fightResult := fight.FightExec(20)

	// 从战斗结果中提取胜利者
	winner := userID
	if winnerStr == "P1" {
		winner = userID
	} else {
		winner = oppUserID
	}

	model.DelDalilyPkUserID(ctx, userID)
	model.DelDalilyPkUserID(ctx, oppUserID)

	// 计算积分变化
	myScore := model.GetMyTutengScore(ctx, userID)
	oppScore := model.GetMyTutengScore(ctx, oppUserID)

	winScore := 0
	loseScore := 0
	winnerFlag := 1

	var rewardItems []*table.ItemsCollection
	if winner == userID { // 赢了
		winnerFlag = 1
		if myScore < oppScore {
			randomNum := rand.Intn(6) + 13
			winScore = randomNum + rand.Intn(4) + 2
			loseScore = -randomNum
		} else {
			randomNum := rand.Intn(6) + 1
			winScore = randomNum + rand.Intn(3) + 1
			loseScore = -randomNum
		}
		model.IncrTutengScore(ctx, userID, float64(winScore))
		model.IncrTutengScore(ctx, oppUserID, float64(loseScore))
		rewardItems = logic.GetRandCollectionItem(41, 1)
	} else {
		winnerFlag = 2
		rewardItems = logic.GetRandCollectionItem(40, 1)
	}

	// 奖励（PHP 仅取首项，且首项 items_id 非空时才插入）
	rewards := make([]util.TypeNum, 0)
	if len(rewardItems) > 0 {
		first := rewardItems[0]
		rewards = append(rewards, util.TypeNum{
			Type: first.ItemsId,
			Num:  first.Number,
		})
	}
	model.GiveReward(userID, rewards...)

	model.IncrRewardTimes(ctx, userID, logic.RewardCopyIDPKCount)

	// 插入PK记录（对齐 PHP：win_score / lose_score 均入库）
	model.InsertUserTutengPkDetail(ctx, &table.UserTutengPkDetail{
		UserId:    userID,
		OppUserId: oppUserID,
		Winner:    winnerFlag,
		PkTime:    time.Now().Format(time.DateTime),
		WinScore:  winScore,
		LoseScore: loseScore,
	})

	// 日常任务
	model.SetDailyTaskFinish(ctx, userID, 10004, 1)

	// 获取用户信息
	bothIDs := []int64{userID, oppUserID}
	userInfos := model.GetUsersWithDetail(ctx, bothIDs, 2, config.AttrLv)

	data := types.Map{
		"P1": userID, "P2": oppUserID, "winner": winner,
		"my_hero":      fightHero,
		"opp_hero":     opFightHero,
		"fight_result": fightResult,
		"rewards":      rewards,
		"my_score":     model.GetMyTutengScore(ctx, userID),
		"opp_score":    model.GetMyTutengScore(ctx, oppUserID),
		"win_score":    winScore,
		"lose_score":   loseScore,
	}
	if info, ok := userInfos[userID]; ok {
		data["my_lv"] = info["lv"]
		data["my_nickname"] = info["nickname"]
		data["my_avatar_url"] = info["avatar_url"]
	}
	if info, ok := userInfos[oppUserID]; ok {
		data["opp_lv"] = info["lv"]
		data["opp_nickname"] = info["nickname"]
		data["opp_avatar_url"] = info["avatar_url"]
	}

	model.GuideTaskHandle(ctx, userID, 37, 1)
	model.GuideTaskHandleIncr(ctx, userID, 51, 1)

	return c.ResponseSuccessToMe(data)
}

// PK记录列表
func (c *ShinelightController) PkListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	tmp, err := model.GetUserTutengPkDetail(ctx, userID)
	if err != nil {
		return c.ResponseError(99, err.Error())
	}

	data := types.ObjectsToMaps(tmp)

	for k, v := range data {
		if v.GetInt64E("user_id") == userID {
			data[k]["beatflag"] = 1
			data[k]["score"] = v.GetIntE("win_score")
		} else {
			data[k]["beatflag"] = 2
			data[k]["opp_user_id"] = v["user_id"]
			data[k]["score"] = v.GetIntE("lose_score")
		}
		delete(data[k], "win_score")
		delete(data[k], "lose_score")
		delete(data[k], "user_id")

		oppUID := data[k].GetInt64E("opp_user_id")
		oppInfo := model.GetUsersWithDetail(ctx, []int64{oppUID}, 1)
		if info, ok := oppInfo[oppUID]; ok {
			data[k]["nickname"] = info.GetStringE("nickname")
			data[k]["avatar_url"] = info.GetStringE("avatar_url")
		} else {
			data[k]["nickname"] = "陪练员"
			data[k]["avatar_url"] = ""
		}
	}

	return c.ResponseSuccessToMe(types.Map{"list": data})
}

// 当前PK排名
func (c *ShinelightController) PkRankAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	hideList := c.Params.GetIntE("hidelist")

	result := types.Map{
		"rank":  model.GetMyTutengRank(ctx, userID),
		"score": model.GetMyTutengScore(ctx, userID),
	}
	if hideList == 0 {
		result["list"] = model.GetTutengRankList(ctx, true, 0, 99)
	}
	return c.ResponseSuccessToMe(result)
}

// 我的PK排名
func (c *ShinelightController) MyPkRankAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	myRank := model.GetMyTutengRank(ctx, userID)
	return c.ResponseSuccessToMe(types.Map{
		"left_time": util.LeftTimeToWeekend2100(),
		"myrank":    myRank,
	})
}

// allocatePKRanks 分配PK对手排名
func allocatePKRanks(myRank, maxRank int64) []int64 {
	ranks := make([]int64, 0, 3)
	if myRank == 1 {
		ranks = []int64{2, 3, 4}
	} else if myRank == 2 {
		ranks = []int64{1, 3, 4}
	} else if myRank == 3 {
		ranks = []int64{1, 2, 4}
	} else if myRank <= 20 {
		ranks = []int64{myRank - 3, myRank - 2, myRank - 1}
	} else if myRank > maxRank-11 {
		r := int64(rand.Intn(8)) + myRank - 10
		ranks = []int64{r, myRank - 2, myRank - 1}
	} else if myRank < 100 {
		ranks = []int64{
			myRank - int64(rand.Intn(5)) - 6,
			myRank - int64(rand.Intn(4)) - 2,
			myRank + int64(rand.Intn(10)) + 1,
		}
		if ranks[2] >= maxRank {
			ranks[2] = myRank - 1
		}
	} else {
		ranks = []int64{
			int64(float64(myRank) * (0.91 + rand.Float64()*0.04)),
			int64(float64(myRank) * (0.96 + rand.Float64()*0.03)),
			int64(float64(myRank) * (1.01 + rand.Float64()*0.14)),
		}
		if ranks[2] >= maxRank {
			ranks[2] = myRank - 1
		}
	}
	// 确保排名 >= 1
	for i, r := range ranks {
		if r < 1 {
			ranks[i] = 1
		}
	}
	return ranks
}
