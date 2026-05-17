// 英雄属性计算模块 - 等级/星级/阶属性，升星消耗等
package logic

import (
	"context"
	"fmt"

	"server_golang/repo/table"

	"server_golang/repo/info"
)

// 英雄属性全局数据
var (
	// 英雄属性表 hero_id => data
	HeroAttrDatas map[int]table.HeroAttr
	// 英雄等级属性缓存 key => data
	HeroLvAttrDatas map[string]Hero
	// 英雄等级基础属性 hero_id => data
	HeroLvBasicAttrDatas map[int]table.HeroLvBasicAttr
	// 英雄升阶成长属性 stage => BaseAttr
	HeroStageUpAttrDatas map[int]table.HeroStageUpAttr
	// 英雄升阶基础属性（固定） stage => BaseAttr
	HeroStageBasicAttrDatas map[int]table.HeroStageUpAttr
	// 英雄升星成长属性 star => BaseAttr
	HeroStarUpAttrDatas map[int]table.HeroStarUpAttr
	// 英雄升星基础属性（固定） star => BaseAttr
	HeroStarBasicAttrDatas map[int]table.HeroStarUpAttr
)

// InitHeroattr 初始化英雄属性数据
func InitHeroattr(ctx context.Context) {
	HeroAttrDatas = make(map[int]table.HeroAttr)
	HeroLvAttrDatas = make(map[string]Hero)
	HeroLvBasicAttrDatas = make(map[int]table.HeroLvBasicAttr)
	HeroStarUpAttrDatas = make(map[int]table.HeroStarUpAttr)
	HeroStarBasicAttrDatas = make(map[int]table.HeroStarUpAttr)
	HeroStageUpAttrDatas = make(map[int]table.HeroStageUpAttr)
	HeroStageBasicAttrDatas = make(map[int]table.HeroStageUpAttr)

	heroAttrRows, err := info.GetAllHeroAttr(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_attr error: %v", err))
	}
	for _, v := range heroAttrRows {
		HeroAttrDatas[v.HeroInfo] = *v
	}

	// 英雄等级基础属性
	lvBasicRows, err := info.GetAllHeroLvBasicAttr(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_lv_basic_attr error: %v", err))
	}

	for _, v := range lvBasicRows {
		HeroLvBasicAttrDatas[v.HeroId] = *v
	}

	// 英雄升阶基础属性（固定）
	stageBasicRows, err := info.GetAllHeroStageBasicAttr(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_stage_basic_attr error: %v", err))
	}
	for _, v := range stageBasicRows {
		prev := HeroStageBasicAttrDatas[v.HeroStage-1]
		HeroStageBasicAttrDatas[v.HeroStage] = table.HeroStageUpAttr{
			Hp:    prev.Hp + v.Hp,
			Atk:   prev.Atk + v.Atk,
			Def:   prev.Def + v.Def,
			Speed: prev.Speed + v.Speed,
		}
	}

	// 英雄升阶成长属性
	HeroStageUpAttrDatas[0] = table.HeroStageUpAttr{Hp: 1, Atk: 1, Def: 1, Speed: 1}
	stageUpRows, err := info.GetAllHeroStageUpAttr(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_stage_up_attr error: %v", err))
	}

	for _, v := range stageUpRows {
		prev := HeroStageUpAttrDatas[v.HeroStage-1]
		HeroStageUpAttrDatas[v.HeroStage] = table.HeroStageUpAttr{
			Hp:    prev.Hp * (1 + v.Hp/100),
			Atk:   prev.Atk * (1 + v.Atk/100),
			Def:   prev.Def * (1 + v.Def/100),
			Speed: prev.Speed * (1 + v.Speed/100),
		}
	}

	// 英雄升星基础属性（固定）
	starBasicRows, err := info.GetAllHeroStarBasicAttr(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_star_basic_attr error: %v", err))
	}
	for _, v := range starBasicRows {
		prev := HeroStarBasicAttrDatas[v.HeroStar-1]
		HeroStarBasicAttrDatas[v.HeroStar] = table.HeroStarUpAttr{
			Hp:    prev.Hp + v.Hp,
			Atk:   prev.Atk + v.Atk,
			Def:   prev.Def + v.Def,
			Speed: prev.Speed + v.Speed,
		}
	}

	// 英雄升星成长属性
	starUpRows, err := info.GetAllHeroStarUpAttr(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_star_up_attr error: %v", err))
	}
	for _, v := range starUpRows {
		if v.HeroStar <= 5 {
			HeroStarUpAttrDatas[v.HeroStar] = table.HeroStarUpAttr{Hp: 1, Atk: 1, Def: 1, Speed: 1}
		} else {
			prev := HeroStarUpAttrDatas[v.HeroStar-1]
			HeroStarUpAttrDatas[v.HeroStar] = table.HeroStarUpAttr{
				Hp:    prev.Hp * (1 + v.Hp/100),
				Atk:   prev.Atk * (1 + v.Atk/100),
				Def:   prev.Def * (1 + v.Def/100),
				Speed: prev.Speed * (1 + v.Speed/100),
			}
		}
	}

}

// GetHeroLvAttr 获取英雄等级属性（含缓存）
func GetHeroLvAttr(heroID, star, stage, lv int) *Hero {
	key := fmt.Sprintf("hero_attr_%d_%d_%d_%d", heroID, star, stage, lv)

	if cached, ok := HeroLvAttrDatas[key]; ok {
		return cached.Clone()
	}

	heroAttr := getHeroBaseAttr(heroID, star, stage, lv)
	heroAttr.HeroId = heroID
	heroAttr.Star = star
	heroAttr.Stage = stage
	heroAttr.Lv = lv
	heroLv := HeroLvDatas[lv]
	heroAttr.UpLvExpNeed = heroLv.Exp
	heroAttr.UpLvGoldNeed = heroLv.Gold
	heroAttr.MaxStage = 6
	heroAttr.LvMax = GetHeroMaxLv(star, stage)
	heroAttr.Location = HeroLocation(heroID)
	heroAttr.Property = HeroProperty(heroID)
	heroAttr.SkinID = HeroSkin(heroID)
	heroAttr.Name = HeroName(heroID)

	HeroLvAttrDatas[key] = *heroAttr
	return heroAttr.Clone()
}

// CalFightPoint 战力计算
func CalFightPoint(heroAttr *table.HeroAttr) int {
	fightPoint := 0.0
	fightPoint += float64(heroAttr.Hp) * 0.4
	fightPoint += float64(heroAttr.Atk) * 2
	fightPoint += float64(heroAttr.Def) * 3
	fightPoint += float64(heroAttr.Speed) * 6
	fightPoint += float64(heroAttr.Crt) * 140

	baoHarm := float64(heroAttr.BaoHarm)
	if int(baoHarm) > 150 {
		fightPoint += (baoHarm - 150) * 20
	}

	fightPoint += float64(heroAttr.Control) * 80
	fightPoint += float64(heroAttr.OppControl) * 80
	fightPoint += float64(heroAttr.OppBao) * 140

	hit := float64(heroAttr.Hit)
	if int(hit) > 100 {
		fightPoint += (hit - 100) * 140
	}

	fightPoint += float64(heroAttr.NoHarm) * 180
	fightPoint += float64(heroAttr.Avd) * 140
	fightPoint += float64(heroAttr.CureAdd) * 180
	fightPoint += float64(heroAttr.BeCureAdd) * 180
	fightPoint += float64(heroAttr.HarmAdd) * 180
	fightPoint += float64(heroAttr.MagicHarmAdd) * 100
	fightPoint += float64(heroAttr.MagicNoHarm) * 100
	fightPoint += float64(heroAttr.PhysicHarmAdd) * 100
	fightPoint += float64(heroAttr.PhysicNoHarm) * 100

	return int(fightPoint)
}

// 计算英雄基础属性
func getHeroBaseAttr(heroID, star, stage, lv int) *Hero {
	attr := &Hero{}

	// hero_attr 固有属性
	attr.HeroAttr = HeroAttrDatas[heroID]

	// hero_lv_basic_attr * lv * hero_star_up_attr * hero_stage_up_attr
	lvBasic := HeroLvBasicAttrDatas[heroID]
	starUp := HeroStarUpAttrDatas[star]
	stageUp := HeroStageUpAttrDatas[stage]

	hp := lvBasic.Hp * float64(lv) * starUp.Hp * stageUp.Hp
	atk := lvBasic.Atk * float64(lv) * starUp.Atk * stageUp.Atk
	def := lvBasic.Def * float64(lv) * starUp.Def * stageUp.Def
	speed := lvBasic.Speed * float64(lv) * starUp.Speed * stageUp.Speed

	// hero_star_basic_attr
	starBasic := HeroStarBasicAttrDatas[star]
	hp += starBasic.Hp
	atk += starBasic.Atk
	def += starBasic.Def
	speed += starBasic.Speed

	// hero_stage_basic_attr
	stageBasic := HeroStageBasicAttrDatas[stage]
	hp += stageBasic.Hp
	atk += stageBasic.Atk
	def += stageBasic.Def
	speed += stageBasic.Speed

	attr.Hp = int(hp) + attr.Hp
	attr.Atk = int(atk) + attr.Atk
	attr.Def = int(def) + attr.Def
	attr.Speed = int(speed) + attr.Speed

	attr.FightPoint = CalFightPoint(&attr.HeroAttr)
	return attr
}
