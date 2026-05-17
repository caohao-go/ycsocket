// 装备数据模块 - 符文属性、符文消耗
package item

import (
	"context"
	"math/rand"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/repo/info"
)

// 装备全局数据
var (
	// 符文属性 item_id => id => data
	RunePropDatas map[int]map[int]util.FuProp
	// 符文属性概率 item_id => []id
	RunePropProbablitys map[int][]int
	// 符文消耗 rune_type => []TypeNum
	RuneConsumeDatas map[int][]util.TypeNum
)

// InitEquipment 初始化装备数据
func InitEquipment(ctx context.Context) {
	RunePropDatas = make(map[int]map[int]util.FuProp)
	RunePropProbablitys = make(map[int][]int)
	RuneConsumeDatas = make(map[int][]util.TypeNum)

	propRows, err := info.GetAllRuneProp(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init rune_prop error: %v", err)
	}
	for _, data := range propRows {
		id := data.Id
		itemID := data.ItemId
		probability := data.Probability

		propMap := util.FuProp{
			MaxNum: data.MaxNum,
			Num:    data.Num,
			Prop:   data.Prop,
			Type:   data.Type,
		}

		if RunePropDatas[itemID] == nil {
			RunePropDatas[itemID] = make(map[int]util.FuProp)
		}
		RunePropDatas[itemID][id] = propMap
		for i := 0; i < probability; i++ {
			RunePropProbablitys[itemID] = append(RunePropProbablitys[itemID], id)
		}
	}

	consumeRows, err := info.GetAllRuneConsume(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init rune_consume error: %v", err)
	}
	for _, data := range consumeRows {
		runeType := string(data.RuneType)
		rt := types.ToIntE(runeType)

		consume := util.ToTypeNums(string(data.Consume))
		RuneConsumeDatas[rt] = append(RuneConsumeDatas[rt], consume...)
	}
}

// InitRandRuneProp 随机取符文属性
func InitRandRuneProp(itemID int) []util.FuProp {
	propNum := 1
	r := rand.Intn(100) + 1

	if (itemID == 40201 && r <= 30) || (itemID == 40301 && r <= 50) || itemID == 40401 || itemID == 40501 {
		propNum = 2
	}
	if (itemID == 40401 && r <= 15) || (itemID == 40501 && r <= 30) {
		propNum = 3
	}

	probList := RunePropProbablitys[itemID]
	maxProb := len(probList)
	if maxProb == 0 {
		return nil
	}

	props := make([]util.FuProp, 0, propNum)
	for i := 0; i < propNum; i++ {
		randKey := rand.Intn(maxProb)
		id := probList[randKey]
		if propData, ok := RunePropDatas[itemID][id]; ok {
			props = append(props, propData)
		}
	}
	return props
}
