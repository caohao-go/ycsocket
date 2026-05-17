package model

import (
	"context"
	"time"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/repo/mem/item"
	"server_golang/repo/table"
	"server_golang/repo/user"
)

// BuyStaff 购买物品处理
// 对应 PHP: ShinelightModel::buyStaff
// type: 0-充值 1-成长基金 2-特权商城 3-每日礼包 4-周超值礼包 5-月度超值礼包 6-礼包抢购 7-月基金 8-限时一元礼包
func BuyStaff(ctx context.Context, userID int64, buyType int, amt int, orderID int64) (types.Map, error) {
	// 获取用户VIP内容
	userVipContent := GetUserVipContents(userID)
	userVipContentDay := GetVipContentsCurDay(ctx, userID)

	// 累计消费
	userVipContent["leiji_xiaofei"] = userVipContent.GetIntE("leiji_xiaofei") + amt
	userVipContentDay["leiji_xiaofei"] = userVipContentDay.GetIntE("leiji_xiaofei") + amt

	// 荣耀月卡判断（类型1: 30元解锁）
	if GetRongyaoCardExpire(ctx, userID, 1) == 0 {
		currAmt := IncrRongyaoCardAmt(ctx, userID, 1, amt)
		if currAmt >= 30 {
			SetRongyaoCardExpire(ctx, userID, 1)
		}
	}

	// 至尊月卡判断（类型2: 98元解锁）
	if GetRongyaoCardExpire(ctx, userID, 2) == 0 {
		currAmt := IncrRongyaoCardAmt(ctx, userID, 2, amt)
		if currAmt >= 98 {
			SetRongyaoCardExpire(ctx, userID, 2)
		}
	}

	// 计算VIP等级
	vipLevel := logic.GetVipLevelByCostMoney(userVipContent.GetIntE("leiji_xiaofei"))

	var result types.Map

	switch buyType {
	case 0: // 充值
		userVipContent["leiji_chong"] = userVipContent.GetIntE("leiji_chong") + amt
		userVipContentDay["leiji_chong"] = userVipContentDay.GetIntE("leiji_chong") + amt

		// 首次充值判断
		firstChong, _ := types.ToMap(userVipContent["first_chong"], "")
		if firstChong == nil {
			firstChong = types.Map{}
		}
		amtKey := types.ToString(amt)
		addZuanshi := 0
		if firstChong.GetIntE(amtKey) == 0 {
			firstChong[amtKey] = 1
			addZuanshi = amt * 20
		} else {
			addZuanshi = amt * 10
		}
		userVipContent["first_chong"] = firstChong

		userZuanshi := item.Zuan(userID)
		item.AddZuan(userID, addZuanshi)
		result = types.Map{
			"add_zuan":    addZuanshi,
			"before_zuan": userZuanshi,
			"zuan":        userZuanshi + addZuanshi,
			"amt":         amt,
			"vip_level":   vipLevel,
		}

	case 1: // 成长基金
		jijin, _ := types.ToMap(userVipContent["jijin"], "")
		if jijin == nil {
			jijin = types.Map{}
		}
		jijin["status"] = 1
		userVipContent["jijin"] = jijin
		result = types.Map{"amt": amt}

	case 2: // 特权商城
		if amt == 30 { // 先知水晶礼包
			rewards := logic.TequanShop[1]
			GiveReward(userID, rewards...)
			weekContent := GetVipContentsCurWeek(ctx, userID)
			xianzhi, _ := types.ToMap(weekContent["xianzhi"], "")
			if xianzhi == nil {
				xianzhi = types.Map{}
			}
			xianzhi["limit"] = xianzhi.GetIntE("limit") + 1
			weekContent["xianzhi"] = xianzhi
			SetVipContentsCurWeek(ctx, userID, weekContent)
			result = types.Map{"rewards": rewards}
		} else if amt == 68 { // 快速作战礼包
			rewards := logic.TequanShop[2]
			GiveReward(userID, rewards...)
			monContent := GetVipContentsCurMon(ctx, userID)
			kuaisu, _ := types.ToMap(monContent["kuaisu"], "")
			if kuaisu == nil {
				kuaisu = types.Map{}
			}
			kuaisu["limit"] = kuaisu.GetIntE("limit") + 1
			kuaisu["expire"] = time.Now().Unix() + 86400*31
			monContent["kuaisu"] = kuaisu
			SetVipContentsCurMon(ctx, userID, monContent)
			result = types.Map{"rewards": rewards}
		}

	case 3: // 每日礼包
		rewards := logic.DayLibaoRewards[amt]
		GiveReward(userID, rewards...)
		dayLibao, _ := types.ToMap(userVipContentDay["day_libao"], "")
		if dayLibao == nil {
			dayLibao = types.Map{}
		}
		limit, _ := types.ToMap(dayLibao["limit"], "")
		if limit == nil {
			limit = types.Map{}
		}
		amtKey := types.ToString(amt)
		limit[amtKey] = limit.GetIntE(amtKey) + 1
		dayLibao["limit"] = limit
		userVipContentDay["day_libao"] = dayLibao
		result = types.Map{"rewards": rewards}

	case 4: // 周超值礼包
		zhouqi, _ := logic.CurrentWeekZhouqi()
		vipContentsLimitW := GetVipContentsLimitWeek(ctx, userID, zhouqi)
		rewards := logic.WeekLibaoRewards[amt]
		GiveReward(userID, rewards...)
		weekLibao, _ := types.ToMap(vipContentsLimitW["week_libao"], "")
		if weekLibao == nil {
			weekLibao = types.Map{}
		}
		limit, _ := types.ToMap(weekLibao["limit"], "")
		if limit == nil {
			limit = types.Map{}
		}
		amtKey := types.ToString(amt)
		limit[amtKey] = limit.GetIntE(amtKey) + 1
		weekLibao["limit"] = limit
		vipContentsLimitW["week_libao"] = weekLibao
		SetVipContentsLimitWeek(ctx, userID, zhouqi, vipContentsLimitW)
		result = types.Map{"rewards": rewards}

	case 5: // 月度超值礼包
		zhouqi, _ := logic.CurrentMonthZhouqi()
		vipContentsLimitMon := GetVipContentsLimitMon(ctx, userID, zhouqi)
		rewards := logic.YueduLibaoRewards[amt]
		GiveReward(userID, rewards...)
		yueduLibao, _ := types.ToMap(vipContentsLimitMon["yuedu_libao"], "")
		if yueduLibao == nil {
			yueduLibao = types.Map{}
		}
		limit, _ := types.ToMap(yueduLibao["limit"], "")
		if limit == nil {
			limit = types.Map{}
		}
		amtKey := types.ToString(amt)
		limit[amtKey] = limit.GetIntE(amtKey) + 1
		yueduLibao["limit"] = limit
		vipContentsLimitMon["yuedu_libao"] = yueduLibao
		SetVipContentsLimitMon(ctx, userID, zhouqi, vipContentsLimitMon)
		result = types.Map{"rewards": rewards}

	case 6: // 礼包抢购
		zhouqi, _ := logic.CurrentWeekZhouqi()
		vipContentsLimitW := GetVipContentsLimitWeek(ctx, userID, zhouqi)
		rewards := logic.LibaoQianggouRewards[amt]
		GiveReward(userID, rewards...)
		qianggou, _ := types.ToMap(vipContentsLimitW["qianggou"], "")
		if qianggou == nil {
			qianggou = types.Map{}
		}
		limit, _ := types.ToMap(qianggou["limit"], "")
		if limit == nil {
			limit = types.Map{}
		}
		amtKey := types.ToString(amt)
		limit[amtKey] = limit.GetIntE(amtKey) + 1
		qianggou["limit"] = limit
		vipContentsLimitW["qianggou"] = qianggou
		SetVipContentsLimitWeek(ctx, userID, zhouqi, vipContentsLimitW)
		result = types.Map{"rewards": rewards}

	case 7: // 月基金
		monJijin, _ := types.ToMap(userVipContent["mon_jijin"], "")
		if monJijin == nil {
			monJijin = types.Map{}
		}
		amtKey := types.ToString(amt)
		entry, _ := types.ToMap(monJijin[amtKey], "")
		if entry == nil {
			entry = types.Map{}
		}
		entry["status"] = 1
		entry["days"] = 0
		entry["last_ling_date"] = 0
		monJijin[amtKey] = entry
		userVipContent["mon_jijin"] = monJijin

		addZuanshi := amt * 10
		userZuanshi := item.Zuan(userID)
		item.AddZuan(userID, addZuanshi)
		result = types.Map{
			"add_zuan":    addZuanshi,
			"before_zuan": userZuanshi,
			"zuan":        userZuanshi + addZuanshi,
			"amt":         amt,
			"vip_level":   vipLevel,
		}

	case 8: // 限时一元礼包
		yiyuanlibao, _ := types.ToMap(userVipContent["yiyuanlibao"], "")
		if yiyuanlibao == nil {
			yiyuanlibao = types.Map{}
		}
		yiyuanlibao["buy_flag"] = 1
		userVipContent["yiyuanlibao"] = yiyuanlibao
		rewards := logic.YiyuanLibaoRewards
		GiveReward(userID, rewards...)
		result = types.Map{"rewards": rewards}
	}

	// 任意充值礼包检测
	anychong, _ := types.ToMap(userVipContent["anychong"], "")
	if anychong != nil && anychong.GetIntE("status") == 0 {
		if time.Now().Unix() <= int64(anychong.GetIntE("timestamp"))+7200 {
			anychong["status"] = 1
			userVipContent["anychong"] = anychong
		}
	}

	// 积天豪礼
	jitian, _ := types.ToMap(userVipContent["jitian"], "")
	if jitian == nil {
		jitian = types.Map{}
	}
	today := util.DateYmd()
	if jitian.GetIntE("leiji") <= 15 && jitian.GetStringE("last_date") < today && userVipContentDay.GetIntE("leiji_xiaofei") >= 6 {
		jitian["leiji"] = jitian.GetIntE("leiji") + 1
		jitian["last_date"] = today
		userVipContent["jitian"] = jitian
	}

	// 更新VIP等级
	SetVipLevel(userID, vipLevel)

	// 保存VIP内容
	UpdateUserVipContents(userID, userVipContent)
	SetVipContentsCurDay(ctx, userID, userVipContentDay)

	if result == nil {
		result = types.Map{}
	}

	return result, nil
}

// ========================= 订单/交易相关 =========================

// InsertNewTradeInfo 插入新订单
func InsertNewTradeInfo(ctx context.Context, data *table.NewTradeInfo) (int64, error) {
	return user.InsertNewTradeInfo(ctx, data)
}

// GetNewTradeInfoById 根据订单ID获取订单信息
func GetNewTradeInfoById(ctx context.Context, id int64) (*table.NewTradeInfo, error) {
	return user.GetNewTradeInfoById(ctx, id)
}

// UpdateNewTradeInfoById 根据订单ID更新订单信息
func UpdateNewTradeInfoById(ctx context.Context, id int64, updateData types.Map) error {
	return user.UpdateNewTradeInfoById(ctx, id, updateData)
}
