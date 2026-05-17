package table

// ========================= shine_world_zone1 库 - 区服动态数据表 =========================

import (
	"time"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
)

// ========================= 公会 =========================

// Guild 对应 guild 表
type Guild struct {
	Id        int    `orm:"id,int,omitempty" json:"id"`
	GuildName string `orm:"guild_name,varchar" json:"guild_name"`
	GuildLv   int    `orm:"guild_lv,int,omitempty" json:"guild_lv"`
	Exp       int    `orm:"exp,int" json:"exp"`
	PeopleNum int    `orm:"people_num,int,omitempty" json:"people_num"`
	Creator   int64  `orm:"creator,bigint" json:"creator"`
	OwnUser   int64  `orm:"own_user,bigint" json:"own_user"`
	LvLimit   int    `orm:"lv_limit,int" json:"lv_limit"`
	NeedCheck int    `orm:"need_check,int" json:"need_check"`
	Declar    string `orm:"declar,varchar" json:"declar"`
}

// UsersGuild 对应 users_guild 表
type UsersGuild struct {
	Id      int   `orm:"id,int,omitempty" json:"id"`
	UserId  int64 `orm:"user_id,bigint" json:"user_id"`
	GuildId int   `orm:"guild_id,int" json:"guild_id"`
	Zhiwei  int   `orm:"zhiwei,int" json:"zhiwei"`
}

// GuildApply 对应 guild_apply 表
type GuildApply struct {
	Id      int   `orm:"id,int,omitempty" json:"id"`
	GuildId int64 `orm:"guild_id,bigint" json:"guild_id"`
	UserId  int64 `orm:"user_id,bigint" json:"user_id"`
}

// GuildChapterBlood 对应 guild_chapter_blood 表
type GuildChapterBlood struct {
	Id           int `orm:"id,int" json:"id"`
	Chapter      int `orm:"chapter,int" json:"chapter"`
	ChapterBlood int `orm:"chapter_blood,int" json:"chapter_blood"`
}

// ========================= 用户数据 =========================

// UserClimbtowerRecord 对应 user_climbtower_record 表
type UserClimbtowerRecord struct {
	Layer            int    `orm:"layer,int" json:"layer"`
	LowestUserId     int64  `orm:"lowest_user_id,bigint" json:"lowest_user_id"`
	LowestNickname   string `orm:"lowest_nickname,mediumblob" json:"lowest_nickname"`
	LowestFightPoint int    `orm:"lowest_fight_point,int" json:"lowest_fight_point"`
	FastUserId       int64  `orm:"fast_user_id,bigint" json:"fast_user_id"`
	FastNickname     string `orm:"fast_nickname,mediumblob" json:"fast_nickname"`
}

// UserContents 对应 user_contents 表
type UserContents struct {
	UserId     int64             `orm:"user_id,bigint" json:"user_id"`
	Type       string            `orm:"type,varchar" json:"type"`
	Content    string            `orm:"content,varchar" json:"content"`
	Updatetime extorm.NullString `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// UserContentsInt 对应 user_contents_int 表
type UserContentsInt struct {
	UserId     int64             `orm:"user_id,bigint" json:"user_id"`
	Type       string            `orm:"type,varchar" json:"type"`
	Content    int               `orm:"content,int" json:"content"`
	Updatetime extorm.NullString `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// UserEndlessHelpHero 对应 user_endless_help_hero 表
type UserEndlessHelpHero struct {
	Id         int       `orm:"id,int,omitempty" json:"id"`
	UserId     int64     `orm:"user_id,bigint" json:"user_id"`
	HeroId     int       `orm:"hero_id,int" json:"hero_id"`
	Updatetime time.Time `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// UserExpeditionHelpHero 对应 user_expedition_help_hero 表
type UserExpeditionHelpHero struct {
	Id         int       `orm:"id,int,omitempty" json:"id"`
	UserId     int64     `orm:"user_id,bigint" json:"user_id"`
	HeroId     int       `orm:"hero_id,int" json:"hero_id"`
	Updatetime time.Time `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// UsersFriends 对应 users_friends 表
type UsersFriends struct {
	Id     int   `orm:"id,int,omitempty" json:"id"`
	User1  int64 `orm:"user1,bigint" json:"user1"`
	User2  int64 `orm:"user2,bigint" json:"user2"`
	Status int   `orm:"status,int" json:"status"`
}

// UserHero 对应 user_hero 表
type UserHero struct {
	Id         int       `orm:"id,int,omitempty" json:"id"`
	UserId     int64     `orm:"user_id,bigint" json:"user_id"`
	HeroId     int       `orm:"hero_id,int" json:"hero_id"`
	Star       int       `orm:"star,int,omitempty" json:"star"`
	Stage      int       `orm:"stage,int,omitempty" json:"stage"`
	Lv         int       `orm:"lv,int,omitempty" json:"lv"`
	Fit        string    `orm:"fit,varchar" json:"fit"`
	Fu         string    `orm:"fu,varchar" json:"fu"`
	Updatetime time.Time `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// UserMail 对应 user_mail 表
type UserMail struct {
	Id           int       `orm:"id,int,omitempty" json:"id"`
	UserId       int64     `orm:"user_id,bigint" json:"user_id"`
	ReadFlag     int       `orm:"read_flag,tinyint" json:"read_flag"`
	Title        string    `orm:"title,varchar" json:"title"`
	Content      string    `orm:"content,varchar" json:"content"`
	ItemGetFlag  int       `orm:"item_get_flag,tinyint" json:"item_get_flag"`
	AddItems     string    `orm:"add_items,varchar" json:"add_items"`
	SendTime     string    `orm:"send_time,varchar" json:"send_time"`
	DeadlineTime int       `orm:"deadline_time,int" json:"deadline_time"`
	Updatetime   time.Time `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// UserNickname 对应 user_nickname 表
type UserNickname struct {
	UserId   int64  `orm:"user_id,bigint" json:"user_id"`
	Nickname string `orm:"nickname,varchar" json:"nickname"`
}

// UserPosition 对应 user_position 表
type UserPosition struct {
	Id       int   `orm:"id,int,omitempty" json:"id"`
	UserId   int64 `orm:"user_id,int" json:"user_id"`
	PosType  int   `orm:"pos_type,int" json:"pos_type"`
	Position int   `orm:"position,int" json:"position"`
	Pos1Pos  int   `orm:"pos1_pos,int" json:"pos1_pos"`
	Pos1Hero int   `orm:"pos1_hero,int" json:"pos1_hero"`
	Pos2Pos  int   `orm:"pos2_pos,int" json:"pos2_pos"`
	Pos2Hero int   `orm:"pos2_hero,int" json:"pos2_hero"`
	Pos3Pos  int   `orm:"pos3_pos,int" json:"pos3_pos"`
	Pos3Hero int   `orm:"pos3_hero,int" json:"pos3_hero"`
	Pos4Pos  int   `orm:"pos4_pos,int" json:"pos4_pos"`
	Pos4Hero int   `orm:"pos4_hero,int" json:"pos4_hero"`
	Pos5Pos  int   `orm:"pos5_pos,int" json:"pos5_pos"`
	Pos5Hero int   `orm:"pos5_hero,int" json:"pos5_hero"`
}

// UserRoleTitle 对应 user_role_title 表
type UserRoleTitle struct {
	Id         int       `orm:"id,int,omitempty" json:"id"`
	UserId     int64     `orm:"user_id,bigint" json:"user_id"`
	RoleId     int       `orm:"role_id,int" json:"role_id"`
	Updatetime time.Time `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// UserTongguanReward 对应 user_tongguan_reward 表
type UserTongguanReward struct {
	Id         int       `orm:"id,int,omitempty" json:"id"`
	UserId     int64     `orm:"user_id,bigint" json:"user_id"`
	Type       int       `orm:"type,int" json:"type"`
	Copy       int       `orm:"copy,int" json:"copy"`
	Status     int       `orm:"status,int" json:"status"`
	Updatetime time.Time `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// UserTutengPk 对应 user_tuteng_pk 表
type UserTutengPk struct {
	Id     int    `orm:"id,int,omitempty" json:"id"`
	UserId int64  `orm:"user_id,bigint" json:"user_id"`
	Score  int    `orm:"score,int" json:"score"`
	Data   string `orm:"data,varchar" json:"data"`
}

// UserTutengPkDetail 对应 user_tuteng_pk_detail 表
type UserTutengPkDetail struct {
	Id        int    `orm:"id,int,omitempty" json:"id"`
	UserId    int64  `orm:"user_id,bigint" json:"user_id"`
	OppUserId int64  `orm:"opp_user_id,bigint" json:"opp_user_id"`
	Winner    int    `orm:"winner,int" json:"winner"`
	WinScore  int    `orm:"win_score,int" json:"win_score"`
	LoseScore int    `orm:"lose_score,int" json:"lose_score"`
	PkTime    string `orm:"pk_time,varchar" json:"pk_time"`
}

// UserVipContents 对应 user_vip_contents 表
type UserVipContents struct {
	UserId     int64             `orm:"user_id,bigint" json:"user_id"`
	Content    string            `orm:"content,varchar" json:"content"`
	Updatetime extorm.NullString `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// TemplateLevel 对应 template_level 表
type TemplateLevel struct {
	Id int `orm:"id,int" json:"id"`
	Lv int `orm:"lv,int" json:"lv"`
}

// XingheInfo 对应 xinghe_info 表
type XingheInfo struct {
	Id         int    `orm:"id,int,omitempty" json:"id"`
	ZoneId     int    `orm:"zone_id,int" json:"zone_id"`
	Content    string `orm:"content,text" json:"content"`
	UpdateTime string `orm:"update_time,varchar" json:"update_time"`
}
