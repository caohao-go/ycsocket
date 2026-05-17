// 怪物数据模块 - 关卡怪物、爬塔怪物、无尽怪物
package logic

import (
	"context"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/util"
	"server_golang/repo/table"

	"server_golang/common/types"
	"server_golang/repo/info"
)

// 怪物全局数据
var (
	// 关卡怪物数据 id => data
	MonsterDatas map[int]Hero
	// 爬塔怪物数据 id => data
	ClimbtowerMonsterDatas map[int]Hero
	// 无尽怪物数据 id => data
	EndlessMonsterDatas map[int]Hero
)

// InitMonsters 初始化怪物数据
func InitMonsters(ctx context.Context) {
	MonsterDatas = make(map[int]Hero)
	ClimbtowerMonsterDatas = make(map[int]Hero)
	EndlessMonsterDatas = make(map[int]Hero)

	// 关卡怪物
	rows, err := info.GetAllMonsterAttr(ctx)
	if err != nil || len(rows) == 0 {
		log.Errorf(ctx, 0, "init monster_attr error")
	} else {
		for _, data := range rows {
			attr := table.HeroAttr{
				Id:         data.Id,
				HeroInfo:   data.HeroId,
				Hp:         data.Hp,
				Atk:        data.Atk,
				Def:        data.Def,
				Speed:      data.Speed,
				Crt:        data.Crt,
				BaoHarm:    data.BaoHarm,
				OppBao:     data.OppBao,
				Control:    data.Control,
				OppControl: data.OppControl,
				Hit:        data.Hit,
				NoHarm:     data.NoHarm,
				Avd:        data.Avd,
				HarmAdd:    data.HarmAdd,
				CureAdd:    data.CureAdd,
				BeCureAdd:  data.BeCureAdd,
				IgnoreDef:  data.IgnoreDef,
			}

			MonsterDatas[data.Id] = Hero{
				HeroAttr: attr,
				HeroId:   data.HeroId,
				Star:     data.Star,
				Stage:    data.Stage,
				Lv:       data.Lv,
			}
		}
	}

	// 爬塔怪物
	rows, err = info.GetAllClimbtowerMonsterAttr(ctx)
	if err != nil || len(rows) == 0 {
		log.Errorf(ctx, 0, "init climbtower_monster_attr error")
	} else {
		for _, data := range rows {
			attr := table.HeroAttr{
				Id:         data.Id,
				HeroInfo:   data.HeroId,
				Hp:         data.Hp,
				Atk:        data.Atk,
				Def:        data.Def,
				Speed:      data.Speed,
				Crt:        data.Crt,
				BaoHarm:    data.BaoHarm,
				OppBao:     data.OppBao,
				Control:    data.Control,
				OppControl: data.OppControl,
				Hit:        data.Hit,
				NoHarm:     data.NoHarm,
				Avd:        data.Avd,
				HarmAdd:    data.HarmAdd,
				CureAdd:    data.CureAdd,
				BeCureAdd:  data.BeCureAdd,
				IgnoreDef:  data.IgnoreDef,
			}
			ClimbtowerMonsterDatas[data.Id] = Hero{
				HeroAttr: attr,
				HeroId:   data.HeroId,
				Star:     data.Star,
				Stage:    data.Stage,
				Lv:       data.Lv,
			}
		}
	}

	// 无尽怪物
	rows, err = info.GetAllEndlessMonsterAttr(ctx)
	if err != nil || len(rows) == 0 {
		log.Errorf(ctx, 0, "init endless_monster_attr error")
	} else {
		for _, data := range rows {
			attr := table.HeroAttr{
				Id:         data.Id,
				HeroInfo:   data.HeroId,
				Hp:         data.Hp,
				Atk:        data.Atk,
				Def:        data.Def,
				Speed:      data.Speed,
				Crt:        data.Crt,
				BaoHarm:    data.BaoHarm,
				OppBao:     data.OppBao,
				Control:    data.Control,
				OppControl: data.OppControl,
				Hit:        data.Hit,
				NoHarm:     data.NoHarm,
				Avd:        data.Avd,
				HarmAdd:    data.HarmAdd,
				CureAdd:    data.CureAdd,
				BeCureAdd:  data.BeCureAdd,
				IgnoreDef:  data.IgnoreDef,
			}
			EndlessMonsterDatas[data.Id] = Hero{
				HeroAttr: attr,
				HeroId:   data.HeroId,
				Star:     data.Star,
				Stage:    data.Stage,
				Lv:       data.Lv,
			}
		}
	}
}

// BuildBossDetail 通用构建boss详细数据（关卡/爬塔/无尽通用）
func BuildBossDetail(monsterSource map[int]Hero, monsters []*HeroBaseInfo) []*Hero {
	bossDetail := make([]*Hero, 0)
	for _, v := range monsters {
		id := v.HeroId
		src, ok := monsterSource[id]
		if !ok {
			continue
		}

		tmp := src.Clone()
		tmp.Id = id
		tmp.HeroId = src.HeroId
		tmp.UserId = 0
		tmp.Pos = v.Pos

		tmp.Location = HeroLocation(src.HeroId)
		tmp.Property = HeroProperty(src.HeroId)
		tmp.SkinID = HeroSkin(src.HeroId)
		tmp.Name = HeroName(src.HeroId)

		tmp.Fits = types.Map{}
		tmp.Fu = map[string]*util.Fu{
			"left":  {Unlock: 0},
			"right": {Unlock: 0},
		}

		tmp.FightPoint = CalFightPoint(&tmp.HeroAttr)

		// 取技能
		skillSet := GetSkill(src.HeroId, src.Star, src.Stage)
		if skillSet != nil {
			tmp.Skills = skillSet.Skills
			tmp.BaseSkill = skillSet.GetBaseSkillID()
		}

		bossDetail = append(bossDetail, tmp)
	}

	return bossDetail
}

// GetMonstersByCopyBoss 根据怪物 ID 获取 boss 数据
func GetMonstersByCopyBoss(monsterID int) []*Hero {
	if monsterID <= 0 {
		return []*Hero{}
	}
	if data, ok := MonsterDatas[monsterID]; ok {
		return []*Hero{data.Clone()}
	}
	return []*Hero{}
}
