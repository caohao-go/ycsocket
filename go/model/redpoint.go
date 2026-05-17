// 红点数据模块
package model

import (
	"context"

	"server_golang/common/json"
	"server_golang/repo/mem/daily"

	"server_golang/common/types"
	"server_golang/config"
	"server_golang/connector"
)

// RedpointSend 发送红点提醒
// id - 红点编号 1-成就任务 2-邮件 3-试练塔未点击 4-排行榜
// pk - 是否战斗后再加 0-否 1-是
// type - 类型 1-默认
// num - 数量 0-不显示红点 1-显示红点
func RedpointSend(ctx context.Context, uid int64, id int, pk int, rpType int, num int) {
	ret := types.Map{
		"c":    "redpoint",
		"uid":  uid,
		"type": rpType,
		"id":   id,
		"pk":   pk,
		"num":  num,
	}
	connector.Manager.Send(uid, json.Marshal(ret))
}

// GetTodayPoint 获取今日红点状态
func GetTodayPoint(ctx context.Context, uid int64) []types.Map {
	data := daily.GetAllByPrefix(uid, config.DailyTodayPoint)

	num := 1
	if types.ToIntE(data["3"]) != 0 {
		num = 0
	}

	ret := []types.Map{
		{"type": 1, "id": 3, "pk": 0, "num": num},
	}
	return ret
}

// SetTodayPoint 设置今日红点状态
func SetTodayPoint(uid int64, id int, status int) {
	daily.SetByPrefix(uid, config.DailyTodayPoint, id, status)
}
