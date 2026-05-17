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

type Type uint8 // orm type

const ( /* 我们发送请求到数据统一调度服务的时候，绝大多数情况下可以不指定数据类型，服务端也可以正常解析并执行 query 语句，但是在某些
	特殊情况下，比如 clickhouse 对类型有强限制，又或者字段是一个超大 uint64 整数，json 编码之后请求服务端，由于 json 的基础类型只包含
	string、number(当成float64)、bool，数字在服务端会被解析为 float64，存在精度丢失问题，当类型为 time、[]byte、int、 int8~int64、
	uint、uint8~uint64 时，需要在执行单元 data_type 字段里将数据类型带上。*/
	TypeTime   Type = 1 // 类型是 time.Time
	TypeBytes  Type = 2 // 类型是 []byte
	TypeFloat  Type = 3
	TypeDouble Type = 4
	TypeInt    Type = 5
	TypeUint   Type = 6
	TypeInt8   Type = 7
	TypeInt16  Type = 8
	TypeInt32  Type = 9
	TypeInt64  Type = 10
	TypeUint8  Type = 11
	TypeUint16 Type = 12
	TypeUint32 Type = 13
	TypeUint64 Type = 14
	TypeString Type = 15
	TypeBool   Type = 16
	TypeJSON   Type = 17
)

var OrmType = map[string]Type{
	"time":   TypeTime,
	"bytes":  TypeBytes,
	"float":  TypeFloat,
	"double": TypeDouble,
	"int":    TypeInt,
	"uint":   TypeUint,
	"int8":   TypeInt8,
	"int16":  TypeInt16,
	"int32":  TypeInt32,
	"int64":  TypeInt64,
	"uint8":  TypeUint8,
	"uint16": TypeUint16,
	"uint32": TypeUint32,
	"uint64": TypeUint64,
	"string": TypeString,
	"bool":   TypeBool,
	"json":   TypeJSON,
}
