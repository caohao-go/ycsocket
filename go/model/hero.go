package model

import (
	"context"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/repo/mem/userhero"
	"server_golang/repo/table"
)

// ---- 英雄系统 ----

// GetUserHeroList 获取用户英雄列表（返回 struct，用于只读调用方）
func GetUserHeroList(ctx context.Context, userID int64) []*table.UserHero {
	return userhero.GetUserHeroList(userID)
}

// GetUserHeroByID 根据ID获取英雄（返回 struct，用于只读调用方）
func GetUserHeroByID(ctx context.Context, id int) *table.UserHero {
	// 注意：此函数无 userID 参数，遍历内存查找
	return userhero.GetUserHeroById(id)
}

// GetUserHeroByIDs 根据多个ID获取英雄
func GetUserHeroByIDs(ctx context.Context, ids []int) []*table.UserHero {
	return userhero.GetUserHeroByIds(ids)
}

// GetUserCountHero 获取用户英雄数量
func GetUserCountHero(ctx context.Context, userID int64) int {
	return userhero.GetUserCountHero(userID)
}

// InsertNewUserHero 插入新英雄
func InsertNewUserHero(ctx context.Context, userID int64, heroID, star int) (int, error) {
	data := table.UserHero{
		UserId: userID,
		HeroId: heroID,
		Star:   star,
		Stage:  0,
		Lv:     1,
	}

	id, err := userhero.InsertUserHero(userID, &data)
	if err != nil {
		return 0, err
	}

	data.Id = id

	// 刷新用户英雄属性
	GetHeroAttrCore(ctx, &data, userID, 0, 0, 0)
	return id, nil
}

// DeleteUserHeroByIDs 删除英雄
func DeleteUserHeroByIDs(ctx context.Context, userID int64, ids []int) error {
	return userhero.DeleteUserHeroByIds(userID, ids)
}

// UpdateUserHeroFit 更新英雄装备
func UpdateUserHeroFit(ctx context.Context, userID int64, id int, fit interface{}) error {
	return userhero.UpdateUserHeroById(userID, id, types.Map{"fit": json.Marshal(fit)})
}

// UpdateUserHeroFu 更新英雄符文
func UpdateUserHeroFu(ctx context.Context, userID int64, id int, fu interface{}) error {
	return userhero.UpdateUserHeroById(userID, id, types.Map{"fu": json.Marshal(fu)})
}

// UpdateUserHeroID 更新英雄ID（换英雄）
func UpdateUserHeroID(ctx context.Context, userID int64, id int, newHeroID int) error {
	return userhero.UpdateUserHeroById(userID, id, types.Map{"hero_id": newHeroID})
}

// UpdateUserHero 更新英雄等级/阶段/星级
func UpdateUserHero(ctx context.Context, userID int64, id, lv, stage, star int) error {
	if lv == 0 && stage == 0 && star == 0 {
		return nil
	}
	// 增量更新，与 PHP 一致：lv=lv+N, star=star+N, stage=stage+N
	return userhero.IncrUserHeroLvStarStage(userID, id, lv, star, stage)
}
