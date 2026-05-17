// 星河神殿模块 - 6个位置、挑战、属性加成
package logic

import (
	"context"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/repo/info"
	"server_golang/repo/table"
)

// NumPair 通用的 type+num 配对
type NumPair struct {
	Type string `json:"type"`
	Num  int    `json:"num"`
}

// 称号全局数据 id => data
var TitleDatas map[int]*table.RoleTitle

// 星河神殿位置常量
const (
	TemplateWanshan = 1
	TemplateTaitan1 = 2
	TemplateTaitan2 = 3
	TemplateHanhai1 = 4
	TemplateHanhai2 = 5
	TemplateHanhai3 = 6
)

// 星河神殿名称
var TemplateName = map[int]string{
	TemplateWanshan: "万山之巅",
	TemplateTaitan1: "泰坦神耀", TemplateTaitan2: "泰坦神耀",
	TemplateHanhai1: "瀚海星灵", TemplateHanhai2: "瀚海星灵", TemplateHanhai3: "瀚海星灵",
}

// 星河神殿排名要求
var TemplateRank = map[int]int{
	TemplateWanshan: 10,
	TemplateTaitan1: 20, TemplateTaitan2: 20,
	TemplateHanhai1: 50, TemplateHanhai2: 50, TemplateHanhai3: 50,
}

// 星河神殿英雄配置 pos => []{ hero_id, pos }
var TemplateHeros = map[int][][2]int{
	TemplateWanshan: {{3201, 2}, {4301, 4}, {1201, 6}, {2402, 7}, {1503, 9}},
	TemplateTaitan1: {{1502, 1}, {1203, 3}, {3102, 4}, {3203, 6}, {1201, 8}},
	TemplateTaitan2: {{3302, 1}, {4501, 2}, {2301, 3}, {4301, 5}, {1304, 8}},
	TemplateHanhai1: {{2101, 2}, {1502, 5}, {1101, 7}, {1101, 8}, {4101, 9}},
	TemplateHanhai2: {{2101, 1}, {1502, 3}, {1101, 5}, {1101, 7}, {4101, 9}},
	TemplateHanhai3: {{2101, 2}, {1502, 4}, {1101, 5}, {1101, 6}, {4101, 8}},
}

// 星河神殿信息
var TemplateInfo []int64

// 当前等级缓存
var TemplateCurrentLv = map[int]int{1: 0, 2: 0, 3: 0, 4: 0, 5: 0, 6: 0}
var TemplateCurrentHeros = map[int][]*Hero{1: nil, 2: nil, 3: nil, 4: nil, 5: nil, 6: nil}

// InitTitle 初始化称号数据
func InitTitle(ctx context.Context) {
	TitleDatas = make(map[int]*table.RoleTitle)

	rows, err := info.GetAllRoleTitle(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init role_title error: %v", err)
		return
	}

	for _, val := range rows {
		TitleDatas[val.Id] = val
	}
}

// GetTemplateAttrAdd 获取星河神殿属性加成
func GetTemplateAttrAdd(pos int) []NumPair {
	ret := make([]NumPair, 0, 2)
	if pos <= 1 {
		ret = append(ret, NumPair{Type: "hp", Num: 3}, NumPair{Type: "atk", Num: 3})
	} else if pos <= 3 {
		ret = append(ret, NumPair{Type: "hp", Num: 2}, NumPair{Type: "atk", Num: 2})
	} else {
		ret = append(ret, NumPair{Type: "hp", Num: 1}, NumPair{Type: "atk", Num: 1})
	}
	return ret
}
