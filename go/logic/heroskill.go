package logic

import (
	"context"
	"fmt"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/util"
	"server_golang/repo/info"
	"server_golang/repo/table"
	"server_golang/pk"
)

var (
	// 技能详情缓存 skill_info_id => *SkillDetail
	HeroSkillDatas = map[int]*Skill{}
	// 技能等级数据 skill_id => []SkillLevelInfo (按level降序)
	SkillLevelDatas = map[int][]table.SkillInfo{}
)

// SkillSet 技能集
type SkillSet struct {
	Skills    []*Skill `json:"skills"`
	BaseSkill *Skill   `json:"base_skill"`
}

// GetBaseSkillID 获取基础技能 ID（兼容旧代码访问 base_skill["id"]）
func (h *SkillSet) GetBaseSkillID() int {
	if h.BaseSkill != nil {
		return h.BaseSkill.ID
	}
	return 0
}

// SkillBaseInfo 技能基础信息
type SkillBaseInfo struct {
	SkillID  int    `json:"skill_id"`
	Name     string `json:"name"`
	Level    int    `json:"level"`
	MaxLevel int    `json:"max_level"`
	Active   int    `json:"active"`
}

// SkillBaseSet 技能集基础信息
type SkillBaseSet struct {
	Skills    []*SkillBaseInfo `json:"skills"`
	BaseSkill int              `json:"base_skill"`
}

func InitSkill(ctx context.Context) {
	// 技能等级数据（按level降序）
	skillRows, err := info.GetAllSkillInfoOrderByLevelDesc(ctx)
	if err != nil {
		panic(fmt.Errorf("init skill_info error: %v", err))
	}

	for _, v := range skillRows {
		SkillLevelDatas[v.SkillId] = append(SkillLevelDatas[v.SkillId], *v)
	}

	for _, skillInfoRow := range skillRows {
		skill := Skill{
			ID:                      skillInfoRow.Id,
			SkillID:                 skillInfoRow.SkillId,
			Name:                    skillInfoRow.Name,
			Type:                    int(skillInfoRow.Type),
			Level:                   int(skillInfoRow.Level),
			MaxLevel:                skillInfoRow.MaxLevel,
			Stage:                   skillInfoRow.Stage,
			Star:                    int(skillInfoRow.Star),
			TriggerType:             int(skillInfoRow.TriggerType),
			TriggerCd:               int(skillInfoRow.TriggerCd),
			TriggerDe:               skillInfoRow.TriggerDe,
			AtkNum:                  int(skillInfoRow.AtkNum),
			AtkType:                 skillInfoRow.AtkType,
			AtkValue:                skillInfoRow.AtkValue,
			BakTrigger:              int(skillInfoRow.BakTrigger),
			BakSkill:                skillInfoRow.BakSkill,
			TargetID:                skillInfoRow.TargetId,
			Sk1Scale:                skillInfoRow.Sk1Scale,
			AttaScale:               skillInfoRow.AttaScale,
			HurtTime:                skillInfoRow.HurtTime,
			Detail1:                 skillInfoRow.Detail1,
			AttackAnimationPosition: skillInfoRow.AttackAnimationPosition,
			HitAnimationPosition:    skillInfoRow.HitAnimationPosition,
			SkillName:               skillInfoRow.SkillName,
		}

		// 合并 SkillTargetChoose 数据
		targetRow, err := info.GetSkillTargetChooseById(ctx, skillInfoRow.TargetId)
		if err == nil && targetRow != nil {
			skill.TargetType = int(targetRow.TargetType)
			skill.TargetChoose = targetRow.TargetChoose
			skill.TargetPercent = targetRow.TargetPercent
			skill.TargetNum = int(targetRow.TargetNum)
		}

		// 解析 buffs
		buffIDs := util.ToIDs(skillInfoRow.Buffs)

		for _, buffID := range buffIDs {
			if buffID == 0 {
				continue
			}

			buffInfoRow, err := info.GetSkillBuffById(ctx, buffID)
			if err != nil || buffInfoRow == nil {
				continue
			}

			bd := &pk.Buff{
				ID:              buffInfoRow.Id,
				Name:            buffInfoRow.Name,
				Rate:            buffInfoRow.Rate,
				Duration:        buffInfoRow.Duration,
				Remark:          buffInfoRow.Remark,
				SpeRate:         buffInfoRow.SpeRate,
				BuffEffectID:    buffInfoRow.BuffEffectId,
				FromSkillTarget: int(buffInfoRow.FromSkillTarget),
				TargetID:        buffInfoRow.TargetId,
				SpeTargetID:     buffInfoRow.SpeTargetId,
				BuffID:          buffInfoRow.BuffId,
				Position:        buffInfoRow.Position,
				Icon:            buffInfoRow.Icon,
			}

			// 合并 buff 的 SkillTargetChoose
			bTargetRow, _ := info.GetSkillTargetChooseById(ctx, buffInfoRow.TargetId)
			if bTargetRow != nil {
				bd.TargetType = int(bTargetRow.TargetType)
				bd.TargetChoose = bTargetRow.TargetChoose
				bd.TargetPercent = bTargetRow.TargetPercent
				bd.TargetNum = int(bTargetRow.TargetNum)
			}

			// 合并 spe_target
			if buffInfoRow.SpeRate > 0 {
				speTargetRow, _ := info.GetSkillTargetChooseById(ctx, buffInfoRow.SpeTargetId)
				if speTargetRow != nil {
					bd.SpeTargetType = int(speTargetRow.TargetType)
					bd.SpeTargetChoose = speTargetRow.TargetChoose
					bd.SpeTargetPercent = speTargetRow.TargetPercent
					bd.SpeTargetNum = int(speTargetRow.TargetNum)
				}
			}

			// SkillBuffEffect 只有 id 字段，合并后仅 delete(id)，无实际新字段
			// 因此这里不再处理 buffEffect

			skill.Buffs = append(skill.Buffs, bd)
		}

		if skill.Buffs == nil {
			skill.Buffs = []*pk.Buff{}
		}

		HeroSkillDatas[skillInfoRow.Id] = &skill
	}
}

// GetSkillBaseInfo 获取技能基础信息
func GetSkillBaseInfo(heroID, star, stage int) *SkillBaseSet {
	skillSet := GetSkill(heroID, star, stage)
	if skillSet == nil {
		return nil
	}
	result := &SkillBaseSet{
		BaseSkill: skillSet.GetBaseSkillID(),
	}
	for _, s := range skillSet.Skills {
		result.Skills = append(result.Skills, &SkillBaseInfo{
			SkillID:  s.SkillID,
			Name:     s.Name,
			Level:    s.Level,
			MaxLevel: s.MaxLevel,
			Active:   s.Active,
		})
	}
	return result
}

// GetSkill 根据英雄信息获取技能
func GetSkill(heroID, star, stage int) *SkillSet {
	baseData, ok := HeroBaseDatas[heroID]
	if !ok {
		log.Errorf(context.Background(), 0, "GetSkill: hero_id=%d NOT found in HeroBaseDatas (size=%d)", heroID, len(HeroBaseDatas))
		return nil
	}

	skills := util.ToIDs(baseData.Skills)
	ret := &SkillSet{
		Skills: make([]*Skill, 0, len(skills)),
	}

	log.Infof(context.Background(), "GetSkill: hero_id=%d star=%d stage=%d skills=%s base_skill=%d",
		heroID, star, stage, baseData.Skills, baseData.BaseSkill)

	for i, skillID := range skills {
		tmp := getSkill(skillID, star, stage)
		if tmp != nil {
			ret.Skills = append(ret.Skills, tmp)
		} else {
			log.Errorf(context.Background(), 34382, "GetSkill: hero_id=%d skill[%d] id=%d returned nil", heroID, i, skillID)
		}
	}

	ret.BaseSkill = getSkill(baseData.BaseSkill, star, stage)
	if ret.BaseSkill == nil {
		log.Warnf(context.Background(), "GetSkill: hero_id=%d base_skill=%d returned nil", heroID, baseData.BaseSkill)
	}

	log.Infof(context.Background(), "GetSkill: hero_id=%d result: %d skills found, base_skill=%v",
		heroID, len(ret.Skills), ret.BaseSkill)
	return ret
}

// getSkill 获取单个技能
func getSkill(skillID, star, stage int) *Skill {
	skill, active := getIdBySkillStarStage(skillID, star, stage)
	if skill == nil {
		return nil
	}

	tmp, ok := HeroSkillDatas[skill.Id]
	if !ok {
		return nil
	}

	ret := tmp.Clone()
	ret.Active = active
	return ret
}

// 根据技能ID和星级阶获取技能信息
func getIdBySkillStarStage(skillID, star, stage int) (*table.SkillInfo, int) {
	levels, ok := SkillLevelDatas[skillID]
	if !ok || len(levels) == 0 {
		log.Errorf(context.Background(), 99372, "GetIDBySkillStarStage: skillID=%d NOT FOUND", skillID)
		return nil, 0
	}

	for _, v := range levels {
		if star >= int(v.Star) {
			active := 0
			if stage >= v.Stage {
				active = 1
			}
			return &v, active
		}
	}

	log.Errorf(context.Background(), 48823, "GetIDBySkillID: skillID=%d star=%d no matching levels=%v", skillID, star, levels)
	return nil, 0
}
