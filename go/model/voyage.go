package model

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/repo"
	"server_golang/repo/mem/daily"
)

// GetVoyageList 获取远航列表
func GetVoyageList(ctx context.Context, userID int64, lv int) []types.Map {
	voyageList := getVoyageList(ctx, userID)
	if len(voyageList) == 0 {
		voyageList = logic.GetRandVoyage(10, lv)
		SetVoyageList(ctx, userID, voyageList)
	}

	now := time.Now().Unix()
	for i, v := range voyageList {
		if v.GetIntE("status") == 1 {
			beatTime := int64(v.GetIntE("beat_time"))
			vTime := int64(v.GetIntE("time"))
			leftTime := beatTime + vTime - now
			if leftTime < 0 {
				leftTime = 0
			}
			voyageList[i]["left_time"] = leftTime
			if leftTime <= 0 {
				voyageList[i]["fin"] = 1
			} else {
				voyageList[i]["fin"] = 0
			}

			// 已接任务的加速费用按剩余时间比例折算
			timeNum := int(vTime / 3600)
			accelerateNum := v.GetIntE("accelerate_num")
			var perHourNum int
			if timeNum > 0 {
				perHourNum = accelerateNum / timeNum
			}
			leftHours := int(leftTime / 3600)
			consumedNum := (timeNum - leftHours) * perHourNum
			newAccelerateNum := accelerateNum - consumedNum
			// 少于半小时不可加速
			if leftTime < 1800 {
				newAccelerateNum = 0
			}
			voyageList[i]["accelerate_num"] = newAccelerateNum
		}
	}

	return voyageList
}

// RefreshVoyageList 刷新远航列表（对齐 PHP：删除未接任务 + 追加10个新任务）
func RefreshVoyageList(ctx context.Context, userID int64, lv int) {
	voyageList := getVoyageList(ctx, userID)
	if len(voyageList) == 0 {
		voyageList = logic.GetRandVoyage(10, lv)
	} else {
		// 删除未接任务(status==0)，保留已接的
		var keep []types.Map
		for _, v := range voyageList {
			if v.GetIntE("status") == 1 {
				keep = append(keep, v)
			}
		}
		newList := logic.GetRandVoyage(10, lv)
		voyageList = append(keep, newList...)
	}
	SetVoyageList(ctx, userID, voyageList)
}

// BeatVoyage 派遣英雄执行远航（对齐 PHP 原版）
func BeatVoyage(ctx context.Context, userID int64, id int, heros []int) types.Map {
	voyageList := GetVoyageList(ctx, userID, 0)
	data := types.Map{}
	for k, v := range voyageList {
		if v.GetIntE("id") == id {
			now := time.Now().Unix()
			voyageList[k]["status"] = 1
			voyageList[k]["beat_time"] = now
			voyageList[k]["beat_heros"] = heros
			vTime := int64(v.GetIntE("time"))
			for _, hero := range heros {
				SetVoyageHero(ctx, userID, hero, now+vTime)
			}
			data = v
			break
		}
	}
	SetVoyageList(ctx, userID, voyageList)
	return data
}

// AccelerateVoyage 加速完成远航（回拨 beat_time 使 left_time<=0）
func AccelerateVoyage(ctx context.Context, userID int64, id int) {
	list := getVoyageList(ctx, userID)
	for k, v := range list {
		if v.GetIntE("id") == id && v.GetIntE("status") == 1 {
			// 回拨24小时使剩余时间必然<=0
			list[k]["beat_time"] = time.Now().Unix() - 86400
			break
		}
	}
	SetVoyageList(ctx, userID, list)
}

// LingquVoyage 领取远航奖励（对齐 PHP 原版）
// 返回: (成功?true/false, 领取到的items_id, num, 完整列表用于取数据)
func LingquVoyage(ctx context.Context, userID int64, id int) (bool, int, int) {
	voyageList := GetVoyageList(ctx, userID, 0)
	itemsID, num := 0, 0
	for k, v := range voyageList {
		if v.GetIntE("id") == id {
			beatTime := int64(v.GetIntE("beat_time"))
			vTime := int64(v.GetIntE("time"))
			leftTime := beatTime + vTime - time.Now().Unix()
			if v.GetIntE("status") != 1 || leftTime > 0 {
				return false, 0, 0
			}

			itemsID = v.GetIntE("items_id")
			num = v.GetIntE("num")

			// 解锁英雄：遍历 beat_heros 主动删除英雄锁
			// PHP 中遍历 $v['hero']（配置数据）取 $val['id']，但实际无效（配置只有 prop/star 没有 id）
			// 正确做法是遍历 beat_heros（实际派遣的英雄 ID），主动释放锁
			beatHeros, ok, e := v.GetInt64Array("beat_heros")
			if ok && e == nil {
				DelVoyageHero(ctx, userID, beatHeros...)
			}

			// 从列表中移除该任务
			voyageList = append(voyageList[:k], voyageList[k+1:]...)
			break
		}
	}
	SetVoyageList(ctx, userID, voyageList)
	return true, itemsID, num
}

// ========================= 远航 =========================

func GetVoyageFreeCnt(ctx context.Context, uid int64) int {
	v, _ := daily.Get(uid, config.DailyVoyageFreeCnt)
	return 2 - types.ToIntE(v)
}

func IncrVoyageFreeCnt(ctx context.Context, uid int64) {
	daily.Incr(uid, config.DailyVoyageFreeCnt, 1)
}

func SetVoyageList(ctx context.Context, uid int64, data []types.Map) {
	repo.RedisSet(ctx, fmt.Sprintf(config.KeyVoyageList, uid), data, 30*86400)
}

func getVoyageList(ctx context.Context, uid int64) []types.Map {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyVoyageList, uid))
	if v == "" {
		return []types.Map{}
	}
	var ret []types.Map
	json.Unmarshal(v, &ret)
	if ret == nil {
		return []types.Map{}
	}
	return ret
}

func SetVoyageHero(ctx context.Context, uid int64, heroID int, expireTime int64) {
	repo.RedisHSet(ctx, fmt.Sprintf(config.KeyVoyageHero, uid), heroID, expireTime)
}

// DelVoyageHero 释放远航英雄
func DelVoyageHero(ctx context.Context, uid int64, heroID ...int64) {
	heros := make([]interface{}, len(heroID))
	for k, v := range heroID {
		heros[k] = v
	}

	repo.RedisHDel(ctx, fmt.Sprintf(config.KeyVoyageHero, uid), heros...)
}

// GetVoyageHero 获取远航中的英雄ID列表
func GetVoyageHero(ctx context.Context, uid int64) []string {
	k := fmt.Sprintf(config.KeyVoyageHero, uid)
	data, _ := repo.RedisHGetAll(ctx, k)
	now := time.Now().Unix()
	var delKeys []string
	active := make(types.Map)
	for h, et := range data {
		if types.ToInt64E(et) < now {
			delKeys = append(delKeys, h)
		} else {
			active[h] = et
		}
	}
	if len(active) == 0 {
		repo.RedisDel(ctx, k)
	} else {
		for _, dk := range delKeys {
			repo.RedisHDel(ctx, k, dk)
		}
	}
	ret := make([]string, 0, len(active))
	for h := range active {
		ret = append(ret, h)
	}
	return ret
}
