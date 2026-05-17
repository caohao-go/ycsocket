package util

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"server_golang/common/types"
)

// TypeNum 通用的 type+num 配对
type TypeNum struct {
	Type       int      `json:"type"`
	Num        int      `json:"num"`
	Probablity int      `json:"probablity,omitempty"`
	HeroId     int      `json:"hero_id,omitempty"`
	Star       int      `json:"star,omitempty"`
	Ids        []int    `json:"ids,omitempty"`
	Prop       []FuProp `json:"prop,omitempty"`
}

type Fu struct {
	Id     int64    `json:"id"`
	ItemId int      `json:"item_id"`
	Unlock int8     `json:"unlock"`
	Props  []FuProp `json:"prop"`
}

type FuProp struct {
	MaxNum int `json:"max_num,omitempty"`
	Num    int `json:"num"`
	Prop   int `json:"prop"`
	Type   int `json:"type"`
}

type MinMax struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

// ToTypeNums 解析JSON：[[物品id,数量], ...] => [{type: xx, num: xx}, ...]
func ToTypeNums(typeNumStr string) []TypeNum {
	result := []TypeNum{}

	if typeNumStr == "0" || typeNumStr == "" || typeNumStr == "[]" {
		return result
	}

	if len(typeNumStr) < 2 || typeNumStr[0] != '[' || typeNumStr[len(typeNumStr)-1] != ']' {
		panic("invalid TypeNum: " + typeNumStr)
	}

	if len(typeNumStr) > 4 && typeNumStr[0] == '[' && typeNumStr[1] == '[' &&
		typeNumStr[len(typeNumStr)-1] == ']' && typeNumStr[len(typeNumStr)-2] == ']' {
		var rawTypeNum [][]int
		err := json.Unmarshal([]byte(typeNumStr), &rawTypeNum)
		if err == nil {
			for _, r := range rawTypeNum {
				if len(r) < 2 {
					panic(fmt.Sprintf("invalid TypeNum: %s", typeNumStr))
				}
				typeNum := TypeNum{Type: r[0], Num: r[1], Ids: r}
				if len(r) > 2 {
					typeNum.Probablity = r[2]
				}
				result = append(result, typeNum)
			}
			return result
		}
	} else {
		var rawTypeNum []int
		err := json.Unmarshal([]byte(typeNumStr), &rawTypeNum)
		if err == nil {
			if len(rawTypeNum) < 2 {
				panic(fmt.Sprintf("invalid TypeNum: %v", typeNumStr))
			}
			typeNum := TypeNum{Type: rawTypeNum[0], Num: rawTypeNum[1], Ids: rawTypeNum}
			if len(rawTypeNum) > 2 {
				typeNum.Probablity = rawTypeNum[2]
			}
			return []TypeNum{typeNum}
		}
	}

	err := json.Unmarshal([]byte(typeNumStr), &result)
	if err == nil {
		return result
	}

	panic("invalid TypeNum: " + typeNumStr)
}

func ToIDs(s string) []int {
	s = strings.TrimSpace(s)
	if s == "" || s == "0" || s == "[]" {
		return []int{}
	}

	s = strings.Replace(s, "[", "", 1)
	s = strings.Replace(s, "]", "", 1)
	sArr := strings.Split(s, ",")
	ids := make([]int, len(sArr))
	for i, sv := range sArr {
		ids[i], _ = strconv.Atoi(sv)
	}

	return ids
}

func ToMinMax(minMaxStr string) MinMax {
	if types.IsNumber(minMaxStr) {
		return MinMax{Min: types.ToFloat64E(minMaxStr), Max: types.ToFloat64E(minMaxStr)}
	}

	minMax := []float64{}
	err := json.Unmarshal([]byte(minMaxStr), &minMax)
	if err == nil {
		return MinMax{Min: minMax[0], Max: minMax[1]}
	}

	ret := MinMax{}
	_ = json.Unmarshal([]byte(minMaxStr), &ret)
	return ret
}

// ToPosHeros 解析JSON：[[英雄id,位置], ...] => [英雄id, ...], {英雄id: 位置, ...}
func ToPosHeros(posHeros interface{}) (heroIDs []int, heroPos map[int]int) {
	var parsed [][]int

	switch expr := posHeros.(type) {
	case string:
		json.Unmarshal([]byte(expr), &parsed)
	case []byte:
		json.Unmarshal(expr, &parsed)
	case [][]int:
		parsed = expr
	default:
		posHeroArr, _ := types.ToArray(posHeros)
		if len(posHeroArr) > 0 {
			parsed = make([][]int, len(posHeroArr))
			for k, posHero := range posHeroArr {
				tmps, _ := types.ToArray(posHero)
				for _, tmp := range tmps {
					parsed[k] = append(parsed[k], types.ToIntE(tmp))
				}
			}
		}

	}

	heroIDs = []int{}
	heroPos = map[int]int{}

	for _, h := range parsed {
		if len(h) >= 2 {
			id := h[0]
			pos := h[1]
			heroIDs = append(heroIDs, id)
			heroPos[id] = pos
		}
	}

	return
}

// ToHeroFus 解析JSON：{"left": {...}, "right": {...}} => {left: {...}, right: {...}}
func ToHeroFus(fus interface{}, lv, star int) map[string]*Fu {
	switch v := fus.(type) {
	case map[string]*Fu:
		if v == nil {
			return map[string]*Fu{}
		}
		return v
	case map[string]interface{}:
		ret := map[string]*Fu{
			"left":  mapToHeroFu(v, "left", lv, star),
			"right": mapToHeroFu(v, "right", lv, star),
		}
		return ret
	case types.Map:
		ret := map[string]*Fu{
			"left":  mapToHeroFu(v, "left", lv, star),
			"right": mapToHeroFu(v, "right", lv, star),
		}
		return ret
	case string:
		tmp := types.Map{}
		json.Unmarshal([]byte(v), &tmp)
		if tmp == nil {
			return map[string]*Fu{}
		}
		ret := map[string]*Fu{
			"left":  mapToHeroFu(tmp, "left", lv, star),
			"right": mapToHeroFu(tmp, "right", lv, star),
		}
		return ret
	case []byte:
		tmp := types.Map{}
		json.Unmarshal(v, &tmp)
		if tmp == nil {
			return map[string]*Fu{}
		}
		ret := map[string]*Fu{
			"left":  mapToHeroFu(tmp, "left", lv, star),
			"right": mapToHeroFu(tmp, "right", lv, star),
		}
		return ret
	}

	return map[string]*Fu{}
}

func (h *Fu) Clone() *Fu {
	tmp := &Fu{
		Id:     h.Id,
		ItemId: h.ItemId,
		Unlock: h.Unlock,
		Props:  make([]FuProp, len(h.Props)),
	}

	for k, v := range h.Props {
		tmp.Props[k] = v
	}
	return tmp
}

// Merge 合并两组道具列表
func Merge(items1, items2 []TypeNum) []TypeNum {
	if len(items1) == 0 {
		return items2
	}
	if len(items2) == 0 {
		return items1
	}

	result := []TypeNum{}

	merged := make(map[int]int)
	for _, v := range items1 {
		if len(v.Prop) > 0 {
			result = append(result, v)
		} else {
			merged[v.Type] += v.Num
		}
	}
	for _, v := range items2 {
		if len(v.Prop) > 0 {
			result = append(result, v)
		} else {
			merged[v.Type] += v.Num
		}
	}

	for k, v := range merged {
		result = append(result, TypeNum{Type: k, Num: v})
	}
	return result
}

func mapToHeroFu(heroFus types.Map, typ string, lv, star int) *Fu {
	if heroFus == nil || len(heroFus) == 0 || heroFus[typ] == nil || len(types.ToMapE(heroFus[typ])) == 0 {
		var unlock int8

		if typ == "left" {
			if lv >= 100 {
				unlock = 1
			}
		} else {
			if lv >= 100 && star >= 7 {
				unlock = 1
			}
		}

		return &Fu{Unlock: unlock}
	}

	heroFu := types.ToMapE(heroFus[typ])

	var unlock int8
	if typ == "left" {
		if lv >= 100 {
			unlock = 1
		}
	} else {
		if lv >= 100 && star >= 7 {
			unlock = 1
		}
	}

	ret := Fu{
		Unlock: unlock,
		ItemId: types.ToIntE(heroFu["item_id"]),
		Id:     types.ToInt64E(heroFu["id"]),
		Props:  make([]FuProp, 0),
	}

	props := heroFu.GetMapArrayE("prop")
	if len(props) > 0 {
		ret.Props = mapsToProp(props)
	}
	return &ret
}

func mapsToProp(props []types.Map) []FuProp {
	ret := []FuProp{}
	for _, v := range props {
		ret = append(ret, FuProp{
			MaxNum: types.ToIntE(v["max_num"]),
			Num:    types.ToIntE(v["num"]),
			Prop:   types.ToIntE(v["prop"]),
			Type:   types.ToIntE(v["type"]),
		})
	}
	return ret
}
