package model

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/repo"
	"server_golang/repo/mem/daily"
	"server_golang/repo/table"

	"server_golang/common/types"
	"server_golang/logic"
)

type ShopShell struct {
	logic.ShopFunctionConfig
	Items         []*logic.ItemsCollection `json:"items"`
	LimitTotal    int                      `json:"limit_total"`
	Status        int                      `json:"status"`
	SellCost      int                      `json:"sell_cost"`
	MaxPower      int                      `json:"max_power"`
	Power         int                      `json:"power"`
	PowerNeedTime int                      `json:"power_need_time"`
}

// GetRefreshInfo 获取刷新信息
func GetRefreshInfo(ctx context.Context, userid int64, shopID int, lv int) *ShopShell {
	return getShopByID(ctx, userid, shopID, lv, false)
}

// GetCollectionItem 获取商店随机物品（带缓存）
func GetCollectionItem(ctx context.Context, userid int64, shopID int,
	collectionID int, randNum int, refresh bool, lv int) []*table.ItemsCollection {
	if !refresh {
		isScoreShop := shopID == 10010 || shopID == 10021 ||
			shopID == 10022 || shopID == 10023 || shopID == 10024 || shopID == 10025
		if isScoreShop {
			data := GetDalilyRedisCollectionItem(ctx, userid, collectionID)
			if len(data) > 0 {
				return data
			}
		} else {
			data := GetRedisCollectionItem(ctx, userid, collectionID)
			if len(data) > 0 {
				return data
			}
		}
	}

	DelRedisShopBuynum(ctx, userid, shopID)
	data := logic.GetRandCollectionItem(collectionID, randNum)

	// 新手引导精灵商店第一个位置特殊物品
	if len(data) > 0 && lv <= 3 && shopID == 10040 {
		data[0].ItemsId = 21001
		data[0].Number = 1
		data[0].CostType = 2
		data[0].Price = 150
		data[0].SaleOff = 6
	}

	isScoreShop := shopID == 10010 || shopID == 10021 || shopID == 10022 || shopID == 10023 || shopID == 10024 || shopID == 10025
	if isScoreShop {
		SetDalilyRedisCollectionItem(ctx, userid, collectionID, data)
	} else {
		SetRedisCollectionItem(ctx, userid, collectionID, data)
	}

	return data
}

// GetShopByID 获取商店信息
func GetShopByID(ctx context.Context, userid int64, shopID int, lv int) *ShopShell {
	data := getShopByID(ctx, userid, shopID, lv, false)
	if data == nil {
		return nil
	}
	return formatShopData(ctx, userid, data)
}

// IsAllBuy 是否全部购买
func IsAllBuy(ctx context.Context, userid int64, shopID int, lv int) bool {
	data := getShopByID(ctx, userid, shopID, lv, false)
	if data == nil {
		return false
	}
	totalNumber := GetRedisAllShopBuynum(ctx, userid, shopID)
	return totalNumber >= data.LimitTotal
}

// AllByRefresh 如果全部买完则自动刷新
func AllByRefresh(ctx context.Context, userid int64, shopID int, lv int) {
	data := getShopByID(ctx, userid, shopID, lv, false)
	if data == nil {
		return
	}

	hasBuyAll := false
	if int(data.RefreshType.Min) == logic.ShopRefreshTypeBuyAll ||
		int(data.RefreshType.Max) == logic.ShopRefreshTypeBuyAll {
		hasBuyAll = true
	}

	if hasBuyAll && IsAllBuy(ctx, userid, shopID, lv) {
		RefreshShopBuynum(ctx, userid, shopID, lv)
	}
}

// RefreshShopBuynum 刷新商店购买次数
func RefreshShopBuynum(ctx context.Context, userid int64, shopID int, lv int) {
	DelRedisShopBuynum(ctx, userid, shopID)
	if shopID == PowerType10030 || shopID == PowerType10040 || shopID == PowerType10024 {
		getShopByID(ctx, userid, shopID, lv, true)
	}
}

// ========================= 商店刷新 =========================

func GetRefreshTimes(ctx context.Context, uid int64, shopID int) int {
	v, _ := daily.GetByPrefix(uid, config.DailyRefreshTimes, shopID)
	return types.ToIntE(v)
}

func IncrRefreshTimes(ctx context.Context, uid int64, shopID int) {
	daily.IncrByPrefix(uid, config.DailyRefreshTimes, shopID, 1)
}

// getShopByID 内部获取商店数据
func getShopByID(ctx context.Context, userid int64, shopID int, lv int, refresh bool) *ShopShell {
	data := getShopSellByID(ctx, userid, shopID, refresh, lv)
	if data == nil {
		return nil
	}

	if len(data.Items) > 0 {
		for _, item := range data.Items {
			logic.GlobalPreShopsItems[item.Id] = item
		}
	}

	return data
}

// getShopSellByID 获取商店售卖数据
func getShopSellByID(ctx context.Context, userid int64, shopID int, refresh bool, lv int) *ShopShell {
	shopInfo, ok := logic.ShopFunctionConfigDatas[shopID]
	if !ok {
		return nil
	}

	shopSells := logic.ShopSellDatas[shopID]
	if len(shopSells) == 0 {
		return nil
	}

	result := ShopShell{
		ShopFunctionConfig: shopInfo,
	}

	limitTotal := 0
	items := make([]*logic.ItemsCollection, 0)

	// 判断是否物品集（与 PHP 一致：goods_collection 不是数字时为物品集）
	isCollection := !types.IsNumber(shopSells[0].GoodsCollection)
	if isCollection {
		// 物品集
		for _, shopSell := range shopSells {
			var gc []int
			_ = json.Unmarshal(types.ToString(shopSell.GoodsCollection), &gc)
			if len(gc) < 2 {
				continue
			}
			realItems := GetCollectionItem(ctx, userid, shopID, gc[0], gc[1], refresh, lv)
			for _, ri := range realItems {
				item := logic.ItemsCollection{
					Id:       types.ToIntE(fmt.Sprintf("99999%d", ri.Id)),
					ShopId:   shopID,
					ItemId:   ri.ItemsId,
					Number:   ri.Number,
					CostType: ri.CostType,
					Price:    ri.Price,
					BuyLimit: 1,
					SaleOff:  ri.SaleOff,
				}
				items = append(items, &item)
			}
		}
	} else {
		for _, shopSell := range shopSells {
			itemID := types.ToIntE(shopSell.GoodsCollection)
			item := logic.ItemsCollection{
				Id:       shopSell.Id,
				ShopId:   shopSell.ShopId,
				ItemId:   itemID, // 非collection类商店使用item_id，与PHP保持一致
				Number:   shopSell.GoodsNumber,
				CostType: shopSell.CostType,
				Price:    shopSell.Price,
				BuyLimit: shopSell.BuyLimit,
				SaleOff:  shopSell.SaleOff,
			}
			limitTotal += shopSell.BuyLimit
			items = append(items, &item)
		}
	}

	result.Items = items
	result.LimitTotal = limitTotal
	return &result
}

// formatShopData 格式化商店数据
func formatShopData(ctx context.Context, userid int64, data *ShopShell) *ShopShell {
	data.Status = 1
	data.SellCost = data.BasicCost

	// 体力相关
	if data.ShopId == PowerType10024 || data.ShopId == PowerType10040 {
		var needTime int
		power := GetUserPower(ctx, userid, data.ShopId, &needTime)
		data.MaxPower = ConfMaxPower[data.ShopId]
		data.Power = power
		data.PowerNeedTime = needTime
	} else {
		data.MaxPower = 0
	}

	if len(data.Items) > 0 {
		for key, val := range data.Items {
			data.Items[key].HasBuyNum = GetRedisShopBuynum(ctx, userid, data.ShopId, val.Id)
		}
	}

	return data
}

// ========================= 商店 =========================

// IsNextDayRefresh 次日刷新判断
func IsNextDayRefresh(shopID int) string {
	switch shopID {
	case 10010, 10021, 10022, 10024, 10025, 10030, 10040:
		return "_" + util.DateYmd()
	case 10023:
		return "_" + util.DateW()
	}
	return ""
}

// GetRedisShopBuynum 获取商品已购买次数
func GetRedisShopBuynum(ctx context.Context, uid int64, shopID, id int) int {
	hk := fmt.Sprintf(config.KeyShopBuy, shopID, uid, IsNextDayRefresh(shopID))
	v, _ := repo.RedisHGet(ctx, hk, id)
	return types.ToIntE(v)
}

// AddRedisShopBuynum 增加商品购买次数
func AddRedisShopBuynum(ctx context.Context, uid int64, shopID, id, num int) {
	hk := fmt.Sprintf(config.KeyShopBuy, shopID, uid, IsNextDayRefresh(shopID))
	repo.RedisHIncrBy(ctx, hk, id, int64(num))
	repo.RedisExpire(ctx, hk, 7*86400)
	hk2 := fmt.Sprintf(config.KeyShopTotalBuy, shopID, IsNextDayRefresh(shopID))
	repo.RedisHIncrBy(ctx, hk2, uid, int64(num))
	repo.RedisExpire(ctx, hk2, 7*86400)
}

func DelRedisShopBuynum(ctx context.Context, uid int64, shopID int) {
	repo.RedisDel(ctx, fmt.Sprintf(config.KeyShopBuy, shopID, uid, IsNextDayRefresh(shopID)))
	repo.RedisDel(ctx, fmt.Sprintf(config.KeyShopTotalBuy, shopID, IsNextDayRefresh(shopID)))
}

func GetRedisAllShopBuynum(ctx context.Context, uid int64, shopID int) int {
	v, _ := repo.RedisHGet(ctx, fmt.Sprintf(config.KeyShopTotalBuy, shopID, IsNextDayRefresh(shopID)), uid)
	return types.ToIntE(v)
}

func GetDalilyRedisCollectionItem(ctx context.Context, uid int64, cid int) []*table.ItemsCollection {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyPreDalilyCollectionItem, uid, cid, util.DateYmd()))
	if v == "" {
		return []*table.ItemsCollection{}
	}
	var ret []*table.ItemsCollection
	json.Unmarshal(v, &ret)
	if ret == nil {
		return []*table.ItemsCollection{}
	}
	return ret
}

func SetDalilyRedisCollectionItem(ctx context.Context, uid int64, cid int, data []*table.ItemsCollection) {
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyPreDalilyCollectionItem, uid, cid, util.DateYmd()), data, 86400)
}

func GetRedisCollectionItem(ctx context.Context, uid int64, cid int) []*table.ItemsCollection {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyPreRandCollectionItem, uid, cid))
	if v == "" {
		return []*table.ItemsCollection{}
	}
	var ret []*table.ItemsCollection
	json.Unmarshal(v, &ret)
	if ret == nil {
		return []*table.ItemsCollection{}
	}
	return ret
}

func SetRedisCollectionItem(ctx context.Context, uid int64, cid int, data []*table.ItemsCollection) {
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyPreRandCollectionItem, uid, cid), data, 86400)
}
