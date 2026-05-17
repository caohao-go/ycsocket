package model

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo/cache"
	"server_golang/repo/mem/content"
	"server_golang/repo/mem/item"
	"server_golang/repo/table"
)

// GetHeroAttrCore 获取单个英雄详细属性（核心计算方法）
func GetHeroAttrCore(ctx context.Context, userHero *table.UserHero,
	userID int64, lvOverride, stageOverride, starOverride int) *logic.Hero {
	if userID == 0 {
		userID = userHero.UserId
	}
	heroStar := userHero.Star
	heroStage := userHero.Stage
	heroLv := userHero.Lv
	if starOverride > 0 {
		heroStar = starOverride
	}
	if stageOverride > 0 {
		heroStage = stageOverride
	}
	if lvOverride > 0 {
		heroLv = lvOverride
	}

	heroID := userHero.HeroId

	// 获取英雄等级属性
	heroAttr := logic.GetHeroLvAttr(heroID, heroStar, heroStage, heroLv)
	heroAttr.UserId = userID
	heroAttr.Id = userHero.Id

	// 装备加成
	fits := types.ToMapE(userHero.Fit)

	for _, fitID := range fits {
		if info := item.Table[types.ToIntE(fitID)]; info != nil {
			heroAttr.Hp = heroAttr.Hp + info.HP
			heroAttr.Atk = heroAttr.Atk + info.Atk
			heroAttr.Def = heroAttr.Def + info.Def
			heroAttr.Speed = heroAttr.Speed + info.Speed
		}
	}
	heroAttr.Fits = fits

	// 符文加成
	heroAttr.Fu = util.ToHeroFus(userHero.Fu, heroLv, heroStar)
	if len(heroAttr.Fu) > 0 {
		for _, fu := range heroAttr.Fu {
			for _, prop := range fu.Props {
				addProp(heroAttr, types.ObjectToMap(prop))
			}
		}
	}

	// 玩家额外加成（公会技能、活跃度、星河神殿）
	if userID > 1000 {
		profType := logic.HeroLocation(heroID)
		addProps := getAddProps(ctx, userID, profType)
		for _, prop := range addProps {
			addProp(heroAttr, prop)
		}
	}

	// 计算战斗力
	heroAttr.FightPoint = logic.CalFightPoint(&heroAttr.HeroAttr)

	// 保存战斗力 + 更新英雄战力排行（与PHP一致：go协程异步执行）
	// 使用 Clone 副本，避免调用者后续 CombinationAttrAdd 修改 heroAttr 导致缓存数据被污染
	SaveFightpointHero(userID, heroAttr.Clone())
	return heroAttr
}

// GetUserHeroAttrByIDs 按ID获取英雄详细信息
func GetUserHeroAttrByIDs(ctx context.Context, ids []int, userID int64, combination bool) map[int]*logic.Hero {
	if len(ids) == 0 {
		return nil
	}
	userHeros := GetUserHeroByIDs(ctx, ids)

	ret := make(map[int]*logic.Hero, len(userHeros))
	for _, hero := range userHeros {
		heroData := GetHeroAttrCore(ctx, hero, userID, 0, 0, 0)
		ret[hero.Id] = heroData
	}

	// 组合加成（与PHP一致：combination=true 时应用组合加成）
	if combination && len(ret) == 5 {
		heroList := make([]*logic.Hero, 0, len(ret))
		for _, v := range ret {
			heroList = append(heroList, v)
		}
		logic.CombinationAttrAdd(heroList)
	}

	return ret
}

// GetUserHeroAttrWithSkillByIDs 按ID获取英雄详情带技能
func GetUserHeroAttrWithSkillByIDs(ctx context.Context, ids []int, userID int64, combination bool) map[int]*logic.Hero {
	userHeros := GetUserHeroAttrByIDs(ctx, ids, userID, combination)
	for _, hero := range userHeros {
		skills := logic.GetSkill(hero.HeroInfo, hero.Star, hero.Stage)
		if skills != nil {
			hero.Skills = skills.Skills
			hero.BaseSkill = skills.GetBaseSkillID()
		}
	}
	return userHeros
}

// GetFightHeroAttrWithSkill 根据hero信息数组(含hero_id/star/stage/lv)获取完整英雄属性（含技能）
// 用于怪物对手等非真实用户英雄的场景，从hero_attr配置表计算属性而非查询user_hero表
func GetFightHeroAttrWithSkill(ctx context.Context, heros []*logic.HeroBaseInfo) []*logic.Hero {
	userHeros, pos := heroBaseToUserHero(heros)
	userAttrs := getUserHeroAttrWithSkill(ctx, userHeros)
	for k, v := range userAttrs {
		v.Pos = pos[k]
	}
	return userAttrs
}

func getUserHeroAttrWithSkill(ctx context.Context, heros []*table.UserHero) []*logic.Hero {
	if len(heros) == 0 {
		return nil
	}

	ret := make([]*logic.Hero, len(heros))

	for i, hero := range heros {
		data := GetHeroAttrCore(ctx, hero, 0, 0, 0, 0)

		// 获取技能
		skillSet := logic.GetSkill(data.HeroInfo, data.Star, data.Stage)
		if skillSet != nil {
			data.Skills = skillSet.Skills
			data.BaseSkill = skillSet.GetBaseSkillID()
		}

		ret[i] = data
	}

	// 组合加成（与PHP一致）
	logic.CombinationAttrAdd(ret)
	return ret
}

// GetAddProps 获取附加属性（公会技能、活跃等级、星河神殿加成）
func getAddProps(ctx context.Context, userID int64, professionType int) []types.Map {
	cacheKey := fmt.Sprintf(config.CacheAddProp, userID)
	if propsStr, ok := cache.HGet(cacheKey, professionType); ok {
		if propsStr == config.EmptyString {
			return nil
		}
		result := types.ToMapArrayE(propsStr)
		if len(result) > 0 {
			return result
		}
	}

	// 公会技能加成
	propsMap := make(map[int]map[int]int) // prop => type => num

	guildID := GetUsersGuildID(ctx, userID)
	if guildID != 0 {
		// 公会技能加成（与 PHP 一致：遍历 attr_lv，查 guild_skill_datas 配置计算）
		guildSkillsRaw := content.GetMap(userID, "guild_skills")
		ptKey := types.ToString(professionType - 1)
		if skillDataRaw := guildSkillsRaw[ptKey]; skillDataRaw != nil {
			skillData, _ := types.ToMap(skillDataRaw, "")
			if attrLvArr, err := types.ToArray(skillData["attr_lv"]); err == nil {
				for attrKeyIdx, attrLvVal := range attrLvArr {
					attrLv := types.ToIntE(attrLvVal)
					if attrLv != 0 {
						skillAttr, ok := logic.GuildSkillDatas[professionType][attrKeyIdx]
						if ok {
							increaseType := skillAttr.IncreaseType
							attrType := skillAttr.AttrType
							prop := logic.SkillAttrType2Prop(attrType)
							num := types.ToIntE(skillAttr.AttrPerRangeNumber) * attrLv
							if propsMap[prop] == nil {
								propsMap[prop] = make(map[int]int)
							}
							propsMap[prop][increaseType] += num
						}
					}
				}
			}
		}

		// 公会活跃等级加成
		guildActive := content.GetInt(userID, "guild_active")
		if guildActive > 0 {
			activeLvData := logic.GetActiveLv(guildActive)
			if activeLvData != nil {
				if activeAttr, err := types.ToMapArray(activeLvData["active_attr"], ""); err == nil && len(activeAttr) >= 2 {
					if propsMap[1] == nil {
						propsMap[1] = make(map[int]int)
					}
					propsMap[1][1] += activeAttr[0].GetIntE("num")
					if propsMap[2] == nil {
						propsMap[2] = make(map[int]int)
					}
					propsMap[2][1] += activeAttr[1].GetIntE("num")
				}
			}
		}
	}

	// 星河神殿加成
	for posIdx, tempUid := range logic.TemplateInfo {
		if tempUid == userID {
			pos := posIdx + 1
			attrAdd := logic.GetTemplateAttrAdd(pos)
			if len(attrAdd) >= 2 {
				if propsMap[1] == nil {
					propsMap[1] = make(map[int]int)
				}
				propsMap[1][2] += attrAdd[0].Num
				if propsMap[2] == nil {
					propsMap[2] = make(map[int]int)
				}
				propsMap[2][2] += attrAdd[1].Num
			}
		}
	}

	// 格式化结果
	var result []types.Map
	for prop, valMap := range propsMap {
		for propType, num := range valMap {
			result = append(result, types.Map{
				"prop": prop,
				"type": propType,
				"num":  num,
			})
		}
	}

	// 缓存
	if result == nil {
		cache.HSet(cacheKey, professionType, config.EmptyString, 600)
	} else {
		cache.HSet(cacheKey, professionType, json.Marshal(result), 600)
	}
	return result
}

// 处理符文属性增益
func addProp(heroAttr *logic.Hero, prop types.Map) {
	propNum := types.ToIntE(prop["num"])
	propType := types.ToIntE(prop["type"])

	switch types.ToIntE(prop["prop"]) {
	case 1:
		heroAttr.Hp = addAttr("hp", heroAttr.Hp, propType, propNum)
	case 2:
		heroAttr.Atk = addAttr("atk", heroAttr.Atk, propType, propNum)
	case 3:
		heroAttr.Def = addAttr("def", heroAttr.Def, propType, propNum)
	case 4:
		heroAttr.Speed = addAttr("speed", heroAttr.Speed, propType, propNum)
	case 5:
		heroAttr.Crt = addAttr("crt", heroAttr.Crt, propType, propNum)
	case 6:
		heroAttr.BaoHarm = addAttr("bao_harm", heroAttr.BaoHarm, propType, propNum)
	case 7:
		heroAttr.OppBao = addAttr("opp_bao", heroAttr.OppBao, propType, propNum)
	case 8:
		heroAttr.NoHarm = addAttr("no_harm", heroAttr.NoHarm, propType, propNum)
	case 9:
		heroAttr.HarmAdd = addAttr("harm_add", heroAttr.HarmAdd, propType, propNum)
	case 10:
		heroAttr.MagicHarmAdd = addAttr("magic_harm_add", heroAttr.MagicHarmAdd, propType, propNum)
	case 11:
		heroAttr.PhysicHarmAdd = addAttr("physic_harm_add", heroAttr.PhysicHarmAdd, propType, propNum)
	case 13:
		heroAttr.Control = addAttr("control", heroAttr.Control, propType, propNum)
	case 14:
		heroAttr.OppControl = addAttr("opp_control", heroAttr.OppControl, propType, propNum)
	default:
		return
	}

}

func addAttr(column string, source, propType, propNum int) int {
	var dest int

	if propType == 1 { // 按数量
		dest = source + propNum
	} else { // 按百分比
		// 百分比属性直接加
		percentAttrs := map[string]bool{
			"crt":             true,
			"bao_harm":        true,
			"opp_bao":         true,
			"no_harm":         true,
			"harm_add":        true,
			"magic_harm_add":  true,
			"physic_harm_add": true,
			"control":         true,
			"opp_control":     true,
		}

		if percentAttrs[column] {
			dest = source + propNum
		} else {
			dest = source + propNum*source/100
		}
	}

	return dest
}

func heroBaseToUserHero(tmp interface{}) (userHeros []*table.UserHero, pos []int) {
	switch input := tmp.(type) {
	case []*logic.HeroBaseInfo:
		userHeros = make([]*table.UserHero, 0, len(input))
		pos = make([]int, 0, len(input))

		for _, m := range input {
			userHeros = append(userHeros, &table.UserHero{
				Id:     m.Id,
				UserId: m.UserId,
				HeroId: m.HeroId,
				Star:   m.Star,
				Stage:  m.Stage,
				Lv:     m.Lv,
			})

			pos = append(pos, m.Pos)
		}
	default:
		heros := types.ToMapArrayE(tmp)
		userHeros = make([]*table.UserHero, 0, len(heros))
		pos = make([]int, 0, len(heros))

		for _, m := range heros {
			userHeros = append(userHeros, &table.UserHero{
				Id:     m.GetIntE("id"),
				UserId: m.GetInt64E("user_id"),
				HeroId: m.GetIntE("hero_id"),
				Star:   m.GetIntE("star"),
				Stage:  m.GetIntE("stage"),
				Lv:     m.GetIntE("lv"),
			})

			pos = append(pos, m.GetIntE("pos"))
		}
	}

	return
}
