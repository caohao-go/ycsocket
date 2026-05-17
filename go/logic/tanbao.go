// tanbao 包实现了探宝系统
package logic

import (
	"context"
	"fmt"
	"math/rand"

	"server_golang/common/util"
	"server_golang/repo/info"
	"server_golang/repo/table"
)

type LuckReward struct {
	Id        int            `json:"id"`
	CdTime    int            `json:"cd_time"`
	LuckItems util.TypeNum   `json:"luck_items"`
	Cost      []util.TypeNum `json:"cost_type_number"`
	ExtraGain []util.TypeNum `json:"extra_gain"`
}

var (
	TanbaoData     = make(map[int]*LuckReward)
	TanbaoLuckData = make(map[int]map[int][]util.TypeNum)
	TanbaoRandData = make(map[int]*table.TaobaoRand) // id_k => {type, id, probability}
)

// InitTanbao 初始化探宝数据（与 PHP Tanbao::initTanbao 一致）
func InitTanbao(ctx context.Context) {
	rows, err := info.GetAllLuckRewardConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("init tanbao err: %v", err))
	}

	for _, row := range rows {
		m := &LuckReward{
			Id:     row.Id,
			CdTime: 3600 * row.CdTime,
		}

		// 解析 luck_items（与 PHP 一致：查询 items_collection 构建完整的 items/probabilitys 结构）
		var luckItems = util.ToTypeNums(row.LuckItems)
		if len(luckItems) > 0 {
			m.LuckItems = luckItems[0]
		}

		m.Cost = util.ToTypeNums(row.CostTypeNumber) // 解析 cost（与 PHP 一致）
		m.ExtraGain = util.ToTypeNums(row.ExtraGain) // 解析 extra_gain（与 PHP 一致，之前缺失）
		TanbaoData[row.Id] = m
	}

	// 幸运积分奖励
	luckRows, _ := info.GetAllLuckIntegralReward(ctx)
	for _, row := range luckRows {
		if TanbaoLuckData[row.IntegralType] == nil {
			TanbaoLuckData[row.IntegralType] = map[int][]util.TypeNum{}
		}
		TanbaoLuckData[row.IntegralType][row.Integral] = util.ToTypeNums(row.Reward)
	}

	// 探宝随机数据 (对应 PHP: Tanbao::$tanbao_datas)
	randRows, _ := info.GetAllTaobaoRand(ctx)
	for _, row := range randRows {
		TanbaoRandData[row.IdK] = row
	}
}

// GetTanbaoItemRand 获取探宝随机物品（与 PHP Tanbao::getTanbaoItemRand 一致）
// 参数: id=探宝类型, randNum=抽取数量, lv=等级（新手引导用）
func GetTanbaoItemRand(id, randNum, lv int) []*table.ItemsCollection {
	conf, ok := TanbaoData[id]
	if !ok {
		return nil
	}

	// PHP: $luck_items = self::$datas[$id]['luck_items'][0];
	luckItem := conf.LuckItems
	if luckItem.Num == 0 {
		return nil
	}

	ret := GetRandCollectionItem(luckItem.Type, randNum)

	// 新手引导硬编码（与 PHP 一致）
	if len(ret) > 0 {
		if lv == 1 && id == 2 && rand.Intn(100) < 20 {
			ret[0] = &table.ItemsCollection{
				Id:           540,
				ItemsId:      50217,
				CollectionId: 6,
				Probability:  20,
				Number:       1,
				Price:        1,
				SaleOff:      0,
				CostType:     20901,
			}
		} else if len(ret) > 0 && lv == 2 && id == 6 && rand.Intn(100) < 10 {
			ret[0] = &table.ItemsCollection{
				Id:           138,
				ItemsId:      50209,
				CollectionId: 8,
				Probability:  10,
				Number:       1,
				Price:        1,
				SaleOff:      0,
				CostType:     21001,
			}
		}
	}

	return ret
}
