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

package proto

import (
	"strconv"

	"server_golang/common/errs"
)

// PageResult 当 page >= 1 时会返回分页结果
type PageResult struct {
	Detail *Detail       `orm:"detail,omitempty" json:"detail,omitempty"` // 查询细节信息
	Data   []interface{} `orm:"data,omitempty" json:"data,omitempty"`     // 分页结果
}

// ParallelResult 并行查询返回结果
type ParallelResult struct {
	Detail map[string]*Detail     `json:"detail"` // 返回详细信息
	Data   map[string]interface{} `json:"data"`   // 返回数据
}

// CompResult 混合查询返回结果
type CompResult struct {
	Detail *Detail     `json:"detail"` // 返回详细信息
	Data   interface{} `json:"data"`   // 返回数据
}

// Detail 其他查询细节信息，例如 分页信息、滚动翻页信息、其他信息等。
type Detail struct {
	Total     uint64                 `orm:"total,omitempty" json:"total"`             // 总数
	TotalPage uint32                 `orm:"total_page,omitempty" json:"total_page"`   // 总页数
	Page      int                    `orm:"page,omitempty" json:"page"`               // 当前分页
	Size      int                    `orm:"size,omitempty" json:"size"`               // 每页大小
	Scroll    *Scroll                `orm:"scroll,omitempty" json:"scroll,omitempty"` // 滚动翻页信息
	Extras    map[string]interface{} `orm:"extras,omitempty" json:"extras,omitempty"` // 更多详细信息

	// for parallel、complex
	IsNil bool   `json:"is_nil,omitempty"` // 是否为空（针对并行/混合执行）
	Error *Error `json:"error,omitempty"`  // 错误返回（针对并行/混合执行）
}

// ModRet 新增/更新返回信息
type ModRet struct {
	ID          ID                     `orm:"id,omitempty" json:"id,omitempty"`                       // id 主键，可能是 mysql 的最后自增id，last_insert_id 或 elastic 的 _id 等，类型可能是 int64、string
	RowAffected int64                  `orm:"rows_affected,omitempty" json:"rows_affected,omitempty"` // 影响行数
	Version     int64                  `orm:"version,omitempty" json:"version,omitempty"`             // 数据版本
	Status      int                    `orm:"status,omitempty" json:"status,omitempty"`               // 返回状态码
	Reason      string                 `orm:"reason,omitempty" json:"reason,omitempty"`               // mod 失败原因
	Extras      map[string]interface{} `orm:"extras,omitempty" json:"extras,omitempty"`               // 更多详细信息
}

// MemberScore redis 集合成员及其分数信息。
type MemberScore struct {
	Member []string  `orm:"member,omitempty" json:"member,omitempty"`
	Score  []float64 `orm:"score,omitempty" json:"score,omitempty"`
}

type ID string

func (id ID) String() string {
	return string(id)
}

func (id ID) Float64() float64 {
	f, _ := strconv.ParseFloat(string(id), 64)
	return f
}

func (id ID) Int() int {
	return int(id.Int64())
}

func (id ID) Int64() int64 {
	i, _ := strconv.ParseInt(string(id), 10, 64)
	return i
}

func (id ID) Uint() uint {
	return uint(id.Uint64())
}

func (id ID) Uint64() uint64 {
	ui, _ := strconv.ParseUint(string(id), 10, 64)
	return ui
}

func (x *Error) ToError() *errs.Error {
	return &errs.Error{
		Type: errs.EType(x.Type),
		Code: int(x.Code),
		Msg:  x.Msg,
		Sql:  x.Sql,
	}
}
