package model

import (
	"context"
	"fmt"

	"server_golang/common/types"
	"server_golang/logic"
	"server_golang/repo/table"
	"server_golang/repo/world"
)

type UserPosition struct {
	Id       int         `json:"id"`
	UserId   int64       `json:"user_id"`
	PosType  int         `json:"pos_type"`
	Position int         `json:"position"`
	HeroPos  map[int]int `json:"hero_pos"`
}

// ReplaceUserPosition 替换用户阵位
func ReplaceUserPosition(ctx context.Context, data *table.UserPosition) error {
	return world.ReplaceUserPosition(ctx, data)
}

// GetUserPositionByID 根据阵位类型获取用户阵位信息
func GetUserPositionByID(ctx context.Context, userID int64, posType int) *UserPosition {
	data, _ := world.GetUserPosition(ctx, userID, posType)
	if len(data) == 0 {
		return nil
	}

	tmp := UserPosition{
		Id:       data.GetIntE("id"),
		UserId:   data.GetInt64E("user_id"),
		PosType:  data.GetIntE("pos_type"),
		Position: data.GetIntE("position"),
		HeroPos:  getPositionHeroIDs(data),
	}

	return &tmp
}

// GetUserPositionWithHeroAttrs 获取阵位信息+英雄完整属性(含技能)
// 对应 PHP: get_user_position_by_id($userId, $posType, true)
func GetUserPositionWithHeroAttrs(ctx context.Context, userID int64, posType int) []*logic.Hero {
	data := GetUserPositionByID(ctx, userID, posType)
	if data == nil {
		return nil
	}

	if len(data.HeroPos) == 0 {
		return []*logic.Hero{}
	}

	ids := make([]int, 0, len(data.HeroPos))
	for id := range data.HeroPos {
		ids = append(ids, id)
	}

	heroAttrs := GetUserHeroAttrWithSkillByIDs(ctx, ids, userID, true)

	// 设置 pos
	for id, pos := range data.HeroPos {
		if heroAttr, ok := heroAttrs[id]; ok {
			heroAttr.Pos = pos
		}
	}

	heroAttrList := make([]*logic.Hero, 0, len(heroAttrs))
	for _, hero := range heroAttrs {
		heroAttrList = append(heroAttrList, hero)
	}

	return heroAttrList
}

// GetUserPosition 获取用户所有阵位
func GetUserPosition(ctx context.Context, userID int64) ([]*UserPosition, error) {
	ret, err := world.GetUserPositionByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}

	result := []*UserPosition{}
	for _, v := range ret {
		result = append(result, &UserPosition{
			Id:       v.Id,
			UserId:   v.UserId,
			PosType:  v.PosType,
			Position: v.Position,
			HeroPos:  getPositionHeroIDs(types.ObjectToMap(v)),
		})
	}

	return result, nil
}

// 从阵位数据中解析英雄ID
func getPositionHeroIDs(data types.Map) map[int]int {
	heroPos := make(map[int]int)
	for i := 1; i <= 5; i++ {
		heroID := data.GetIntE(fmt.Sprintf("pos%d_hero", i))
		if heroID > 0 {
			pos := data.GetIntE(fmt.Sprintf("pos%d_pos", i))
			heroPos[heroID] = pos
		}
	}
	return heroPos
}
