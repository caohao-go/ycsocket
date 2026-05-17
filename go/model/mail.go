package model

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

// 邮件系统

// ---- 邮件系统 ----

// GetUserMail 获取用户邮件
// 与 PHP getUserMail 对齐：
//   - add_items 由 JSON 字符串反序列化为数组（空字符串返回空数组）
//   - send_time 由 "Y-m-d H:i:s" 转为 Unix 秒级时间戳
//   - 移除 updatetime 字段
func GetUserMail(ctx context.Context, userID int64) []types.Map {
	rows := world.GetUserMailListWithCache(ctx, userID)
	if len(rows) == 0 {
		return nil
	}
	result := types.ObjectsToMaps(rows)
	for i := range result {
		// add_items: string -> []interface{}
		rawItems, _ := result[i]["add_items"].(string)
		if rawItems == "" {
			result[i]["add_items"] = []types.Map{}
		} else {
			items := types.ToMapArrayE(rawItems)
			if len(items) > 0 {
				result[i]["add_items"] = items
			} else {
				result[i]["add_items"] = []types.Map{}
			}
		}
		// send_time: "Y-m-d H:i:s" -> unix
		if sendStr, ok := result[i]["send_time"].(string); ok && sendStr != "" {
			if t, err := time.ParseInLocation("2006-01-02 15:04:05", sendStr, time.Local); err == nil {
				result[i]["send_time"] = t.Unix()
			}
		}
		// 去除 updatetime 字段（与 PHP unset 一致）
		delete(result[i], "updatetime")
	}
	return result
}

// GetMailsByIDs 根据ID列表获取邮件
func GetMailsByIDs(ctx context.Context, userID int64, ids []int64) []*table.UserMail {
	rows := world.GetUserMailByIds(ctx, userID, ids)
	if rows == nil {
		return nil
	}
	return rows
}

// ReadUserMail 标记邮件已读（与 PHP readUserMail 一致：更新 read_flag=1 并清缓存）
func ReadUserMail(ctx context.Context, userID, id int64) error {
	return world.UpdateUserMailByIdWithCache(ctx, userID, id, types.Map{"read_flag": 1})
}

// DelUserMail 删除邮件
func DelUserMail(ctx context.Context, userID int64, ids []int64) error {
	return world.DeleteUserMailByIdsWithCache(ctx, userID, ids)
}

// LingquMail 领取邮件附件 - 批量标记为已领取
// 与 PHP lingquMail 一致：单条 SQL 批量更新 item_get_flag=2 并清缓存
func LingquMail(ctx context.Context, userID int64, ids []int64) {
	if len(ids) == 0 {
		return
	}
	_ = world.LingquMailByIdsWithCache(ctx, userID, ids)
}

// InsertUserMail 插入邮件（与 PHP insertUserMail 一致：插入DB + 清缓存 + 发红点）
// addItems 格式: [{"type": 1, "num": 1000000}, {"type": 2, "num": 300}]
func InsertUserMail(ctx context.Context, userID int64, title, content string, addItems []types.Map) {
	itemGetFlag := 0
	addItemsJSON := ""
	if len(addItems) > 0 {
		itemGetFlag = 1 // 0-无物品 1-未领取 2-已领取
		addItemsJSON = json.Marshal(addItems)
	}

	world.InsertUserMail(ctx, userID, title, content, time.Now().Format("2006-01-02 15:04:05"), itemGetFlag, addItemsJSON)

	// 发送红点提醒（id=2 即邮件红点，与 PHP Redpoint::send($userid, 2) 一致）
	RedpointSend(ctx, userID, 2, 0, 1, 1)
}

// SendGuildBossKillMail 发送公会副本击杀奖励邮件（与 PHP upGuildCopyAction/saodangGuildCopyAction 一致）
func SendGuildBossKillMail(ctx context.Context, userID int64, guildID, chapter int, hitReward interface{}) {
	// 1. 击杀奖励邮件发给击杀者
	var hitRewardItems []types.Map
	switch r := hitReward.(type) {
	case []util.TypeNum:
		for _, pair := range r {
			hitRewardItems = append(hitRewardItems, types.Map{"type": pair.Type, "num": pair.Num})
		}
	case []types.Map:
		hitRewardItems = r
	}

	InsertUserMail(ctx, userID, "行会副本击杀奖励",
		fmt.Sprintf("恭喜你，击杀了第%d章节BOSS，奖励如下：\n", chapter), hitRewardItems)

	// 2. 排行奖励邮件发给前50名
	rankKey := fmt.Sprintf("%s_%d_%d", config.RankGuildHarm, guildID, chapter)
	rankList := GetRankList(ctx, rankKey, true, 0, 49)
	for index, rankUser := range rankList {
		rank := index + 1
		rankRewards := logic.GetCopyRankRewards(rank, chapter)
		if len(rankRewards) > 0 {
			rankUserID := rankUser.GetInt64E("user_id")
			rewardsMap := make([]types.Map, 0, len(rankRewards))
			for _, rr := range rankRewards {
				rewardsMap = append(rewardsMap, types.Map{"type": rr.Type, "num": rr.Num})
			}
			InsertUserMail(ctx, rankUserID, "章节排行奖励",
				fmt.Sprintf("恭喜你，在第%d章节BOSS伤害排行中排第%d名，奖励如下：\n", chapter, rank), rewardsMap)
		}
	}
}
