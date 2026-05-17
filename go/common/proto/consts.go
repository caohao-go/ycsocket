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

const ( // http 头部
	HeaderTenant       = "head-tenant"     // 租户id
	HeaderVersion      = "head-version"    // 客户端版本
	HeaderRequestID    = "head-request-id" // 请求唯一id
	HeaderTraceID      = "head-trace-id"   // trace-id
	HeaderTimestamp    = "head-timestamp"  // 请求时间戳（精确到毫秒）
	HeaderTimeout      = "head-timeout"    // 请求超时时间，单位ms
	HeaderCaller       = "head-caller"     // 主调服务的名称 app.server.service
	HeaderAppid        = "head-appid"      // appid
	HeaderAuthRand     = "head-auth-rand"  // 随机生成 0-9999999 的数字，相同 timestamp 不允许出现同样的 ip、auth_rand。为了避免碰撞，0-9999999，单机理论最大支持 100 亿/秒的并发。
	HeaderSign         = "head-sign"       // sign 签名，为 md5(appid+secret+version+request_type+query_mode+request_id+trace_id+timestamp+timeout+caller+compress+ip+auth_rand)
	HeaderIsNil        = "head-is-nil"     // 返回是否为空（针对单执行单元）
	HeaderErrorType    = "head-error-type" // 错误类型
	HeaderErrorCode    = "head-error-code" // 错误码
	HeaderErrorMessage = "head-error-msg"  // 错误消息
)

const ( // web 头部
	WebVersion   = "web-version"    // 客户端版本
	WebRequestID = "web-request-id" // 请求唯一id
	WebTenant    = "web-tenant-id"  // 租户id
	WebHormToken = "horm-token"     // 平台token
)
