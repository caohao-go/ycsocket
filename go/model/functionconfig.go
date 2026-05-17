package model

import (
	"context"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo/mem/daily"
)

type FunctionInfo struct {
	Counts   *FunctionCounts `json:"counts"`
	List     []*FunctionData `json:"list"`
	VipLevel int             `json:"vip_level"`
}
type FunctionCounts struct {
	Times        int `json:"times"`
	FreeCount    int `json:"free_count"`
	VipCount     int `json:"vip_count"`
	NextVipCount int `json:"next_vip_count"`
	LeftTimes    int `json:"left_times"`
}

type FunctionData struct {
	CopyID    int            `json:"copy_id"`
	Status    int            `json:"status"`
	Name      string         `json:"name"`
	CostType  int            `json:"cost_type"`
	Layer     int            `json:"layer"`
	BasicCost int            `json:"basic_cost"`
	Reward    []util.TypeNum `json:"reward"`
	OpenType  []util.TypeNum `json:"open_type"`
}

// GetFunctionByCopyID 根据copy_id获取副本信息
func GetFunctionByCopyID(ctx context.Context, userid int64,
	copyID int, userGrade types.Map, functionFights types.Map) *FunctionData {
	id := logic.CopyIDsID[copyID]
	result := GetFunctionByID(ctx, userid, id, userGrade, functionFights, copyID)
	if len(result.List) > 0 {
		return result.List[0]
	}
	return nil
}

// GetFunctionByID 根据功能ID获取日常副本列表
func GetFunctionByID(ctx context.Context, userid int64, id int,
	userGrade types.Map, functionFights types.Map, needCopyID int) *FunctionInfo {
	copyIDs := logic.FunctionIDs[id]
	counts := FunctionCounts{}

	v, _ := daily.GetByPrefix(userid, config.DailyFunctionTimes, id)
	times := types.ToIntE(v)

	counts.Times = times

	dataList := make([]*FunctionData, 0)

	for _, copyID := range copyIDs {
		if needCopyID != -1 && needCopyID != copyID {
			continue
		}

		fc, ok := logic.FunctionConfigs[copyID]
		if !ok {
			continue
		}

		tmp := FunctionData{}
		tmp.CopyID = copyID

		freeCount := fc.FreeCount
		vipLevel := types.ToIntE(userGrade["vip_level"])
		vipCount := logic.GetVipInfoLv(vipLevel, "copy_number")
		nextVipCount := logic.GetVipInfoLv(vipLevel+1, "copy_number")

		counts.FreeCount = freeCount
		counts.VipCount = vipCount
		counts.NextVipCount = nextVipCount
		leftTimes := freeCount + vipCount - times
		if leftTimes < 0 {
			leftTimes = 0
		}
		counts.LeftTimes = leftTimes

		openTypes := fc.OpenType
		if meetCondition(ctx, openTypes, userid, userGrade) || functionFights.GetIntE(copyID) != 0 {
			if functionFights.GetIntE(copyID) != 0 {
				// 已通关：可扫荡或次数用完
				if leftTimes <= 0 {
					tmp.Status = 2
				} else {
					tmp.Status = 1
				}
			} else {
				// 未通关：需挑战（无论次数是否用完，都显示挑战状态）
				tmp.Status = 3
			}
		} else {
			tmp.Status = 0
			tmp.OpenType = openTypes
		}

		tmp.Name = fc.Name
		tmp.CostType = fc.CostType
		tmp.Layer = fc.Layer
		if times < freeCount {
			tmp.BasicCost = 0
		} else {
			tmp.BasicCost = fc.BasicCost
		}
		tmp.Reward = fc.Reward

		if needCopyID != -1 {
			return &FunctionInfo{Counts: &counts, List: []*FunctionData{&tmp}}
		}
		dataList = append(dataList, &tmp)
	}

	return &FunctionInfo{Counts: &counts, List: dataList}
}

// ========================= 副本次数 =========================

func IncrRedisUserFunctionTimes(ctx context.Context, uid int64, copyID int) {
	daily.IncrByPrefix(uid, config.DailyFunctionTimes, copyID, 1)
}

// meetCondition 检查是否满足开启条件
func meetCondition(ctx context.Context, openTypes []util.TypeNum, userID int64, userGrade types.Map) bool {
	for _, ot := range openTypes {
		switch ot.Type {
		case 1: // 等级
			if userGrade.GetIntE("lv") < ot.Num {
				return false
			}
		case 2: // 战力
			fightPoint := GetUserFightPoint(ctx, userID, 1)
			if fightPoint < ot.Num {
				return false
			}
		}
	}
	return true
}
