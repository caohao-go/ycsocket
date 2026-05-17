package model

import (
	"context"
	"time"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/types"
	"server_golang/repo/mem/content"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

// InitContents 初始化用户 content 数据（对应 PHP init_contents）
func InitContents(ctx context.Context, userID int64) {
	// 初始化公会技能（4个属性初始化为0）
	content.SetMap(userID, "guild_skills", types.Map{})
	// 初始化副本奖励统计
	content.SetArray(userID, "copy_rewards_stat", []interface{}{})
	// 初始化任务完成计数
	content.SetMap(userID, "task_finish_count", types.Map{})
	// 初始化任务领取
	content.SetMap(userID, "task_lingqu", types.Map{})
	// 初始化功能战斗
	content.SetMap(userID, "function_fights", types.Map{})
	// 初始化邀请玩家
	content.SetMap(userID, "invite_player", types.Map{"1": 0, "3": 0, "5": 0})

	// 初始化英雄（2301号英雄，5星）
	heroID, err := InsertNewUserHero(ctx, userID, 2301, 5)
	if err != nil || heroID <= 0 {
		log.Errorf(context.Background(), 0, "初始化英雄失败: userID=%d, err=%v", userID, err)
		return
	}

	time.Sleep(5 * time.Millisecond) // 首次等待战力异步更新完成

	// 插入主线剧情阵位
	world.ReplaceUserPosition(ctx, &table.UserPosition{
		UserId:   userID,
		PosType:  1,
		Position: 101,
		Pos1Pos:  2,
		Pos1Hero: heroID,
	})

	// 初始化 boss 阵型（初始英雄放在位置2，与 PHP [[$id, 2]] 一致）
	SetFightHeros(ctx, userID, "copy_fight", map[int]int{heroID: 2}, 101)
}

// InitContentsInt 初始化用户整数 content 数据
func InitContentsInt(ctx context.Context, userID int64) {
	content.SetInt(userID, "cost_money", 0)
	content.SetInt(userID, "climbtower_layer", 0)
	content.SetInt(userID, "endless_layer", 0)
	content.SetInt(userID, "guild_active", 0)
}

// 将 controller 对 world 包的直接调用迁移到 model 层

// GetUsersContentInt 获取用户整数 content
func GetUsersContentInt(ctx context.Context, userID int64, contentType string) int {
	return content.GetInt(userID, contentType)
}

// IncrUsersContentInt 自增用户整数 content
func IncrUsersContentInt(ctx context.Context, userID int64, contentType string, num int) error {
	content.IncrInt(userID, contentType, num)
	return nil
}

// UpdateUsersContentInt 更新用户整数 content
func UpdateUsersContentInt(ctx context.Context, userID int64, contentType string, num int) error {
	content.SetInt(userID, contentType, num)
	return nil
}

// GetUsersContent 获取用户 JSON content（Map 类型）
func GetUsersContent(ctx context.Context, userID int64, contentType string) types.Map {
	return content.GetMap(userID, contentType)
}

// UpdateUsersContent 更新用户 JSON content（Map 类型）
func UpdateUsersContent(ctx context.Context, userID int64, contentType string, c types.Map) error {
	content.SetMap(userID, contentType, c)
	return nil
}

// GetUsersContentArray 获取用户 JSON content（数组类型）
func GetUsersContentArray(ctx context.Context, userID int64, contentType string) []interface{} {
	return content.GetArray(userID, contentType)
}

// UpdateUsersContentArray 更新用户 JSON content（数组类型）
func UpdateUsersContentArray(ctx context.Context, userID int64, contentType string, c []interface{}) error {
	content.SetArray(userID, contentType, c)
	return nil
}
