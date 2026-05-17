// 商店系统模块 - 随机物品集、商店刷新、体力商店
package logic

import (
	"context"
	"math/rand"

	"server_golang/common/util"
	"server_golang/repo/table"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"

	"server_golang/repo/info"
)

const (
	ShopRefreshTypeManual = 2 // 手动刷新
	ShopRefreshTypeBuyAll = 4 // 全部购买刷新
)

// ShopFunctionConfig 对应 shop_function_config 表
type ShopFunctionConfig struct {
	Id            int         `json:"id"`
	ShopId        int         `json:"shop_id"`
	SellType      int         `json:"sell_type"`
	BasicCost     int         `json:"basic_cost"`
	RefreshType   util.MinMax `json:"refresh_type"`
	RefreshNumber util.MinMax `json:"refresh_number"`
}

// ItemsCollection 对应 items_collection 表
type ItemsCollection struct {
	Id              int `json:"id"`
	ShopId          int `json:"shop_id"`
	CollectionId    int `json:"collection_id"`
	ItemId          int `json:"item_id"`
	Number          int `json:"number"`
	CostType        int `json:"cost_type"`
	Price           int `json:"price"`
	BuyLimit        int `json:"buy_limit"`
	SaleOff         int `json:"sale_off"`
	HasBuyNum       int `json:"has_buy_num"`
	Probability     int `json:"probability,omitempty"`
	OrigProbability int `json:"orig_probability,omitempty"`
	MaxProbability  int `json:"max_probability,omitempty"`
}

// 商店全局数据
var (
	// 物品集数据 collection_id => id => data
	CollectionItemDatas map[int][]*table.ItemsCollection
	// 商店功能配置 shop_id => data
	ShopFunctionConfigDatas map[int]ShopFunctionConfig
	// 商店出售数据 shop_id => []data
	ShopSellDatas map[int][]*table.ShopSell
	// 全局预加载商品 item_id => item
	GlobalPreShopsItems map[int]*ItemsCollection
)

// InitItemsCollection 初始化物品集和商店配置
func InitItemsCollection(ctx context.Context) {
	CollectionItemDatas = make(map[int][]*table.ItemsCollection)
	ShopFunctionConfigDatas = make(map[int]ShopFunctionConfig)
	ShopSellDatas = make(map[int][]*table.ShopSell)
	GlobalPreShopsItems = make(map[int]*ItemsCollection)

	// 物品集
	icRows, err := info.GetAllItemsCollection(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init items_collection error: %v", err)
		return
	}

	collectionItems := make(map[int][]*table.ItemsCollection)
	probability := make(map[int]int)
	for _, item := range icRows {
		probability[item.CollectionId] += item.Probability
		if collectionItems[item.CollectionId] == nil {
			collectionItems[item.CollectionId] = make([]*table.ItemsCollection, 0)
		}
		collectionItems[item.CollectionId] = append(collectionItems[item.CollectionId], item)
		item.OrigProbability = item.Probability
		item.Probability = probability[item.CollectionId]
	}

	for collectionID, items := range collectionItems {
		maxProb := probability[collectionID]
		for k := range items {
			items[k].MaxProbability = maxProb
		}
		CollectionItemDatas[collectionID] = items
	}

	// 商店功能配置
	sfcRows, err := info.GetAllShopFunctionConfig(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init shop_function_config error: %v", err)
	}
	for _, sfc := range sfcRows {
		shopInfo := ShopFunctionConfig{
			ShopId: sfc.ShopId, SellType: sfc.SellType, BasicCost: sfc.BasicCost,
		}

		shopInfo.RefreshType = util.ToMinMax(sfc.RefreshType)
		shopInfo.RefreshNumber = util.ToMinMax(sfc.RefreshNumber)

		ShopFunctionConfigDatas[sfc.ShopId] = shopInfo
	}

	// 商店出售
	ssRows, err := info.GetAllShopSell(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init shop_sell error: %v", err)
	}
	for _, ss := range ssRows {
		ShopSellDatas[ss.ShopId] = append(ShopSellDatas[ss.ShopId], ss)
	}
}

// GetRandCollectionItem 随机取物品集
func GetRandCollectionItem(collectionID int, randNum int) []*table.ItemsCollection {
	items := CollectionItemDatas[collectionID]
	if len(items) == 0 {
		return nil
	}

	if len(items) <= randNum {
		return items
	}

	maxProbability := items[len(items)-1].MaxProbability

	remaining := make([]*table.ItemsCollection, len(items))
	for k := range items {
		remaining[k] = items[k]
	}

	ret := make([]*table.ItemsCollection, 0, randNum)
	for i := 0; i < randNum; {
		if len(remaining) == 0 {
			break
		}
		r := rand.Intn(maxProbability) + 1
		end := true
		for k, v := range remaining {
			if r <= v.Probability {
				ret = append(ret, v)
				remaining = append(remaining[:k], remaining[k+1:]...)
				end = false
				i++
				break
			}
		}
		if end && len(remaining) > 0 {
			maxProbability = remaining[len(remaining)-1].Probability
		}
	}

	return ret
}
