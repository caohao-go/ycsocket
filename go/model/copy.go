package model

import (
	"context"
	"fmt"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo"
)

type FightHerosData struct {
	Heros    map[int]int `json:"heros"` // 英雄位置 pos => heroid
	Position int         `json:"position"`
}

// IncrRewardTimes 完成次数+1
func IncrRewardTimes(ctx context.Context, userID int64, copyID int) {
	datas := logic.GetRewardByCopyID(ctx, copyID)
	if datas == nil || datas[0] == nil {
		return
	}
	IncrRedisRewardTimes(ctx, userID, copyID, datas[0].Refresh)
}

// GetAllRewardStatus 获取所有奖励领取状态
func GetAllRewardStatus(ctx context.Context, userID int64, copyID int) map[int]int {
	hVal := GetRedisAllRewardStatus(ctx, userID, copyID)
	ret := make(map[int]int)
	datas := logic.GetRewardByCopyID(ctx, copyID)
	for key := range datas {
		ret[key] = types.ToIntE(hVal[types.ToString(key)])
	}
	return ret
}

// SetRewardStatus 设置奖励领取状态
func SetRewardStatus(ctx context.Context, userID int64, copyID, seq int) {
	datas := logic.GetRewardByCopyID(ctx, copyID)
	if datas == nil || datas[0] == nil {
		return
	}
	SetRedisRewardStatus(ctx, userID, copyID, seq, datas[0].Refresh)
}

// GetRewardFinish 获取完成状态
func GetRewardFinish(ctx context.Context, userID int64, copyID, seq int) int {
	times := GetRedisRewardTimes(ctx, userID, copyID)
	datas := logic.GetRewardByCopyID(ctx, copyID)
	data := datas[seq]
	if data == nil {
		return 0
	}
	if times < int(data.RankCount.Max) {
		return 0
	}
	return 1
}

// ========================= pika =========================

func GetFightHeros(ctx context.Context, uid int64, typ string) *FightHerosData {
	var fightHeros FightHerosData

	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyFightHero, typ, uid))
	if v != "" {
		json.Unmarshal(v, &fightHeros)
	}

	if len(fightHeros.Heros) == 0 && typ == "copy_fight" {
		posData := GetUserPositionByID(ctx, uid, 1)
		if posData != nil {
			position := posData.Position
			if position == 0 {
				position = 101
			}
			if len(posData.HeroPos) > 0 {
				SetFightHeros(ctx, uid, "copy_fight", posData.HeroPos, position)
				fightHeros.Heros = posData.HeroPos
				fightHeros.Position = position
			}
		}
	}

	return &fightHeros
}

func SetFightHeros(ctx context.Context, uid int64, typ string, heros map[int]int, position int) {
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyFightHero, typ, uid), json.Marshal(FightHerosData{heros, position}), 7*86400)
}

func GetRedisRewardTimes(ctx context.Context, uid int64, copyID int) int {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyRewardTimes, uid, copyID))
	return types.ToIntE(v)
}

func IncrRedisRewardTimes(ctx context.Context, uid int64, copyID, refresh int) {
	k := fmt.Sprintf(config.KeyRewardTimes, uid, copyID)
	repo.RedisIncr(ctx, k)
	switch {
	case refresh == 1:
		repo.RedisExpire(ctx, k, util.LeftTimeToTomorrow())
	case refresh == 2:
		repo.RedisExpire(ctx, k, util.LeftTimeToNextMonday())
	case refresh >= 3:
		repo.RedisExpire(ctx, k, refresh)
	}
}

func GetRedisRewardStatus(ctx context.Context, uid int64, copyID, seq int) int {
	v, _ := repo.RedisHGet(ctx, fmt.Sprintf(config.KeyRewardStatus, uid, copyID), seq)
	return types.ToIntE(v)
}

func GetRedisAllRewardStatus(ctx context.Context, uid int64, copyID int) types.Map {
	v, _ := repo.RedisHGetAll(ctx, fmt.Sprintf(config.KeyRewardStatus, uid, copyID))
	return v
}

func SetRedisRewardStatus(ctx context.Context, uid int64, copyID, seq, refresh int) {
	hk := fmt.Sprintf(config.KeyRewardStatus, uid, copyID)
	repo.RedisHSet(ctx, hk, seq, "1")
	switch {
	case refresh == 1:
		repo.RedisExpire(ctx, hk, util.LeftTimeToTomorrow())
	case refresh == 2:
		repo.RedisExpire(ctx, hk, util.LeftTimeToNextMonday())
	case refresh >= 3:
		repo.RedisExpire(ctx, hk, refresh)
	}
}
