package model

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo/mem/daily"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

type ExpeditionLayerOpHero struct {
	Nickname  string                `json:"nickname"`
	AvatarUrl string                `json:"avatar_url"`
	Gender    int                   `json:"gender"`
	Lv        int                   `json:"lv"`
	Type      string                `json:"type"`
	Heros     []*logic.HeroBaseInfo `json:"heros"`
	HeroAttr  []*logic.Hero         `json:"hero_attr"`
}

type ExpeditionOpHero struct {
	Type   string `json:"type"`
	UserID int64  `json:"user_id"`
	Layer  int    `json:"layer"`
}

type ExpeditionInfo struct {
	Choose       int               `json:"choose"`
	CurrentLayer int               `json:"current_layer"`
	Baoxiang     []*BaoxiangInfo   `json:"baoxiang"`
	Copy         map[int]*CopyInfo `json:"copy"`
	Rewards      []util.TypeNum    `json:"current_rewards"`
}

type BaoxiangInfo struct {
	Pos    int `json:"pos"`
	Status int `json:"status"`
}
type CopyInfo struct {
	Status int `json:"status"`
}

// GetExpeditionInfo 获取远征信息
func GetExpeditionInfo(ctx context.Context, userID int64, userGrade types.Map) *ExpeditionInfo {
	open := 1
	if userGrade.GetIntE("lv") < 30 {
		open = 0
	}
	todayInfo := daily.GetAllByPrefix(userID, config.DailyExpeditionToday)
	choose := todayInfo.GetIntE("choose")
	currentLayer := todayInfo.GetIntE("current_layer")

	data := ExpeditionInfo{
		Choose:       choose,
		CurrentLayer: currentLayer,
	}

	// 宝箱信息
	if choose != 0 {
		data.Baoxiang = GetBaoxiangInfo(ctx, userID, choose, currentLayer)
	} else {
		data.Baoxiang = []*BaoxiangInfo{}
	}

	fightPoint := GetUserFightPoint(ctx, userID, 1)
	data.Copy = map[int]*CopyInfo{
		20201: {open},
		20202: {types.BoolToInt(open == 1 && fightPoint > 1000000)},
		20203: {types.BoolToInt(open == 1 && fightPoint > 1800000)},
	}

	// 当前奖励
	data.Rewards = getCurrentRewards(choose, data.Baoxiang, currentLayer)

	return &data
}

// GetExpeditionLayerOpHero 获取远征某层对手信息（对齐 PHP getExpeditionLayerOpHero）
// 根据图腾排名决定使用 NPC 还是真实玩家：排名 >100 用爬塔 NPC，<=100 用真实玩家
func GetExpeditionLayerOpHero(ctx context.Context,
	userID int64, expeditionInfo *ExpeditionInfo, layer int) *ExpeditionLayerOpHero {
	myRank := GetMyTutengRank(ctx, userID)

	var opData *ExpeditionOpHero
	if myRank > 100 {
		opData = getExpeditionLayerOpHeroByClimbtower(ctx, userID, expeditionInfo, layer)
	} else {
		opData = getExpeditionLayerOpHero(ctx, userID, expeditionInfo, layer)
	}

	opUserID := opData.UserID

	// 获取对手基本信息
	userGrade := GetUserAttr(opUserID)
	op := ExpeditionLayerOpHero{
		Nickname:  userGrade.GetStringE("nickname"),
		AvatarUrl: userGrade.GetStringE("avatar_url"),
		Gender:    userGrade.GetIntE("gender"),
		Lv:        userGrade.GetIntE("lv"),
	}

	if opData.Type == "rank" {
		// 真实玩家 — 获取其阵位英雄属性
		posData := GetUserPositionWithHeroAttrs(ctx, opUserID, 2)
		if posData == nil || len(posData) == 0 {
			posData = GetUserPositionWithHeroAttrs(ctx, opUserID, 1)
		}
		op.HeroAttr = posData
		op.Type = "rank"
	} else {
		// NPC — 从爬塔配置取英雄
		opLayer := opData.Layer
		op.Heros = logic.GetExpeditionHeros(opLayer)
		op.Type = "climb"
	}

	return &op
}

// 纯 NPC 对手（对齐 PHP _get_expedition_layer_op_hero_by_Climbtower）
func getExpeditionLayerOpHeroByClimbtower(ctx context.Context,
	userID int64, expeditionInfo *ExpeditionInfo, layer int) *ExpeditionOpHero {
	// 跳过宝箱层
	adjLayer := adjustExpeditionLayer(layer)

	// 读缓存
	data := GetExpeditionOpHeros(ctx, userID)
	if data != nil && len(data) > adjLayer-1 {
		return data[adjLayer-1]
	}

	var climbStart int
	switch expeditionInfo.Choose {
	case 20201:
		climbStart = 1
	case 20202:
		climbStart = 16
	case 20203:
		climbStart = 45
	default:
		climbStart = 1
	}

	ret := make([]*ExpeditionOpHero, 0, 15)
	for i := climbStart; i < climbStart+15; i++ {
		ret = append(ret, &ExpeditionOpHero{
			Type:   "climb",
			UserID: int64(rand.Intn(100)) + 1,
			Layer:  i,
		})
	}
	SetExpeditionOpHeros(ctx, userID, ret)

	if adjLayer-1 < len(ret) {
		return ret[adjLayer-1]
	}
	return ret[0]
}

// 混合对手：根据图腾排名匹配真实玩家+NPC（对齐 PHP _get_expedition_layer_op_hero）
func getExpeditionLayerOpHero(ctx context.Context,
	userID int64, expeditionInfo *ExpeditionInfo, layer int) *ExpeditionOpHero {
	// 跳过宝箱层
	adjLayer := adjustExpeditionLayer(layer)

	// 读缓存
	data := GetExpeditionOpHeros(ctx, userID)
	if len(data) > adjLayer-1 {
		return data[adjLayer-1]
	}

	myRank := GetMyTutengRank(ctx, userID)
	maxRank := GetTutengMaxRank(ctx)

	pkRanksOpHeros := make([]int64, 0)

	if expeditionInfo.Choose == 20203 && myRank > 5 {
		// 地狱模式: 取当前玩家的后10条+前5条
		for i := myRank - 5; i < myRank; i++ {
			pkRanksOpHeros = append(pkRanksOpHeros, i)
		}
		for i := myRank + 1; i <= myRank+10; i++ {
			if i >= maxRank {
				break
			}
			pkRanksOpHeros = append(pkRanksOpHeros, i)
		}
	} else {
		var pkRanksStart int64
		if expeditionInfo.Choose == 20201 { // 普通模式
			if myRank <= 5 {
				pkRanksStart = 10
			} else {
				pkRanksStart = 15
			}
		} else { // 困难模式
			pkRanksStart = 1
		}

		for i := myRank + pkRanksStart; i < myRank+pkRanksStart+15; i++ {
			if i >= maxRank {
				break
			}
			pkRanksOpHeros = append(pkRanksOpHeros, i)
		}
	}

	// 反转
	for i, j := 0, len(pkRanksOpHeros)-1; i < j; i, j = i+1, j-1 {
		pkRanksOpHeros[i], pkRanksOpHeros[j] = pkRanksOpHeros[j], pkRanksOpHeros[i]
	}

	pkCount := len(pkRanksOpHeros)
	climbtowerHeroCount := 15 - pkCount

	var climbStart int
	switch expeditionInfo.Choose {
	case 20201:
		climbStart = rand.Intn(15) + 1
	case 20202:
		climbStart = rand.Intn(15) + 16
	case 20203:
		climbStart = rand.Intn(15) + 31
	default:
		climbStart = 1
	}

	ret := make([]*ExpeditionOpHero, 0, 15)
	for i := climbStart; i < climbStart+climbtowerHeroCount; i++ {
		ret = append(ret, &ExpeditionOpHero{
			Type:   "climb",
			UserID: int64(rand.Intn(100)) + 1,
			Layer:  i,
		})
	}

	for _, rank := range pkRanksOpHeros {
		rankInfo := GetRankUserid(ctx, rank)
		if rankInfo != nil {
			ret = append(ret, &ExpeditionOpHero{
				Type:   "rank",
				UserID: rankInfo.GetInt64E("user_id"),
			})
		}
	}

	SetExpeditionOpHeros(ctx, userID, ret)
	if adjLayer-1 < len(ret) {
		return ret[adjLayer-1]
	}
	if len(ret) > 0 {
		return ret[0]
	}

	return &ExpeditionOpHero{Type: "climb", UserID: 1, Layer: 1}
}

// GetBaoxiangInfo 获取宝箱信息
func GetBaoxiangInfo(ctx context.Context, userID int64, chooseCopyID, currentLayer int) []*BaoxiangInfo {
	baoxiangOpen := daily.GetAllByPrefix(userID, config.DailyExpeditionBaoxiangOpen)
	positions := logic.BaoxiangPos[chooseCopyID]
	var ret = []*BaoxiangInfo{}
	for _, pos := range positions {
		if currentLayer < pos {
			ret = append(ret, &BaoxiangInfo{Pos: pos, Status: 0})
		} else if i := baoxiangOpen.GetIntE(pos); i > 0 {
			ret = append(ret, &BaoxiangInfo{Pos: pos, Status: 2})
		} else {
			ret = append(ret, &BaoxiangInfo{Pos: pos, Status: 1})
		}
	}
	return ret
}

// 获取远征助战英雄
func GetUserExpeditionHelpHero(ctx context.Context, userIDs []int64) []*table.UserExpeditionHelpHero {
	return world.GetUserExpeditionHelpHeroByUserIdsWithCache(ctx, userIDs)
}

// 替换远征助战英雄
func ReplaceUserExpeditionHelpHero(ctx context.Context, data *table.UserExpeditionHelpHero) error {
	return world.ReplaceUserExpeditionHelpHero(ctx, data)
}

// ========================= pika =========================

// IncrExpeditionScoreRank 远征积分排名递增
// 对应 PHP ShinelightModel::incrExpeditionScoreRank
// 根据 copyID 映射积分（20201→1, 20202→2, 其他→3），按星期分排行，TTL 8 天
func IncrExpeditionScoreRank(ctx context.Context, userID int64, copyID int) {
	w := time.Now().Weekday() // 0=Sunday, 与 PHP date('w') 一致

	var addScore float64
	switch copyID {
	case 20201:
		addScore = 1
	case 20202:
		addScore = 2
	default:
		addScore = 3
	}

	rankKey := fmt.Sprintf("%s_%d", config.RankExpeditionScore, w)
	IncrRankScore(ctx, rankKey, userID, addScore, 86400*8)
}

func HsetExpeditionTodayInfo(ctx context.Context, uid int64, typ string, cnt int) {
	daily.SetByPrefix(uid, config.DailyExpeditionToday, typ, cnt)
}

func GetExpeditionHelpChoose(ctx context.Context, uid int64) types.Map {
	return daily.GetAllByPrefix(uid, config.DailyExpeditionHelpChooseAll)
}

func SetExpeditionHelpChoose(ctx context.Context, uid int64, heroID string) {
	daily.SetByPrefix(uid, config.DailyExpeditionHelpChooseAll, heroID, "1")
}

func GetExpeditionHerosHP(ctx context.Context, uid int64, heroIDs []int) map[int]int {
	tmp := daily.HmGetByPrefix(uid, config.DailyExpeditionHerosHp, heroIDs)

	ret := make(map[int]int)
	for k, v := range tmp {
		ret[types.ToIntE(k)] = types.ToIntE(v)
	}

	return ret
}

func GetAllExpeditionHerosHP(ctx context.Context, uid int64) map[int]int {
	tmp := daily.GetAllByPrefix(uid, config.DailyExpeditionHerosHp)

	ret := make(map[int]int)
	for k, v := range tmp {
		ret[types.ToIntE(k)] = types.ToIntE(v)
	}

	return ret
}

func SetExpeditionHerosHP(ctx context.Context, uid int64, hp map[int]int) {
	ret := types.Map{}
	for k, v := range hp {
		ret[types.ToString(k)] = v
	}
	daily.HmSetByPrefix(uid, config.DailyExpeditionHerosHp, ret)
}

func GetExpeditionOpHerosHP(ctx context.Context, uid int64, heroIDs []int) types.Map {
	return daily.HmGetByPrefix(uid, config.DailyExpeditionOpHerosHp, heroIDs)
}

func SetExpeditionOpHerosHP(ctx context.Context, uid int64, hp types.Map) {
	daily.HmSetByPrefix(uid, config.DailyExpeditionOpHerosHp, hp)
}

func DelExpeditionOpHerosHP(ctx context.Context, uid int64) {
	daily.DelAllByPrefix(uid, config.DailyExpeditionOpHerosHp)
}

func SetExpeditionBaoxiangOpen(ctx context.Context, uid int64, pos int) {
	daily.SetByPrefix(uid, config.DailyExpeditionBaoxiangOpen, pos, "1")
}

func GetExpeditionOpHeros(ctx context.Context, uid int64) []*ExpeditionOpHero {
	v, _ := daily.Get(uid, config.DailyExpeditionOpHeros)
	if v == nil {
		return nil
	}

	s := types.ToString(v)
	if s == "" {
		return nil
	}

	var ret []*ExpeditionOpHero
	json.Unmarshal(s, &ret)
	return ret
}

func SetExpeditionOpHeros(ctx context.Context, uid int64, data []*ExpeditionOpHero) {
	daily.Set(uid, config.DailyExpeditionOpHeros, json.Marshal(data))
}

// adjustExpeditionLayer 调整远征层数（跳过宝箱层）— 对齐 PHP 层数调整逻辑
func adjustExpeditionLayer(layer int) int {
	if layer > 16 {
		layer -= 4
	} else if layer > 12 {
		layer -= 3
	} else if layer > 8 {
		layer -= 2
	} else if layer > 4 {
		layer -= 1
	}
	return layer
}

// 获取当前奖励
func getCurrentRewards(copyID int, baoxiang []*BaoxiangInfo, currentLayer int) []util.TypeNum {
	currentReward := []util.TypeNum{
		{Type: 1, Num: 0},
		{Type: 12, Num: 0},
	}
	if copyID == 0 {
		return currentReward
	}

	baoxiangMap := make(map[int]int)
	for _, v := range baoxiang {
		baoxiangMap[v.Pos] = v.Status
	}

	for i := 1; i < currentLayer; i++ {
		if status, ok := baoxiangMap[i]; ok && status != 2 {
			continue
		}
		rewards := logic.ExpeditionRewards[copyID][i]
		for _, r := range rewards {
			if r.Type == 1 {
				currentReward[0].Num += r.Num
			} else if r.Type == 12 {
				currentReward[1].Num += r.Num
			}
		}
	}

	return currentReward
}
