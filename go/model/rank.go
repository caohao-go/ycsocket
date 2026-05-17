// CoreModel 实现了数据模型基类，封装通用 pika 辅助和排名（Pika ZSet）操作。
// 具体表（user_grade、user_contents、user_contents_int）的 DB+Cache 操作已迁移到 dao 包。
package model

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/cache"
)

// ======================== 排名（Pika ZSet）操作 ========================

// SetRankScore 设置排名分数（覆盖）
func SetRankScore(ctx context.Context, projectName string, userID int64, score float64, expire int) {
	repo.RedisZAdd(ctx, fmt.Sprintf("pre_rank_%s", projectName), score, userID)
	cache.RankDelAll(projectName)
	if expire > 0 {
		repo.RedisExpire(ctx, fmt.Sprintf("pre_rank_%s", projectName), expire)
	}
}

// GetMyRank 获取我的排名
func GetMyRank(ctx context.Context, projectName string, userID int64) int64 {
	rank, err := repo.RedisZRevRank(ctx, fmt.Sprintf("pre_rank_%s", projectName), userID)
	if err != nil {
		return 0
	}
	return rank + 1
}

// GetMyRankScore 获取我的分数
func GetMyRankScore(ctx context.Context, projectName string, userID int64) int {
	score, err := repo.RedisZScore(ctx, fmt.Sprintf("pre_rank_%s", projectName), userID)
	if err != nil {
		return 0
	}
	return int(score)
}

// IncrRankScore 分数新增
func IncrRankScore(ctx context.Context, projectName string, userID int64, addScore float64, expire int) {
	repo.RedisZIncrBy(ctx, fmt.Sprintf("pre_rank_%s", projectName), userID, addScore)
	cache.RankDelAll(projectName)
	if expire > 0 {
		repo.RedisExpire(ctx, fmt.Sprintf("pre_rank_%s", projectName), expire)
	}
}

// DelRankScore 删除排名
func DelRankScore(ctx context.Context, projectName string) {
	repo.RedisDel(ctx, fmt.Sprintf("pre_rank_%s", projectName))
	cache.RankDelAll(projectName)
}

// GetRankList 获取排名列表
// usersGradeKey 可选参数：传入需要从 user_grade 表额外获取的字段列表（如 "lv", "fight_point", "vip_level"）
// 对齐 PHP CoreModel::getRankList 的 $users_grade_key 参数逻辑
func GetRankList(ctx context.Context, projectName string, returnUserInfo bool, start, end int64, usersGradeKey ...string) []types.Map {
	// 非 fight_point 排行尝试读缓存
	if projectName != "fight_point" {
		cacheField := fmt.Sprintf("%d_%d", start, end)
		if cacheData, ok := cache.RankGet(projectName, cacheField); ok {
			if cacheData == config.EmptyString {
				return []types.Map{}
			}
			result := types.ToMapArrayE(cacheData)
			if len(result) > 0 {
				return result
			}
		}
	}

	// 从 pika ZSet 获取排名
	scoreRanks, err := repo.RedisZRevRangeWithScores(ctx, fmt.Sprintf("pre_rank_%s", projectName), start, end)
	if err != nil || len(scoreRanks) == 0 {
		return []types.Map{}
	}

	// 收集 userIDs（保持 Redis 返回的排序）
	userIDs := make([]int64, 0, len(scoreRanks))
	for _, sm := range scoreRanks {
		userIDs = append(userIDs, types.ToInt64E(sm.Member))
	}

	// 获取用户信息（对齐 PHP：传入 usersGradeKey 时调用 GetMultiUsersDetailWithGrade）
	var userInfos map[int64]types.Map
	if returnUserInfo {
		userInfos = GetUsersWithDetail(ctx, userIDs, 1, usersGradeKey...)
	}

	result := make([]types.Map, 0, len(scoreRanks))
	for i, uid := range userIDs {
		tmp := types.Map{
			"user_id":    uid,
			"zone_id":    0,
			"nickname":   "",
			"avatar_url": "",
			"score":      scoreRanks[i].Score,
		}
		if info, ok := userInfos[uid]; ok {
			tmp["nickname"] = info["nickname"]
			tmp["avatar_url"] = info["avatar_url"]
			// 附加 usersGradeKey 中请求的额外字段
			for _, key := range usersGradeKey {
				if v, exists := info[key]; exists {
					tmp[key] = types.ToIntE(v)
				}
			}
		}
		result = append(result, tmp)
	}

	// 缓存结果 60 秒
	cacheField := fmt.Sprintf("%d_%d", start, end)
	cache.RankSet(projectName, cacheField, json.Marshal(result), 60)

	return result
}

// ClearRank 清理排名
func ClearRank(ctx context.Context, projectName string) {
	repo.RedisDel(ctx, fmt.Sprintf("pre_rank_%s", projectName))
	cache.RankDelAll(projectName)
}

// ClearMyRank 清除我的排行
func ClearMyRank(ctx context.Context, projectName string, userID int64) {
	repo.RedisZRem(ctx, fmt.Sprintf("pre_rank_%s", projectName), userID)
	cache.RankDelAll(projectName)
}
