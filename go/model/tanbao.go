package model

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo"
	"server_golang/repo/mem/daily"
)

// GetRand8Item 获取随机8个探宝物品
func GetRand8Item(ctx context.Context, userID int64, id int) []*logic.ItemsCollection {
	v, ok := daily.GetByPrefix(userID, config.DailyUserTanbao, id)
	if ok {
		data := []*logic.ItemsCollection{}
		json.Unmarshal(v, &data)
		return data
	}

	// 从 TanbaoData 配置中获取 luck_items 概率池
	conf, ok := logic.TanbaoData[id]
	if !ok {
		return []*logic.ItemsCollection{}
	}

	luckItem := conf.LuckItems
	if luckItem.Num == 0 {
		return []*logic.ItemsCollection{}
	}

	tmps := logic.GetRandCollectionItem(luckItem.Type, 8)

	rets := []*logic.ItemsCollection{}
	for _, item := range tmps {
		rets = append(rets, &logic.ItemsCollection{
			Id:           item.Id,
			CollectionId: item.CollectionId,
			ItemId:       item.ItemsId,
			Number:       item.Number,
			CostType:     item.CostType,
			Price:        item.Price,
			BuyLimit:     item.BuyLimit,
			SaleOff:      item.SaleOff,
		})
	}

	daily.SetByPrefix(userID, config.DailyUserTanbao, id, json.Marshal(rets))
	return rets
}

// DelUserTanbaoCache 删除探宝缓存（刷新用）
func DelUserTanbaoCache(userID int64, id int) {
	daily.DelByPrefix(userID, config.DailyUserTanbao, id)
}

// ========================= 探宝 =========================

// HgetUserTanbaoLuckLingqu 获取探宝幸运积分领取状态
func HgetUserTanbaoLuckLingqu(ctx context.Context, uid int64, id, integral int) int {
	v, _ := repo.RedisHGet(ctx, fmt.Sprintf(config.KeyTanbaoLuckLingqu, uid, id), integral)
	return types.ToIntE(v)
}

// SetUserTanbaoLuckLingqu 设置探宝幸运积分已领取
func SetUserTanbaoLuckLingqu(ctx context.Context, uid int64, id, integral int) {
	repo.RedisHSet(ctx, fmt.Sprintf(config.KeyTanbaoLuckLingqu, uid, id), integral, "1")
}

// DelUserTanbaoLuckLingqu 删除探宝幸运积分领取记录
func DelUserTanbaoLuckLingqu(ctx context.Context, uid int64, id int) {
	repo.RedisDel(ctx, fmt.Sprintf(config.KeyTanbaoLuckLingqu, uid, id))
}

// AddTanbaoHistory 添加探宝历史
func AddTanbaoHistory(ctx context.Context, id int, result types.Map) {
	lk := fmt.Sprintf(config.KeyShinelightTanbaoHistory, id)
	repo.RedisRPush(ctx, lk, json.Marshal(result))
	length, _ := repo.RedisLLen(ctx, lk)
	if length > 50 {
		repo.RedisLPop(ctx, lk)
	}
}

// GetTanbaoHistory 获取探宝历史
func GetTanbaoHistory(ctx context.Context, id int) []types.Map {
	data, _ := repo.RedisLRange(ctx, fmt.Sprintf(config.KeyShinelightTanbaoHistory, id), 0, 5)
	if len(data) == 0 {
		return []types.Map{}
	}
	ret := make([]types.Map, 0, len(data))
	for _, v := range data {
		var it = types.Map{}
		json.Unmarshal(v, &it)
		ret = append(ret, it)
	}
	return ret
}
