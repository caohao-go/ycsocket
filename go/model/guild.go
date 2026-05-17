package model

import (
	"context"
	"fmt"
	"sort"
	"time"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo"
	"server_golang/repo/cache"
	"server_golang/repo/mem/content"
	"server_golang/repo/mem/daily"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

// ---- 公会系统 ----

type GuildInfo struct {
	*table.Guild
	MemberLimit  int    `json:"member_limit"`
	OwnNickname  string `json:"own_nickname"`
	OwnAvatarURL string `json:"own_avatar_url"`
	FightPoint   int    `json:"fight_point"`
}

// CreateGuild 创建公会
func CreateGuild(ctx context.Context, data *table.Guild) (int, error) {
	id, err := world.InsertGuild(ctx, data)
	return int(id), err
}

// GetGuildInfoByID 获取公会信息（返回 struct，用于只读调用方）
func GetGuildInfoByID(ctx context.Context, id int) *table.Guild {
	return world.GetGuildById(ctx, id)
}

// AddGuildPeopleNum 增加公会人数
func AddGuildPeopleNum(ctx context.Context, id int) error {
	cache.Del(fmt.Sprintf(config.CacheGuildInfo, id))
	return world.IncrGuildPeopleNum(ctx, id)
}

// DeleteGuildPeopleNum 减少公会人数
func DeleteGuildPeopleNum(ctx context.Context, id int) error {
	cache.Del(fmt.Sprintf(config.CacheGuildInfo, id))
	return world.DecrGuildPeopleNum(ctx, id)
}

// UpdateGuildInfoByID 更新公会信息
func UpdateGuildInfoByID(ctx context.Context, id int, data types.Map) error {
	return world.UpdateGuildByIdWithCache(ctx, id, data)
}

// GetGuildListByName 按名称搜索公会（带会长信息和成员上限，与 PHP 一致）
func GetGuildListByName(ctx context.Context, name string) ([]*GuildInfo, error) {
	rows, err := world.GetGuildListByName(ctx, name)
	if err != nil {
		return nil, err
	}
	return enrichGuildList(ctx, rows), nil
}

// GetGuildList 获取公会列表（带会长信息和成员上限，与 PHP 一致）
func GetGuildList(ctx context.Context) ([]*GuildInfo, error) {
	rows, err := world.GetGuildListTop50(ctx)
	if err != nil {
		return nil, err
	}
	return enrichGuildList(ctx, rows), nil
}

// InsertUsersGuild 插入用户公会关系（与 PHP 一致：先检查人数上限，再加人数）
func InsertUsersGuild(ctx context.Context, data *table.UsersGuild, userID int64) (bool, error) {
	guildID := data.GuildId
	guildInfo := GetGuildInfoByID(ctx, guildID)
	if guildInfo == nil {
		return false, fmt.Errorf("guild not found")
	}

	// 检查人数上限（与 PHP 一致）
	guildLv := guildInfo.GuildLv
	peopleNum := guildInfo.PeopleNum
	memberLimit := 0
	if lvData, ok := logic.GuildLvDatas[guildLv]; ok {
		memberLimit = lvData.MemberNum
	}
	if memberLimit > 0 && peopleNum >= memberLimit {
		return false, nil
	}

	// 增加公会人数
	AddGuildPeopleNum(ctx, guildID)

	_, err := world.InsertUsersGuild(ctx, data)
	if err != nil {
		return false, err
	}
	return true, nil
}

// UserQuitGuild 用户退出公会（与 PHP 一致：先减人数再删记录）
func UserQuitGuild(ctx context.Context, userID int64) error {
	usersGuild := GetUsersGuildInfo(ctx, userID)
	if usersGuild == nil {
		return nil
	}
	guildID := usersGuild.GuildId
	// 减掉一个人
	DeleteGuildPeopleNum(ctx, guildID)
	// 删除个人公会信息
	return world.DeleteUsersGuildByUserId(ctx, userID)
}

// GetUsersGuildDetail 获取用户公会详情（与 PHP 一致：JOIN 查询）
func GetUsersGuildDetail(ctx context.Context, userID int64) *table.Guild {
	usersGuild := GetUsersGuildInfo(ctx, userID)
	if usersGuild == nil {
		return nil
	}
	return GetGuildInfoByID(ctx, usersGuild.GuildId)
}

// GetUsersGuildInfo 获取用户公会信息
func GetUsersGuildInfo(ctx context.Context, userID int64) *table.UsersGuild {
	return world.GetUsersGuildByUserId(ctx, userID)
}

// GetUsersGuildID 获取用户公会ID（与 PHP 一致：返回 guild_id 整数）
func GetUsersGuildID(ctx context.Context, userID int64) int {
	ug := world.GetUsersGuildByUserId(ctx, userID)
	if ug == nil {
		return 0
	}
	return ug.GuildId
}

// UpdateUsersGuild 更新用户公会信息
func UpdateUsersGuild(ctx context.Context, userID int64, data types.Map) error {
	return world.UpdateUsersGuildByUserId(ctx, userID, data)
}

// GetGuildFuhuizhangCount 获取公会副会长数量
func GetGuildFuhuizhangCount(ctx context.Context, guildID int) int {
	cnt, _ := world.GetGuildFuhuizhangCount(ctx, guildID)
	return cnt
}

// GetGuildsUser 获取公会成员列表（与 PHP 一致：富化用户详情、贡献、活跃等级）
func GetGuildsUser(ctx context.Context, guildID int, onlyID bool, meUserID int64) []types.Map {
	members, err := world.GetGuildsUserByGuildId(ctx, guildID, onlyID)
	if err != nil || len(members) == 0 {
		return nil
	}

	if onlyID {
		// 仅返回 uid 列表
		result := make([]types.Map, 0, len(members))
		for _, m := range members {
			result = append(result, types.Map{"user_id": m.UserId})
		}
		return result
	}

	// 收集所有用户 ID
	uids := make([]int64, 0, len(members))
	for _, m := range members {
		uids = append(uids, m.UserId)
	}

	// 批量获取用户详情
	userDetail := GetUsersWithDetail(ctx, uids, 1, config.AttrLv, config.AttrFightPoint, config.AttrOffTime)

	// 获取公会今日贡献和总贡献
	todayGongxian := GetTodayGuildGongxian(ctx, guildID)
	gongxian := GetGuildGongxian(ctx, guildID)

	// 批量获取公会活跃值
	guildActiveMap := content.GetMoreInt(uids, "guild_active")

	// 分成三类：我自己、在线、离线
	meInfo := make([]types.Map, 0)
	online := make([]types.Map, 0)
	notOnline := make([]types.Map, 0)

	zoneID := logic.GetShineZoneID()

	for _, member := range members {
		uid := member.UserId
		v := types.Map{
			"id":      member.Id,
			"user_id": uid,
			"zhiwei":  member.Zhiwei,
			"zone_id": zoneID,
		}

		if ud, ok := userDetail[uid]; ok {
			v["nickname"] = ud.GetStringE("nickname")
			v["avatar_url"] = ud.GetStringE("avatar_url")
			v["lv"] = ud.GetIntE("lv")
			v["fight_point"] = ud.GetIntE("fight_point")
			v["off_time"] = ud.GetStringE("off_time")
		}

		// 活跃等级
		guildActive := guildActiveMap[uid]
		activeLvData := logic.GetActiveLv(guildActive)
		v["active_lv"] = activeLvData.GetIntE("current_lv")

		// 今日贡献和总贡献
		v["today_gongxian"] = todayGongxian.GetIntE(uid)
		v["gongxian"] = gongxian.GetIntE(uid)

		if uid == meUserID {
			meInfo = append(meInfo, v)
		} else if v.GetStringE("off_time") == "" || v.GetStringE("off_time") == "0" {
			online = append(online, v)
		} else {
			notOnline = append(notOnline, v)
		}
	}

	// 在线按战斗力降序排列
	sort.Slice(online, func(i, j int) bool {
		return online[i].GetIntE("fight_point") > online[j].GetIntE("fight_point")
	})

	// 合并: 我自己 + 在线 + 离线
	result := make([]types.Map, 0, len(members))
	result = append(result, meInfo...)
	result = append(result, online...)
	result = append(result, notOnline...)

	return result
}

// GetGuildRank 获取公会排名（与 PHP 一致：计算所有成员总战斗力，按战斗力降序）
func GetGuildRank(ctx context.Context) ([]*GuildInfo, error) {
	rows, err := world.GetGuildAll(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]*GuildInfo, 0, len(rows))

	for _, guild := range rows {
		gMap := GuildInfo{
			Guild: guild,
		}

		// 成员上限
		memberLimit := 0
		if lvData, ok := logic.GuildLvDatas[guild.GuildLv]; ok {
			memberLimit = lvData.MemberNum
		}
		gMap.MemberLimit = memberLimit

		// 计算公会总战斗力（所有成员的战斗力之和）
		guildMembers, _ := world.GetGuildsUserByGuildId(ctx, guild.Id, true)
		totalFP := 0
		for _, member := range guildMembers {
			totalFP += GetUserFightPoint(ctx, member.UserId, 1)
		}
		gMap.FightPoint = totalFP

		// 会长信息
		ownUserInfo := GetUserAttr(guild.OwnUser)
		if ownUserInfo != nil {
			gMap.OwnNickname = ownUserInfo.GetStringE("nickname")
			gMap.OwnAvatarURL = ownUserInfo.GetStringE("avatar_url")
		} else {
			gMap.OwnNickname = ""
			gMap.OwnAvatarURL = ""
		}

		result = append(result, &gMap)
	}

	// 按战斗力降序排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].FightPoint > result[j].FightPoint
	})

	return result, nil
}

// GuildChangeOwner 公会转让（与 PHP 一致：更新 guild 表 own_user + 更新两人的 zhiwei）
func GuildChangeOwner(ctx context.Context, guildID int, userID, toUserID int64) error {
	UpdateGuildInfoByID(ctx, guildID, types.Map{"own_user": toUserID})
	UpdateUsersGuild(ctx, userID, types.Map{"zhiwei": 0})
	UpdateUsersGuild(ctx, toUserID, types.Map{"zhiwei": 1})
	return nil
}

// GetGuildsApplyUser 获取公会申请列表（与 PHP 一致：从 Pika 获取，富化用户信息）
func GetGuildsApplyUser(ctx context.Context, guildID int) []types.Map {
	datas := GetUsersGuildApply(ctx, guildID)
	if len(datas) == 0 {
		return []types.Map{}
	}

	uids := make([]int64, 0, len(datas))
	for _, sm := range datas {
		uids = append(uids, types.ToInt64E(sm.Member))
	}

	userDetails := GetUsersWithDetail(ctx, uids, 1, config.AttrLv, config.AttrFightPoint, config.AttrOffTime)

	result := make([]types.Map, 0, len(datas))
	for _, uid := range uids {
		if ud, ok := userDetails[uid]; ok {
			result = append(result, ud)
		}
	}
	return result
}

// AddUsersGuildApply 添加公会申请（使用 Pika，与 PHP 一致）
func AddUsersGuildApply(ctx context.Context, guildID int, userID int64) {
	AddUsersApplyGuild(ctx, guildID, userID)
}

// DelUsersGuildApply 删除公会申请（使用 Pika，与 PHP 一致）
func DelUsersGuildApply(ctx context.Context, guildID int, userID int64) {
	DelUsersApplyGuild(ctx, guildID, userID)
}

// enrichGuildList 为公会列表添加会长信息和成员上限（与 PHP getGuildList/getGuildListByName 一致）
func enrichGuildList(ctx context.Context, rows []*table.Guild) []*GuildInfo {
	result := make([]*GuildInfo, 0, len(rows))
	// 收集会长 UID
	ownerUIDs := make([]int64, 0, len(rows))
	for _, g := range rows {
		ownerUIDs = append(ownerUIDs, g.OwnUser)
	}
	// 批量获取会长信息
	ownerInfos := GetUsersWithDetail(ctx, ownerUIDs, 1)

	for _, g := range rows {
		gMap := GuildInfo{Guild: g}

		// 成员上限
		memberLimit := 0
		if lvData, ok := logic.GuildLvDatas[g.GuildLv]; ok {
			memberLimit = lvData.MemberNum
		}
		gMap.MemberLimit = memberLimit

		// 会长昵称和头像
		if oi, ok := ownerInfos[g.OwnUser]; ok {
			gMap.OwnNickname = oi.GetStringE("nickname")
			gMap.OwnAvatarURL = oi.GetStringE("avatar_url")
		} else {
			gMap.OwnNickname = ""
			gMap.OwnAvatarURL = ""
		}
		result = append(result, &gMap)
	}
	return result
}

// GetGuildSkills 获取公会技能（与 PHP 一致：返回复杂结构，含 attr 信息、消耗、状态）
func GetGuildSkills(ctx context.Context, userID int64) []types.Map {
	ret := make([]types.Map, 0, 4)
	datas := content.GetMap(userID, "guild_skills")

	for key := 0; key < 4; key++ {
		data := datas.GetIntE(key)

		skillInfo := types.Map{
			"profession_type": key + 1,
			"max_lv":          40,
		}

		if data == 0 {
			// 初始化
			dataInit := types.Map{types.ToString(key): types.Map{"lv": 1, "attr_lv": []int{0, 0, 0, 0, 0, 0}}}
			content.SetMap(userID, "guild_skills", dataInit)

			skillInfo["status"] = 0
			skillInfo["lv"] = 1
			attrLvs := []int{0, 0, 0, 0, 0, 0}
			skillInfo["attr_lv"] = getAttrInfo(key+1, attrLvs)

			// 获取消耗
			needNum8, needNum1 := getSkillConsume(key+1, 1, 0)
			skillInfo["cost"] = []types.Map{
				{"cost_type": 8, "basic_cost": needNum8},
				{"cost_type": 1, "basic_cost": needNum1},
			}
		} else {
			dataMap, _ := types.ToMap(data, "")
			lv := dataMap.GetIntE("lv")
			attrLvRaw := dataMap["attr_lv"]

			skillInfo["status"] = 1
			skillInfo["lv"] = lv

			// 转换 attr_lv 为数组
			attrLvs, _ := types.ToIntArray(attrLvRaw)
			attrLvInfo := getAttrInfo(key+1, attrLvs)
			skillInfo["attr_lv"] = attrLvInfo

			// 找出下一个要升级的 attr_key
			attrKey := 0
			for k, alv := range attrLvInfo {
				if alv.GetIntE("lv") < lv {
					attrKey = k
					break
				}
			}

			// 判断是否原始状态
			isActive := true
			for _, alv := range attrLvInfo {
				if alv.GetIntE("lv") != 0 {
					isActive = false
					break
				}
			}

			needNum8, needNum1 := 0, 0
			if !isActive {
				needNum8, needNum1 = getSkillConsume(key+1, lv, attrKey)
			} else {
				needNum8, needNum1 = getSkillConsume(key+1, 1, 0)
			}
			skillInfo["cost"] = []types.Map{
				{"cost_type": 8, "basic_cost": needNum8},
				{"cost_type": 1, "basic_cost": needNum1},
			}
		}

		ret = append(ret, skillInfo)
	}
	return ret
}

// ActiveGuildSkill 激活公会技能（与 PHP activeGuildSkills 一致：attr_lv 轮转升级逻辑）
func ActiveGuildSkill(ctx context.Context, userID int64, prop int) {
	key := prop - 1

	datas := content.GetMap(userID, "guild_skills")
	if datas == nil {
		datas = make(types.Map)
	}

	data := datas.GetIntE(key)
	if data == 0 {
		// 初始化: lv=1, attr_lv=[1,0,0,0,0,0]（第一个属性升1级）
		datas[types.ToString(key)] = types.Map{"lv": 1, "attr_lv": []int{1, 0, 0, 0, 0, 0}}
		content.SetMap(userID, "guild_skills", datas)
		cache.ClearAddPropCache(fmt.Sprintf(config.CacheAddProp, userID))
		return
	}

	dataMap, _ := types.ToMap(data, "")
	lv := dataMap.GetIntE("lv")
	attrLvRaw := dataMap["attr_lv"]
	attrLvs, _ := types.ToIntArray(attrLvRaw)

	// 找到第一个 attr_lv < lv 的，升级它
	lastK := 0
	for k := 0; k < 6; k++ {
		if attrLvs[k] < lv {
			if attrLvs[k] < 40 {
				attrLvs[k]++
			}
			lastK = k
			break
		}
	}

	// 如果是第6个（k==5），6个属性全加满了，整体升级
	if lastK == 5 {
		if lv < 40 {
			lv++
		}
	}

	dataMap["lv"] = lv
	dataMap["attr_lv"] = attrLvs
	datas[types.ToString(key)] = dataMap

	content.SetMap(userID, "guild_skills", datas)
	cache.ClearAddPropCache(fmt.Sprintf(config.CacheAddProp, userID))
}

// getAttrInfo 获取技能属性信息（与 PHP getAttrInfo 一致）
func getAttrInfo(professionType int, attrLvs []int) []types.Map {
	ret := make([]types.Map, 0, 6)
	skillData := logic.GuildSkillDatas[professionType]

	for k, lv := range attrLvs {
		info := skillData[k]

		tmp := types.Map{
			"attr_type":             info.AttrType,
			"lv":                    lv,
			"increase_type":         info.IncreaseType,
			"attr_per_range_number": types.ToIntE(info.AttrPerRangeNumber),
			"attr_add_total":        lv * types.ToIntE(info.AttrPerRangeNumber),
		}
		ret = append(ret, tmp)
	}
	return ret
}

// getSkillConsume 获取技能升级消耗（帮贡和金币）
func getSkillConsume(professionType, lv, attrKey int) (int, int) {
	needNum8, needNum1 := 0, 0
	if consumeData, ok := logic.GuildSkillConsumeDatas[professionType]; ok {
		if lvData, ok := consumeData[lv]; ok {
			if consume, ok := lvData[attrKey]; ok {
				for _, c := range consume {
					if types.ToIntE(c.Type) == 8 {
						needNum8 = c.Num
					}
					if types.ToIntE(c.Type) == 1 {
						needNum1 = c.Num
					}
				}
			}
		}
	}
	return needNum8, needNum1
}

// GetGuildContributeActive 获取公会捐献活跃值
func GetGuildContributeActive(ctx context.Context, gid int) int {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyGuildContributeActive, gid, util.DateYmd()))
	return types.ToIntE(v)
}

// IncrGuildContributeActive 增加公会捐献活跃值
func IncrGuildContributeActive(ctx context.Context, gid int, id int) {
	k := fmt.Sprintf(config.KeyGuildContributeActive, gid, util.DateYmd())
	var num int64
	switch id {
	case 3:
		num = 200
	case 2:
		num = 100
	default:
		num = 50
	}
	repo.RedisIncrBy(ctx, k, num)
	repo.RedisExpire(ctx, k, 86400)
}

func GetTodayGuildGongxian(ctx context.Context, gid int) types.Map {
	v, _ := repo.RedisHGetAll(ctx, fmt.Sprintf(config.KeyGuildTodayGongxian, gid, util.DateYmd()))
	return v
}

func IncrTodayGuildGongxian(ctx context.Context, gid, uid int64, num int) {
	hk := fmt.Sprintf(config.KeyGuildTodayGongxian, gid, util.DateYmd())
	repo.RedisHIncrBy(ctx, hk, uid, int64(num))
	repo.RedisExpire(ctx, hk, 86400)
}

// GetGuildGongxian 获取公会贡献数据
func GetGuildGongxian(ctx context.Context, gid int) types.Map {
	v, _ := repo.RedisHGetAll(ctx, fmt.Sprintf(config.KeyGuildGongxian, gid))
	return v
}

func IncrGuildGongxian(ctx context.Context, gid int, uid int64, num int) {
	repo.RedisHIncrBy(ctx, fmt.Sprintf(config.KeyGuildGongxian, gid), uid, int64(num))
}

// HdelGuildGongxian 删除公会成员贡献数据
func HdelGuildGongxian(ctx context.Context, gid int, uid int64) {
	repo.RedisHDel(ctx, fmt.Sprintf(config.KeyGuildGongxian, gid), uid)
}

// QuitWaitTime 设置退出公会等待时间
func QuitWaitTime(ctx context.Context, uid int64) {
	fk := fmt.Sprintf(config.KeyFirstQuit, uid)
	v, _ := repo.RedisGet(ctx, fk)
	if v == "" {
		repo.RedisSet(ctx, fk, "1", 86400*30)
	} else {
		repo.RedisSet(ctx, fmt.Sprintf(config.KeyQuitWaitTime, uid), time.Now().Unix()+12*3600, 12*3600)
	}
}

// GetQuitWaitTime 获取退出公会等待时间
func GetQuitWaitTime(ctx context.Context, uid int64) int {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyQuitWaitTime, uid))
	et := types.ToIntE(v)
	if et == 0 {
		return 0
	}
	left := et - int(time.Now().Unix())
	if left < 0 {
		return 0
	}
	return left
}

// GetGuildContribute 获取公会捐献状态
func GetGuildContribute(ctx context.Context, uid int64, id int) int {
	v, _ := daily.GetByPrefix(uid, config.DailyGuildContribute, id)
	return types.ToIntE(v)
}

// SetGuildContribute 设置公会捐献状态
func SetGuildContribute(ctx context.Context, uid int64, id int) {
	daily.SetByPrefix(uid, config.DailyGuildContribute, id, "1")
}

// GetActiveLingquStatus 获取捐献活跃领取状态
func GetActiveLingquStatus(ctx context.Context, uid int64) types.Map {
	return daily.GetAllByPrefix(uid, config.DailyActiveLingquStatus)
}

// SetGuildContributeActiveLingqu 设置活跃领取状态
func SetGuildContributeActiveLingqu(ctx context.Context, uid int64, id int) {
	var key int
	switch id {
	case 3:
		key = 3000
	case 2:
		key = 2000
	default:
		key = 1000
	}
	daily.SetByPrefix(uid, config.DailyActiveLingquStatus, key, "1")
}

func GetUsersGuildApply(ctx context.Context, gid int) []repo.ScoreMember {
	v, _ := repo.RedisZRevRangeWithScores(ctx, fmt.Sprintf(config.KeyGuildApplyUid, gid), 0, 999)
	return v
}

func AddUsersApplyGuild(ctx context.Context, gid int, uid int64) {
	k := fmt.Sprintf(config.KeyGuildApplyUid, gid)
	repo.RedisZAdd(ctx, k, float64(time.Now().Unix()), uid)
	repo.RedisExpire(ctx, k, 6*3600)
}

// GetUserGuildApplyByUID 获取用户公会申请记录
func GetUserGuildApplyByUID(ctx context.Context, gid int, uid int64) float64 {
	v, _ := repo.RedisZScore(ctx, fmt.Sprintf(config.KeyGuildApplyUid, gid), uid)
	return v
}

func DelUsersApplyGuild(ctx context.Context, gid int, uid int64) {
	repo.RedisZRem(ctx, fmt.Sprintf(config.KeyGuildApplyUid, gid), uid)
}

func SetGuildOwnerLoginTime(ctx context.Context, gid int) {
	k := fmt.Sprintf(config.KeyGuildOwnerLoginTime, gid)
	repo.RedisSet(ctx, k, time.Now().Unix(), 7*86400)
}

func GetGuildOwnerLoginTime(ctx context.Context, gid int) int {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyGuildOwnerLoginTime, gid))
	return types.ToIntE(v)
}
