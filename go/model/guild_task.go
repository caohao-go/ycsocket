package model

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo"
	"server_golang/repo/mem/daily"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

// GetGuildTask 获取公会任务列表
func GetGuildTask(ctx context.Context, userid int64) []table.GuildTask {
	ret := make([]table.GuildTask, 0)
	for taskID, task := range logic.GuildTaskDatas {
		tmp := task
		tmp.FinishCount = GetTaskFinishNumStr(ctx, userid, fmt.Sprint(taskID), task.TaskCdTime)
		ret = append(ret, tmp)
	}
	return ret
}

// FinishGuildTask 完成公会任务
func FinishGuildTask(ctx context.Context, userid int64, taskID int) {
	task, ok := logic.GuildTaskDatas[taskID]
	if !ok {
		return
	}
	cdTime := task.TaskCdTime
	IncrTaskFinishNumStr(ctx, userid, "guild_"+fmt.Sprint(taskID), cdTime)
}

// InsertGuildChapterBlood 插入公会副本章节血量
func InsertGuildChapterBlood(ctx context.Context, data *table.GuildChapterBlood) (int64, error) {
	return world.InsertGuildChapterBlood(ctx, data)
}

// GetGuildChapterBlood 获取公会副本章节血量
func GetGuildChapterBlood(ctx context.Context, guildID int) *table.GuildChapterBlood {
	return world.GetGuildChapterBlood(ctx, guildID)
}

// UpdateGuildChapterBlood 更新公会副本章节血量
func UpdateGuildChapterBlood(ctx context.Context, guildID int, chapter, chapterBlood int) error {
	return world.UpdateGuildChapterBlood(ctx, guildID, chapter, chapterBlood)
}

// GetCopyInfo 获取公会副本信息
func GetCopyInfo(ctx context.Context, userid int64, guildID int, guidCopyChapter int, guidCopyChapterBlood int) types.Map {
	v, _ := daily.Get(userid, config.DailyGuildCopyCount)
	copyCount := types.ToIntE(v)

	data := make(types.Map)
	freeTimesVal := 2 - copyCount
	if freeTimesVal < 0 {
		freeTimesVal = 0
	}
	data["free_times"] = freeTimesVal

	dataTemp := 4 - (copyCount - (2 - freeTimesVal))
	if dataTemp < 0 {
		dataTemp = 0
	}
	data["vip_times"] = dataTemp
	data["cost_type"] = 2
	data["basic_cost"] = 20

	data["last_copy_harm_blood"] = GetLastCopyHarmBlood(ctx, userid, guildID)

	atkAdd, leftTime := GetGuildCopyAtkAdd(ctx, guildID)
	data["atk_add"] = atkAdd
	data["atk_add_left_time"] = leftTime
	data["chapter"] = guidCopyChapter
	data["chapter_blood"] = logic.GetCopyChapterHP(guidCopyChapter)
	data["chapter_current_blood"] = guidCopyChapterBlood

	rewards := logic.GetGuildCopyRewards(1, guidCopyChapter)
	if len(rewards) > 0 {
		data["harm_reward"] = rewards[0].Reward
	}
	rewards = logic.GetGuildCopyRewards(2, guidCopyChapter)
	if len(rewards) > 0 {
		data["hit_reward"] = rewards[0].Reward
	}
	data["rank_reward"] = logic.GetGuildCopyRewards(4, guidCopyChapter)
	return data
}

func GetGuildCopyAtkAdd(ctx context.Context, gid int) (int, int) {
	k := fmt.Sprintf(config.KeyGuildCopyAddAtk, gid)
	data, _ := repo.RedisHGetAll(ctx, k)
	totalAdd, maxTime, now := 0, 0, int(time.Now().Unix())
	for ts, addStr := range data {
		t := types.ToIntE(ts)
		if now > t {
			repo.RedisHDel(ctx, k, ts)
			continue
		}
		totalAdd += types.ToIntE(addStr)
		if t > maxTime {
			maxTime = t
		}
	}
	left := 0
	if maxTime != 0 {
		left = maxTime - now
	}
	return totalAdd, left
}

func AddGuildCopyAtkAdd(ctx context.Context, gid int) {
	k := fmt.Sprintf(config.KeyGuildCopyAddAtk, gid)
	add, maxTime := GetGuildCopyAtkAdd(ctx, gid)
	if maxTime != 0 {
		repo.RedisExpire(ctx, k, maxTime)
	}
	if add >= 20 {
		return
	}
	repo.RedisHIncrBy(ctx, k, time.Now().Unix()+7200, 2)
}

func GetLastCopyHarmBlood(ctx context.Context, uid int64, gid int) int {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyLastCopyHarmBlood, uid, gid, util.DateYmd()))
	return types.ToIntE(v)
}

func SetLastCopyHarmBlood(ctx context.Context, uid int64, gid, blood int) {
	k := fmt.Sprintf(config.KeyLastCopyHarmBlood, uid, gid, util.DateYmd())
	repo.RedisSet(ctx, k, blood, 86400)
}

func GetGuildCopyCount(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyGuildCopyCount)
	return types.ToIntE(v)
}

// IncrGuildCopyCount 增加公会副本次数
func IncrGuildCopyCount(ctx context.Context, uid int64) {
	daily.Incr(uid, config.DailyGuildCopyCount, 1)
}
