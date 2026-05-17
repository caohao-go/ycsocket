// 爬塔系统模块 - 层数怪物、扫荡、奖励
package logic

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/repo/info"
)

// ClimbtowerLayerData 爬塔层数据
type ClimbtowerLayerData struct {
	Position      int             `json:"position"`
	Combination   int             `json:"combination"`
	RecFightPoint int             `json:"rec_fight_point"`
	Addition      int             `json:"addition"`
	Taici         string          `json:"taici"`
	Heros         []*HeroBaseInfo `json:"heros"`
}

type ClimbtowerRewards struct {
	Layer            int            `json:"layer"`
	SaodangRewards   []util.TypeNum `json:"saodang_rewards"`
	FirstRewards     []util.TypeNum `json:"first_rewards"`
	RecFightPoint    int            `json:"rec_fight_point"`
	Taici            string         `json:"taici"`
	LowestFightPoint int            `json:"lowest_fight_point"`
	FastUserId       int64          `json:"fast_user_id"`
	FastNickname     string         `json:"fast_nickname"`
}

// 爬塔全局数据
var (
	// 爬塔层数据 layer => ClimbtowerLayerData
	ClimbtowerData map[int]*ClimbtowerLayerData
	// 远征层数据 layer => ClimbtowerLayerData
	ExpeditionHeroData map[int]*ClimbtowerLayerData
)

// InitClimbtower 初始化爬塔数据
func InitClimbtower(ctx context.Context) {
	ClimbtowerData = make(map[int]*ClimbtowerLayerData)
	ExpeditionHeroData = make(map[int]*ClimbtowerLayerData)

	// 爬塔数据
	rows, err := info.GetAllClimbtowerHero(ctx)
	if err != nil {
		panic(fmt.Errorf("init climbtower_hero error: %v", err))
	}
	for _, v := range rows {
		ClimbtowerData[v.Layer] = &ClimbtowerLayerData{
			Position:      v.Position,
			Combination:   v.Combination,
			RecFightPoint: v.RecFightPoint,
			Addition:      v.Addition,
			Taici:         v.Taici,
		}

		var fightHeros [][]int
		_ = json.Unmarshal(v.FightHeros, &fightHeros)
		for _, fh := range fightHeros {
			ClimbtowerData[v.Layer].Heros = append(ClimbtowerData[v.Layer].Heros, &HeroBaseInfo{
				HeroId: fh[0],
				Pos:    fh[1],
				Star:   v.Star,
				Stage:  v.Stage,
				Lv:     v.Lv,
			})
		}
	}

	// 远征英雄数据
	expRows, err := info.GetAllExpeditionHero(ctx)
	if err != nil {
		panic(fmt.Errorf("init expedition_hero error: %v", err))
	}

	for _, v := range expRows {
		if ExpeditionHeroData[v.Layer] == nil {
			ExpeditionHeroData[v.Layer] = &ClimbtowerLayerData{
				Position:      v.Position,
				Combination:   v.Combination,
				RecFightPoint: v.RecFightPoint,
				Addition:      v.Addition,
				Taici:         v.Taici,
			}
		}
		var fightHeros [][]int
		_ = json.Unmarshal(v.FightHeros, &fightHeros)
		for _, fh := range fightHeros {
			ExpeditionHeroData[v.Layer].Heros = append(ExpeditionHeroData[v.Layer].Heros, &HeroBaseInfo{
				HeroId: fh[0],
				Pos:    fh[1],
				Star:   v.Star,
				Stage:  v.Stage,
				Lv:     v.Lv,
			})
		}
	}
}

// GetClimbtowerRewards 获取爬塔奖励（接下来9层）
func GetClimbtowerRewards(currentLayer int) []*ClimbtowerRewards {
	data := make([]*ClimbtowerRewards, 0)
	for layer := currentLayer + 1; layer < currentLayer+10; layer++ {
		// 与 PHP 一致：Checkpointreward::$datas[type][num] = ['rewards' => [...]]
		// Go 中 CheckpointRewardDatas[type][num] 直接是 []util.TypeNum（即 rewards 数组）
		saodangRewards := CheckpointRewardDatas[1][layer]
		firstRewards := CheckpointRewardDatas[7][layer]
		if saodangRewards == nil {
			saodangRewards = []util.TypeNum{}
		}
		if firstRewards == nil {
			firstRewards = []util.TypeNum{}
		}
		item := ClimbtowerRewards{
			Layer:          layer,
			SaodangRewards: saodangRewards,
			FirstRewards:   firstRewards,
		}
		// 补充层数详情：推荐战力、怪物特点等
		if layerInfo, ok := ClimbtowerData[layer]; ok {
			item.RecFightPoint = layerInfo.RecFightPoint
			item.Taici = layerInfo.Taici
		} else {
			item.RecFightPoint = 0
			item.Taici = ""
		}
		data = append(data, &item)
	}
	return data
}

// GetClimbtowerHeros 获取爬塔层英雄
func GetClimbtowerHeros(layer int) []*HeroBaseInfo {
	if data, ok := ClimbtowerData[layer]; ok {
		return data.Heros
	}
	return nil
}

// GetExpeditionHeros 获取远征层英雄
func GetExpeditionHeros(layer int) []*HeroBaseInfo {
	if data, ok := ExpeditionHeroData[layer]; ok {
		return data.Heros
	}
	return nil
}

// GetClimbSaodang 获取爬塔扫荡信息
func GetClimbSaodang(layer int) types.Map {
	layerInfo := ClimbtowerData[layer]
	if layerInfo == nil || len(layerInfo.Heros) == 0 {
		return nil
	}

	// 与 PHP 一致：Checkpointreward::$datas[1][$layer]['rewards']
	rewards := CheckpointRewardDatas[1][layer]
	if rewards == nil {
		rewards = []util.TypeNum{}
	}
	data := types.Map{
		"hero_id":         layerInfo.Heros[0].HeroId,
		"monster_spe":     layerInfo.Taici,
		"rec_fight_point": layerInfo.RecFightPoint,
		"rewards":         rewards,
	}
	return data
}
