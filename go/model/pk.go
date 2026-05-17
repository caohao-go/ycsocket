package model

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/cache"
	"server_golang/repo/mem/daily"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

// ========================= PK 竞技场 =========================

// GetDalilyPkUserID 获取每日PK对手ID列表
func GetDalilyPkUserID(ctx context.Context, uid int64) []int64 {
	v, _ := daily.Get(uid, config.DailyPkUserIds)
	if v == nil {
		return []int64{}
	}
	s := types.ToString(v)
	if s == "" {
		return []int64{}
	}
	var ret []int64
	json.Unmarshal(s, &ret)
	if ret == nil {
		return []int64{}
	}
	return ret
}

// SetDalilyPkUserID 设置每日PK对手ID列表
func SetDalilyPkUserID(ctx context.Context, uid int64, ids []int64) {
	daily.Set(uid, config.DailyPkUserIds, json.Marshal(ids))
}

// DelDalilyPkUserID 删除每日PK对手ID列表
func DelDalilyPkUserID(ctx context.Context, uid int64) {
	daily.Del(uid, config.DailyPkUserIds)
}

// GetRedisFreeTutengPkTimes 获取免费PK次数
func GetRedisFreeTutengPkTimes(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyFreePkTimes)
	t := 3 - types.ToIntE(v)
	if t < 0 {
		return 0
	}
	return t
}

// UseRedisFreeTutengPkTimes 使用免费PK次数
func UseRedisFreeTutengPkTimes(ctx context.Context, uid int64) {
	daily.Incr(uid, config.DailyFreePkTimes, 1)
}

// GetUserTutengPk 获取用户图腾 PK 信息（带缓存，与 PHP getUserTutengPk 一致）
func GetUserTutengPk(ctx context.Context, userID int64) *table.UserTutengPk {
	return world.GetUserTutengPkByUserId(ctx, userID)
}

// InsertUserTutengPk 插入图腾 PK 信息（带缓存清理，与 PHP insertUserTutengPk 一致）
func InsertUserTutengPk(ctx context.Context, userID int64, data *table.UserTutengPk) (int64, error) {
	return world.InsertUserTutengPk(ctx, userID, data)
}

// UpdateUserTutengPk 更新图腾 PK 信息（带缓存清理，与 PHP updateUserTutengPk 一致）
func UpdateUserTutengPk(ctx context.Context, userID int64, data types.Map) error {
	return world.UpdateUserTutengPkByUserId(ctx, userID, data)
}

// GetUserTutengPkDetail 获取用户PK记录（返回 struct，用于只读调用方）
func GetUserTutengPkDetail(ctx context.Context, userID int64) ([]*table.UserTutengPkDetail, error) {
	return world.GetUserTutengPkDetail(ctx, userID)
}

func InsertUserTutengPkDetail(ctx context.Context, data *table.UserTutengPkDetail) (int64, error) {
	return world.InsertUserTutengPkDetail(ctx, data)
}

////  pk 排行榜

// tutengRankCacheKey 图腾排行缓存 key（与 PHP clearPikaCache("pre_tuteng_pk_rank_cache") 对齐）
const tutengRankCacheKey = "pre_tuteng_pk_rank_cache"

// GetMyTutengRank 获取我的图腾排名（与 PHP getMyTutengRank 一致，使用固定 key pk_rank_keys）
func GetMyTutengRank(ctx context.Context, userID int64) int64 {
	rank, err := repo.RedisZRevRank(ctx, config.PrePkRankKey, userID)
	if err != nil || rank < 0 {
		// 与 PHP 一致：用户不存在时先清排行缓存，再添加初始分数 1000（PK_INIT_SCORE）
		cache.RankDelAll(tutengRankCacheKey)
		repo.RedisZAdd(ctx, config.PrePkRankKey, config.PkInitScore, userID)
		// 重新获取排名
		rank, _ = repo.RedisZRevRank(ctx, config.PrePkRankKey, userID)
	}
	return rank + 1
}

// GetMyTutengScore 获取我的图腾分数（与 PHP getMyTutengScore 一致，直接从 pk_rank_keys 读取）
func GetMyTutengScore(ctx context.Context, userID int64) int {
	score, err := repo.RedisZScore(ctx, config.PrePkRankKey, userID)
	if err != nil {
		return 0
	}
	return int(score)
}

// IncrTutengScore 设置图腾分数（与 PHP modifyTutengScore 一致：使用 zincrby 增量操作 pk_rank_keys）
func IncrTutengScore(ctx context.Context, userID int64, score float64) {
	repo.RedisZIncrBy(ctx, config.PrePkRankKey, userID, score)
	cache.RankDelAll(tutengRankCacheKey)
}

// ClearTutengScore 清空图腾分数（与 PHP clearTutengScore 一致）
func ClearTutengScore(ctx context.Context) {
	repo.RedisDel(ctx, config.PrePkRankKey)
	cache.RankDelAll(tutengRankCacheKey)
}

// GetTutengRankList 获取图腾排行榜（与 PHP getTutengRankList 一致，使用 pk_rank_keys + 额外字段 lv/fight_point/vip_level）
func GetTutengRankList(ctx context.Context, returnUserInfo bool, start, end int64) []types.Map {
	// 尝试读缓存（对齐 PHP：使用内存缓存替代 pika get/set）
	cacheField := fmt.Sprintf("%d_%d", start, end)
	if cacheData, ok := cache.RankGet(tutengRankCacheKey, cacheField); ok {
		if cacheData == config.EmptyString {
			return []types.Map{}
		}
		result := types.ToMapArrayE(cacheData)
		if len(result) > 0 {
			return result
		}
	}

	// 从 Pika ZSet 获取排名
	scoreRanks, err := repo.RedisZRevRangeWithScores(ctx, config.PrePkRankKey, start, end)
	if err != nil || len(scoreRanks) == 0 {
		return []types.Map{}
	}

	// 收集 userIDs（保持 Redis 返回的排序）
	userIDs := make([]int64, 0, len(scoreRanks))
	for _, sm := range scoreRanks {
		userIDs = append(userIDs, types.ToInt64E(sm.Member))
	}

	// 获取用户信息（对齐 PHP：传入 lv/fight_point/vip_level，pos_type=2）
	var userInfos map[int64]types.Map
	if returnUserInfo {
		userInfos = GetUsersWithDetail(ctx, userIDs, 2, config.AttrLv, config.AttrFightPoint, config.AttrVipLevel)
	}

	rankIdx := start
	result := make([]types.Map, 0, len(scoreRanks))
	for i, uid := range userIDs {
		tmp := types.Map{
			"user_id":    uid,
			"nickname":   "",
			"avatar_url": "",
			"score":      scoreRanks[i].Score,
		}
		if info, ok := userInfos[uid]; ok {
			tmp["nickname"] = info["nickname"]
			tmp["avatar_url"] = info["avatar_url"]
			tmp["fight_point"] = types.ToIntE(info["fight_point"])
			tmp["lv"] = types.ToIntE(info["lv"])
			tmp["vip_level"] = types.ToIntE(info["vip_level"])
		}
		rankIdx++
		tmp["rank"] = rankIdx
		result = append(result, tmp)
	}

	// 缓存结果（对齐 PHP 300 秒过期，使用内存缓存替代 pika）
	cache.RankSet(tutengRankCacheKey, cacheField, json.Marshal(result), 300)

	return result
}

// GetTutengMaxRank 获取图腾最大排名（与 PHP getTutengMaxRank 一致，使用 pk_rank_keys）
func GetTutengMaxRank(ctx context.Context) int64 {
	count, err := repo.RedisZCard(ctx, config.PrePkRankKey)
	if err != nil {
		return 0
	}
	return count
}

// GetRankUserid 获取指定排名的用户（与 PHP getRankUserid 一致，使用 pk_rank_keys）
func GetRankUserid(ctx context.Context, rank int64) types.Map {
	members, err := repo.RedisZRevRange(ctx, config.PrePkRankKey, rank-1, rank-1)
	if err != nil || len(members) == 0 {
		return nil
	}

	score, _ := repo.RedisZScore(ctx, config.PrePkRankKey, members[0])
	return types.Map{
		"user_id": types.ToInt64E(members[0]),
		"score":   int(score),
	}
}
