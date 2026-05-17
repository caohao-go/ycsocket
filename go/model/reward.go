package model

import (
	"context"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/repo/mem/item"
)

// GiveReward rewards 格式: [{"type": 1, "num": 100}, {"type": 2, "num": 50}]
// 远航令牌(item_id=6)上限，与 PHP Power::POWER_TYPE_VOYAGE 一致
const voyageExpMax = 22000

func GiveReward(userID int64, rewards ...util.TypeNum) error {
	ctx := context.Background()

	var hasExp bool
	for _, reward := range rewards {
		if reward.Type == item.ItemExp {
			hasExp = true
		}

		itemId := reward.Type
		num := reward.Num
		if itemId <= 0 || num <= 0 {
			continue
		}
		// 远航令牌上限检查：与 PHP 一致，不超过 22000
		if itemId == 6 {
			currentNum := item.Total(userID, 6)
			if currentNum >= voyageExpMax {
				continue
			}
			if currentNum+num > voyageExpMax {
				num = voyageExpMax - currentNum
			}
		}
		_, err := item.Add(userID, itemId, num, reward.Prop)
		if err != nil {
			log.Errorf(ctx, 298481, "发放奖励失败: uid=%d, item_id=%d, num=%d, err=%v", userID, itemId, num, err)
		}
	}

	if hasExp {
		upLv(ctx, userID)
	}

	return nil
}

func AddExp(userId int64, exp int) error {
	return GiveReward(userId, util.TypeNum{Type: item.ItemExp, Num: exp})
}

func upLv(ctx context.Context, userID int64) {
	userGrade := GetUserAttr(userID)
	beforeLv := userGrade.GetIntE("lv")
	// 升级处理（严格对齐 PHP do-while 逻辑）
	exp := item.Exp(userID)
	tempExp := exp
	lvUp := 0               // 对应 PHP 的 ($i - 1)
	guideLv := beforeLv + 1 // 与 PHP 一致：引导任务固定使用 $user_grade['lv'] + 1
	for {
		nextNeedExp := logic.GetLvUpdateExp(beforeLv + lvUp) // 对应 PHP $user_lv_data[$user_grade['lv'] + $i]
		if nextNeedExp <= 0 {
			break
		}
		tempExp -= nextNeedExp
		if tempExp > 0 {
			if beforeLv < 200 {
				lvUp++
				item.Sub(userID, item.ItemExp, nextNeedExp)

				// 引导任务（PHP lv <= 36 判定，参数固定为 $user_grade['lv']+1）
				if guideLv <= 36 {
					guideTaskIDs := []int{23, 30, 35, 40, 46, 54, 60, 64, 68, 69, 74, 81, 88, 92, 104, 106, 110, 113, 115}
					for _, tid := range guideTaskIDs {
						GuideTaskHandle(ctx, userID, tid, guideLv)
					}
				}

				IncrUserLv(userID)
			}
		} else {
			break
		}
	}

	return
}
