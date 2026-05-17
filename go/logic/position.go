// 组合数据模块
package logic

import (
	"context"
	"fmt"
	"sort"

	"server_golang/common/json"
	"server_golang/repo/table"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"

	"server_golang/repo/info"
)

// 位置全局数据 id => data
var PositionDatas map[int]*table.Position

// 组合全局数据 key(排序后property拼接) => data
var CombinationDatas map[string]table.Combination

// InitPositions 初始化位置数据
func InitPositions(ctx context.Context) {
	PositionDatas = make(map[int]*table.Position)

	rows, err := info.GetAllPosition(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init position error: %v", err)
		return
	}
	for _, data := range rows {
		PositionDatas[data.Id] = data
	}
}

// InitCombination 初始化组合数据
func InitCombination(ctx context.Context) {
	CombinationDatas = make(map[string]table.Combination)

	rows, err := info.GetAllCombination(ctx)
	if err != nil || len(rows) == 0 {
		panic(fmt.Errorf("init combination error %v", err))
	}

	for _, row := range rows {
		var propertys []int
		_ = json.Unmarshal(row.Property, &propertys)
		CombinationDatas[getPropertyKey(propertys)] = *row
	}
}

// CombinationAttrAdd 根据英雄组合获取加成
func CombinationAttrAdd(heros []*Hero) {
	if len(heros) != 5 {
		return
	}

	propertys := make([]int, 0, 5)
	for _, h := range heros {
		propertys = append(propertys, h.Property)
	}

	add, ok := CombinationDatas[getPropertyKey(propertys)]
	if !ok {
		return
	}

	for k, v := range heros {
		addHP := v.Hp * add.Hp / 100
		addAtk := v.Atk * add.Atk / 100
		addOppControl := add.OppControl
		addFightPoint := int(float64(addHP)*0.4 + float64(addAtk)*2 + float64(addOppControl)*80)

		heros[k].Hp = v.Hp + addHP
		heros[k].Atk = v.Atk + addAtk
		heros[k].OppControl = v.OppControl + addOppControl
		heros[k].FightPoint = v.FightPoint + addFightPoint
	}

	return
}

func getPropertyKey(propertys []int) string {
	sort.Ints(propertys)
	var sortProperty string
	for _, p := range propertys {
		sortProperty = fmt.Sprintf("%s,%d", sortProperty, p)
	}
	return sortProperty
}
