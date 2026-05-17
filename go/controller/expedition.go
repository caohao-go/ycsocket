package controller

import (
	"context"
	"fmt"

	"server_golang/common/lock"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/table"
)

// 远征信息
func (c *ShinelightController) GetExpeditionInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	userGrade := model.GetUserAttr(userID)

	data := model.GetExpeditionInfo(ctx, userID, userGrade)
	return c.ResponseSuccessToMe(types.ObjectToMap(data))
}

// 远征获取某一层数据（对齐 PHP getExpeditionLayerAction — 支持 rank 和 climb 两种对手类型）
func (c *ShinelightController) GetExpeditionLayerAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	layer := c.Params.GetIntE("layer")

	userGrade := model.GetUserAttr(userID)
	expeditionInfo := model.GetExpeditionInfo(ctx, userID, userGrade)

	choose := expeditionInfo.Choose
	currentLayer := expeditionInfo.CurrentLayer

	if choose == 0 {
		return c.ResponseError(33333, "未选择难度")
	}

	// 获取对手信息（支持 rank 和 climb 两种类型）
	op := model.GetExpeditionLayerOpHero(ctx, userID, expeditionInfo, layer)

	var opHeroDetail []*logic.Hero
	if op.Type == "rank" {
		// 真实玩家 — HeroAttr 已经是完整英雄属性
		opHeroDetail = op.HeroAttr
	} else {
		// NPC — 需要将 HeroBaseInfo 转成完整英雄属性
		opHeroDetail = model.GetFightHeroAttrWithSkill(ctx, op.Heros)
	}

	opHeros := logic.GetBaseFromHero(opHeroDetail)

	// 已通关的层敌方HP设为0
	if layer < currentLayer {
		for i := range opHeros {
			opHeros[i].CurrentHp = 0
		}
	}

	// 计算战力
	fightPoint := 0
	for _, v := range opHeroDetail {
		fightPoint += v.FightPoint
	}

	rewards := logic.ExpeditionRewards[choose][layer]

	data := types.Map{
		"nickname":    op.Nickname,
		"avatar_url":  op.AvatarUrl,
		"gender":      op.Gender,
		"lv":          op.Lv,
		"fight_point": fightPoint,
		"op_heros":    opHeros,
		"rewards":     rewards,
	}
	return c.ResponseSuccessToMe(data)
}

// 远征选择难度（对齐 PHP chooseExpeditionAction）
func (c *ShinelightController) ChooseExpeditionAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	copyID := c.Params.GetIntE("copy_id")
	if copyID != 20201 && copyID != 20202 && copyID != 20203 {
		return c.ResponseError(5523, "无此难度")
	}

	userGrade := model.GetUserAttr(userID)

	// 检查难度条件（与 PHP 一致：$expedition_info['copy'][$copy_id]['status'] != 1 直接报错，
	// 不存在的 key 取到的是 null，比较 != 1 为真，即视为不满足）
	expeditionInfo := model.GetExpeditionInfo(ctx, userID, userGrade)
	status := 0
	if len(expeditionInfo.Copy) > 0 {
		if statusVal, ok := expeditionInfo.Copy[copyID]; ok {
			status = statusVal.Status
		}
	}
	if status != 1 {
		return c.ResponseError(5123, "不满足选择条件")
	}

	// 检查是否已选择过（与 PHP 一致：$expedition_info['choose'] != 0）
	if expeditionInfo.Choose > 0 {
		return c.ResponseError(5333, "已经选择难度")
	}

	// 保存选择的副本ID，重置当前进度为第1层（与 PHP 一致）
	model.HsetExpeditionTodayInfo(ctx, userID, "choose", copyID)
	model.HsetExpeditionTodayInfo(ctx, userID, "current_layer", 1)

	// 与 PHP 一致：返回空数组（宝箱/HP redis key 带日期前缀，天然每日隔离，无需清理）
	return c.ResponseSuccessToMe(types.Map{})
}

// 远征战斗（对齐PHP原版，服务端真实战斗 15回合，支持 rank 和 climb 两种对手类型）
func (c *ShinelightController) UpExpeditionAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	copyID := c.Params.GetIntE("copy_id")
	layer := c.Params.GetIntE("layer")
	position := c.Params.GetIntE("position")
	_, fightHeros := util.ToPosHeros(c.Params.GetArrayE("fight_heros"))

	if len(fightHeros) == 0 {
		return c.ResponseError(6742, "请选择出战英雄")
	}

	userGrade := model.GetUserAttr(userID)
	expeditionInfo := model.GetExpeditionInfo(ctx, userID, userGrade)

	choose := expeditionInfo.Choose
	currentLayer := expeditionInfo.CurrentLayer

	// 校验难度是否已解锁（与 PHP upExpeditionAction 一致：$expedition_info['copy'][$copy_id]['status'] != 1 → 5433）
	cpStatus := 0
	if len(expeditionInfo.Copy) > 0 {
		if statusVal, ok := expeditionInfo.Copy[copyID]; ok {
			cpStatus = statusVal.Status
		}
	}
	if cpStatus != 1 {
		return c.ResponseError(5433, "不满足选择条件")
	}

	// 校验难度
	if choose == 0 {
		return c.ResponseError(3, "请先选择远征难度")
	}
	if copyID != choose {
		return c.ResponseError(9329, "难度异常")
	}
	if layer != currentLayer {
		return c.ResponseError(9679, "关卡异常")
	}
	// 与 PHP 一致：检查 current_layer（已通关）而非 layer
	if currentLayer > logic.MaxPos[choose] {
		return c.ResponseError(3199, "已经通关")
	}

	// 缓存阵容
	model.SetFightHeros(ctx, userID, "expedition", fightHeros, position)

	// 获取己方英雄完整属性
	fightHeroDetails := c.getFightHeroByPosMap(ctx, userID, fightHeros)

	// 支援英雄数量校验
	helpCount := c.getHelpCount(userID, fightHeroDetails)
	if helpCount > 1 {
		return c.ResponseError(7643, "只允许一个支援英雄上阵")
	}

	// 加分布式锁
	lock.Lock(fmt.Sprintf("upExpedition%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("upExpedition%d", userID))

	// 己方起始HP（从Redis持久化读取）
	p1HP := model.GetExpeditionHerosHP(ctx, userID, c.extractHeroIDs(fightHeroDetails))
	c.setStartHP(fightHeroDetails, p1HP)

	// 获取敌方阵容 — 对齐 PHP: 使用 getExpeditionLayerOpHero 支持 rank+climb
	op := model.GetExpeditionLayerOpHero(ctx, userID, expeditionInfo, layer)
	var opHeroDetail []*logic.Hero
	if op.Type == "rank" {
		// 真实玩家 — HeroAttr 已经是完整英雄属性
		opHeroDetail = op.HeroAttr
	} else {
		// NPC — 需要将 HeroBaseInfo 转成完整英雄属性
		opHeroDetail = model.GetFightHeroAttrWithSkill(ctx, op.Heros)
	}

	// 敌方起始HP
	opHeroIDs := make([]int, 0, len(opHeroDetail))
	for _, h := range opHeroDetail {
		if h.Id <= 0 {
			h.Id = h.HeroInfo
		}
		opHeroIDs = append(opHeroIDs, h.Id)
	}

	p2HP := model.GetExpeditionOpHerosHP(ctx, userID, opHeroIDs)
	// 敌方仍为 map，需先设置起始HP再转 Hero
	if p2HP != nil && len(p2HP) > 0 {
		for i, v := range opHeroDetail {
			if hpVal, existsInMap := p2HP[types.ToString(v.Id)]; existsInMap {
				hpNum := types.ToIntE(hpVal)
				maxHP := opHeroDetail[i].Hp
				if hpNum > 0 && hpNum < maxHP {
					opHeroDetail[i].CurrentHP = hpNum
				}
			}
		}
	}

	// ========== 真实战斗（15回合上限）==========
	fight := logic.NewFight(fightHeroDetails, opHeroDetail)
	winner, fightResult := fight.FightExec(15)
	success := 0
	if winner == "P1" {
		success = 1
	}

	// 持久化己方HP
	p1FinalHP := map[int]int{}
	for _, ha := range fight.HerosAttr["P1"] {
		hp := ha.CurrentHP
		if hp < 0 {
			hp = 0
		}
		p1FinalHP[ha.ID] = hp
	}
	model.SetExpeditionHerosHP(ctx, userID, p1FinalHP)

	if success == 1 {
		// 清除敌方HP
		model.DelExpeditionOpHerosHP(ctx, userID)

		// 推进层数
		newLayer := currentLayer + 1
		// 跳过宝箱层
		baoxiangPos := logic.BaoxiangPos[choose]
		for _, pos := range baoxiangPos {
			if newLayer == pos && newLayer < logic.MaxPos[choose] {
				newLayer++
			}
		}
		model.HsetExpeditionTodayInfo(ctx, userID, "current_layer", newLayer)

		// 发放奖励
		rewards := logic.ExpeditionRewards[choose][layer]
		if len(rewards) > 0 {
			model.GiveReward(userID, rewards...)
		}

		// 更新排行榜（按 copyID 映射积分，按星期分排行，TTL 8天）
		model.IncrExpeditionScoreRank(ctx, userID, choose)
		// 日常任务: 远征胜利3次
		model.SetDailyTaskFinish(ctx, userID, 10012, 1)
		// 周任务: 远征（对应 PHP Task::finishWeekTask($userId, "expedition")）
		model.IncrTaskFinishNumStr(ctx, userID, "expedition", 7)
	} else {
		// 失败：持久化敌方HP，下次继续打
		p2FinalHP := types.Map{}
		for _, ha := range fight.HerosAttr["P2"] {
			hp := ha.CurrentHP
			if hp < 0 {
				hp = 0
			}
			p2FinalHP[types.ToString(ha.ID)] = types.ToStr(hp)
		}
		model.SetExpeditionOpHerosHP(ctx, userID, p2FinalHP)
	}

	// 构建返回数据
	myHero := logic.GetBaseFromHero(fightHeroDetails)
	oppHero := logic.GetBaseFromHero(opHeroDetail)

	data := types.Map{
		"success":       success,
		"current_layer": currentLayer,
		"my_hero":       myHero,
		"opp_hero":      oppHero,
		"fight_result":  fightResult,
	}

	if success == 1 {
		data["rewards"] = logic.ExpeditionRewards[choose][layer]
		// 计算宝箱跳过后的实际层数（对齐 PHP $current_layer 在宝箱跳过后的值）
		newRetLayer := currentLayer + 1
		baoxiangPosRet := logic.BaoxiangPos[choose]
		for _, pos := range baoxiangPosRet {
			if newRetLayer == pos && newRetLayer < logic.MaxPos[choose] {
				newRetLayer++
			}
		}
		data["current_layer"] = newRetLayer
	}

	// 引导任务 & 公会任务
	model.GuideTaskHandle(ctx, userID, 114, 1)

	return c.ResponseSuccessToMe(data)
}

// 开启远征宝箱（openBaoxiang）— 对齐PHP原版
func (c *ShinelightController) OpenBaoxiangAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	copyID := c.Params.GetIntE("copy_id")
	pos := c.Params.GetIntE("pos")

	// 校验难度
	if copyID != 20201 && copyID != 20202 && copyID != 20203 {
		return c.ResponseError(5523, "无此难度")
	}

	// 校验是否宝箱位置
	baoxiangPos := logic.BaoxiangPos[copyID]
	isBaoxiang := false
	for _, bp := range baoxiangPos {
		if bp == pos {
			isBaoxiang = true
			break
		}
	}
	if !isBaoxiang {
		return c.ResponseError(5553, "该位置不是宝箱")
	}

	// 分布式锁
	lock.Lock(fmt.Sprintf("openBaoxiang%d", userID), 5)
	defer lock.Unlock(fmt.Sprintf("openBaoxiang%d", userID))

	// 标记已开启（与 PHP 一致）
	model.SetExpeditionBaoxiangOpen(ctx, userID, pos)

	// 发放宝箱奖励（与 PHP 一致：Expedition::getBaoxiangRewards($copy_id, $pos)）

	rewards := logic.ExpeditionRewards[copyID][pos]
	model.GiveReward(userID, rewards...)

	// 与 PHP 一致：response_success_to_me($rewards) — 将 rewards 数组以数字下标方式放入返回体
	// PHP 的 get_result_success 会给数组追加 c/m/reqid/code 字段，数字下标数组被强制转为 object
	resp := make(types.Map)
	for i, r := range rewards {
		resp[types.ToString(i)] = r
	}
	return c.ResponseSuccessToMe(resp)
}

// 远征助战英雄列表 — 对齐 PHP ExpeditionHelpListAction
func (c *ShinelightController) ExpeditionHelpListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := c.expeditionHelpInfo(ctx, userID)
	return c.ResponseSuccessToMe(data)
}

// 保存远征助战英雄 — 对齐 PHP expeditionHelpAddAction
func (c *ShinelightController) ExpeditionHelpAddAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetIntE("hero_id")

	model.ReplaceUserExpeditionHelpHero(ctx, &table.UserExpeditionHelpHero{
		UserId: userID,
		HeroId: heroID,
	})

	return c.ResponseSuccessToMe(types.Map{})
}

// expeditionHelpInfo 获取远征助战信息（对齐 PHP private function expeditionHelpInfo）
func (c *ShinelightController) expeditionHelpInfo(ctx context.Context, userID int64) types.Map {

	// 获取好友列表（与 PHP 一致：getFriendsList($userId) 默认 status=0）
	friends := model.GetFriendsList(ctx, userID, 0)
	if len(friends) == 0 {
		friends = []int64{userID}
	} else {
		friends = append(friends, userID)
	}

	// 获取好友+自己的助战英雄
	helpHeros := model.GetUserExpeditionHelpHero(ctx, friends)
	if len(helpHeros) == 0 {
		return types.Map{"help_me": []interface{}{}, "i_help_other": []interface{}{}}
	}

	// 收集英雄ID，获取英雄属性
	heroIDs := make([]int, 0, len(helpHeros))
	for _, h := range helpHeros {
		heroIDs = append(heroIDs, h.HeroId)
	}
	helpHeroDetails := model.GetUserHeroAttrByIDs(ctx, heroIDs, 0, false)

	// 批量获取用户信息
	userInfos := model.GetUsersWithDetail(ctx, friends, 1)

	// 获取已选择的助战英雄
	helpChooses := model.GetExpeditionHelpChoose(ctx, userID)

	helpMe := make([]types.Map, 0)
	iHelpOther := make([]types.Map, 0)

	for _, heroID := range heroIDs {
		val, ok := helpHeroDetails[heroID]
		if !ok {
			continue
		}
		tmp := types.Map{
			"id":          val.Id,
			"user_id":     val.UserId,
			"hero_id":     val.HeroInfo,
			"star":        val.Star,
			"stage":       val.Stage,
			"lv":          val.Lv,
			"fight_point": val.FightPoint,
			"choosed":     0,
		}
		if _, chosen := helpChooses[types.ToString(val.Id)]; chosen {
			tmp["choosed"] = 1
		}
		// 与 PHP 一致：无条件赋值 nickname（即使查不到也是 nil）
		if info, ok := userInfos[val.UserId]; ok {
			tmp["nickname"] = info.GetStringE("nickname")
		} else {
			tmp["nickname"] = nil
		}

		if val.UserId == userID {
			iHelpOther = append(iHelpOther, tmp)
		} else {
			helpMe = append(helpMe, tmp)
		}
	}

	return types.Map{"help_me": helpMe, "i_help_other": iHelpOther}
}

// 选择远征助战英雄 — 对齐 PHP expeditionChooseHelpAction
func (c *ShinelightController) ExpeditionChooseHelpAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	heroID := c.Params.GetStringE("hero_id")

	// 最多选3个
	helpHeros := model.GetExpeditionHelpChoose(ctx, userID)
	if len(helpHeros) >= 3 {
		return c.ResponseError(5723, "最多只能选3个援助英雄")
	}

	model.SetExpeditionHelpChoose(ctx, userID, heroID)

	return c.ResponseSuccessToMe(types.Map{})
}
