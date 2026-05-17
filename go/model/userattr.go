package model

import (
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo/mem/attr"
)

// ========================= 用户属性 =========================

// GetMultiUserAttr 批量获取用户属性（封装 mem 调用）
func GetMultiUserAttr(uids []int64) map[int64]types.Map {
	ret := make(map[int64]types.Map)
	for _, uid := range uids {
		ret[uid] = attr.GetAll(uid)
	}
	return ret
}

func GetUserAttr(uid int64) types.Map {
	return attr.GetAll(uid)
}

// SetUserInfo 用户升级
func SetUserInfo(userID int64, nickname, avatarUrl string, gender int) {
	attr.Set(userID, config.AttrNickname, nickname)
	attr.Set(userID, config.AttrAvatarUrl, avatarUrl)
	attr.Set(userID, config.AttrGender, gender)
}

// IncrUserLv 用户升级
func IncrUserLv(userID int64) int64 {
	return attr.Incr(userID, config.AttrLv, 1)
}

// IncrUserCopy 副本关卡+1
func IncrUserCopy(userID int64) int64 {
	return attr.Incr(userID, config.AttrCopy, 1)
}

// GetLingqu10Chou 获取领取10连抽标记
func GetLingqu10Chou(uid int64) int {
	v, _ := attr.Get(uid, config.AttrLingqu10Chou)
	return types.ToIntE(v)
}

// SetLingqu10Chou 设置领取10连抽标记
func SetLingqu10Chou(uid int64) {
	attr.Set(uid, config.AttrLingqu10Chou, "1")
}

// GetBegin10Chou 获取开局10连抽标记
func GetBegin10Chou(uid int64) int {
	v, _ := attr.Get(uid, config.AttrBegin10Chou)
	return types.ToIntE(v)
}

// SetBegin10Chou 设置开局10连抽标记
func SetBegin10Chou(uid int64) {
	attr.Set(uid, config.AttrBegin10Chou, "1")
}

// GetUserFinishedGuideID 获取用户已完成的引导任务ID
func GetUserFinishedGuideID(uid int64) int {
	v, _ := attr.Get(uid, config.AttrGuideFinished)
	return types.ToIntE(v)
}

// SetUserFinishedGuideID 设置用户已完成的引导任务ID
func SetUserFinishedGuideID(uid int64, id int) {
	attr.Set(uid, config.AttrGuideFinished, id)
}

// GetFinishFunctionNum 获取完成副本次数
func GetFinishFunctionNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrFinishFunctionNum)
	return types.ToIntE(v)
}

// IncrFinishFunctionNum 增加副本完成数
func IncrFinishFunctionNum(uid int64) {
	attr.Incr(uid, config.AttrFinishFunctionNum, 1)
}

// GetFourstarNum 获取4星英雄数
func GetFourstarNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrFourStarNum)
	return types.ToIntE(v)
}

// IncrFourstarNum 增加4星英雄数
func IncrFourstarNum(uid int64, num int) {
	attr.Incr(uid, config.AttrFourStarNum, int64(num))
}

// GetFivestarNum 获取5星英雄数
func GetFivestarNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrFiveStarNum)
	return types.ToIntE(v)
}

// IncrFivestarNum 增加5星英雄数
func IncrFivestarNum(uid int64, num int) {
	attr.Incr(uid, config.AttrFiveStarNum, int64(num))
}

// GetSixstarNum 获取6星英雄数
func GetSixstarNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrSixStarNum)
	return types.ToIntE(v)
}

// IncrSixstarNum 增加6星英雄数
func IncrSixstarNum(uid int64, num int) {
	attr.Incr(uid, config.AttrSixStarNum, int64(num))
}

// GetGuildbossNum 获取挑战公会boss次数
func GetGuildbossNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrGuildBossNum)
	return types.ToIntE(v)
}

// IncrGuildbossNum 增加公会Boss数
func IncrGuildbossNum(uid int64) {
	attr.Incr(uid, config.AttrGuildBossNum, 1)
}

// GetKillGuildbossNum 获取击杀公会Boss数
func GetKillGuildbossNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrKillGuildBossNum)
	return types.ToIntE(v)
}

// IncrKillGuildbossNum 增加击杀公会Boss数
func IncrKillGuildbossNum(uid int64) {
	attr.Incr(uid, config.AttrKillGuildBossNum, 1)
}

// GetMergeruneNum 获取符文合成次数
func GetMergeruneNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrMergeRuneNum)
	return types.ToIntE(v)
}

// IncrMergeruneNum 增加符文合成次数
func IncrMergeruneNum(uid int64) {
	attr.Incr(uid, config.AttrMergeRuneNum, 1)
}

// GetSacrificeNum 获取祭献数
func GetSacrificeNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrSacrificeNum)
	return types.ToIntE(v)
}

// AddSacrificeNum 增加祭献数
func AddSacrificeNum(uid int64, num int) {
	attr.Incr(uid, config.AttrSacrificeNum, int64(num))
}

// GetUserTanbaoLucky 获取探宝幸运值
func GetUserTanbaoLucky(uid int64, id int) int {
	v, _ := attr.GetByPrefix(uid, config.AttrTanBaoLucky, id)
	return types.ToIntE(v)
}

// AddUserTanbaoLucky 增加探宝幸运值
func AddUserTanbaoLucky(uid int64, id, num int) int64 {
	if num == 0 {
		num = 100
	}
	return attr.IncrByPrefix(uid, config.AttrTanBaoLucky, id, int64(num))
}

// GetVoyageChengNum 获取橙色远航计数
func GetVoyageChengNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrVoyageChengNum)
	return types.ToIntE(v)
}

// IncrVoyageChengNum 增加橙色远航计数
func IncrVoyageChengNum(uid int64) {
	attr.Incr(uid, config.AttrVoyageChengNum, 1)
}

// GetVoyageHongNum 获取红色远航计数
func GetVoyageHongNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrVoyageHongNum)
	return types.ToIntE(v)
}

// IncrVoyageHongNum 增加红色远航计数
func IncrVoyageHongNum(uid int64) {
	attr.Incr(uid, config.AttrVoyageHongNum, 1)
}

// GetXianzhiNum 获取先知召唤次数
func GetXianzhiNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrXianZhiNum)
	return types.ToIntE(v)
}

// IncrXianzhiNum 增加先知召唤次数
func IncrXianzhiNum(uid int64) {
	attr.Incr(uid, config.AttrXianZhiNum, 1)
}

// GetXingheNum 获取星河挑战次数（成就）
func GetXingheNum(uid int64) int {
	v, _ := attr.Get(uid, config.AttrXingHeNum)
	return types.ToIntE(v)
}

// IncrXingheNum 增加星河挑战次数（成就）
func IncrXingheNum(uid int64) {
	attr.Incr(uid, config.AttrXingHeNum, 1)
}

// ChooseRoleTitle 选择称号（封装 mem.Set）
func ChooseRoleTitle(uid int64, roleTitleID int) {
	attr.Set(uid, config.AttrRoleTitle, roleTitleID)
}

func SetVipLevel(uid int64, val int) {
	attr.Set(uid, config.AttrVipLevel, val)
}

func IncrUserDay7(uid int64) int64 {
	return attr.Incr(uid, config.AttrDay7, 1)
}

func SetIGift(uid int64) {
	attr.Set(uid, config.AttrIGift, 1)
}

func IncrUserNewGift(uid int64) int64 {
	return attr.Incr(uid, config.AttrNewGift, 1)
}

func SetInitFlag(uid int64) {
	attr.Set(uid, config.AttrInitFlag, 1)
}
