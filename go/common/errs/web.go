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

package errs

const ( // web 错误
	RetWebNotLogin     = 1000 // 未登录
	RetWebNotFindUser  = 1001 // 未找到用户
	RetWebLoginExpired = 1002 // 登录已过期，请重新登录
	RetWebParamEmpty   = 1003 // 参数不能为空
	RetWebInitTenant   = 1004 // 初始化 tenant 失败

	RetWebEmailSendFailed     = 1010 // 邮件发送失败
	RetWebEmailSendFrequently = 1011 // 邮件发送太频繁
	RetWebCodeIncorrectly     = 1012 // 验证码校验错误

	RetWebAccountExists     = 1030 // 账号已注册
	RetWebAccountNotExists  = 1031 // 账号未注册
	RetWebPasswordIncorrect = 1032 // 密码错误

	RetWebTenantEmpty         = 1040 // 租户为空
	RetWebTenantNotExists     = 1041 // 租户不存在
	RetWebTenantExists        = 1042 // 租户已存在
	RetWebNotIndicateTenant   = 1043 // 当前租户不是指定租户
	RetWebNotTenantMember     = 1044 // 不是租户成员
	RetWebTenantMemberExpired = 1045 // 租户成员权限已过期
	RetWebTenantNotMatch      = 1046 // 租户不匹配

	RetWebMemberNotManager       = 1061 // 无管理员权限
	RetWebIsMember               = 1062 // 已经是成员
	RetWebIsNotApply             = 1063 // 未申请权限
	RetWebIsNotMember            = 1064 // 不是成员
	RetWebMemberExpired          = 1065 // 成员权限已过期
	RetWebMemberNotUnderApproval = 1066 // 并非审批状态
	RetWebMemberUnderApproval    = 1067 // 正在审批中

	RetWebCantCreateDB = 1071 // 无数据库创建权限
	RetWebNotFindDB    = 1072 // 未找到 db
	RetWebNotDBManager = 1073 // 非库管理员

	RetWebNotFindTable    = 1081 // 未找到表
	RetWebCantCreateTable = 1082 // 无建表权限

	RetWebNotFindPlugin = 1091 // 未找到插件
	RetWebIsFirstPlugin = 1092 // 这是第一个插件，front 必须为0

	RetWebNotFindTablePlugin = 1101

	RetWebNotFindApp = 1201

	RetWebNotFindAccessInfo       = 1301
	RetWebAccessStatusNormal      = 1302
	RetWebAccessStatusChecking    = 1303
	RetWebAccessStatusNotChecking = 1310
	RetWebAccessPermissionDeny    = 1321
)
