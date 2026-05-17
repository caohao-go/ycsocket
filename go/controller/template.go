package controller

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/cache"
)

// ======================== 星河神殿 ========================

// 星河神殿信息
func (c *ShinelightController) templateInfoAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	myRank := model.GetMyTutengRank(ctx, userID)
	data := model.GetTemplateInfo(ctx, userID, int(myRank))
	return c.ResponseSuccessToMe(types.Map{"myrank": myRank, "data": data})
}

// 星河神殿详情（templateDetail）
func (c *ShinelightController) templateDetailAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	pos := c.Params.GetIntE("pos")
	if pos < 1 || pos > 6 {
		return c.ResponseError(7822, "pos error")
	}

	lv := model.GetTemplateLv(ctx, pos)
	myRank := int(model.GetMyTutengRank(ctx, userID))

	data := types.Map{
		"lv":       lv,
		"myrank":   myRank,
		"add_attr": logic.GetTemplateAttrAdd(pos),
	}

	// 当前占领者信息
	var insideUID int64
	if pos-1 < len(logic.TemplateInfo) {
		insideUID = logic.TemplateInfo[pos-1]
	}
	data["inside"] = types.Map{"uid": insideUID}
	if insideUID > 0 {
		insideGrade := model.GetUserAttr(insideUID)
		data["inside"] = types.Map{
			"uid":      insideUID,
			"nickname": insideGrade.GetStringE("nickname"),
			"lv":       insideGrade.GetIntE("lv"),
		}
	}

	// 对手英雄
	heroDetails := model.GetTemplateOpHero(pos, lv)
	heros := make([]types.Map, 0)
	skills := make([]types.Map, 0)
	for _, v := range heroDetails {
		heros = append(heros, types.Map{
			"id":          v.Id,
			"hero_id":     v.HeroInfo,
			"star":        v.Star,
			"stage":       v.Stage,
			"lv":          v.Lv,
			"pos":         v.Pos,
			"fight_point": v.FightPoint,
		})
		// 技能
		for _, skill := range v.Skills {
			skills = append(skills, types.Map{
				"id":       skill.ID,
				"skill_id": skill.SkillID,
				"name":     skill.Name,
				"level":    skill.Level,
				"type":     skill.Type,
			})
		}
	}
	data["position"] = 101
	data["heros"] = heros
	data["skills"] = skills

	return c.ResponseSuccessToMe(data)
}

// 升级星河神殿（upTemplate）— 对齐PHP原版，服务端真实战斗 15回合
func (c *ShinelightController) UpTemplateAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	pos := c.Params.GetIntE("pos")
	addLv := c.Params.GetIntE("add_lv")
	if addLv <= 0 {
		addLv = 1
	}
	position := c.Params.GetIntE("position")
	_, fightHeros := util.ToPosHeros(c.Params.GetArrayE("fight_heros"))

	myRank := model.GetMyTutengRank(ctx, userID)

	// 检查等待时间（返回的是过期时间戳，需转为剩余秒数）
	waitTimeout := model.GetTemplateFailTimeout(ctx, userID)
	if waitTimeout > 0 {
		remainSec := waitTimeout - int(time.Now().Unix())
		if remainSec > 0 {
			return c.ResponseError(54123, fmt.Sprintf("需要等待%d分钟", remainSec/60+1))
		}
		// 已过期，清除脏数据
		model.DelTemplateFailTimeout(ctx, userID)
	}

	// 检查状态
	var templateUser int64
	if pos-1 < len(logic.TemplateInfo) {
		templateUser = logic.TemplateInfo[pos-1]
	}
	detail := model.GetTemplateDetail(ctx, userID, pos, templateUser, int(myRank))
	if detail == nil || detail.GetIntE("status") != 0 {
		return c.ResponseError(54153, "状态异常")
	}

	if len(fightHeros) == 0 {
		return c.ResponseError(6742, "请选择出战英雄")
	}

	// 保存战斗阵容
	model.SetFightHeros(ctx, userID, "template", fightHeros, position)

	// 获取我方英雄完整属性（含技能）
	fightHeroDetails := c.getFightHeroByPosMap(ctx, userID, fightHeros)

	// 获取对手NPC完整属性（直接从配置表构建，不依赖回调查用户表）
	lv := model.GetTemplateLv(ctx, pos)
	heroLv := lv / 2
	if heroLv < 1 {
		heroLv = 1
	}

	// 从 TemplateHeros 配置表构建 NPC 基础配置
	opHeroDetail := model.GetTemplateOpHero(pos, lv)
	if len(opHeroDetail) == 0 {
		return c.ResponseError(54153, "对手配置不存在")
	}

	// ========== 服务端真实战斗（15回合上限）==========
	fight := logic.NewFight(fightHeroDetails, opHeroDetail)
	winner, fightResult := fight.FightExec(15)
	success := 0
	if winner == "P1" {
		success = 1
	}

	data := types.Map{"success": success}
	if success == 1 {
		// 确保 TemplateInfo 切片长度至少为6（防御性保护）
		for len(logic.TemplateInfo) < 6 {
			logic.TemplateInfo = append(logic.TemplateInfo, 0)
		}

		// 清除旧的占领位置，记录旧pos用于同区域判断
		oldPos := 0
		for k, v := range logic.TemplateInfo {
			if v == userID {
				logic.TemplateInfo[k] = 0
				oldPos = k + 1 // 转为1-based
			}
		}
		// 设置新位置
		logic.TemplateInfo[pos-1] = userID

		// 持久化到 Redis（多实例共享）
		model.SaveTemplateInfo(ctx, logic.TemplateInfo)

		oldFightPoint := model.GetUserFightPoint(ctx, userID, 1)

		// 清除属性缓存
		cache.ClearAddPropCache(fmt.Sprintf(config.CacheAddProp, userID))

		// 触发英雄属性重算（对齐 PHP：$this->shinelight_model->getUserPositionById($userId, 1, true) 刷新战力）
		model.GetUserPositionWithHeroAttrs(ctx, userID, 1)

		// 计算战力变化
		newFightPoint := model.GetUserFightPoint(ctx, userID, 1)
		data["old_fight_point"] = oldFightPoint
		data["new_fight_point"] = newFightPoint

		// 同区域不加分（对齐PHP原版）
		sameArea := (pos == 1 && oldPos == 1) ||
			((pos == 2 || pos == 3) && (oldPos == 2 || oldPos == 3)) ||
			((pos >= 4 && pos <= 6) && (oldPos >= 4 && oldPos <= 6))
		if sameArea {
			data["add_fight_point"] = 0
		} else {
			data["add_fight_point"] = newFightPoint - oldFightPoint
			if data.GetIntDefault("add_fight_point", 0) < 0 {
				data["add_fight_point"] = 0
			}
		}

		// 增加等级
		model.AddTemplateLv(ctx, pos, addLv)
	} else {
		model.SetTemplateFailTimeout(ctx, userID)
		data["add_fight_point"] = 0
	}

	// 构建战斗展示数据
	myHero := logic.GetBaseFromHero(fightHeroDetails)
	oppHero := logic.GetBaseFromHero(opHeroDetail)

	data["my_hero"] = myHero
	data["opp_hero"] = oppHero
	data["fight_result"] = fightResult

	// 成就任务
	model.IncrXingheNum(userID)
	xingheNum := model.GetXingheNum(userID)
	model.AchieveTaskHandle(ctx, userID, 17, xingheNum, 12001, 12007)

	// 公会任务
	model.GuideTaskHandle(ctx, userID, 116, 1)

	return c.ResponseSuccessToMe(data)
}
