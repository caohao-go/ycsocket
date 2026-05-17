// Copyright (c) 2024 The horm-database Authors. All rights reserved.
// This file Author:  CaoHao <18500482693@163.com> .
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package util

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"server_golang/common/consts"
	"server_golang/common/types"
)

// Alias 别名
func Alias(src string) (name, alias string) {
	start := strings.Index(src, "(")

	if start != -1 {
		end := strings.Index(src, ")")
		if end != -1 {
			return strings.TrimSpace(src[:start]), strings.TrimSpace(src[start+1 : end])
		}
	}

	return strings.TrimSpace(src), ""
}

// Namespace 命名空间
func Namespace(src string) (namespace, name string) {
	start := strings.Index(src, "::")

	if start != -1 {
		end := strings.Index(src, "::")
		if end != -1 {
			return strings.TrimSpace(src[:start]), strings.TrimSpace(src[start+2 : end])
		}
	}

	return "", strings.TrimSpace(src)
}

// RemoveComments 去掉注释
func RemoveComments(key string) string {
	index := strings.Index(key, "#")
	if index == -1 {
		return strings.TrimSpace(key)
	}

	return strings.TrimSpace(key[0:index])
}

// IsRelation 是否关系连接词
func IsRelation(dbType consts.DBType, key string) int8 {
	switch key { //为了性能，先不做 ToUpper，先大小都做比较，如果匹配，再 ToUpper
	case "AND", "OR", "NOT",
		"and", "or", "not":
		return consts.AndOrNot
	}

	switch dbType {
	case consts.DBTypeElastic:
		switch key {
		case "NESTED", "PARENT_ID", "HAS_CHILD", "HAS_PARENT",
			"nested", "parent_id", "has_child", "has_parent":
			return consts.EsKeyword
		}
	}

	return consts.NotOP
}

// GetRelation 获取关系连接词
func GetRelation(dbType consts.DBType, key string, v reflect.Value) (bool, bool, string) {
	nk := RemoveComments(key)

	k := v.Kind()

	if k == reflect.Map || k == reflect.Array || k == reflect.Slice {
		isRelation := IsRelation(dbType, nk)
		if k == reflect.Map && isRelation != consts.NotOP {
			return true, false, strings.ToUpper(nk)
		} else if (k == reflect.Array || k == reflect.Slice) && isRelation == consts.AndOrNot {
			if v.Len() > 0 {
				v0 := v.Index(0)
				if !v0.IsNil() && types.IsMap(v0) {
					return true, true, strings.ToUpper(nk)
				}
			}
		}
	}

	return false, false, ""
}

// OperatorMatch 匹配操作符及操作属性
func OperatorMatch(key string, isElastic bool) (column, operator,
	opAttr, minShouldMatch string, boost float64, slop int) {
	key = RemoveComments(key)

	if key == "" {
		return
	}

	if key[0] == '~' { //函数
		operator = "FUNC"
		column = key[1:]
		return
	}

	if isElastic {
		tmp, attrStr := Alias(key)
		if attrStr != "" {
			attrs := strings.Split(attrStr, ",")
			for _, attr := range attrs {
				kv := strings.Split(attr, "=")
				if len(kv) == 1 {
					opAttr = strings.TrimSpace(attr)
				} else if len(kv) == 2 {
					switch kv[0] {
					case "boost":
						boost, _ = strconv.ParseFloat(strings.TrimSpace(kv[1]), 64)
					case "slop":
						slop, _ = strconv.Atoi(strings.TrimSpace(kv[1]))
					case "minimum_should_match":
						minShouldMatch = strings.TrimSpace(kv[1])
					}
				}
			}

			key = strings.TrimSpace(tmp)
		}
	}

	lastThreeWord := types.LastWord(key, 3)
	if lastThreeWord == consts.OPNotBetween {
		operator = lastThreeWord
		column = strings.TrimSpace(types.CutLast(key, 3))
		return
	}

	lastTwoWord := types.LastWord(key, 2)
	switch lastTwoWord {
	case consts.OPBetween, consts.OPGte, consts.OPLte, consts.OPNotLike, consts.OPNotMatchPhrase, consts.OPNotMatch:
		operator = lastTwoWord
		column = strings.TrimSpace(types.CutLast(key, 2))
		return
	}

	lastWord := types.LastWord(key, 1)
	switch lastWord {
	case consts.OPEqual, consts.OPGt, consts.OPLt, consts.OPNot, consts.OPLike, consts.OPMatchPhrase, consts.OPMatch:
		operator = lastWord
		column = strings.TrimSpace(types.CutLast(key, 1))
		return
	}

	column = key
	return
}

// CalcTotalPage 根据数据总数和每页大小计算总页数。
func CalcTotalPage(total uint64, size int) uint32 {
	if total%uint64(size) == 0 {
		return uint32(total / uint64(size))
	} else {
		return uint32(total/uint64(size)) + 1
	}
}

// Order 排序
type Order struct {
	Field     string
	Ascending bool
}

// FormatOrders 格式化排序
func FormatOrders(orders []string) []*Order {
	var ret = []*Order{}

	for _, order := range orders {
		var field = strings.TrimSpace(order)
		if field == "" {
			continue
		}

		var ascending = true

		lastWord := types.LastWord(field, 5)
		if lastWord == " desc" || lastWord == " DESC" {
			ascending = false
			field = strings.TrimSpace(field[:len(field)-4])
		} else {
			lastWord = types.LastWord(field, 4)
			if lastWord == " asc" || lastWord == " ASC" {
				field = strings.TrimSpace(field[:len(field)-3])
			}
		}

		switch field[0] {
		case '+':
			field = field[1:]
		case '-':
			field = field[1:]
			ascending = false
		}

		ret = append(ret, &Order{strings.TrimSpace(field), ascending})
	}

	return ret
}

// FormatArgs change args bytes to string
func FormatArgs(args []interface{}) []interface{} {
	if types.HasBytes(args) {
		var newArgs = make([]interface{}, len(args))

		for k, arg := range args {
			switch v := arg.(type) {
			case []byte:
				newArgs[k] = types.BytesToString(v)
			case *[]byte:
				newArgs[k] = types.BytesToString(*v)
			default:
				newArgs[k] = arg
			}
		}

		return newArgs
	}

	return args
}

// ArgRefererEscape arg 参数如果包含引用符，可以用该函数转义。
func ArgRefererEscape(arg string) string {
	tmp := strings.TrimSpace(arg)

	l := len(tmp)
	if l <= 4 {
		return arg
	}

	if tmp[:2] == "@{" && tmp[l-1:] == "}" {
		return "\\" + tmp
	}

	return arg
}

// ArgRefererUnEscape arg 参数如果包含引用符，可以用该函数反转义。
func ArgRefererUnEscape(arg string) string {
	tmp := strings.TrimSpace(arg)

	l := len(tmp)
	if l <= 5 {
		return arg
	}

	if tmp[:3] == "\\@{" && tmp[l-1:] == "}" {
		return tmp[1:]
	}

	return arg
}

// MatchTable 匹配表
func MatchTable(table, verifyRule string) bool {
	rules := strings.Split(verifyRule, ",")
	for _, rule := range rules {
		if rule == table {
			return true
		}

		re := regexp.MustCompile(`(\d+)(\.\.\.)(\d+)`)
		matchs := re.FindAllStringSubmatch(rule, -1)
		if matchs != nil && len(matchs) == 1 && len(matchs[0]) == 4 {
			ruleArr := strings.Split(rule, matchs[0][0])
			startIndex, err := strconv.Atoi(matchs[0][1])
			if err != nil {
				continue
			}
			endIndex, err := strconv.Atoi(matchs[0][3])
			if err != nil {
				continue
			}
			if endIndex < startIndex {
				continue
			}
			for i := startIndex; i <= endIndex; i++ {
				ruleStr := ruleArr[0] + fmt.Sprint(i) + ruleArr[1]
				if table == ruleStr {
					return true
				}
			}
		}

		if strings.HasPrefix(rule, "regex/") && strings.HasSuffix(rule, "/") {
			ruleStr := rule[6 : len(rule)-1]
			if regexp.MustCompile(ruleStr).MatchString(table) {
				return true
			}
		}
	}

	return false
}
