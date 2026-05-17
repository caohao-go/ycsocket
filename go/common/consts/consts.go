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

package consts

const (
	FALSE = 0
	TRUE  = 1
)

const ( // 查询类型
	QueryTypeApp = 1 // app
	QueryTypeWeb = 2 // web
)

const (
	QueryModeSingle   = 0 //单个执行单元
	QueryModeParallel = 1 //多个执行单元并发
	QueryModeCompound = 2 //复合查询
)

const (
	NoCompression = 0
	Compression   = 1
)

const ( //request type
	RequestTypeRPC  = 0
	RequestTypeHTTP = 1
	RequestTypeWeb  = 2
)

const (
	StatusOnline  = 1 // 正常
	StatusOffline = 2 // 下线
)
