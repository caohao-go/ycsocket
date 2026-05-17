// 祭祀数据模块
package logic

import (
	"context"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/repo/info"
)

// 祭祀全局数据
var (
	// 祭祀奖励 sacrifice_type => hero_type => star => []TypeNum
	SacrificeDatas map[int]map[int]map[int][]util.TypeNum
	// 升星返还 type => level => []TypeNum
	HeroStarUpReturnDatas map[int]map[int][]util.TypeNum
)

// InitSacrifice 初始化祭祀数据
func InitSacrifice(ctx context.Context) {
	SacrificeDatas = make(map[int]map[int]map[int][]util.TypeNum)
	HeroStarUpReturnDatas = make(map[int]map[int][]util.TypeNum)

	// 祭祀基础数据
	rows, err := info.GetAllSacrificeBase(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init sacrifice_base error: %v", err)
	}
	for _, val := range rows {
		if SacrificeDatas[val.SacrificeType] == nil {
			SacrificeDatas[val.SacrificeType] = make(map[int]map[int][]util.TypeNum)
		}
		if SacrificeDatas[val.SacrificeType][val.HeroType] == nil {
			SacrificeDatas[val.SacrificeType][val.HeroType] = make(map[int][]util.TypeNum)
		}
		SacrificeDatas[val.SacrificeType][val.HeroType][val.Star] = util.ToTypeNums(val.Reward)
	}

	// 升星返还
	returnRows, err := info.GetAllHeroStarUpReturn(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init hero_star_up_return error: %v", err)
	}
	for _, v := range returnRows {
		returnList := util.ToTypeNums(v.ReturnNum)

		if HeroStarUpReturnDatas[v.Type] == nil {
			HeroStarUpReturnDatas[v.Type] = make(map[int][]util.TypeNum)
		}

		// type != 1 时需要累加前一级
		if v.Type != 1 {
			if prev, ok := HeroStarUpReturnDatas[v.Type][v.Level-1]; ok {
				returnList = util.Merge(prev, returnList)
			}
		}
		HeroStarUpReturnDatas[v.Type][v.Level] = returnList
	}
}

// GetReturnItems 获取升星返还物品
func GetReturnItems(star, stage, lv int) []util.TypeNum {
	returns := make([]util.TypeNum, 0)
	returns = util.Merge(returns, HeroStarUpReturnDatas[1][star])
	returns = util.Merge(returns, HeroStarUpReturnDatas[2][stage])
	returns = util.Merge(returns, HeroStarUpReturnDatas[3][lv])
	return returns
}

// GetEquipmentReturn 获取装备返还
func GetEquipmentReturn(fit, fu string, lv, star int) []util.TypeNum {
	unload := make([]util.TypeNum, 0)

	var fits types.Map
	_ = json.Unmarshal(fit, &fits)
	if fits != nil {
		for _, slot := range []string{"weapon", "dress", "head", "shoes"} {
			if fits.GetIntE(slot) > 0 {
				unload = append(unload, util.TypeNum{Type: fits.GetIntE(slot), Num: 1})
			}
		}
	}

	fus := util.ToHeroFus(fu, lv, star)
	for _, tmp := range fus {
		if tmp.Unlock == 1 && tmp.ItemId > 0 {
			unload = append(unload, util.TypeNum{Type: tmp.ItemId, Num: 1, Prop: tmp.Props})
		}
	}

	return unload
}
