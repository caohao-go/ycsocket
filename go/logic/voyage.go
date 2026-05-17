// voyage 包实现了远航系统
package logic

import (
	"context"
	"math/rand"

	"server_golang/common/util"
	"server_golang/repo/table"

	"server_golang/common/types"
	"server_golang/repo/info"
)

// voyageHero 远航英雄需求（对齐 PHP 的 prop/star 字段名）
type voyageHero struct {
	Prop int `json:"prop"`
	Star int `json:"star"`
}

type VoyageFunction struct {
	Id            int          `json:"id"`
	VoyageHero    []voyageHero `json:"voyage_hero"`
	VoyageTime    int          `json:"voyage_time"`
	VoyageQuicken util.TypeNum `json:"voyage_quicken"`
}

var (
	VoyageIDProbability   []int
	VoyageIDProbabilityLv []int // 新手等级过滤后的概率池（lv<=12 时剔除橙红）
	VoyageProbabilityData = make(map[int]*table.VoyageProbability)
	VoyageFunctionData    = make(map[int]*VoyageFunction)
)

// InitVoyage 初始化远航数据
func InitVoyage(ctx context.Context) {
	vpRows, _ := info.GetAllVoyageProbability(ctx)
	for _, row := range vpRows {
		for i := 0; i < row.Probability; i++ {
			VoyageIDProbability = append(VoyageIDProbability, row.Id)
		}
		VoyageProbabilityData[row.Id] = row
	}

	vfRows, _ := info.GetAllVoyageFunction(ctx)
	for _, row := range vfRows {
		// voyage_hero 数据库存储格式: [[prop, star], ...] 如 [[1,3],[0,0]]
		// PHP 解析为 [{prop: 1, star: 3}, ...]，Go 端需对齐
		rawHeros := util.ToTypeNums(row.VoyageHero)
		heros := make([]voyageHero, len(rawHeros))
		for i, h := range rawHeros {
			heros[i] = voyageHero{Prop: h.Type, Star: h.Num}
		}
		VoyageFunctionData[row.Id] = &VoyageFunction{
			Id:            row.Id,
			VoyageTime:    row.VoyageTime * 3600,
			VoyageHero:    heros,
			VoyageQuicken: util.ToTypeNums(row.VoyageQuicken)[0],
		}
	}
}

// GetRandVoyage 随机生成远航任务（对齐 PHP 原版逻辑）
func GetRandVoyage(num int, lv int) []types.Map {
	if num <= 0 {
		num = 10
	}
	ret := make([]types.Map, 0)
	if len(VoyageIDProbability) == 0 {
		return ret
	}

	// 新手保护：12级之前不能刷出橙色(19)、红色(20)
	probPool := VoyageIDProbability
	if lv <= 12 {
		if len(VoyageIDProbabilityLv) == 0 {
			buildFilteredProbPool()
		}
		probPool = VoyageIDProbabilityLv
	}

	gets := make(map[int]int) // 各品质(item_colection)已刷出的数量

	for len(ret) < num {
		idx := rand.Intn(len(probPool))
		id := probPool[idx]
		vp := VoyageProbabilityData[id]
		if vp == nil {
			continue
		}
		itemColection := vp.ItemColection
		numMax := vp.NumMax

		// 品质数量上限检查：该品质已达到 num_max 则跳过重新随机
		if gets[itemColection] >= numMax {
			continue
		}
		gets[itemColection]++

		items := GetRandCollectionItem(itemColection, 1)
		if len(items) > 0 {
			vfData := VoyageFunctionData[id]
			var vTime int
			var vHero []voyageHero
			var vQuicken util.TypeNum
			if vfData != nil {
				vTime = vfData.VoyageTime
				vHero = vfData.VoyageHero
				vQuicken = vfData.VoyageQuicken
			}
			if vTime == 0 {
				vTime = 3600
			}
			if vHero == nil {
				vHero = []voyageHero{}
			}

			item := types.Map{
				"id":              rand.Intn(999999999) + 111,
				"voyage_id":       id,
				"items_id":        items[0].ItemsId,
				"num":             items[0].Number,
				"cost_exp":        2000,
				"status":          0,
				"time":            vTime,
				"left_time":       0,
				"hero":            vHero,
				"accelerate_type": vQuicken.Type,
				"accelerate_num":  vQuicken.Num,
				"item_colection":  itemColection,
			}
			ret = append(ret, item)
		}
	}
	return ret
}

// buildFilteredProbPool 构建新手过滤概率池（lv<=12 剔除橙红品质）
func buildFilteredProbPool() {
	VoyageIDProbabilityLv = make([]int, 0)
	for _, id := range VoyageIDProbability {
		vp := VoyageProbabilityData[id]
		if vp != nil {
			itemColection := vp.ItemColection
			if itemColection != 19 && itemColection != 20 {
				VoyageIDProbabilityLv = append(VoyageIDProbabilityLv, id)
			}
		}
	}
	// 如果过滤后为空则用原池兜底
	if len(VoyageIDProbabilityLv) == 0 {
		VoyageIDProbabilityLv = VoyageIDProbability
	}
}
