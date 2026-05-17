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

package types

import (
	"fmt"
	"strconv"
	"strings"
)

func SplitInt(s, sep string) []int {
	s = strings.TrimSpace(s)
	if s == "" {
		return []int{}
	}

	tmp := strings.Split(s, sep)
	ret := make([]int, len(tmp))
	for k, v := range tmp {
		ret[k], _ = strconv.Atoi(strings.TrimSpace(v))
	}

	return ret
}

func SplitInt8(s, sep string) []int8 {
	s = strings.TrimSpace(s)
	if s == "" {
		return []int8{}
	}

	tmp := strings.Split(s, sep)
	ret := make([]int8, len(tmp))
	for k, v := range tmp {
		i, _ := strconv.Atoi(strings.TrimSpace(v))
		ret[k] = int8(i)
	}

	return ret
}

func SplitInt64(s, sep string) []int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return []int64{}
	}

	tmp := strings.Split(s, sep)
	ret := make([]int64, len(tmp))
	for k, v := range tmp {
		ret[k], _ = strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	}

	return ret
}

func SplitUint64(s, sep string) []uint64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return []uint64{}
	}

	tmp := strings.Split(s, sep)
	ret := make([]uint64, len(tmp))
	for k, v := range tmp {
		ret[k], _ = strconv.ParseUint(strings.TrimSpace(v), 10, 64)
	}

	return ret
}

func JoinUint64(s []uint64, sep string) string {
	if len(s) == 0 {
		return ""
	}

	var ret strings.Builder

	for k, v := range s {
		if k > 0 {
			ret.WriteString(sep)
		}

		ret.WriteString(strings.TrimSpace(fmt.Sprint(v)))
	}

	return ret.String()
}

func JoinInt64(s []int64, sep string) string {
	if len(s) == 0 {
		return ""
	}

	var ret strings.Builder

	for k, v := range s {
		if k > 0 {
			ret.WriteString(sep)
		}

		ret.WriteString(strings.TrimSpace(fmt.Sprint(v)))
	}

	return ret.String()
}

func JoinInt8(s []int8, sep string) string {
	if len(s) == 0 {
		return ""
	}

	var ret strings.Builder

	for k, v := range s {
		if k > 0 {
			ret.WriteString(sep)
		}

		ret.WriteString(strings.TrimSpace(fmt.Sprint(v)))
	}

	return ret.String()
}

// CutString 以 sep 作为分隔符切割字符串
func CutString(s, sep string) (found bool, s1, s2 string) {
	if s == sep {
		return true, "", ""
	}

	i := strings.Index(s, sep)

	switch i {
	case -1:
		found = false
		s2 = s
	default:
		found = true
		s1 = s[:i]
		s2 = s[i+len(sep):]
	}

	return
}

// FirstWord 开始 n 个字符
func FirstWord(key string, n int) string {
	if key == "" || n <= 0 {
		return ""
	}

	if n >= len(key) {
		return key
	}

	return key[0:n]
}

// CutLast 移除最后 n 个字符
func CutLast(key string, n int) string {
	if len(key) <= n {
		return ""
	}

	return key[0 : len(key)-n]
}

// LastWord 末尾 n 个字符
func LastWord(key string, n int) string {
	if key == "" || n <= 0 {
		return ""
	}

	l := len(key)

	if n >= l {
		return key
	}

	return key[l-n:]
}

var replacerLFCRToQuote = strings.NewReplacer("\n", "\\n", "\r", "\\r")

// QuickReplaceLFCR 替换 \r(回车)、\n(换行) 为字符串 `\r`、`\n`
func QuickReplaceLFCR(str string) string {
	return replacerLFCRToQuote.Replace(str)
}

var replacerLFCRToSpace = strings.NewReplacer("\r\n", " ", "\n", " ", "\r", " ")

// QuickReplaceLFCR2Space 替换 \r(回车)、\n(换行) 为空格
func QuickReplaceLFCR2Space(b []byte) string {
	return replacerLFCRToSpace.Replace(string(b))
}

var replacerLFCRToEmpty = strings.NewReplacer("\n", "", "\r", "")

// QuickRemoveLFCR 去掉 \r(回车)、\n(换行)
func QuickRemoveLFCR(b []byte) string {
	return replacerLFCRToEmpty.Replace(string(b))
}

type Str string

func (v Str) String() string {
	return string(v)
}

func (v Str) Int() int {
	return ToIntE(string(v))
}

func (v Str) Int64() int64 {
	return ToInt64E(string(v))
}

func (v Str) Uint() uint {
	uiv, _ := ToUint(string(v))
	return uiv
}

func (v Str) Uint64() uint64 {
	uiv, _ := ToUint64(string(v))
	return uiv
}

func (v Str) Float() float64 {
	fiv, _ := ToFloat64(string(v))
	return fiv
}

func ToStr(value interface{}) Str {
	return Str(ToString(value))
}
