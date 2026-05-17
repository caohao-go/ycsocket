// 区服信息模块
package logic

import (
	"context"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"

	"server_golang/common/types"
	"server_golang/repo/user"
)

// 区服全局数据
var (
	// 区服信息
	ZoneinfoData types.Map
	// 游戏版本号
	GameVersion string
	// 区服ID（从配置读取）
	ShineZoneID int
)

// InitZoneinfo 初始化区服信息（从配置文件加载）
func InitZoneinfo(ctx context.Context) {
	// 区服信息需要从配置文件加载，这里提供骨架
	// 实际项目中需要从trpc_go.yaml或zoneinfo配置读取
	ZoneinfoData = make(types.Map)
	ShineZoneID = 1 // 默认值，实际从配置读取

	// 初始化游戏版本
	initGameVersion(ctx)
}

// initGameVersion 从数据库获取游戏版本
func initGameVersion(ctx context.Context) {
	rows, err := user.GetAllGameVersion(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init game_version error: %v", err)
		return
	}
	if len(rows) > 0 {
		GameVersion = string(rows[0].Ver)
	}
}

// ZoneInfo 获取区服信息列表
func ZoneInfo() []types.Map {
	if zi, err := types.ToMapArray(ZoneinfoData["zone_info"], ""); err == nil {
		return zi
	}
	return nil
}

// RecommendZone 获取推荐区服
func RecommendZone() interface{} {
	return ZoneinfoData["recommend_zone"]
}

// SourceZone 获取源区服
func SourceZone() interface{} {
	return ZoneinfoData["source_zone"]
}

// Hequ 获取合区信息
func Hequ(zoneID int) interface{} {
	zi := ZoneInfo()
	if zoneID-1 >= 0 && zoneID-1 < len(zi) {
		return zi[zoneID-1]["hequ"]
	}
	return nil
}

// GetGameVersion 获取游戏版本
func GetGameVersion() string {
	return GameVersion
}

// GetShineZoneID 获取当前区服ID
func GetShineZoneID() int {
	return ShineZoneID
}

// SetZoneinfo 设置区服信息（供外部配置调用）
func SetZoneinfo(data types.Map) {
	ZoneinfoData = data
}
