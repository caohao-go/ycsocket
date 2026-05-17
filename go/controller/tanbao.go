package controller

import (
	"context"
	"fmt"
	"math/rand"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// 探宝系统

func (c *ShinelightController) TanbaoInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")
	if id != 10 && id != 12 {
		return c.ResponseError(46388, "id is error")
	}

	costType := 21301
	if id == 12 {
		costType = 21302
	}

	var needTime int
	powerType := model.PowerTypeBaseTanbaoRefresh
	if id == 12 {
		powerType = model.PowerTypeHighTanbaoRefresh
	}
	power := model.GetUserPower(ctx, userID, powerType, &needTime)

	freshTime := 0
	freshBasicCost := 0
	if power == 0 {
		freshTime = needTime
		freshBasicCost = 30
	}

	data := types.Map{
		"rewards":          model.GetRand8Item(ctx, userID, id),
		"tanbaoquan_num":   item.Total(userID, costType),
		"lucky":            model.GetUserTanbaoLucky(userID, id),
		"lucky_lingqu":     getTanbaoLuckyLingqu(ctx, userID, id),
		"history":          model.GetTanbaoHistory(ctx, id),
		"fresh_cost_type":  2,
		"fresh_basic_cost": freshBasicCost,
		"fresh_time":       freshTime,
	}

	return c.ResponseSuccessToMe(data)
}

// getTanbaoLuckyLingqu 获取探宝幸运积分领取状态列表
// 与 PHP RedisProxy::getUserTanbaoLuckLingqu 逻辑一致
func getTanbaoLuckyLingqu(ctx context.Context, userID int64, id int) []types.Map {
	ret := make([]types.Map, 0)
	luckData, ok := logic.TanbaoLuckData[id]
	if !ok || len(luckData) == 0 {
		return ret
	}

	for integral, rewards := range luckData {
		status := model.HgetUserTanbaoLuckLingqu(ctx, userID, id, integral)
		entry := types.Map{
			"integral": integral,
			"rewards":  rewards,
			"status":   status,
		}
		ret = append(ret, entry)
	}

	// 判断是否全部都领完了（与 PHP tanbaoInfoAction 一致）
	allLingqu := true
	for _, v := range ret {
		if v.GetIntE("status") == 0 {
			allLingqu = false
			break
		}
	}

	if allLingqu && len(ret) > 0 {
		// PHP 顺序：先 -1000，再 del，再检查 <0 回滚
		newLucky := model.AddUserTanbaoLucky(userID, id, -1000)
		model.DelUserTanbaoLuckLingqu(ctx, userID, id)
		if newLucky < 0 {
			model.AddUserTanbaoLucky(userID, id, 1000)
		}
		// 重新获取领取状态（与 PHP 一致：$data['lucky_lingqu'] = RedisProxy::getUserTanbaoLuckLingqu()）
		ret = make([]types.Map, 0)
		for integral, rewards := range luckData {
			status := model.HgetUserTanbaoLuckLingqu(ctx, userID, id, integral)
			entry := types.Map{
				"integral": integral,
				"rewards":  rewards,
				"status":   status,
			}
			ret = append(ret, entry)
		}
	}
	return ret
}

// 探宝
func (c *ShinelightController) TanbaoAction(ctx context.Context) *Result {
	userID, _, userInfo, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")
	num := c.Params.GetIntE("num")
	if id != 10 && id != 12 {
		return c.ResponseError(46388, "id is error")
	}

	userGrade := model.GetUserAttr(userID)
	vipLv := userGrade.GetIntE("vip_level")
	if vipLv < 2 && (num == 15 || num == 10) {
		return c.ResponseError(46387, "vip2以上玩家专享")
	}

	lock.Lock(fmt.Sprintf("tanbao%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("tanbao%d", userID))

	var costType, costNum, addNum, tanbaoNum int
	if id == 10 {
		costType = 21301
		if num == 1 {
			costNum, addNum, tanbaoNum = 1, 1, 10
		} else {
			costNum, addNum, tanbaoNum = 12, 15, 150
		}
	} else {
		costType = 21302
		if num == 1 {
			costNum, addNum, tanbaoNum = 1, 1, 10
		} else {
			costNum, addNum, tanbaoNum = 10, 10, 100
		}
	}

	if item.NotEnough(userID, costType, costNum) {
		return c.ResponseError(999999, item.Name(costType)+"不够")
	}

	item.Sub(userID, costType, costNum)

	allRewards := model.GetRand8Item(ctx, userID, id)

	// 从 TanbaoRandData 构建概率映射（与 PHP $rand_temp = Tanbao::$tanbao_datas 一致）
	reward10 := make(map[int]int) // id => probability
	reward12 := make(map[int]int)
	for _, v := range logic.TanbaoRandData {
		vType := v.Type
		vID := v.Id
		vProb := v.Probability
		if vType == 10 {
			reward10[vID] = vProb
		} else if vType == 12 {
			reward12[vID] = vProb
		}
	}

	// 更新 allRewards 的 probability 并构建概率池（与 PHP 一致）
	probPool := make([]int, 0)
	for k := range allRewards {
		recordID := allRewards[k].Id
		if id == 10 {
			if prob, exists := reward10[recordID]; exists {
				allRewards[k].Probability = prob
			}
		}
		if id == 12 {
			if prob, exists := reward12[recordID]; exists {
				allRewards[k].Probability = prob
			}
		}
		probability := allRewards[k].Probability
		for i := 0; i < probability; i++ {
			probPool = append(probPool, k)
		}
	}

	// 幸运探宝不重复物品 ID（与 PHP 一致）
	notAgain10 := map[int]bool{167: true, 168: true, 550: true, 552: true, 581: true, 585: true, 586: true}
	notAgain12 := map[int]bool{170: true, 175: true, 194: true, 587: true, 187: true, 589: true, 590: true}

	keyVal := make([]int, 0, addNum)
	rewards := make([]*logic.ItemsCollection, 0, addNum)
	nickname := userInfo.Nickname

	for i := 0; i < addNum; i++ {
		if len(probPool) == 0 {
			break
		}
		randKey := rand.Intn(len(probPool))
		idx := probPool[randKey]
		keyVal = append(keyVal, idx)
		rewards = append(rewards, allRewards[idx])

		// 记录探宝历史（与 PHP 一致）
		history := types.Map{
			"items_id": allRewards[idx].ItemId,
			"nickname": nickname,
		}
		model.AddTanbaoHistory(ctx, id, history)

		// 稀有物品不重复（与 PHP 一致）
		// PHP: if ($id == 10 && in_array($all_rewards[$rand_key]['id'], $not_agin_10))
		// PHP 中 $rand_key 是 probabilitys 数组的索引，当 >= 8 时 $all_rewards[$rand_key] 为 null
		// 所以只在 randKey < len(allRewards) 时才检查
		if randKey < len(allRewards) {
			rewardID := allRewards[randKey].Id
			if id == 10 && notAgain10[rewardID] {
				newPool := make([]int, 0, len(probPool))
				for _, p := range probPool {
					if p != randKey {
						newPool = append(newPool, p)
					}
				}
				probPool = newPool
			} else if id == 12 && notAgain12[rewardID] {
				newPool := make([]int, 0, len(probPool))
				for _, p := range probPool {
					if p != randKey {
						newPool = append(newPool, p)
					}
				}
				probPool = newPool
			}
		}
	}

	// 增加幸运值（与 PHP 一致）
	model.AddUserTanbaoLucky(userID, id, 10*addNum)

	// 增加探宝积分（与 PHP 一致）
	item.Add(userID, 13, tanbaoNum, nil)

	// 给予奖励（与 PHP 一致）
	rewardsAll := make([]util.TypeNum, 0)
	for _, v := range rewards {
		rewardsAll = append(rewardsAll, util.TypeNum{
			Type: v.ItemId,
			Num:  v.Number,
		})
	}
	model.GiveReward(userID, rewardsAll...)

	ret := types.Map{
		"tanbaoquan_num": item.Total(userID, costType),
		"lucky":          model.GetUserTanbaoLucky(userID, id),
		"rewards":        rewards,
		"key":            keyVal,
	}

	// 引导任务（与 PHP 一致：先 22 和 30，再 id==10 时额外 22+1）
	lv := userGrade.GetIntE("lv")
	model.GuideTaskHandle(ctx, userID, 22, lv+1)
	model.GuideTaskHandle(ctx, userID, 30, lv+1)
	if id == 10 {
		model.GuideTaskHandle(ctx, userID, 22, 1)
	}

	return c.ResponseSuccessToMe(ret)
}

// 探宝幸运领取
func (c *ShinelightController) TanbaoLingquAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")
	integral := c.Params.GetIntE("integral")

	// 参数校验
	if id != 10 && id != 12 {
		return c.ResponseError(46388, "id is error")
	}

	// 配置存在性检查
	rewards, ok := logic.TanbaoLuckData[id]
	if !ok {
		return c.ResponseError(5689, "luck config not exist")
	}
	integralRewards, ok := rewards[types.ToIntE(integral)]
	if !ok || len(integralRewards) == 0 {
		return c.ResponseError(5689, "luck config not exist")
	}

	// 加锁
	lock.Lock(fmt.Sprintf("tanbaoLingqu%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("tanbaoLingqu%d", userID))

	// 重复领取检查
	if model.HgetUserTanbaoLuckLingqu(ctx, userID, id, integral) != 0 {
		return c.ResponseError(56849, "奖励已经领取")
	}

	// 标记已领取（与 PHP 一致）
	model.SetUserTanbaoLuckLingqu(ctx, userID, id, integral)

	// 判断是否全部都领完了（与 PHP tanbaoLingquAction 一致）
	allLingqu := true
	luckData := logic.TanbaoLuckData[id]
	for integ := range luckData {
		status := model.HgetUserTanbaoLuckLingqu(ctx, userID, id, integ)
		if status == 0 {
			allLingqu = false
			break
		}
	}

	if allLingqu {
		newLucky := model.AddUserTanbaoLucky(userID, id, -1000)
		if newLucky < 0 {
			model.AddUserTanbaoLucky(userID, id, 1000)
			return c.ResponseError(56849, "重复领取")
		}
		model.DelUserTanbaoLuckLingqu(ctx, userID, id)
	}

	// 发放奖励

	model.GiveReward(userID, integralRewards...)

	return c.ResponseSuccessToMe(types.Map{"rewards": integralRewards})
}

// 探宝刷新
func (c *ShinelightController) TanbaoRefreshAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	id := c.Params.GetIntE("id")
	if id != 10 && id != 12 {
		return c.ResponseError(46388, "id is error")
	}

	data := types.Map{
		"fresh_cost_type": 2,
	}

	powerType := model.PowerTypeBaseTanbaoRefresh
	if id == 12 {
		powerType = model.PowerTypeHighTanbaoRefresh
	}
	var needTime int
	power := model.GetUserPower(ctx, userID, powerType, &needTime)

	if power <= 0 {
		data["fresh_basic_cost"] = 30
		if item.NotEnough(userID, 2, 30) {
			return c.ResponseError(999999, "钻石不够")
		}
		data["fresh_time"] = needTime
	} else {
		data["fresh_basic_cost"] = 0
	}

	if power <= 0 {
		item.Sub(userID, 2, 30)
	} else {
		model.SubUserPower(ctx, userID, powerType, &needTime, 1)
		data["fresh_time"] = needTime
	}

	// 先删除旧数据再重新生成（与 PHP 一致：delUserTanbao + getRand8Item）
	model.DelUserTanbaoCache(userID, id)
	data["rewards"] = model.GetRand8Item(ctx, userID, id)

	return c.ResponseSuccessToMe(data)
}
