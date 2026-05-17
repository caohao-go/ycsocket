package model

import (
	"context"
	"fmt"

	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo"
	"server_golang/repo/mem/heroattr"
	"server_golang/repo/mem/userhero"
)

// GetUserFightPoint 获取用户英雄总战斗力
func GetUserFightPoint(ctx context.Context, uid int64, posType int) int {
	userPosition := GetUserPositionByID(ctx, uid, posType)
	if userPosition == nil {
		return 0
	}

	heros := []*logic.Hero{}
	for heroID := range userPosition.HeroPos {
		hero, _ := heroattr.Get(uid, heroID)
		if hero != nil && hero.FightPoint > 0 {
			heros = append(heros, hero)
		} else {
			userHero := userhero.GetUserHeroById(heroID)
			if userHero != nil {
				heroCalc := GetHeroAttrCore(ctx, userHero, uid, 0, 0, 0)
				if heroCalc != nil && heroCalc.FightPoint > 0 {
					heros = append(heros, heroCalc)
				}
			}
		}
	}

	if len(heros) == 0 {
		return 0
	}

	logic.CombinationAttrAdd(heros)

	total := 0
	for _, attr := range heros {
		total += attr.FightPoint
	}

	return total
}

// SaveFightpointHero 更新英雄属性与战力
func SaveFightpointHero(userID int64, heroAttr *logic.Hero) {
	ctx := context.Background()

	id := heroAttr.Id
	if id < 100000000 { // 用户英雄 ID 从 1亿开始，战力只存用户英雄
		return
	}

	heroattr.Set(userID, id, heroAttr)

	userPosition := GetUserPositionByID(ctx, userID, 1)
	if userPosition != nil { //英雄在主线剧情阵位中，需要更新战斗力排行
		if _, ok := userPosition.HeroPos[id]; ok {
			SetFightPointRank(ctx, userID)
		}
	}

	// 单个英雄战斗力排行（与PHP一致）
	repo.RedisZAdd(ctx, config.RankHeroFightPoint, float64(heroAttr.FightPoint), id)
	repo.RedisZAdd(ctx, fmt.Sprintf(config.RankHeroFightPointProp, heroAttr.Property), float64(heroAttr.FightPoint), id)
}

// GetTopFightpointHero 获取战斗力最高的英雄
func GetTopFightpointHero(ctx context.Context) []repo.ScoreMember {
	scoreRanks, _ := repo.RedisZRevRangeWithScores(ctx, config.RankHeroFightPoint, 0, 1)
	return scoreRanks
}

func GetFightpointRank(ctx context.Context, prop int) []repo.ScoreMember {
	var rankKey string
	if prop == 0 {
		rankKey = config.RankHeroFightPoint
	} else {
		rankKey = fmt.Sprintf(config.RankHeroFightPointProp, prop)
	}

	// 从 pika ZSet 获取排名数据（对齐 PHP: zrevrange key 0 120 WITHSCORES）
	scoreRanks, _ := repo.RedisZRevRangeWithScores(ctx, rankKey, 0, 120)
	return scoreRanks
}

// SetFightPointRank 更新主阵位战力总排行
func SetFightPointRank(ctx context.Context, userID int64) {
	totalPoint := GetUserFightPoint(ctx, userID, 1)
	SetRankScore(ctx, config.RankFightPoint, userID, float64(totalPoint), 0)
}
