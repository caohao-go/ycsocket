package table

// ========================= shine_info 库 - 静态配置表 =========================

import (
	"time"
)

// ========================= 签到/章节/阵位/称号 =========================

// CheckinReward 对应 checkin_reward 表
type CheckinReward struct {
	CheckinDays int    `orm:"checkin_days,int" json:"checkin_days"`
	Reward      string `orm:"reward,varchar" json:"reward"`
}

// Combination 对应 combination 表
type Combination struct {
	Id         int    `orm:"id,int,omitempty" json:"id"`
	Property   string `orm:"property,varchar" json:"property"`
	Hp         int    `orm:"hp,int" json:"hp"`
	Atk        int    `orm:"atk,int" json:"atk"`
	OppControl int    `orm:"opp_control,int" json:"opp_control"`
}

// Chapter 对应 chapter 表
type Chapter struct {
	Id    int    `orm:"id,int,omitempty" json:"id"`
	Name  string `orm:"name,varchar" json:"name"`
	Copy  string `orm:"copy,varchar" json:"copy"`
	MapId int    `orm:"map_id,int" json:"map_id"`
}

// Position 对应 position 表
type Position struct {
	Id      int    `orm:"id,int,omitempty" json:"id"`
	PosType int    `orm:"pos_type,int" json:"pos_type"`
	Name    string `orm:"name,varchar" json:"name"`
}

// RoleTitle 对应 role_title 表
type RoleTitle struct {
	Id        int    `orm:"id,int,omitempty" json:"id"`
	Name      string `orm:"name,varchar" json:"name"`
	Type      int    `orm:"type,int" json:"type"`
	Condition string `orm:"condition,varchar" json:"condition"`
	Prop      string `orm:"prop,varchar" json:"prop"`
}

// CheckpointReward 对应 checkpoint_reward 表
type CheckpointReward struct {
	Id             int    `orm:"id,int,omitempty" json:"id"`
	CheckpointType int    `orm:"checkpoint_type,int" json:"checkpoint_type"`
	CheckpointNum  int    `orm:"checkpoint_num,int" json:"checkpoint_num"`
	Reward         string `orm:"reward,varchar" json:"reward"`
	OnlineTime     int    `orm:"online_time,int" json:"online_time"`
}

// ========================= 符文 =========================

// RuneProp 对应 rune_prop 表
type RuneProp struct {
	Id          int `orm:"id,int,omitempty" json:"id"`
	ItemId      int `orm:"item_id,int" json:"item_id"`
	Prop        int `orm:"prop,int" json:"prop"`
	Type        int `orm:"type,int" json:"type"`
	Num         int `orm:"num,int" json:"num"`
	MaxNum      int `orm:"max_num,int" json:"max_num"`
	Probability int `orm:"probability,int" json:"probability"`
}

// RuneConsume 对应 rune_consume 表
type RuneConsume struct {
	Id       int    `orm:"id,int,omitempty" json:"id"`
	RuneType string `orm:"rune_type,varchar" json:"rune_type"`
	Consume  string `orm:"consume,varchar" json:"consume"`
}

// RuneMerge 对应 rune_merge 表
type RuneMerge struct {
	Id           int    `orm:"id,int,omitempty" json:"id"`
	OriginalId   int    `orm:"original_id,int" json:"original_id"`
	FinalId      int    `orm:"final_id,int" json:"final_id"`
	MergeNum     int    `orm:"merge_num,int" json:"merge_num"`
	MergeConsume int    `orm:"merge_consume,int" json:"merge_consume"`
	SuccessRate  int    `orm:"success_rate,int" json:"success_rate"`
	FailGet      string `orm:"fail_get,varchar" json:"fail_get"`
}

// MergeConfig 对应 merge_config 表
type MergeConfig struct {
	Id         int    `orm:"id,int,omitempty" json:"id"`
	OriginalId int    `orm:"original_id,int" json:"original_id"`
	FinalId    int    `orm:"final_id,int" json:"final_id"`
	Consume    string `orm:"consume,varchar" json:"consume"`
}

// ========================= 怪物属性 =========================

// MonsterAttr 对应 monster_attr 表
type MonsterAttr struct {
	Id         int `orm:"id,int,omitempty" json:"id"`
	HeroId     int `orm:"hero_id,int" json:"hero_id"`
	Star       int `orm:"star,int" json:"star"`
	Stage      int `orm:"stage,int" json:"stage"`
	Lv         int `orm:"lv,int" json:"lv"`
	Hp         int `orm:"hp,int" json:"hp"`
	Atk        int `orm:"atk,int" json:"atk"`
	Def        int `orm:"def,int" json:"def"`
	Speed      int `orm:"speed,int" json:"speed"`
	Crt        int `orm:"crt,int" json:"crt"`
	BaoHarm    int `orm:"bao_harm,int" json:"bao_harm"`
	OppBao     int `orm:"opp_bao,int" json:"opp_bao"`
	Control    int `orm:"control,int" json:"control"`
	OppControl int `orm:"opp_control,int" json:"opp_control"`
	Hit        int `orm:"hit,int" json:"hit"`
	NoHarm     int `orm:"no_harm,int" json:"no_harm"`
	Avd        int `orm:"avd,int" json:"avd"`
	HarmAdd    int `orm:"harm_add,int" json:"harm_add"`
	CureAdd    int `orm:"cure_add,int" json:"cure_add"`
	BeCureAdd  int `orm:"be_cure_add,int" json:"be_cure_add"`
	IgnoreDef  int `orm:"ignore_def,int" json:"ignore_def"`
}

// ClimbtowerMonsterAttr 同 MonsterAttr 结构，对应 climbtower_monster_attr 表
type ClimbtowerMonsterAttr = MonsterAttr

// EndlessMonsterAttr 同 MonsterAttr 结构，对应 endless_monster_attr 表
type EndlessMonsterAttr = MonsterAttr

// ========================= 等级/关卡/挂机 =========================

// LvInfo 对应 lv_info 表
type LvInfo struct {
	Lv  int `orm:"lv,int" json:"lv"`
	Exp int `orm:"exp,int" json:"exp"`
}

// Copy 对应 copy 表
type Copy struct {
	Lv           int    `orm:"lv,int" json:"lv"`
	OpenLv       int    `orm:"open_lv,int" json:"open_lv"`
	Time         string `orm:"time,varchar" json:"time"`
	Monster      string `orm:"monster,varchar" json:"monster"`
	MonsterStar  int    `orm:"monster_star,int" json:"monster_star"`
	MonsterStage int    `orm:"monster_stage,int" json:"monster_stage"`
	MonsterLv    int    `orm:"monster_lv,int" json:"monster_lv"`
}

// OnhookReward 对应 onhook_reward 表
type OnhookReward struct {
	Id         int    `orm:"id,int" json:"id"`
	OnhookMin  string `orm:"onhook_min,varchar" json:"onhook_min"`
	OnhookHour string `orm:"onhook_hour,varchar" json:"onhook_hour"`
}

// CopyBossReward 对应 copy_boss_reward 表
type CopyBossReward struct {
	Id         int    `orm:"id,int,omitempty" json:"id"`
	Reward     string `orm:"reward,varchar" json:"reward"`
	RandNum    string `orm:"rand_num,varchar" json:"rand_num"`
	RandReward string `orm:"rand_reward,varchar" json:"rand_reward"`
}

// CopyReward 对应 copy_reward 表
type CopyReward struct {
	Id        int    `orm:"id,int,omitempty" json:"id"`
	CopyId    int    `orm:"copy_id,int" json:"copy_id"`
	Seq       int    `orm:"seq,int" json:"seq"`
	RankCount string `orm:"rank_count,varchar" json:"rank_count"`
	Reward    string `orm:"reward,varchar" json:"reward"`
	Refresh   int    `orm:"refresh,int" json:"refresh"`
}

// CheckpointDaliyReward 对应 checkpoint_daliy_reward 表
type CheckpointDaliyReward struct {
	Id         int    `orm:"id,int,omitempty" json:"id"`
	CopyId     int    `orm:"copy_id,int" json:"copy_id"`
	Checkpoint int    `orm:"checkpoint,int" json:"checkpoint"`
	Reward     string `orm:"reward,varchar" json:"reward"`
}

// RandomCountConfig 对应 random_count_config 表
type RandomCountConfig struct {
	Id           int    `orm:"id,int,omitempty" json:"id"`
	CopyRange    string `orm:"copy_range,varchar" json:"copy_range"`
	Probability  int    `orm:"probability,int" json:"probability"`
	Num          string `orm:"num,varchar" json:"num"`
	CollectionId int    `orm:"collection_id,int" json:"collection_id"`
	TimeRange    string `orm:"time_range,char" json:"time_range"`
}

// ========================= 无尽试炼 =========================

// EndlessHero 对应 endless_hero 表
type EndlessHero struct {
	Id            int    `orm:"id,int,omitempty" json:"id"`
	Layer         int    `orm:"layer,int" json:"layer"`
	Position      int    `orm:"position,int" json:"position"`
	Combination   int    `orm:"combination,int" json:"combination"`
	RecFightPoint int    `orm:"rec_fight_point,int" json:"rec_fight_point"`
	Addition      int    `orm:"addition,int" json:"addition"`
	Taici         string `orm:"taici,varchar" json:"taici"`
	Star          int    `orm:"star,int" json:"star"`
	Stage         int    `orm:"stage,int" json:"stage"`
	Lv            int    `orm:"lv,int" json:"lv"`
	FightHeros    string `orm:"fight_heros,varchar" json:"fight_heros"`
}

// EndlessBuff 对应 endless_buff 表
type EndlessBuff struct {
	Id   int `orm:"id,int,omitempty" json:"id"`
	Type int `orm:"type,int" json:"type"`
	Num  int `orm:"num,int" json:"num"`
}

// EndlessLayerReward 对应 endless_layer_reward 表
type EndlessLayerReward struct {
	Id     int    `orm:"id,int,omitempty" json:"id"`
	Layer  int    `orm:"layer,int" json:"layer"`
	Reward string `orm:"reward,varchar" json:"reward"`
}

// ========================= 爬塔 =========================

// ClimbtowerHero 对应 climbtower_hero 表
type ClimbtowerHero struct {
	Id            int    `orm:"id,int,omitempty" json:"id"`
	Layer         int    `orm:"layer,int" json:"layer"`
	Position      int    `orm:"position,int" json:"position"`
	Combination   int    `orm:"combination,int" json:"combination"`
	RecFightPoint int    `orm:"rec_fight_point,int" json:"rec_fight_point"`
	Addition      int    `orm:"addition,int" json:"addition"`
	Taici         string `orm:"taici,varchar" json:"taici"`
	Star          int    `orm:"star,int" json:"star"`
	Stage         int    `orm:"stage,int" json:"stage"`
	Lv            int    `orm:"lv,int" json:"lv"`
	FightHeros    string `orm:"fight_heros,varchar" json:"fight_heros"`
}

// ExpeditionHero 对应 expedition_hero 表（结构同 ClimbtowerHero）
type ExpeditionHero = ClimbtowerHero

// ========================= 英雄/技能 =========================

// HeroLv 对应 hero_lv 表
type HeroLv struct {
	Lv   int `orm:"lv,int" json:"lv"`
	Exp  int `orm:"exp,int" json:"exp"`
	Gold int `orm:"gold,int" json:"gold"`
}

// HeroAttr 对应 hero_attr 表
type HeroAttr struct {
	Id            int    `orm:"id,int,omitempty" json:"id"`
	HeroInfo      int    `orm:"hero_info,int" json:"hero_info"`
	Name          string `orm:"name,varchar" json:"name"`
	Hp            int    `orm:"hp,double" json:"hp"`
	Atk           int    `orm:"atk,double" json:"atk"`
	Def           int    `orm:"def,double" json:"def"`
	Speed         int    `orm:"speed,double" json:"speed"`
	Crt           int    `orm:"crt,double" json:"crt"`
	BaoHarm       int    `orm:"bao_harm,double" json:"bao_harm"`
	OppBao        int    `orm:"opp_bao,double" json:"opp_bao"`
	Control       int    `orm:"control,double" json:"control"`
	OppControl    int    `orm:"opp_control,double" json:"opp_control"`
	Hit           int    `orm:"hit,double" json:"hit"`
	NoHarm        int    `orm:"no_harm,double" json:"no_harm"`
	Avd           int    `orm:"avd,double" json:"avd"`
	HarmAdd       int    `orm:"harm_add,double" json:"harm_add"`
	MagicHarmAdd  int    `orm:"magic_harm_add,double" json:"magic_harm_add"`
	MagicNoHarm   int    `orm:"magic_no_harm,double" json:"magic_no_harm"`
	PhysicHarmAdd int    `orm:"physic_harm_add,double" json:"physic_harm_add"`
	PhysicNoHarm  int    `orm:"physic_no_harm,double" json:"physic_no_harm"`
	CureAdd       int    `orm:"cure_add,double" json:"cure_add"`
	BeCureAdd     int    `orm:"be_cure_add,double" json:"be_cure_add"`
	IgnoreDef     int    `orm:"ignore_def,double" json:"ignore_def"`
}

// HeroStageConsume 对应 hero_stage_consume 表
type HeroStageConsume struct {
	Stage      int `orm:"stage,int" json:"stage"`
	LvMax      int `orm:"lv_max,int" json:"lv_max"`
	StageStone int `orm:"stage_stone,varchar" json:"stage_stone"`
	StageGold  int `orm:"stage_gold,varchar" json:"stage_gold"`
}

// HeroStarUpAttr 对应 hero_star_up_attr、hero_star_basic_attr 表
type HeroStarUpAttr struct {
	HeroStar int     `orm:"hero_star,int" json:"hero_star"`
	Hp       float64 `orm:"hp,double" json:"hp"`
	Atk      float64 `orm:"atk,double" json:"atk"`
	Def      float64 `orm:"def,double" json:"def"`
	Speed    float64 `orm:"speed,double" json:"speed"`
}

// HeroStageUpAttr 对应 hero_stage_up_attr、hero_stage_basic_attr 表
type HeroStageUpAttr struct {
	HeroStage int     `orm:"hero_stage,int" json:"hero_stage"`
	Hp        float64 `orm:"hp,double" json:"hp"`
	Atk       float64 `orm:"atk,double" json:"atk"`
	Def       float64 `orm:"def,double" json:"def"`
	Speed     float64 `orm:"speed,double" json:"speed"`
}

// HeroLvBasicAttr 对应 hero_lv_basic_attr 表
type HeroLvBasicAttr struct {
	Id     int     `orm:"id,int,omitempty" json:"id"`
	HeroId int     `orm:"hero_id,int" json:"hero_id"`
	Hp     float64 `orm:"hp,double" json:"hp"`
	Atk    float64 `orm:"atk,double" json:"atk"`
	Def    float64 `orm:"def,double" json:"def"`
	Speed  float64 `orm:"speed,double" json:"speed"`
}

// HeroStarConsume 对应 hero_star_consume 表
type HeroStarConsume struct {
	Star           int    `orm:"star,int" json:"star"`
	StageStone     int    `orm:"stage_stone,int" json:"stage_stone"`
	LvMax          int    `orm:"lv_max,int" json:"lv_max"`
	Self           string `orm:"self,varchar" json:"self"`
	StarUpProperty string `orm:"star_up_property,varchar" json:"star_up_property"`
}

// HeroStar 对应 hero_star 表
type HeroStar struct {
	Star       int    `orm:"star,int" json:"star"`
	Hero       int    `orm:"hero,int" json:"hero"`
	StarUpHero string `orm:"star_up_hero,varchar" json:"star_up_hero"`
}

// HeroInfo 对应 hero_info 表
type HeroInfo struct {
	Id               int    `orm:"id,int,omitempty" json:"id"`
	Name             string `orm:"name,varchar" json:"name"`
	Location         int    `orm:"location,int" json:"location"`
	Property         int    `orm:"property,int" json:"property"`
	HeroOriginalStar int    `orm:"hero_original_star,int" json:"hero_original_star"`
	SkinId           int    `orm:"skin_id,int" json:"skin_id"`
	Skills           string `orm:"skills,varchar" json:"skills"`
	BaseSkill        int    `orm:"base_skill,int" json:"base_skill"`
}

// SkillInfo 对应 skill_info 表
type SkillInfo struct {
	Id                      int     `orm:"id,int,omitempty" json:"id"`
	SkillId                 int     `orm:"skill_id,int" json:"skill_id"`
	Name                    string  `orm:"name,varchar" json:"name"`
	Type                    int8    `orm:"type,tinyint" json:"type"`
	Level                   int8    `orm:"level,tinyint" json:"level"`
	MaxLevel                int     `orm:"max_level,int" json:"max_level"`
	Stage                   int     `orm:"stage,int" json:"stage"`
	Star                    int8    `orm:"star,tinyint" json:"star"`
	TriggerType             int8    `orm:"trigger_type,tinyint" json:"trigger_type"`
	TriggerCd               int8    `orm:"trigger_cd,tinyint" json:"trigger_cd"`
	TriggerDe               int     `orm:"trigger_de,int" json:"trigger_de"`
	TargetId                int     `orm:"target_id,int" json:"target_id"`
	AtkNum                  int8    `orm:"atk_num,tinyint" json:"atk_num"`
	AtkType                 int     `orm:"atk_type,int" json:"atk_type"`
	AtkValue                int     `orm:"atk_value,int" json:"atk_value"`
	Buffs                   string  `orm:"buffs,varchar" json:"buffs"`
	BakTrigger              int8    `orm:"bak_trigger,tinyint" json:"bak_trigger"`
	BakSkill                int     `orm:"bak_skill,int" json:"bak_skill"`
	AttackAnimationPath     string  `orm:"attackAnimation_path,varchar" json:"attackAnimation_path"`
	SkillPath1              string  `orm:"skill_path1,varchar" json:"skill_path1"`
	Position1               int8    `orm:"position1,tinyint" json:"position1"`
	SkillPath2              string  `orm:"skill_path2,varchar" json:"skill_path2"`
	Position2               int     `orm:"position2,int" json:"position2"`
	HitAnimationPath        string  `orm:"hitAnimation_path,varchar" json:"hitAnimation_path"`
	PlayOverTrigger         string  `orm:"play_over_trigger,varchar" json:"play_over_trigger"`
	Sk1Scale                float32 `orm:"sk1Scale,float" json:"sk1Scale"`
	AttaScale               float32 `orm:"attaScale,float" json:"attaScale"`
	HurtTime                int     `orm:"hurt_time,int" json:"hurt_time"`
	Detail                  string  `orm:"detail,varchar" json:"detail"`
	Detail1                 string  `orm:"detail1,varchar" json:"detail1"`
	Updatetime              string  `orm:"updatetime,varchar" json:"updatetime"`
	AttackAnimationPosition int     `orm:"attackAnimation_position,int" json:"attackAnimation_position"`
	HitAnimationPosition    int     `orm:"hitAnimation_position,int" json:"hitAnimation_position"`
	SkillName               int     `orm:"skill_name,int" json:"skill_name"`
}

// SkillTargetChoose 对应 skill_target_choose 表
type SkillTargetChoose struct {
	Id            int  `orm:"id,int,omitempty" json:"id"`
	TargetType    int8 `orm:"target_type,tinyint" json:"target_type"`
	TargetChoose  int  `orm:"target_choose,int" json:"target_choose"`
	TargetPercent int  `orm:"target_percent,int" json:"target_percent"`
	TargetNum     int8 `orm:"target_num,tinyint" json:"target_num"`
}

// SkillBuff 对应 skill_buff 表
type SkillBuff struct {
	Id              int    `orm:"id,int,omitempty" json:"id"`
	Name            string `orm:"name,varchar" json:"name"`
	TargetId        int    `orm:"target_id,int" json:"target_id"`
	FromSkillTarget int8   `orm:"from_skill_target,tinyint" json:"from_skill_target"`
	Rate            int    `orm:"rate,int" json:"rate"`
	SpeTargetId     int    `orm:"spe_target_id,int" json:"spe_target_id"`
	SpeRate         int    `orm:"spe_rate,int" json:"spe_rate"`
	BuffEffectId    int    `orm:"buff_effect_id,int" json:"buff_effect_id"`
	Duration        int    `orm:"duration,int" json:"duration"`
	Remark          string `orm:"remark,varchar" json:"remark"`
	Updatetime      string `orm:"updatetime,varchar" json:"updatetime"`
	BuffId          int    `orm:"buff_id,int" json:"buff_id"`
	Position        int    `orm:"position,int" json:"position"`
	Icon            int    `orm:"icon,int" json:"icon"`
}

// SkillBuffEffect 对应 skill_buff_effect 表
type SkillBuffEffect struct {
	Id int `orm:"id,int,omitempty" json:"id"`
}

// HeroStarUpReturn 对应 hero_star_up_return 表
type HeroStarUpReturn struct {
	Id        int    `orm:"id,int,omitempty" json:"id"`
	Type      int    `orm:"type,int" json:"type"`
	Level     int    `orm:"level,int" json:"level"`
	ReturnNum string `orm:"return_num,varchar" json:"return_num"`
}

// ========================= 道具 =========================

// Items 对应 items 表
type Items struct {
	Id         int       `orm:"id,int,omitempty" json:"id"`
	Name       string    `orm:"name,varchar" json:"name"`
	NameColor  string    `orm:"name_color,varchar" json:"name_color"`
	Type       int       `orm:"type,tinyint" json:"type"`
	TypeSub    int       `orm:"type_sub,tinyint" json:"type_sub"`
	Stack      int       `orm:"stack,int" json:"stack"`
	Lv         int       `orm:"lv,int" json:"lv"`
	Pro        int       `orm:"pro,int" json:"pro"`
	Price      string    `orm:"price,varchar" json:"price"`
	Hp         int       `orm:"hp,int" json:"hp"`
	Atk        int       `orm:"atk,int" json:"atk"`
	Def        int       `orm:"def,int" json:"def"`
	Speed      int       `orm:"speed,int" json:"speed"`
	Open       string    `orm:"open,varchar" json:"open"`
	Suit       int       `orm:"suit,tinyint" json:"suit"`
	Icon       string    `orm:"icon,varchar" json:"icon"`
	Value      string    `orm:"value,varchar" json:"value"`
	Remark     string    `orm:"remark,varchar" json:"remark"`
	Updatetime time.Time `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// ItemsCollection 对应 items_collection 表
type ItemsCollection struct {
	Id              int `orm:"id,int,omitempty" json:"id"`
	CollectionId    int `orm:"collection_id,int" json:"collection_id"`
	ItemsId         int `orm:"items_id,int" json:"items_id"`
	Number          int `orm:"number,int" json:"number"`
	CostType        int `orm:"cost_type,int" json:"cost_type"`
	Price           int `orm:"price,int" json:"price"`
	BuyLimit        int `orm:"buy_limit,int" json:"buy_limit"`
	SaleOff         int `orm:"sale_off,int" json:"sale_off"`
	Probability     int `orm:"probability,int" json:"probability"`
	OrigProbability int `json:"orig_probability"`
	MaxProbability  int `json:"max_probability"`
}

// ========================= 商店 =========================

// ShopSell 对应 shop_sell 表
type ShopSell struct {
	Id              int    `orm:"id,int,omitempty" json:"id"`
	ShopId          int    `orm:"shop_id,int" json:"shop_id"`
	GoodsCollection string `orm:"goods_collection,varchar" json:"goods_collection"`
	GoodsNumber     int    `orm:"goods_number,int" json:"goods_number"`
	CostType        int    `orm:"cost_type,int" json:"cost_type"`
	Price           int    `orm:"price,int" json:"price"`
	BuyLimit        int    `orm:"buy_limit,int" json:"buy_limit"`
	SaleOff         int    `orm:"sale_off,int" json:"sale_off"`
}

// ShopFunctionConfig 对应 shop_function_config 表
type ShopFunctionConfig struct {
	Id            int    `orm:"id,int,omitempty" json:"id"`
	ShopId        int    `orm:"shop_id,int" json:"shop_id"`
	SellType      int    `orm:"sell_type,int" json:"sell_type"`
	BasicCost     int    `orm:"basic_cost,int" json:"basic_cost"`
	RefreshType   string `orm:"refresh_type,varchar" json:"refresh_type"`
	RefreshNumber string `orm:"refresh_number,varchar" json:"refresh_number"`
}

// ========================= 功能配置 =========================

// FunctionConfig 对应 function_config 表
type FunctionConfig struct {
	Id         int    `orm:"id,int,omitempty" json:"id"`
	CopyId     int    `orm:"copy_id,int" json:"copy_id"`
	Name       string `orm:"name,varchar" json:"name"`
	OpenType   string `orm:"open_type,varchar" json:"open_type"`
	CdTime     int    `orm:"cd_time,int" json:"cd_time"`
	FreeCount  int    `orm:"free_count,int" json:"free_count"`
	CostType   int    `orm:"cost_type,int" json:"cost_type"`
	BasicCost  int    `orm:"basic_cost,int" json:"basic_cost"`
	IncreaCost int    `orm:"increa_cost,int" json:"increa_cost"`
	LimitCost  int    `orm:"limit_cost,int" json:"limit_cost"`
	Layer      int    `orm:"layer,int" json:"layer"`
}

// ========================= 祭献 =========================

// SacrificeBase 对应 sacrifice_base 表
type SacrificeBase struct {
	Id            int    `orm:"id,int,omitempty" json:"id"`
	SacrificeType int    `orm:"sacrifice_type,int" json:"sacrifice_type"`
	HeroType      int    `orm:"hero_type,int" json:"hero_type"`
	Star          int    `orm:"star,int" json:"star"`
	Reward        string `orm:"reward,varchar" json:"reward"`
}

// ========================= 抽奖/探宝 =========================

// LuckRewardConfig 对应 luck_reward_config 表
type LuckRewardConfig struct {
	Id             int    `orm:"id,int,omitempty" json:"id"`
	CdTime         int    `orm:"cd_time,int" json:"cd_time"`
	LuckItems      string `orm:"luck_items,varchar" json:"luck_items"`
	CostTypeNumber string `orm:"cost_type_number,varchar" json:"cost_type_number"`
	ExtraGain      string `orm:"extra_gain,varchar" json:"extra_gain"`
}

// LuckIntegralReward 对应 luck_integral_reward 表
type LuckIntegralReward struct {
	Id           int    `orm:"id,int,omitempty" json:"id"`
	Integral     int    `orm:"integral,int" json:"integral"`
	IntegralType int    `orm:"integral_type,int" json:"integral_type"`
	Reward       string `orm:"reward,varchar" json:"reward"`
}

// TaobaoRand 对应 taobao_rand 表
type TaobaoRand struct {
	IdK         int `orm:"id_k,int,omitempty" json:"id_k"`
	Type        int `orm:"type,int" json:"type"`
	Id          int `orm:"id,int" json:"id"`
	Probability int `orm:"probability,int" json:"probability"`
}

// ========================= 礼包 =========================

// LibaoRewards 对应 libao_rewards 表
type LibaoRewards struct {
	Id      int    `orm:"id,int,omitempty" json:"id"`
	Code    string `orm:"code,varchar" json:"code"`
	Rewards string `orm:"rewards,varchar" json:"rewards"`
}

// ========================= 任务 =========================

// TaskAchieve 对应 task_achieve 表（成就任务配置）
type TaskAchieve struct {
	Id       int    `orm:"id,int" json:"id"`
	Type     int    `orm:"type,tinyint" json:"type"`
	Num      int    `orm:"num,int" json:"num"`
	Reward   string `orm:"reward,varchar" json:"reward"`
	ExtraNum int    `orm:"extra_num,int" json:"extra_num"`
}

// TaskActiveConfig 对应 task_active_config 表（活跃度奖励配置）
type TaskActiveConfig struct {
	Id         string `orm:"id,char" json:"id"`
	Type       int    `orm:"type,int" json:"type"`
	TaskReward string `orm:"task_reward,varchar" json:"task_reward"`
	Active     int    `orm:"active,int" json:"active"`
}

// TaskConfig 对应 task_config 表（任务配置）
type TaskConfig struct {
	Id          string `orm:"id,char" json:"id"`
	Type        int    `orm:"type,int" json:"type"`
	TaskType    int    `orm:"task_type,int" json:"task_type"`
	TaskCount   string `orm:"task_count,varchar" json:"task_count"`
	TaskReward  string `orm:"task_reward,varchar" json:"task_reward"`
	Active      int    `orm:"active,int" json:"active"`
	OnOff       int    `orm:"on_off,int" json:"on_off"`
	OnTime      string `orm:"on_time,varchar" json:"on_time"`
	OfTime      string `orm:"of_time,varchar" json:"of_time"`
	OnCondition string `orm:"on_condition,varchar" json:"on_condition"`
}

// TaskGuide 对应 task_guide 表（引导任务配置）
type TaskGuide struct {
	Id     int    `orm:"id,int" json:"id"`
	Type   int    `orm:"type,tinyint" json:"type"`
	Num    int    `orm:"num,tinyint" json:"num"`
	Reward string `orm:"reward,varchar" json:"reward"`
	Detail string `orm:"detail,varchar" json:"detail"`
}

// TaskWeeklyConfig 对应 task_weekly_config 表（周任务配置）
type TaskWeeklyConfig struct {
	Id     int    `orm:"id,int" json:"id"`
	Name   string `orm:"name,varchar" json:"name"`
	Type   int    `orm:"type,int" json:"type"`
	Number int    `orm:"number,int" json:"number"`
	Reward string `orm:"reward,char" json:"reward"`
}

// ========================= VIP =========================

// VipConfig 对应 vip_config 表
type VipConfig struct {
	VipLv               int `orm:"vip_lv,int" json:"vip_lv"`
	RmbTatal            int `orm:"rmb_tatal,int" json:"rmb_tatal"`
	PkNumber            int `orm:"pk_number,int" json:"pk_number"`
	QuickBattleNumber   int `orm:"quick_battle_number,int" json:"quick_battle_number"`
	ClimbtowerNumber    int `orm:"climbtower_number,int" json:"climbtower_number"`
	SpiritshopNumber    int `orm:"spiritshop_number,int" json:"spiritshop_number"`
	TwoTimesSpeedBattle int `orm:"two_times_speed_battle,int" json:"two_times_speed_battle"`
	ProshopNumber       int `orm:"proshop_number,int" json:"proshop_number"`
	LuckyHuntNumber     int `orm:"lucky_hunt_number,int" json:"lucky_hunt_number"`
	HeroIntegralNumber  int `orm:"hero_integral_number,int" json:"hero_integral_number"`
	OfflineExp          int `orm:"offline_exp,int" json:"offline_exp"`
	OfflineCoin         int `orm:"offline_coin,int" json:"offline_coin"`
	OfflineHeroExp      int `orm:"offline_hero_exp,int" json:"offline_hero_exp"`
	OfflineTime         int `orm:"offline_time,int" json:"offline_time"`
	GoldenTouchCoin     int `orm:"golden_touch_coin,int" json:"golden_touch_coin"`
	HeroPackgeNumber    int `orm:"hero_packge_number,int" json:"hero_packge_number"`
	CopyNumber          int `orm:"copy_number,int" json:"copy_number"`
}

// ========================= 远航 =========================

// VoyageFunction 对应 voyage_function 表
type VoyageFunction struct {
	Id            int    `orm:"id,int,omitempty" json:"id"`
	VoyageHero    string `orm:"voyage_hero,varchar" json:"voyage_hero"`
	VoyageTime    int    `orm:"voyage_time,int" json:"voyage_time"`
	VoyageQuicken string `orm:"voyage_quicken,varchar" json:"voyage_quicken"`
}

// VoyageProbability 对应 voyage_probability 表
type VoyageProbability struct {
	Id            int `orm:"id,int,omitempty" json:"id"`
	Probability   int `orm:"probability,int" json:"probability"`
	ItemColection int `orm:"item_colection,int" json:"item_colection"`
	NumMax        int `orm:"num_max,int" json:"num_max"`
}

// ========================= 公会配置 =========================

// GuildMemberLimit 对应 guild_member_limit 表
type GuildMemberLimit struct {
	GuildLv   int `orm:"guild_lv,int" json:"guild_lv"`
	MemberNum int `orm:"member_num,int" json:"member_num"`
	Exp       int `orm:"exp,int" json:"exp"`
}

// GuildSkillAttr 对应 guild_skill_attr 表
type GuildSkillAttr struct {
	Id                 int    `orm:"id,int,omitempty" json:"id"`
	GuildSkillLv       int    `orm:"guild_skill_lv,int" json:"guild_skill_lv"`
	ProfessionType     int    `orm:"profession_type,int" json:"profession_type"`
	AttrKey            int    `orm:"attr_key,tinyint" json:"attr_key"`
	AttrType           int    `orm:"attr_type,int" json:"attr_type"`
	IncreaseType       int    `orm:"increase_type,int" json:"increase_type"`
	AttrPerRangeNumber string `orm:"attr_per_range_number,varchar" json:"attr_per_range_number"`
}

// GuildskillConsume 对应 guildskill_consume 表
type GuildskillConsume struct {
	Id               int    `orm:"id,int,omitempty" json:"id"`
	SkillLv          int    `orm:"skill_lv,int" json:"skill_lv"`
	Lv               int    `orm:"lv,int" json:"lv"`
	GongxianConsume  string `orm:"gongxian_consume,varchar" json:"gongxian_consume"`
	ProfessionalType int    `orm:"professional_type,int" json:"professional_type"`
	AttrKey          int    `orm:"attr_key,int" json:"attr_key"`
}

// GuildTask 对应 guild_task 表
type GuildTask struct {
	TaskId         int `orm:"task_id,char" json:"task_id"`
	TaskCountLimit int `orm:"task_count_limit,int" json:"task_count_limit"`
	TaskActiveExp  int `orm:"task_active_exp,int" json:"task_active_exp"`
	TaskCdTime     int `orm:"task_cd_time,int" json:"task_cd_time"`
	FinishCount    int `orm:"finish_count,int,omitempty" json:"finish_count"`
}

// GuildActiveAttr 对应 guild_active_attr 表
type GuildActiveAttr struct {
	ActiveLv     int    `orm:"active_lv,int" json:"active_lv"`
	ActiveNum    int    `orm:"active_num,int" json:"active_num"`
	ActiveAttr   string `orm:"active_attr,varchar" json:"active_attr"`
	ActiveReward string `orm:"active_reward,varchar" json:"active_reward"`
}

// GuildCopyReward 对应 guild_copy_reward 表
type GuildCopyReward struct {
	Id                 int    `orm:"id,int,omitempty" json:"id"`
	RewardType         int    `orm:"reward_type,int" json:"reward_type"`
	Rank               string `orm:"rank,varchar" json:"rank"`
	Reward             string `orm:"reward,varchar" json:"reward"`
	ItemPerRangeNumber string `orm:"item_per_range_number,varchar" json:"item_per_range_number"`
}

// RecordCalculate 对应 record_calculate 表
type RecordCalculate struct {
	CombatRank int `orm:"combat_rank,int" json:"combat_rank"`
	Star       int `orm:"star,int" json:"star"`
	RecordNum  int `orm:"record_num,int" json:"record_num"`
}

// RecordReward 对应 record_reward 表
type RecordReward struct {
	Id     int    `orm:"id,int,omitempty" json:"id"`
	Rank   string `orm:"rank,varchar" json:"rank"`
	Reward string `orm:"reward,varchar" json:"reward"`
}

// GuildBossConfig 对应 guild_boss_config 表
type GuildBossConfig struct {
	Chapter int   `orm:"chapter,int" json:"chapter"`
	BossId  int   `orm:"boss_id,int" json:"boss_id"`
	BossLv  int   `orm:"boss_lv,int" json:"boss_lv"`
	BossHp  int64 `orm:"boss_hp,bigint" json:"boss_hp"`
	Star    int   `orm:"star,int" json:"star"`
	Stage   int   `orm:"stage,int" json:"stage"`
	Atk     int   `orm:"atk,int" json:"atk"`
	Def     int   `orm:"def,int" json:"def"`
}
