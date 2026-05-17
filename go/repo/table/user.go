package table

// ========================= shine_user 库 - 用户信息表 =========================

import (
	"time"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
)

// GameVersion 对应 game_version 表
type GameVersion struct {
	Id          int    `orm:"id,int,omitempty" json:"id"`
	Ver         string `orm:"ver,varchar" json:"ver"`
	ForceUpdate int    `orm:"force_update,int" json:"force_update"`
	Url         string `orm:"url,varchar" json:"url"`
	UpdateTime  string `orm:"update_time,varchar" json:"update_time"`
}

// NewTradeInfo 对应 new_trade_info 表
type NewTradeInfo struct {
	Id            int       `orm:"id,int,omitempty" json:"id"`
	Type          int       `orm:"type,int" json:"type"`
	ZoneId        int       `orm:"zone_id,varchar" json:"zone_id"`
	UserId        int64     `orm:"user_id,bigint" json:"user_id"`
	OpenId        string    `orm:"open_id,varchar" json:"open_id"`
	Status        string    `orm:"status,varchar" json:"status"`
	TotalFee      int       `orm:"total_fee,int" json:"total_fee"`
	Ip            string    `orm:"ip,varchar" json:"ip"`
	TransactionId string    `orm:"transaction_id,varchar" json:"transaction_id"`
	OutTradeNo    string    `orm:"out_trade_no,varchar" json:"out_trade_no"`
	Pf            string    `orm:"pf,varchar" json:"pf"`
	Ts            int       `orm:"ts,int" json:"ts"`
	Channel       string    `orm:"channel,varchar" json:"channel"`
	Createtime    time.Time `orm:"createtime,timestamp,oncreatetime" json:"createtime"`
	Updatetime    time.Time `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// Sequence 对应 sequence 表
type Sequence struct {
	SequenceNo int `orm:"sequence_no,int,omitempty" json:"sequence_no"`
	Time       int `orm:"time,int" json:"time"`
}

// UserInfo 对应 user_info 表
type UserInfo struct {
	UserId        int64     `orm:"user_id,int" json:"user_id"`
	Appid         string    `orm:"appid,varchar" json:"appid"`
	OpenId        string    `orm:"open_id,varchar" json:"open_id"`
	Union         string    `orm:"union,varchar" json:"union"`
	SessionKey    string    `orm:"session_key,varchar" json:"session_key"`
	Nickname      string    `orm:"nickname,varchar" json:"nickname"`
	City          string    `orm:"city,varchar" json:"city"`
	AvatarUrl     string    `orm:"avatar_url,varchar" json:"avatar_url"`
	Gender        int       `orm:"gender,int" json:"gender"`
	FormId        string    `orm:"form_id,varchar" json:"form_id"`
	Token         string    `orm:"token,varchar" json:"token"`
	LastLoginTime string    `orm:"last_login_time,varchar" json:"last_login_time"`
	LoginTimes    int       `orm:"login_times,int" json:"login_times"`
	RegistTime    string    `orm:"regist_time,varchar" json:"regist_time"`
	Updatetime    time.Time `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// InviteInfo 对应 invite_info 表（邀请信息）
type InviteInfo struct {
	UserId        int64             `orm:"user_id,int" json:"user_id"`
	InviteZoneUid int64             `orm:"invite_zone_uid,bigint" json:"invite_zone_uid"`
	Updatetime    extorm.NullString `orm:"updatetime,timestamp,onupdatetime" json:"updatetime"`
}

// ZoneInfoTable 对应 zone_info 表
type ZoneInfoTable struct {
	Id            int    `orm:"id,int,omitempty" json:"id"`
	ZoneName      string `orm:"zone_name,varchar" json:"zone_name"`
	Status        int    `orm:"status,int" json:"status"`
	MaintainStart string `orm:"maintain_start,varchar" json:"maintain_start"`
	MaintainEnd   string `orm:"maintain_end,varchar" json:"maintain_end"`
}
