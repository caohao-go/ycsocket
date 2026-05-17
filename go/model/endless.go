package model

import (
	"context"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo/mem/daily"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

// EndlessFightData 无尽试炼战斗数据
type EndlessFightData struct {
	Heros    map[int]int `json:"heros"`
	Position int         `json:"position"`
	Addition int         `json:"addition"`
	P1HP     map[int]int `json:"p1_hp"`
	Success  int         `json:"success"`
}

// GetEndlessInfo 获取无尽试炼信息
func GetEndlessInfo(ctx context.Context, userID int64, endlessLayer int) types.Map {
	todayLayerInfo := daily.GetAllByPrefix(userID, config.DailyEndlessToday)
	if len(todayLayerInfo) == 0 {
		startLayer := endlessLayer - 10
		if endlessLayer < 10 {
			startLayer = 1
		}

		todayLayerInfo = types.Map{
			"start_layer":       startLayer, //今天从哪层开始
			"already_layer":     startLayer, //今天最高打到了多少层（reset之后最高层不变）
			"current_layer":     startLayer, //当前打到了多少层（reset之后重新计算）
			"first_fight":       "1",
			"reset":             "0",
			"today_cross_layer": "0",
		}

		daily.HmSetByPrefix(userID, config.DailyEndlessToday, todayLayerInfo)
	}

	startLayer := todayLayerInfo.GetIntE("start_layer")
	alreadyLayer := todayLayerInfo.GetIntE("already_layer")

	thistimeAlreadyLayer := GetEndlessThistimeAlreadyLayer(ctx, userID)
	ret := types.Map{
		"endless_layer":           endlessLayer,
		"reset":                   todayLayerInfo.GetIntE("reset"),
		"first_fight":             todayLayerInfo.GetIntE("first_fight"),
		"start_layer":             startLayer,
		"current_layer":           todayLayerInfo.GetIntE("current_layer"),
		"already_layer":           alreadyLayer,
		"today_cross_layer":       todayLayerInfo.GetIntE("today_cross_layer"),
		"rank_reward":             []util.TypeNum{{Type: 20301, Num: 500}, {Type: 7, Num: 500000}},
		"cross_reward":            logic.GetLeijiCrossLayerReward(alreadyLayer, startLayer),
		"next_first_layer":        logic.GetNextFirstLayer(endlessLayer),
		"next_first_layer_reward": logic.GetFirstCrossLayerReward(logic.GetNextFirstLayer(endlessLayer)),
		"thistime_already_layer":  thistimeAlreadyLayer,
	}
	return ret
}

// 获取无尽模式助战英雄
func GetUserEndlessHelpHero(ctx context.Context, userIDs []int64) []*table.UserEndlessHelpHero {
	return world.GetUserEndlessHelpHeroByUserIdsWithCache(ctx, userIDs)
}

// 替换无尽模式助战英雄
func ReplaceUserEndlessHelpHero(ctx context.Context, data *table.UserEndlessHelpHero) error {
	return world.ReplaceUserEndlessHelpHero(ctx, data)
}

// ---- 通关奖励 ----

// GetUserTongguanReward 获取通关奖励
func GetUserTongguanReward(ctx context.Context, userID int64, rewardType int) []*table.UserTongguanReward {
	return world.GetUserTongguanRewardList(ctx, userID, rewardType)
}

// ReplaceUserTongguanReward 替换通关奖励
func ReplaceUserTongguanReward(ctx context.Context, userID int64, rewardType, copyID, status int) error {
	return world.ReplaceUserTongguanReward(ctx, &table.UserTongguanReward{
		UserId: userID, Type: rewardType, Copy: copyID, Status: status,
	})
}

// DeleteUserTongguanReward 删除通关奖励
func DeleteUserTongguanReward(ctx context.Context, userID int64, rewardType, copyID int) error {
	return world.DeleteUserTongguanReward(ctx, userID, rewardType, copyID)
}

func HsetEndlessTodayLayer(ctx context.Context, uid int64, typ string, layer int) {
	daily.SetByPrefix(uid, config.DailyEndlessToday, typ, layer)
}

func GetEndlessThistimeAlreadyLayer(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyEndlessThistime)
	return types.ToIntE(v)
}

func SetEndlessThistimeAlreadyLayer(ctx context.Context, uid int64, layer int) {
	daily.Set(uid, config.DailyEndlessThistime, layer)
}

func GetEndlessHelpChoose(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyEndlessHelpChoose)
	return types.ToIntE(v)
}

func SetEndlessHelpChoose(ctx context.Context, uid int64, heroID string) {
	daily.Set(uid, config.DailyEndlessHelpChoose, heroID)
}

func GetEndlessFightHeros(ctx context.Context, uid int64) *EndlessFightData {
	v, _ := daily.Get(uid, config.DailyEndlessFightHero)
	if v == nil {
		return nil
	}
	s := types.ToString(v)
	if s == "" {
		return nil
	}

	var ret EndlessFightData
	json.Unmarshal(s, &ret)
	return &ret
}

func ResetEndlessFightHeros(ctx context.Context, uid int64) {
	daily.Del(uid, config.DailyEndlessFightHero)
}

func SetEndlessFightHeros(ctx context.Context, uid int64, heros, p1HP map[int]int, position, addition, success int) {
	daily.Set(uid, config.DailyEndlessFightHero, json.Marshal(EndlessFightData{heros, position, addition, p1HP, success}))
}
