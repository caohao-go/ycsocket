package model

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

// InitGuildFightChangci 初始化公会战场次
func InitGuildFightChangci(ctx context.Context) {
	// 初始化公会战场次
	changci := GetGuildFightChangci(ctx)
	logic.GuildFightChangci = changci
}

// AddZhanji 增加战绩
func AddZhanji(ctx context.Context, guildID int, userID int64, zhanji int) {
	changci := logic.GetGuildFightChangci()
	projectName := fmt.Sprintf("%s_%d_%d", config.RankGuildFight, guildID, changci)
	IncrRankScore(ctx, projectName, userID, float64(zhanji), 259200)
}

// GetGuildFightInfo 获取公会战信息
// 翻译自 PHP ShinelightModel::get_guild_fight_info
func GetGuildFightInfo(ctx context.Context, userID int64, guildID int) types.Map {
	ret := types.Map{}

	w := time.Now().Weekday()
	h := time.Now().Hour()
	status := GetGuildFightStatus(ctx)
	ret["status"] = status

	// 周一三五晚上21点后战斗结束
	if status == 1 && h >= 21 {
		SetGuildFightStatus(ctx, 0)
		ret["status"] = 0
	}

	// 获取我方战斗信息
	changci := logic.GetGuildFightChangci()
	myGuildFight := GetRedisGuildFightInfo(ctx, guildID, changci)

	if len(myGuildFight) > 0 {
		// 计算 left_time（周一三五 21:00 剩余时间）
		leftTime := int64(0)
		if w == time.Monday || w == time.Wednesday || w == time.Friday {
			targetTime := time.Date(time.Now().Year(), time.Now().Month(), time.Now().Day(), 21, 0, 0, 0, time.Local)
			leftTime = targetTime.Unix() - time.Now().Unix()
		}
		ret["left_time"] = leftTime
		if leftTime <= 0 {
			ret["status"] = 3
		} else {
			ret["status"] = status
		}
		ret["my_pos"] = -1
		ret["my_fight_count"] = 0

		totalStar := 0
		users, _ := myGuildFight["users"].([]types.Map)
		if users == nil {
			// 尝试从 map 转换
			if usersMap, ok := myGuildFight["users"].(types.Map); ok {
				for _, v := range usersMap {
					if um, ok := v.(types.Map); ok {
						users = append(users, um)
					}
				}
			}
		}
		for _, fightUser := range users {
			if fightUser.GetInt64E("user_id") == userID {
				// 找到当前用户的 position（在 map 中的 key）
				for k, fu := range users {
					if fu.GetInt64E("user_id") == userID {
						ret["my_pos"] = types.ToIntE(k) // map 的 key 作为位置
						break
					}
				}
				ret["my_fight_count"] = fightUser.GetIntE("fight_count")
			}
			totalStar += fightUser.GetIntE("star")
		}
		myGuildFight["total_star"] = totalStar
		myGuildFight["incre_lv"] = 0

		// 获取对手战斗信息
		opGuildID := myGuildFight.GetIntE("op_guild_id")
		opGuildFight := GetRedisGuildFightInfo(ctx, opGuildID, changci)
		opTotalStar := 0
		if len(opGuildFight) > 0 {
			opUsers, _ := opGuildFight["users"].([]types.Map)
			if opUsers == nil {
				if opUsersMap, ok := opGuildFight["users"].(types.Map); ok {
					for _, v := range opUsersMap {
						if oum, ok := v.(types.Map); ok {
							opUsers = append(opUsers, oum)
						}
					}
				}
			}
			for _, ofu := range opUsers {
				opTotalStar += ofu.GetIntE("star")
			}
		}
		opGuildFight["total_star"] = opTotalStar
		opGuildFight["incre_lv"] = 0

		ret["my_guild"] = myGuildFight
		ret["op_guild"] = opGuildFight
	} else {
		ret["status"] = 0
	}

	return ret
}

// AssignGuildFight 分配公会战（核心逻辑）
// 翻译自 PHP ShinelightModel::assign_guild_fight (约170行)
func AssignGuildFight(ctx context.Context) error {
	// 1. 找出大于5个人且公会等级>=3的所有公会
	guilds, err := world.GetGuildsForFight(ctx)
	if err != nil || len(guilds) == 0 {
		return err
	}

	type guildOp struct {
		GuildID int
		Guild   types.Map
		Users   []types.Map // []map[user_k]user_data
	}
	type userInfo struct {
		UserID     int64
		ZoneID     int
		Nickname   string
		AvatarURL  string
		Lv         int
		FightPoint int
		Position   int
		Star       int
		BeStar     int
		SucDef     int
		FightCount int
		BeFightCnt int
		Zhanji     int
		Heros      []types.Map
	}

	ops := make(map[int]*guildOp)
	var guildIDs []int

	for _, guild := range guilds {
		gID := guild.Id
		gCopy := types.ObjectToMap(guild)
		delete(gCopy, "creator")
		delete(gCopy, "exp")

		ops[gID] = &guildOp{GuildID: gID, Guild: gCopy}
		guildIDs = append(guildIDs, gID)
	}

	// 2. 找出公会用户
	guildsUsersRaw, err := world.GetGuildsUsersByGuildIds(ctx, guildIDs)
	if err != nil {
		return err
	}

	// PHP: $guilds_users[$v['guild_id']][] = ['user_id' => $v['user_id']];
	guildsUsers := make(map[int][]userInfo)
	var userIDs []int64
	for _, gu := range guildsUsersRaw {
		guildsUsers[gu.GuildId] = append(guildsUsers[gu.GuildId], userInfo{UserID: gu.UserId})
		userIDs = append(userIDs, gu.UserId)
	}

	// 3. 获取参战选手详细信息（等级、战斗力）
	userinfos := GetUsersWithDetail(ctx, userIDs, 1, config.AttrLv, config.AttrFightPoint)

	changci := logic.GetGuildFightChangci()
	rankTypePrefix := fmt.Sprintf("%s_%d", config.RankGuildFight, changci)

	// 4. 处理每个公会的用户
	for k, v := range guildsUsers {
		guildTotalFP := 0
		filtered := make([]userInfo, 0)

		for userK, userV := range v {
			uid := userV.UserID

			// 初始化战绩排名
			SetRankScore(ctx, fmt.Sprintf("%s_%d", rankTypePrefix, k), uid, 0, 259200)

			// 用户详情
			tmp := userInfo{UserID: uid}
			if ui, ok := userinfos[uid]; ok {
				tmp.Lv = ui.GetIntE("lv")
				tmp.FightPoint = ui.GetIntE("fight_point")
				tmp.ZoneID = GetUserZoneID(uid)
				tmp.Nickname = ui.GetStringE("nickname")
				tmp.AvatarURL = ui.GetStringE("avatar_url")
			} else {
				tmp.Nickname = ""
			}

			// 取出剧情英雄阵型
			userPosHeros := GetUserPositionByID(ctx, uid, 1)
			if userPosHeros == nil {
				continue // 无阵型的用户不参战
			}
			tmp.Position = userPosHeros.Position
			tmp.Star = 0
			tmp.BeStar = 0
			tmp.SucDef = 0
			tmp.FightCount = 0
			tmp.BeFightCnt = 0

			// 防守英雄 - 从阵型中取 hero_ids
			heroPos := userPosHeros.HeroPos
			if len(heroPos) > 0 {
				heroIDList := make([]int, 0, len(heroPos))
				for id := range heroPos {
					heroIDList = append(heroIDList, id)
				}
				userHeros := GetUserHeroByIDs(ctx, heroIDList)
				heroesMap := make(map[int]*table.UserHero)
				for _, h := range userHeros {
					heroesMap[h.Id] = h
				}

				for heroID, pos := range heroPos {
					hInfo := heroesMap[heroID]
					heroTmp := types.Map{
						"id":      heroID,
						"pos":     pos,
						"hero_id": hInfo.HeroId,
						"star":    hInfo.Star,
						"stage":   hInfo.Stage,
						"lv":      hInfo.Lv,
					}
					heroTmp["fit"] = types.ToMapE(hInfo.Fit)
					heroTmp["fu"] = types.ToMapE(hInfo.Fu)
					tmp.Heros = append(tmp.Heros, heroTmp)
				}
			}
			guildTotalFP += tmp.FightPoint
			v[userK] = tmp
			filtered = append(filtered, tmp)
		}

		guildsUsers[k] = filtered
		if op, ok := ops[k]; ok {
			// 将 []userInfo 转为 []types.Map
			slice := make([]types.Map, len(filtered))
			for i, u := range filtered {
				slice[i] = types.Map{
					"user_id":        u.UserID,
					"zone_id":        u.ZoneID,
					"nickname":       u.Nickname,
					"avatar_url":     u.AvatarURL,
					"lv":             u.Lv,
					"fight_point":    u.FightPoint,
					"position":       u.Position,
					"star":           u.Star,
					"be_star":        u.BeStar,
					"suc_def":        u.SucDef,
					"fight_count":    u.FightCount,
					"be_fight_count": u.BeFightCnt,
					"zhanji":         u.Zhanji,
					"heros":          u.Heros,
				}
			}
			op.Users = slice
			ops[k] = op
		}
	}

	// 5. 每个公会的用户按战斗力降序排序
	for _, v := range guildsUsers {
		sort.Slice(v, func(i, j int) bool {
			return v[i].FightPoint > v[j].FightPoint
		})
	}

	// 6. 清理位置日志 - 分配战绩值
	zhanjiDatas := logic.GetGuildRecordCalDatas()
	for k, v := range guildsUsers {
		pos := 0
		for uk := range v {
			if pos < len(zhanjiDatas) {
				v[uk].Zhanji = zhanjiDatas[pos]
			}
			pos++
		}
		guildsUsers[k] = v
	}

	// 7. 所有公会按总战力降序排序
	sortedOps := make([]*guildOp, 0, len(ops))
	for _, op := range ops {
		sortedOps = append(sortedOps, op)
	}
	sort.Slice(sortedOps, func(i, j int) bool {
		fpI := sortedOps[i].Guild.GetIntE("fight_point")
		fpJ := sortedOps[j].Guild.GetIntE("fight_point")
		return fpI > fpJ
	})

	// 8. 4个一组分组
	var tmpFenzu [][]*guildOp
	groupIdx := 0
	currentGroup := make([]*guildOp, 0)
	for i, v := range sortedOps {
		currentGroup = append(currentGroup, v)
		if i%4 == 3 { // 每4个一组
			tmpFenzu = append(tmpFenzu, currentGroup)
			currentGroup = make([]*guildOp, 0)
			groupIdx++
		}
	}
	if len(currentGroup) > 0 {
		tmpFenzu = append(tmpFenzu, currentGroup)
	}

	// 9. 随机分组对战
	rand.Seed(time.Now().UnixNano())
	ret := make(map[int][]*guildOp)
	for k, v := range tmpFenzu {
		if len(v) <= 1 {
			break
		} else if len(v) == 2 {
			ret[k] = []*guildOp{v[0], v[1]}
		} else if len(v) == 3 {
			idxs := rand.Perm(len(v))[:2]
			ret[k] = []*guildOp{v[idxs[0]], v[idxs[1]]}
		} else if len(v) == 4 {
			idxs := rand.Perm(len(v))[:2]
			ret[k] = []*guildOp{v[idxs[0]], v[idxs[1]]}
			k++
			remaining := make([]*guildOp, 0)
			for i, vv := range v {
				found := false
				for _, idx := range idxs {
					if i == idx {
						found = true
						break
					}
				}
				if !found {
					remaining = append(remaining, vv)
				}
			}
			ret[k] = remaining
		}
	}

	// 10. 将战斗信息设置到 pika
	duizhanInfo := types.Map{"duizhan": []types.Map{}, "members": []int64{}}
	for _, v := range ret {
		if len(v) < 2 {
			break
		}

		guildID0 := v[0].GuildID
		guildID1 := v[1].GuildID

		duizhanList, _ := duizhanInfo["duizhan"].([]types.Map)
		duizhanList = append(duizhanList, types.Map{"p1": guildID0, "p2": guildID1})
		duizhanInfo["duizhan"] = duizhanList

		members, _ := duizhanInfo["members"].([]int)
		members = append(members, guildID0, guildID1)
		duizhanInfo["members"] = members

		// 构建战斗信息并保存到 pika
		fightInfo0 := types.CopyMap(v[0].Guild)
		fightInfo0["op_guild_id"] = guildID1
		fightInfo0["users"] = guildsUsers[guildID0]
		// 与 PHP 一致：初始化 increate 属性加成
		fightInfo0["increate"] = types.Map{"atk": 0, "def": 0, "hp": 0, "speed": 0}
		SetGuildFightInfo(ctx, guildID0, changci, fightInfo0)

		fightInfo1 := types.CopyMap(v[1].Guild)
		fightInfo1["op_guild_id"] = guildID0
		fightInfo1["users"] = guildsUsers[guildID1]
		// 与 PHP 一致：初始化 increate 属性加成
		fightInfo1["increate"] = types.Map{"atk": 0, "def": 0, "hp": 0, "speed": 0}
		SetGuildFightInfo(ctx, guildID1, changci, fightInfo1)
	}

	SetDuizhanInfo(ctx, duizhanInfo)
	log.Infof(context.Background(), "公会战分配完成，共 %d 个公会参战", len(ops))
	return nil
}

func GetGuildFightChangci(ctx context.Context) int {
	v, _ := repo.RedisGet(ctx, "guild_fight_changci")
	return types.ToIntE(v)
}

func SetGuildFightChangci(ctx context.Context, changci int) {
	repo.RedisSet(ctx, "guild_fight_changci", changci, 0)
}

func GetGuildFightStatus(ctx context.Context) int {
	v, _ := repo.RedisGet(ctx, "guild_fight_status")
	return types.ToIntE(v)
}

func SetGuildFightStatus(ctx context.Context, status int) {
	repo.RedisSet(ctx, "guild_fight_status", status, 0)
}

func SetGuildFightInfo(ctx context.Context, gid int, changci int, data interface{}) {
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyGuildFightInfo, gid, changci), data, 259200)
}

func GetRedisGuildFightInfo(ctx context.Context, gid int, changci int) types.Map {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyGuildFightInfo, gid, changci))
	if v == "" {
		return types.Map{}
	}
	ret := types.ToMapE(v)
	return ret
}

func GetGuildStars(ctx context.Context, gid int, changci int) types.Map {
	v, _ := repo.RedisHGetAll(ctx, fmt.Sprintf(config.KeyGuildStars, gid, changci))
	return v
}

func HgetGuildStars(ctx context.Context, gid int, changci int, uid int64) int {
	v, _ := repo.RedisHGet(ctx, fmt.Sprintf(config.KeyGuildStars, gid, changci), uid)
	return types.ToIntE(v)
}

func SetGuildStars(ctx context.Context, gid int, changci int, uid int64, star int) {
	k := fmt.Sprintf(config.KeyGuildStars, gid, changci)
	repo.RedisHSet(ctx, k, uid, star)
	repo.RedisExpire(ctx, k, 259200)
}

func AddPosGuildLog(ctx context.Context, gid, pos, changci int, data interface{}) {
	k := fmt.Sprintf(config.KeyPosGuildLog, gid, pos, changci)
	repo.RedisHSet(ctx, k, time.Now().Unix(), data)
	repo.RedisExpire(ctx, k, 86400)
}

func GetPosGuildLog(ctx context.Context, gid int, pos, changci int) []interface{} {
	data, _ := repo.RedisHGetAll(ctx, fmt.Sprintf(config.KeyPosGuildLog, gid, pos, changci))
	if len(data) == 0 {
		return []interface{}{}
	}
	ret := make([]interface{}, 0, len(data))
	for _, v := range data {
		var it interface{}
		json.Unmarshal(v, &it)
		ret = append(ret, it)
	}
	return ret
}

func AddGuildFightLog(ctx context.Context, gid int, changci int, typ string, content types.Map) {
	k := fmt.Sprintf(config.KeyGuildFightLog, gid, changci)
	content["type"] = typ
	repo.RedisZAdd(ctx, k, float64(time.Now().Unix()), json.Marshal(content))
	repo.RedisExpire(ctx, k, 259200)
}

func GetGuildFightLog(ctx context.Context, gid int, changci int) []types.Map {
	ranks, _ := repo.RedisZRevRangeWithScores(ctx, fmt.Sprintf(config.KeyGuildFightLog, gid, changci), 0, 200)
	if len(ranks) == 0 {
		return []types.Map{}
	}
	ret := make([]types.Map, 0, len(ranks))
	for _, sm := range ranks {
		var f types.Map
		json.Unmarshal(sm.Member, &f)
		if f != nil {
			f["time"] = time.Now().Format("2006-01-02 15:04:05")
			ret = append(ret, f)
		}
	}
	return ret
}

func AddMyGuildFightLog(ctx context.Context, uid int64, changci int, typ string, content types.Map) {
	k := fmt.Sprintf(config.KeyMyGuildFightLog, uid, changci)
	content["type"] = typ
	repo.RedisZAdd(ctx, k, float64(time.Now().Unix()), json.Marshal(content))
	repo.RedisExpire(ctx, k, 259200)
}

func GetMyGuildFightLog(ctx context.Context, uid int64, changci int) []types.Map {
	ranks, _ := repo.RedisZRevRangeWithScores(ctx, fmt.Sprintf(config.KeyMyGuildFightLog, uid, changci), 0, 200)
	if len(ranks) == 0 {
		return []types.Map{}
	}
	ret := make([]types.Map, 0, len(ranks))
	for _, sm := range ranks {
		f := types.ToMapE(sm.Member)
		if f != nil {
			f["time"] = time.Now().Format("2006-01-02 15:04:05")
			ret = append(ret, f)
		}
	}
	return ret
}

func SetDuizhanInfo(ctx context.Context, data interface{}) {
	repo.RedisSet(ctx, "duizhan_info", data, 259200)
}

func GetDuizhanInfo(ctx context.Context) interface{} {
	v, _ := repo.RedisGet(ctx, "duizhan_info")
	if v == "" {
		return []interface{}{}
	}
	var ret interface{}
	json.Unmarshal(v, &ret)
	return ret
}
