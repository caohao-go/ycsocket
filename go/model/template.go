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
)

// InitTemplate 初始化星河神殿
func InitTemplate(ctx context.Context) {
	// 固定6个位置，确保切片长度始终为6
	logic.TemplateInfo = make([]int64, 6)

	v, err := repo.RedisGet(ctx, config.KeyTemplateInfo)
	if err != nil {
		panic(fmt.Errorf("init template err: %v", err))
	}
	if v == "" {
		return
	}

	// 尝试 JSON 数组格式解析（Go 写入的格式）
	var ret []int64
	if json.Unmarshal(v, &ret) == nil && len(ret) > 0 {
		// JSON 数组成功：可能是 0-based [uid1, uid2, ...] 或 PHP 的 1-based 对象
		for i := 0; i < len(ret) && i < 6; i++ {
			logic.TemplateInfo[i] = ret[i]
		}
		return
	}

	// 尝试 JSON 对象格式解析（PHP serialize 转 JSON 或 PHP 的 1-based 关联数组）
	var retMap map[string]int64
	if json.Unmarshal(v, &retMap) == nil && len(retMap) > 0 {
		// PHP 的 $template_info 是 1-based: {"1": uid, "2": uid, ...}
		for i := 1; i <= 6; i++ {
			if uid, ok := retMap[fmt.Sprintf("%d", i)]; ok {
				logic.TemplateInfo[i-1] = uid
			}
		}
		return
	}
}

// GetTemplateInfo 获取星河神殿信息
func GetTemplateInfo(ctx context.Context, userid int64, myrank int) map[int]types.Map {
	data := make(map[int]types.Map)

	for i := 1; i <= 6; i++ {
		var templateUser int64
		if i-1 < len(logic.TemplateInfo) {
			templateUser = logic.TemplateInfo[i-1]
		}
		data[i] = GetTemplateDetail(ctx, userid, i, templateUser, myrank)
	}
	return data
}

// GetTemplateDetail 获取星河神殿位置详情
func GetTemplateDetail(ctx context.Context, userid int64, templatePos int, templateUser int64, myrank int) types.Map {
	data := make(types.Map)
	data["pos"] = templatePos
	data["name"] = logic.TemplateName[templatePos]

	if data["name"] == "" {
		return nil
	}

	templateUserGrade := GetUserAttr(templateUser)
	data["user_id"] = templateUser
	data["nickname"] = templateUserGrade.GetStringE("nickname")
	data["gender"] = templateUserGrade.GetStringE("gender")
	data["avatar_url"] = templateUserGrade.GetStringE("avatar_url")
	data["lv"] = templateUserGrade.GetStringE("lv")

	if myrank == 0 || myrank > logic.TemplateRank[templatePos] {
		data["status"] = -1
	} else {
		waitTimeout := GetTemplateFailTimeout(ctx, userid)
		if myrank > 0 && myrank <= logic.TemplateRank[templatePos] && waitTimeout == 0 {
			data["status"] = 0
		} else if waitTimeout > 0 {
			// PHP: $data['status'] = $wait_timeout - time(); 返回剩余秒数
			data["status"] = waitTimeout - int(time.Now().Unix())
		} else {
			data["status"] = -1
		}
	}

	return data
}

// GetTemplateLv 获取星河神殿等级
func GetTemplateLv(ctx context.Context, pos int) int {
	ret, _ := repo.RedisHGet(ctx, config.KeyTemplateLv, pos)
	return types.ToIntE(ret)
}

// AddTemplateLv 增加星河神殿等级
func AddTemplateLv(ctx context.Context, pos, addLv int) int {
	ret, _ := repo.RedisHIncrBy(ctx, config.KeyTemplateLv, pos, int64(addLv))
	return int(ret)
}

// GetTemplateOpHero 获取星河神殿对手英雄
func GetTemplateOpHero(pos, lv int) []*logic.Hero {
	heroLv := lv / 2
	if heroLv < 1 {
		heroLv = 1
	}

	if heroLv == logic.TemplateCurrentLv[pos] && len(logic.TemplateCurrentHeros[pos]) > 0 {
		return logic.TemplateCurrentHeros[pos]
	}

	logic.TemplateCurrentLv[pos] = heroLv

	stage := 0
	if heroLv >= 30 && heroLv < 40 {
		stage = 1
	} else if heroLv < 50 {
		stage = 2
	} else if heroLv < 60 {
		stage = 3
	} else if heroLv < 80 {
		stage = 4
	} else {
		stage = 5
	}

	star := 1
	if heroLv >= 100 && heroLv < 145 {
		star = 2
	} else if heroLv < 165 {
		star = 3
	} else if heroLv < 185 {
		star = 4
	} else if heroLv < 205 {
		star = 5
	} else if heroLv < 250 {
		star = 6
	} else {
		star = 7
	}

	opHeros := make([]*logic.HeroBaseInfo, 0)
	heros := logic.TemplateHeros[pos]
	i := 1
	for _, v := range heros {
		opHeros = append(opHeros, &logic.HeroBaseInfo{
			Id:     i,
			UserId: 1,
			HeroId: v[0],
			Star:   star,
			Stage:  stage,
			Lv:     heroLv,
			Pos:    v[1],
		})
		i++
	}

	// 计算星河神殿英雄属性（含技能、组合）
	// 对应 PHP: $op_heros_detail = ShinelightModel::get_user_hero_attr_with_skill_by_heros($op_heros);
	opHeroAttrs := GetFightHeroAttrWithSkill(context.Background(), opHeros)
	logic.TemplateCurrentHeros[pos] = opHeroAttrs
	return opHeroAttrs
}

// ========================= 星河神殿 =========================

// SaveTemplateInfo 持久化星河神殿占领信息到 pika
func SaveTemplateInfo(ctx context.Context, data []int64) {
	repo.RedisSet(ctx, config.KeyTemplateInfo, data, 0)
}

// GetTemplateFailTimeout 获取星河神殿挑战失败等待时间戳
func GetTemplateFailTimeout(ctx context.Context, uid int64) int {
	v, _ := repo.RedisGet(ctx, fmt.Sprintf(config.KeyTemplateInfoTimeout, uid))
	return types.ToIntE(v)
}

// SetTemplateFailTimeout 设置星河神殿挑战失败等待时间
func SetTemplateFailTimeout(ctx context.Context, uid int64) {
	k := fmt.Sprintf(config.KeyTemplateInfoTimeout, uid)
	repo.RedisSet(ctx, k, time.Now().Unix()+1200, 1200)
}

// DelTemplateFailTimeout 清除星河神殿挑战失败等待时间
func DelTemplateFailTimeout(ctx context.Context, uid int64) {
	k := fmt.Sprintf(config.KeyTemplateInfoTimeout, uid)
	repo.RedisDel(ctx, k)
}
