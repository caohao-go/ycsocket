package logic

import (
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/pk"
	"server_golang/repo/table"
)

// Hero 英雄战斗对象（输入数据），对应 NewHero + NewHeroAttr 所需字段的并集
type Hero struct {
	table.HeroAttr

	// hero_id（与 PHP array_merge 后保留的 user_hero.hero_id 一致，客户端依赖此字段）
	HeroId int `json:"hero_id"`
	// 所属用户
	UserId int64 `json:"user_id"`

	// 基础身份信息
	Location int `json:"location"` // 1 战士 2 法师 3 坦克 4 辅助
	Property int `json:"property"` // 属性相克用
	SkinID   int `json:"skin_id"`  // 皮肤 ID

	// 等级信息
	Star     int `json:"star"`
	Stage    int `json:"stage"`
	Lv       int `json:"lv"`
	MaxStage int `json:"max_stage"`
	LvMax    int `json:"lv_max"`

	// 升级信息
	UpLvExpNeed  int `json:"uplv_exp_need"`
	UpLvGoldNeed int `json:"uplv_gold_need"`

	// 战斗属性
	Pos        int `json:"pos"`        // 阵位 1~6）
	CurrentHP  int `json:"current_hp"` // 起始当前血量（可选，无则以 HP 初始化
	FightPoint int `json:"fight_point"`

	// 次级/增益属性
	BeHarmAdd  int `json:"be_harm_add"`
	BeCrt      int `json:"be_crt"`
	BeBaoHarm  int `json:"be_bao_harm"`
	BackRate   int `json:"back_rate"`
	BackHarm   int `json:"back_harm"`
	ReturnRate int `json:"return_rate"`
	ReturnHarm int `json:"return_harm"`

	// 装备
	Fits types.Map           `json:"fit"`
	Fu   map[string]*util.Fu `json:"fu"`

	// 技能
	BaseSkill int      `json:"base_skill"` // 基础（普通攻击）技能 ID
	Skills    []*Skill `json:"skills"`     // 技能列表
}

// ToPKHeroAttr 从 Hero 构造战斗中的动态属性
// Hero 中 CurrentHP 字段若为 0 且 map 原始数据也没有 "current_hp"，
// 需要以 MaxHP 初始化。为此 Hero 将 current_hp 转换拆出 — 见 object.go
func (o *Hero) ToPKHeroAttr() *pk.HeroAttr {
	ha := &pk.HeroAttr{
		ID:                   o.Id,
		MaxHP:                o.Hp,
		CurrentAtk:           o.Atk,
		CurrentDef:           o.Def,
		CurrentSpeed:         o.Speed,
		CurrentCrt:           o.Crt,
		CurrentBaoHarm:       o.BaoHarm,
		CurrentOppBao:        o.OppBao,
		CurrentControl:       o.Control,
		CurrentOppControl:    o.OppControl,
		CurrentHit:           o.Hit,
		CurrentNoHarm:        o.NoHarm,
		CurrentAvd:           o.Avd,
		CurrentHarmAdd:       o.HarmAdd,
		CurrentMagicHarmAdd:  o.MagicHarmAdd,
		CurrentPhysicHarmAdd: o.PhysicHarmAdd,
		CurrentMagicNoHarm:   o.MagicNoHarm,
		CurrentPhysicNoHarm:  o.PhysicNoHarm,
		CurrentCureAdd:       o.CureAdd,
		CurrentBeCureAdd:     o.BeCureAdd,
		CurrentIgnoreDef:     o.IgnoreDef,
		CurrentBeHarmAdd:     o.BeHarmAdd,
		CurrentBeCrt:         o.BeCrt,
		CurrentBeBaoHarm:     o.BeBaoHarm,
		CurrentBackRate:      o.BackRate,
		CurrentBackHarm:      o.BackHarm,
		CurrentReturnRate:    o.ReturnRate,
		CurrentReturnHarm:    o.ReturnHarm,
	}

	// current_hp 为 0 表示未指定起始血量（连续战斗时才会有非零值）
	if o.CurrentHP > 0 {
		ha.CurrentHP = o.CurrentHP
	} else {
		ha.CurrentHP = ha.MaxHP
	}

	return ha
}

// ToPKHero 从 Hero 构造战斗英雄
// Hero 直接承载了英雄的全部战斗字段（参见 object.go），
// 相较以前的 types.Map 入参，避免了每字段反射/转换的开销。
func (o *Hero) ToPKHero() *pk.Hero {
	hero := &pk.Hero{
		ID:              o.Id,
		Pos:             o.Pos,
		Name:            o.Name,
		Lv:              o.Lv,
		Location:        o.Location,
		Property:        o.Property,
		HP:              o.Hp,
		Atk:             o.Atk,
		Def:             o.Def,
		Speed:           o.Speed,
		Crt:             o.Crt,
		BaoHarm:         o.BaoHarm,
		OppBao:          o.OppBao,
		Control:         o.Control,
		OppControl:      o.OppControl,
		Hit:             o.Hit,
		NoHarm:          o.NoHarm,
		Avd:             o.Avd,
		HarmAdd:         o.HarmAdd,
		MagicHarmAdd:    o.MagicHarmAdd,
		PhysicHarmAdd:   o.PhysicHarmAdd,
		MagicNoHarm:     o.MagicNoHarm,
		PhysicNoHarm:    o.PhysicNoHarm,
		CureAdd:         o.CureAdd,
		BeCureAdd:       o.BeCureAdd,
		IgnoreDef:       o.IgnoreDef,
		BaseSkill:       o.BaseSkill,
		HeroID:          o.HeroInfo,
		BackRate:        o.BackRate,
		BackHarm:        o.BackHarm,
		ReturnRate:      o.ReturnRate,
		ReturnHarm:      o.ReturnHarm,
		DeathSkillKey:   -1,
		BakKillSkillKey: -1,
	}

	// lv 为空时默认 1（与 PHP hero.zep:57 一致）
	if hero.Lv <= 0 {
		hero.Lv = 1
	}

	for i, s := range o.Skills {
		if s.Active == 0 {
			continue
		}
		skill := s.ToPKSkill()
		hero.Skills = append(hero.Skills, skill)

		if skill.TriggerType == pk.TriggerTypeDeath {
			hero.DeathSkillKey = i
		}
		if skill.SkillID == 13025 || skill.SkillID == 15035 {
			hero.BakKillSkillKey = len(hero.Skills) - 1
		}
	}

	return hero
}

func (o *Hero) Clone() *Hero {
	fits := types.Map{}
	for k, v := range o.Fits {
		fits[k] = v
	}

	fu := make(map[string]*util.Fu, len(o.Fu))
	for k, v := range o.Fu {
		fu[k] = v.Clone()
	}

	skills := make([]*Skill, len(o.Skills))
	for i, s := range o.Skills {
		skills[i] = s
	}

	return &Hero{
		HeroAttr: o.HeroAttr,

		// 所属用户
		UserId: o.UserId,

		// hero_id
		HeroId: o.HeroId,

		// 基础身份信息
		Location: o.Location,
		Property: o.Property,
		SkinID:   o.SkinID,

		// 等级信息
		Star:     o.Star,
		Stage:    o.Stage,
		Lv:       o.Lv,
		MaxStage: o.MaxStage,
		LvMax:    o.LvMax,

		// 升级信息
		UpLvExpNeed:  o.UpLvExpNeed,
		UpLvGoldNeed: o.UpLvGoldNeed,

		// 战斗属性
		Pos:        o.Pos,
		CurrentHP:  o.CurrentHP,
		FightPoint: o.FightPoint,

		// 次级/增益属性
		BeHarmAdd:  o.BeHarmAdd,
		BeCrt:      o.BeCrt,
		BeBaoHarm:  o.BeBaoHarm,
		BackRate:   o.BackRate,
		BackHarm:   o.BackHarm,
		ReturnRate: o.ReturnRate,
		ReturnHarm: o.ReturnHarm,

		// 装备
		Fits: fits,
		Fu:   fu,

		// 技能
		BaseSkill: o.BaseSkill,
		Skills:    skills,
	}
}

func (o *Hero) ToMap() types.Map {
	ret := types.Map{}
	tmp := json.Marshal(o)
	json.Unmarshal(tmp, &ret)
	return ret
}

type HeroBaseInfo struct {
	Id        int   `json:"id"`
	UserId    int64 `json:"user_id"`
	HeroId    int   `json:"hero_id"`
	Star      int   `json:"star"`
	Stage     int   `json:"stage"`
	Lv        int   `json:"lv"`
	Pos       int   `json:"pos"`
	Hp        int   `json:"hp"`
	CurrentHp int   `json:"current_hp"`
}

// GetBaseFromHero 从 []pk.Hero 从完整英雄属性中提取精简的战斗展示数据（战斗专用）
func GetBaseFromHero(objs []*Hero) []*HeroBaseInfo {
	ret := make([]*HeroBaseInfo, 0, len(objs))
	for _, o := range objs {
		tmp := HeroBaseInfo{
			Id:     o.Id,
			HeroId: o.HeroId,
			Star:   o.Star,
			Stage:  o.Stage,
			Lv:     o.Lv,
			Pos:    o.Pos,
			Hp:     o.Hp,
		}
		if o.CurrentHP > 0 {
			tmp.CurrentHp = o.CurrentHP
		} else {
			tmp.CurrentHp = o.Hp
		}
		ret = append(ret, &tmp)
	}
	return ret
}

// Skill 技能完整数据（数据库拼装后的结构，缓存在 HeroInfoDatas 中，是 pk.SkillObject 的超集）
type Skill struct {
	ID            int        `json:"id"`
	SkillID       int        `json:"skill_id"`
	Name          string     `json:"name"`
	Type          int        `json:"type"`
	Level         int        `json:"level"`
	MaxLevel      int        `json:"max_level"`
	Stage         int        `json:"stage"`
	Star          int        `json:"star"`
	TriggerType   int        `json:"trigger_type"`
	TriggerCd     int        `json:"trigger_cd"`
	TriggerDe     int        `json:"trigger_de"`
	TargetType    int        `json:"target_type"`
	TargetChoose  int        `json:"target_choose"`
	TargetPercent int        `json:"target_percent"`
	TargetNum     int        `json:"target_num"`
	AtkNum        int        `json:"atk_num"`
	AtkType       int        `json:"atk_type"`
	AtkValue      int        `json:"atk_value"`
	BakTrigger    int        `json:"bak_trigger"`
	BakSkill      int        `json:"bak_skill"`
	Buffs         []*pk.Buff `json:"buffs"`

	Active int `json:"active"` // 1 表示已解锁

	// 以下为数据库/展示用字段
	TargetID                int     `json:"target_id,omitempty"`
	Sk1Scale                float32 `json:"sk1Scale,omitempty"`
	AttaScale               float32 `json:"attaScale,omitempty"`
	HurtTime                int     `json:"hurt_time,omitempty"`
	Detail1                 string  `json:"detail1,omitempty"`
	AttackAnimationPosition int     `json:"attackAnimation_position,omitempty"`
	HitAnimationPosition    int     `json:"hitAnimation_position,omitempty"`
	SkillName               int     `json:"skill_name,omitempty"`
}

// ToPKSkill 从 SkillObject 构造战斗技能
func (s *Skill) ToPKSkill() *pk.Skill {
	skill := &pk.Skill{
		ID:            s.ID,
		SkillID:       s.SkillID,
		Name:          s.Name,
		Type:          s.Type,
		Level:         s.Level,
		MaxLevel:      s.MaxLevel,
		Stage:         s.Stage,
		Star:          s.Star,
		TriggerType:   s.TriggerType,
		TriggerCd:     s.TriggerCd,
		TriggerDe:     s.TriggerDe,
		TargetType:    s.TargetType,
		TargetChoose:  s.TargetChoose,
		TargetPercent: s.TargetPercent,
		TargetNum:     s.TargetNum,
		AtkNum:        s.AtkNum,
		AtkType:       s.AtkType,
		AtkValue:      s.AtkValue,
		BakTrigger:    s.BakTrigger,
		BakSkill:      s.BakSkill,
		Buffs:         s.Buffs,
		CdTime:        0,
	}

	return skill
}

func (s *Skill) Clone() *Skill {
	buffs := make([]*pk.Buff, len(s.Buffs))
	for i, b := range s.Buffs {
		tmp := *b
		buffs[i] = &tmp
	}

	return &Skill{
		ID:                      s.ID,
		SkillID:                 s.SkillID,
		Name:                    s.Name,
		Type:                    s.Type,
		Level:                   s.Level,
		MaxLevel:                s.MaxLevel,
		Stage:                   s.Stage,
		Star:                    s.Star,
		TriggerType:             s.TriggerType,
		TriggerCd:               s.TriggerCd,
		TriggerDe:               s.TriggerDe,
		TargetType:              s.TargetType,
		TargetChoose:            s.TargetChoose,
		TargetPercent:           s.TargetPercent,
		TargetNum:               s.TargetNum,
		AtkNum:                  s.AtkNum,
		AtkType:                 s.AtkType,
		AtkValue:                s.AtkValue,
		BakTrigger:              s.BakTrigger,
		BakSkill:                s.BakSkill,
		Buffs:                   buffs,
		Active:                  s.Active,
		TargetID:                s.TargetID,
		Sk1Scale:                s.Sk1Scale,
		AttaScale:               s.AttaScale,
		HurtTime:                s.HurtTime,
		Detail1:                 s.Detail1,
		AttackAnimationPosition: s.AttackAnimationPosition,
		HitAnimationPosition:    s.HitAnimationPosition,
		SkillName:               s.SkillName,
	}
}

// NewFight 创建新的战斗实例
// herosA 为 P1 方（攻方/己方），herosB 为 P2 方（守方/对手）
func NewFight(herosA, herosB []*Hero) *pk.Fight {
	f := &pk.Fight{
		Heros:     map[string]map[int]*pk.Hero{"P1": {}, "P2": {}},
		HerosAttr: map[string]map[int]*pk.HeroAttr{"P1": {}, "P2": {}},
	}

	for _, ha := range herosA {
		h := ha.ToPKHero()
		f.Heros["P1"][h.Pos] = h
		f.HerosAttr["P1"][h.Pos] = ha.ToPKHeroAttr()
	}

	for _, hb := range herosB {
		h := hb.ToPKHero()
		f.Heros["P2"][h.Pos] = h
		f.HerosAttr["P2"][h.Pos] = hb.ToPKHeroAttr()
	}

	return f
}
