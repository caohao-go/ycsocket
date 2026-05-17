// 英雄基础信息模块 - 属性、位置、皮肤、技能、战力计算
package logic

import (
	"context"
	"fmt"
	"math/rand"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	daoInfo "server_golang/repo/info"
	"server_golang/repo/table"
)

// StarUpProperty 升星属性需求
type StarUpProperty struct {
	Property int `json:"property"`
	Star     int `json:"star"`
	Num      int `json:"num"`
}

// StarUpSelf 升星自身消耗
type StarUpSelf struct {
	Star int `json:"star"`
	Num  int `json:"num"`
}

// StarIndicateHero 指定英雄消耗
type StarIndicateHero struct {
	HeroID int `json:"hero_id"`
	Star   int `json:"star"`
	Num    int `json:"num"`
}

// StarConsumeData 升星消耗数据
type StarConsumeData struct {
	StageStone     int                      `json:"stage_stone"`
	LvMax          int                      `json:"lv_max"`
	Self           []StarUpSelf             `json:"self"`
	StarUpProperty []StarUpProperty         `json:"star_up_property"`
	Indicate       map[int]StarIndicateHero `json:"indicate"` // hero_id => StarIndicateHero
}

// 英雄基础信息全局数据
var (
	// 英雄基础数据 hero_id => HeroBaseInfo
	HeroBaseDatas map[int]table.HeroInfo
	// 英雄等级数据 lv => data
	HeroLvDatas map[int]table.HeroLv
	// 英雄阶数据 stage => data
	HeroStageDatas map[int]table.HeroStageConsume
	// 英雄升星消耗 star => data
	HeroStarConsumeDatas map[int]StarConsumeData
	// 属性->英雄列表 property => []hero_id
	PropertyHeros map[int][]int
)

// InitHeroinfo 初始化英雄基础信息
func InitHeroinfo(ctx context.Context) {
	HeroBaseDatas = make(map[int]table.HeroInfo)
	HeroLvDatas = make(map[int]table.HeroLv)
	HeroStarConsumeDatas = make(map[int]StarConsumeData)
	HeroStageDatas = make(map[int]table.HeroStageConsume)
	PropertyHeros = make(map[int][]int)

	// 英雄基础基础信息
	rows, err := daoInfo.GetAllHeroInfo(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_info error: %v", err))
	}
	log.Infof(ctx, "InitHeroinfo: loaded %d hero_info records", len(rows))
	for _, v := range rows {
		HeroBaseDatas[v.Id] = *v
		PropertyHeros[v.Property] = append(PropertyHeros[v.Property], v.Id)
	}

	// 英雄等级数据
	heroLvRows, err := daoInfo.GetAllHeroLv(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_lv error: %v", err))
	}
	for _, val := range heroLvRows {
		HeroLvDatas[val.Lv] = *val
	}

	// 英雄阶消耗
	stageRows, err := daoInfo.GetAllHeroStageConsume(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_stage_consume error: %v", err))
	}

	for _, v := range stageRows {
		HeroStageDatas[v.Stage] = *v
	}

	// 英雄升星消耗数据
	starConsumeRows, err := daoInfo.GetAllHeroStarConsume(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_star_consume error: %v", err))
	}
	for _, v := range starConsumeRows {
		tmp := StarConsumeData{
			StageStone:     v.StageStone,
			LvMax:          v.LvMax,
			Self:           []StarUpSelf{},
			StarUpProperty: []StarUpProperty{},
			Indicate:       map[int]StarIndicateHero{},
		}

		// 解析self
		selfArr := util.ToTypeNums(v.Self)
		for _, s := range selfArr {
			tmp.Self = append(tmp.Self, StarUpSelf{Star: s.Type, Num: s.Num})
		}

		// 解析star_up_property
		var propArr [][]int
		_ = json.Unmarshal(v.StarUpProperty, &propArr)
		for _, p := range propArr {
			tmp.StarUpProperty = append(tmp.StarUpProperty, StarUpProperty{
				Property: p[0], Star: p[1], Num: p[2],
			})
		}

		HeroStarConsumeDatas[v.Star] = tmp
	}

	// 指定英雄升星消耗
	heroStarRows, err := daoInfo.GetAllHeroStar(ctx)
	if err != nil {
		panic(fmt.Errorf("init hero_star error: %v", err))
	}
	for _, v := range heroStarRows {
		if _, ok := HeroStarConsumeDatas[v.Star]; !ok {
			continue
		}

		var starUpHero []int
		_ = json.Unmarshal(v.StarUpHero, &starUpHero)
		HeroStarConsumeDatas[v.Star].Indicate[v.Hero] = StarIndicateHero{
			HeroID: starUpHero[0],
			Star:   starUpHero[1],
			Num:    starUpHero[2],
		}
	}

}

// HeroProperty 获取英雄属性类型
func HeroProperty(heroID int) int {
	if info, ok := HeroBaseDatas[heroID]; ok {
		return info.Property
	}
	return 0
}

// HeroLocation 获取英雄位置
func HeroLocation(heroID int) int {
	if info, ok := HeroBaseDatas[heroID]; ok {
		return info.Location
	}
	return 0
}

// HeroSkin 获取英雄皮肤ID
func HeroSkin(heroID int) int {
	if info, ok := HeroBaseDatas[heroID]; ok {
		return info.SkinId
	}
	return 0
}

// HeroName 获取英雄名称
func HeroName(heroID int) string {
	if info, ok := HeroBaseDatas[heroID]; ok {
		return info.Name
	}
	return ""
}

// GetStarUpList 获取升星列表
// 与 PHP getStarUpList 对齐：使用计数器+扣减方式准确计算 has，避免同一英雄被重复统计
func GetStarUpList(userHeros []*table.UserHero) []types.Map {
	ret := make([]types.Map, 0)

	// 预先统计各属性各星级的英雄数量，以及各英雄ID各星级的数量
	// propertyNum[property][star] = count
	propertyNum := make(map[int]map[int]int)
	// herosIDNum[heroID][star] = count
	herosIDNum := make(map[int]map[int]int)

	for _, uh := range userHeros {
		heroID := uh.HeroId
		star := uh.Star
		propertyID := HeroProperty(heroID)

		if propertyNum[propertyID] == nil {
			propertyNum[propertyID] = make(map[int]int)
		}
		propertyNum[propertyID][star]++

		if herosIDNum[heroID] == nil {
			herosIDNum[heroID] = make(map[int]int)
		}
		herosIDNum[heroID][star]++

		// property=0 表示任意属性的统计
		if propertyNum[0] == nil {
			propertyNum[0] = make(map[int]int)
		}
		propertyNum[0][star]++
	}

	settedHero := make(map[int]map[int]bool)
	for _, tmp := range userHeros {
		heroID := tmp.HeroId
		star := tmp.Star
		property := HeroProperty(heroID)

		if star <= 3 || star >= 6 {
			continue
		}
		if settedHero[heroID] != nil && settedHero[heroID][star] {
			continue
		}
		if settedHero[heroID] == nil {
			settedHero[heroID] = make(map[int]bool)
		}
		settedHero[heroID][star] = true

		userHero := types.ObjectToMap(tmp)
		userHero["property"] = property

		// 初始 need=1, has=1 与 PHP 一致（主体本身算1）
		totalNeed := 1
		totalHas := 1

		// PHP 中 $property_num 和 $heros_id_num 按值传递，每次调用互不影响
		// Go 中 map 是引用类型，需要传入深拷贝副本
		pnCopy := types.CopyIntMapMap(propertyNum)
		hnCopy := types.CopyIntMapMap(herosIDNum)
		needs := getStarUpInfo(heroID, star, property, pnCopy, hnCopy)
		for _, need := range needs {
			totalNeed += need.Num
			totalHas += need.Has
		}

		userHero["need"] = totalNeed
		userHero["has"] = totalHas

		delete(userHero, "stage")
		delete(userHero, "lv")
		ret = append(ret, userHero)
	}

	return ret
}

// GetStarUpDetail 获取英雄升星详情
func GetStarUpDetail(userHeroTmp *table.UserHero, userHeros []*table.UserHero) types.Map {
	heroID := userHeroTmp.HeroId
	star := userHeroTmp.Star
	property := HeroProperty(heroID)

	userHero := types.ObjectToMap(userHeroTmp)
	userHero["property"] = property

	consumeData, ok := HeroStarConsumeDatas[star]
	if !ok {
		userHero["needs"] = []types.Map{}
		return userHero
	}

	needs := make([]types.Map, 0)

	// 1. 主体消耗（同英雄同星级）
	consumeMaster := types.Map{
		"star":    star,
		"num":     1,
		"hero_id": heroID,
	}
	masterHas := make([]*table.UserHero, 0)
	for _, val := range userHeros {
		if val.HeroId == heroID && val.Star == star {
			masterHas = append(masterHas, val)
		}
	}
	consumeMaster["has"] = masterHas
	needs = append(needs, consumeMaster)

	// 2. Self 消耗（同英雄不同星级）
	for _, cs := range consumeData.Self {
		item := types.Map{
			"star":     cs.Star,
			"num":      cs.Num,
			"hero_id":  heroID,
			"property": property,
		}
		has := make([]*table.UserHero, 0)
		for _, val := range userHeros {
			if val.HeroId == heroID && val.Star == cs.Star {
				has = append(has, val)
			}
		}
		item["has"] = has
		needs = append(needs, item)
	}

	// 3. Indicate 消耗（指定英雄）
	if indicate, ok := consumeData.Indicate[heroID]; ok {
		item := types.Map{
			"star":    indicate.Star,
			"num":     indicate.Num,
			"hero_id": indicate.HeroID,
		}
		has := make([]*table.UserHero, 0)
		for _, val := range userHeros {
			if val.HeroId == indicate.HeroID && val.Star == indicate.Star {
				has = append(has, val)
			}
		}
		item["has"] = has
		needs = append(needs, item)
	}

	// 4. Property 消耗（同属性英雄）
	for _, cp := range consumeData.StarUpProperty {
		cpProperty := cp.Property
		if cpProperty == 6 {
			cpProperty = property
		}
		item := types.Map{
			"star":     cp.Star,
			"num":      cp.Num,
			"property": cpProperty,
		}
		has := make([]*table.UserHero, 0)
		for _, val := range userHeros {
			valProperty := HeroProperty(val.HeroId)
			if (cpProperty == 0 || property == valProperty) && cp.Star == val.Star {
				has = append(has, val)
			}
		}
		item["has"] = has
		needs = append(needs, item)
	}

	userHero["needs"] = needs

	return userHero
}

// GetHeroMaxLv 获取英雄最大等级
func GetHeroMaxLv(star, stage int) int {
	if stage < 6 {
		if data, ok := HeroStageDatas[stage]; ok {
			return data.LvMax
		}
		return 0
	}
	switch star {
	case 6:
		return 145
	case 7:
		return 165
	case 8:
		return 185
	case 9:
		return 205
	case 10:
		return 255
	case 11:
		return 280
	case 12:
		return 310
	default:
		return 0
	}
}

// GetRandomSamePropertyHero 变身：随机获取同属性同星级的另一个英雄
func GetRandomSamePropertyHero(heroID, star int) int {
	property := HeroProperty(heroID)
	if property == 0 {
		return heroID
	}

	candidates := make([]int, 0)
	for _, hid := range PropertyHeros[property] {
		info, ok := HeroBaseDatas[hid]
		if !ok {
			continue
		}
		if info.HeroOriginalStar == star && hid != heroID {
			candidates = append(candidates, hid)
		}
	}
	if len(candidates) == 0 {
		return heroID
	}
	return candidates[rand.Intn(len(candidates))]
}

// starUpInfoItem 升星信息条目
type starUpInfoItem struct {
	Num int
	Has int
}

// 计算升星需求的 has 数量（与 PHP 对齐）
// 使用计数器+扣减方式，确保同一英雄不会被多个消耗需求重复统计
func getStarUpInfo(heroID, star, property int, propertyNum map[int]map[int]int, herosIDNum map[int]map[int]int) []starUpInfoItem {
	needStar := make([]starUpInfoItem, 0)

	consumeData, ok := HeroStarConsumeDatas[star]
	if !ok {
		return needStar
	}

	// 扣除主体自身（与 PHP 一致）
	safeDecr(herosIDNum, heroID, star, 1)
	safeDecr(propertyNum, property, star, 1)

	// 1. Self 消耗（同英雄不同星级）
	for _, cs := range consumeData.Self {
		consumeStar := cs.Star
		has := safeGet(herosIDNum, heroID, consumeStar)
		if has > cs.Num {
			has = cs.Num
		}
		needStar = append(needStar, starUpInfoItem{Num: cs.Num, Has: has})

		// 扣减已使用的数量
		safeDecr(herosIDNum, heroID, consumeStar, has)
		safeDecr(propertyNum, property, consumeStar, has)
	}

	// 2. Indicate 消耗（指定英雄）
	if indicate, ok := consumeData.Indicate[heroID]; ok {
		indicateHeroID := indicate.HeroID
		indicateStar := indicate.Star
		indicateProperty := HeroProperty(indicateHeroID)

		has := safeGet(herosIDNum, indicateHeroID, indicateStar)
		if has > indicate.Num {
			has = indicate.Num
		}
		needStar = append(needStar, starUpInfoItem{Num: indicate.Num, Has: has})

		// 扣减已使用的数量
		safeDecr(herosIDNum, indicateHeroID, indicateStar, has)
		safeDecr(propertyNum, indicateProperty, indicateStar, has)
	}

	// 3. Property 消耗（同属性英雄）
	for _, cp := range consumeData.StarUpProperty {
		cpProperty := cp.Property
		if cpProperty == 6 {
			cpProperty = property
		}
		has := safeGet(propertyNum, cpProperty, cp.Star)
		if has > cp.Num {
			has = cp.Num
		}
		needStar = append(needStar, starUpInfoItem{Num: cp.Num, Has: has})
	}

	return needStar
}

// safeGet 安全读取二级 map 值，key 不存在时返回 0
func safeGet(m map[int]map[int]int, k1, k2 int) int {
	if inner, ok := m[k1]; ok {
		return inner[k2]
	}
	return 0
}

// safeDecr 安全扣减二级 map 值，自动初始化不存在的 key
func safeDecr(m map[int]map[int]int, k1, k2, val int) {
	if m[k1] == nil {
		m[k1] = make(map[int]int)
	}
	m[k1][k2] -= val
}
