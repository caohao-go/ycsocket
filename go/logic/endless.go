// endless 包实现了无尽试炼系统
package logic

import (
	"context"

	"server_golang/common/json"
	"server_golang/common/util"
	"server_golang/repo/info"
	"server_golang/repo/table"
)

// EndlessLayerData 无尽试炼层数据
type EndlessLayerData struct {
	Position      int
	Combination   int
	RecFightPoint int
	Addition      int
	Taici         string
	Heros         []*HeroBaseInfo
}

var (
	EndlessData       = make(map[int]*EndlessLayerData)
	EndlessBuffMap    = make(map[int]*table.EndlessBuff)
	EndlessRewardsMap = make(map[int][]util.TypeNum)
)

// InitEndless 初始化无尽试炼数据
func InitEndless(ctx context.Context) {
	heroRows, _ := info.GetAllEndlessHero(ctx)
	for _, row := range heroRows {
		layer := row.Layer
		if EndlessData[layer] == nil {
			EndlessData[layer] = &EndlessLayerData{}
		}
		EndlessData[layer].Position = row.Position
		EndlessData[layer].Combination = row.Combination
		EndlessData[layer].RecFightPoint = row.RecFightPoint
		EndlessData[layer].Addition = row.Addition
		EndlessData[layer].Taici = row.Taici

		var fightHeros [][]int
		json.Unmarshal(row.FightHeros, &fightHeros)
		for _, fh := range fightHeros {
			EndlessData[layer].Heros = append(EndlessData[layer].Heros, &HeroBaseInfo{
				HeroId: fh[0],
				Pos:    fh[1],
				Star:   row.Star,
				Stage:  row.Stage,
				Lv:     row.Lv,
			})
		}
	}

	buffRows, _ := info.GetAllEndlessBuff(ctx)
	for _, row := range buffRows {
		EndlessBuffMap[row.Id] = row
	}

	rewardRows, _ := info.GetAllEndlessLayerReward(ctx)
	for _, row := range rewardRows {
		EndlessRewardsMap[row.Layer] = util.ToTypeNums(row.Reward)
	}
}

// GetNextFirstLayer 获取下一个首通奖励关卡
func GetNextFirstLayer(endlessLayer int) int {
	mod := (endlessLayer + 5) % 5
	return endlessLayer + 5 - mod
}

// GetFirstCrossLayerReward 获取首通奖励
func GetFirstCrossLayerReward(layer int) []util.TypeNum {
	if layer <= 5 {
		return []util.TypeNum{{Type: 7, Num: 30000}}
	} else if layer%15 == 0 {
		return []util.TypeNum{{Type: 21001, Num: 1}}
	} else {
		i := 0
		if layer%5 == 0 {
			i = layer/5 - 1
		}
		num := 30000 + 20000*i
		return []util.TypeNum{{Type: 7, Num: num}}
	}
}

// GetLeijiCrossLayerReward 获取累计通关奖励
func GetLeijiCrossLayerReward(alreadyLayer, startLayer int) []util.TypeNum {
	layer := alreadyLayer - startLayer
	if layer <= 0 {
		return []util.TypeNum{{Type: 1, Num: 0}, {Type: 7, Num: 0}}
	}
	coinNum := 0
	expNum := 0
	for i := startLayer + 1; i <= alreadyLayer; i++ {
		if rewards, ok := EndlessRewardsMap[i]; ok {
			for _, r := range rewards {
				rType := r.Type
				rNum := r.Num
				if rType == 1 {
					coinNum += rNum
				} else if rType == 7 {
					expNum += rNum
				}
			}
		}
	}
	return []util.TypeNum{{Type: 1, Num: coinNum}, {Type: 7, Num: expNum}}
}

// GetEndlessHeros 获取指定层的敌人阵容
func GetEndlessHeros(layer int) []*HeroBaseInfo {
	if data, ok := EndlessData[layer]; ok {
		return data.Heros
	}
	return nil
}

// GetCrossLayerReward 获取通关单层奖励（增量）
func GetCrossLayerReward(layer int, startLayer int) []util.TypeNum {
	layerReward := GetLeijiCrossLayerReward(layer, startLayer)
	layerPrev := GetLeijiCrossLayerReward(layer-1, startLayer)
	coinNum := layerReward[0].Num - layerPrev[0].Num
	expNum := layerReward[1].Num - layerPrev[1].Num
	if coinNum < 0 {
		coinNum = 0
	}
	if expNum < 0 {
		expNum = 0
	}
	return []util.TypeNum{
		{Type: 1, Num: coinNum},
		{Type: 7, Num: expNum},
	}
}

// AdditionHandle 处理无尽增益（Hero 版本，战斗专用）
func AdditionHandle(objs []*Hero, addition int) {
	buff := EndlessBuffMap[addition]
	if buff == nil {
		return
	}
	for i := range objs {
		o := objs[i]
		switch addition {
		case 1: // hp
			if buff.Type == 1 {
				o.Hp += buff.Num
			} else {
				o.Hp = o.Hp * (100 + buff.Num) / 100
			}
		case 2: // atk
			if buff.Type == 1 {
				o.Atk += buff.Num
			} else {
				o.Atk = o.Atk * (100 + buff.Num) / 100
			}
		case 3: // def
			if buff.Type == 1 {
				o.Def += buff.Num
			} else {
				o.Def = o.Def * (100 + buff.Num) / 100
			}
		case 4: // speed
			if buff.Type == 1 {
				o.Speed += buff.Num
			} else {
				o.Speed = o.Speed * (100 + buff.Num) / 100
			}
		case 5: // crt
			if buff.Type == 1 {
				o.Crt += buff.Num
			} else {
				o.Crt = o.Crt * (100 + buff.Num) / 100
			}
		case 6: // bao_harm
			if buff.Type == 1 {
				o.BaoHarm += buff.Num
			} else {
				o.BaoHarm = o.BaoHarm * (100 + buff.Num) / 100
			}
		case 7: // opp_bao
			if buff.Type == 1 {
				o.OppBao += buff.Num
			} else {
				o.OppBao = o.OppBao * (100 + buff.Num) / 100
			}
		case 8: // no_harm
			if buff.Type == 1 {
				o.NoHarm += buff.Num
			} else {
				o.NoHarm = o.NoHarm * (100 + buff.Num) / 100
			}
		case 9: // harm_add
			if buff.Type == 1 {
				o.HarmAdd += buff.Num
			} else {
				o.HarmAdd = o.HarmAdd * (100 + buff.Num) / 100
			}
		case 10: // magic_harm_add
			if buff.Type == 1 {
				o.MagicHarmAdd += buff.Num
			} else {
				o.MagicHarmAdd = o.MagicHarmAdd * (100 + buff.Num) / 100
			}
		case 11: // physic_harm_add
			if buff.Type == 1 {
				o.PhysicHarmAdd += buff.Num
			} else {
				o.PhysicHarmAdd = o.PhysicHarmAdd * (100 + buff.Num) / 100
			}
		case 12: // current_hp
			if o.CurrentHP > 0 {
				newHP := o.CurrentHP + o.Hp*buff.Num/100
				if newHP > o.Hp {
					newHP = o.Hp
				}
				o.CurrentHP = newHP
			}
		}
	}
}
