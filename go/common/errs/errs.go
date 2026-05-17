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

// Package errs provides error code type, which contains errcode errmsg.
// These definitions are multi-language universal.
package errs

import (
	"fmt"
	"io"
)

const (
	Success    = 0 // 成功
	ErrSystem  = 1 // 服务端系统异常
	ErrPanic   = 8 // panic
	ErrUnknown = 9 // 未知错误

	// 客户端错误
	ErrClientReadFrame = 11 // 客户端帧读取失败
	ErrClientTimeout   = 12 // 请求在客户端调用超时
	ErrClientConnect   = 13 // 客户端连接错误
	ErrClientEncode    = 14 // 客户端编码错误
	ErrClientDecode    = 15 // 客户端解码错误
	ErrClientRoute     = 16 // 客户端选ip路由错误
	ErrClientNet       = 17 // 客户端网络错误
	ErrClientCanceled  = 18 // 上游调用方提前取消请求
	ErrClientNotInit   = 19 // server 未被初始化

	// 请求参数错误
	ErrReqParamInvalid   = 50 // 请求参数非法
	ErrReqUnitNameEmpty  = 51 // 执行单元名不能为空
	ErrRequestIDNotMatch = 52 // 请求与返回 id 不匹配

	// 网关错误网关错误
	ErrAppCallLimited = 60 // 限制时间内调用失败次数 网关错误

	// 服务器错误
	ErrServerReadFrame  = 101 // 服务端帧读取失败
	ErrServerDecompress = 102 // 服务端解压错误
	ErrServerDecode     = 103 // 服务端解码错误
	ErrServerEncode     = 104 // 服务端编码错误
	ErrServerNoService  = 105 // 服务端没有调用相应的service实现
	ErrServerNoFunc     = 106 // 服务端没有调用相应的接口实现
	ErrServerTimeout    = 107 // 请求在服务端队列超时
	ErrServerOverload   = 108 // 请求在服务端过载

	// 参数错误
	ErrParamInvalid    = 301 // 请求参数非法
	ErrParamEmpty      = 302 // 请求参数不得为空
	ErrParamMiss       = 303 // 请求参数未上传
	ErrParamType       = 304 // 请求参数类型错误
	ErrParamValue      = 305 // 请求参数取值错误
	ErrNotFindName     = 310 // 未找到 name 对应表配置
	ErrUnitNameEmpty   = 312 // 执行单元名不能为空
	ErrRepeatNameAlias = 313 // 在同一层级有重复的 name 或 alias
	ErrFormatData      = 314 // 数据格式化失败
	ErrSameTransaction = 315 // 事务重复定义

	ErrRefererNotFound      = 320 // 未找到被引用的执行单元
	ErrRefererMustBeString  = 321 // 引用必须是 string
	ErrRefererUnitFailed    = 322 // 被引用的执行单元查询失败
	ErrRefererUnitNotExec   = 323 // 被引用的执行单元未执行
	ErrRefererResultType    = 324 // 被引用的执行单元结果类型不符
	ErrRefererFieldNotExist = 325 // 被引用的执行单元结果中不包含引用字段

	// 权限错误
	ErrAuthFail        = 401 // 鉴权失败
	ErrHasNoTableRight = 402 // 无该表访问权限
	ErrHasNoDBRight    = 403 // 无数据库访问权限
	ErrAppidNotFound   = 404 // 未找到 appid
	ErrTableVerify     = 405 // 表校验失败

	// 数据库错误
	ErrQueryNotImp = 501 // 未找到数据库的 Query 实现
	ErrTransaction = 502 // 事务错误
	ErrDBParams    = 503 // database request params error

	ErrSQLQuery     = 510 // mysql/postgresql/clickhouse query error
	ErrAffectResult = 512 // 获取影响行数信息失败

	ErrClickhouseInsert = 530 // clickhouse insert error
	ErrClickhouseCreate = 530 // clickhouse create error

	ErrElasticQuery = 550 // new elastic client error

	ErrRedisDo       = 570 //redis do error
	ErrRedisReqParse = 571 //redis 请求解析 失败
	ErrRedisDecode   = 572 //redis 结果解码 失败

	// 插件错误
	ErrPluginConfigDecode    = 601 // 插件配置解压失败
	ErrPluginNotFound        = 602 // 未找到插件
	ErrPluginFuncNotRegister = 603 // 插件函数未注册
	ErrPluginExec            = 604 // 插件执行异常
	ErrPrefixPluginNotFount  = 605 // 插件先决执行插件未找到

	// 结果编排错误
	ErrOrchFailed      = 701 // 编排失败
	ErrOrchNameEmpty   = 702 // 编排名称为空
	ErrOrchNotFind     = 703 // 未找到编排函数
	ErrOrchPathInvalid = 704 // 编排路径解析失败

	// 其他错误
	ErrOpNotSupport  = 801 // 该数据库不支持该操作
	ErrNameAmbiguity = 802 // 表有歧义，需要加 namespace
	ErrConfig        = 803 // 配置异常

	ErrDBConfigNotFound = 851 // 未找到数据库配置
	ErrDBTypeInvalid    = 852 // 数据库类型错误
	ErrDBAddressParse   = 853 // 数据库地址解析错误
)

// EType 错误类型
type EType int8

const (
	ETypeSystem   EType = 0 //系统错误
	ETypePlugin   EType = 1 //插件错误
	ETypeDatabase EType = 2 //数据库错误
)

func typeDesc(t EType) string {
	switch t {
	case ETypePlugin:
		return "plugin"
	case ETypeDatabase:
		return "database"
	default:
		return "system"
	}
}

// New 创建一个系统错误
func New(code int, msg string) error {
	return &Error{
		Type: ETypeSystem,
		Code: code,
		Msg:  msg,
	}
}

// Newf 创建一个格式化系统错误
func Newf(code int, format string, params ...interface{}) error {
	return &Error{
		Type: ETypeSystem,
		Code: code,
		Msg:  fmt.Sprintf(format, params...),
	}
}

// NewDB 创建一个数据库错误
func NewDB(code int, msg string) error {
	return &Error{
		Type: ETypeDatabase,
		Code: code,
		Msg:  msg,
	}
}

// NewDBf 创建一个格式化数据库错误
func NewDBf(code int, format string, params ...interface{}) error {
	return &Error{
		Type: ETypeDatabase,
		Code: code,
		Msg:  fmt.Sprintf(format, params...),
	}
}

// NewPlugin 创建一个插件错误
func NewPlugin(code int, msg string) error {
	return &Error{
		Type: ETypePlugin,
		Code: code,
		Msg:  msg,
	}
}

// NewPluginf 创建一个格式化插件错误
func NewPluginf(code int, format string, params ...interface{}) error {
	return &Error{
		Type: ETypePlugin,
		Code: code,
		Msg:  fmt.Sprintf(format, params...),
	}
}

// Type 获取错误类型
func Type(e error) EType {
	if e == nil {
		return ETypeSystem
	}

	err, ok := e.(*Error)
	if !ok {
		return ETypeSystem
	}

	return err.Type
}

// Code 获取错误码
func Code(e error) int {
	if e == nil {
		return 0
	}

	err, ok := e.(*Error)
	if !ok {
		return ErrUnknown
	}

	return err.Code
}

// Msg 获取错误信息
func Msg(e error) string {
	if e == nil {
		return "success"
	}

	err, ok := e.(*Error)
	if !ok {
		return e.Error()
	}

	return err.Msg
}

// Sql 获取错误语句
func Sql(e error) string {
	if e == nil {
		return ""
	}

	err, ok := e.(*Error)
	if !ok {
		return ""
	}

	return err.Sql
}

// SetErrorType 设置 error type
func SetErrorType(err error, typ EType) error {
	if err == nil {
		return nil
	}

	e, ok := err.(*Error)
	if !ok {
		return &Error{
			Type: typ,
			Code: ErrUnknown,
			Msg:  err.Error(),
		}
	}

	e.Type = typ
	return err
}

// SetErrorCode 设置 error code
func SetErrorCode(err error, code int) error {
	if err == nil {
		return nil
	}

	e, ok := err.(*Error)
	if !ok {
		return &Error{
			Type: ETypeSystem,
			Code: code,
			Msg:  err.Error(),
		}
	}

	e.Code = code
	return e
}

// SetErrorMsg 设置错误消息
func SetErrorMsg(err error, msg string) error {
	if err == nil {
		return nil
	}

	e, ok := err.(*Error)
	if !ok {
		return &Error{
			Type: ETypeSystem,
			Code: ErrUnknown,
			Msg:  msg,
		}
	}

	e.Msg = msg
	return e
}

// SetErrorSql 设置错误语句
func SetErrorSql(err error, sql string) error {
	if err == nil {
		return nil
	}

	e, ok := err.(*Error)
	if !ok {
		return &Error{
			Type: ETypeDatabase,
			Code: ErrUnknown,
			Msg:  err.Error(),
			Sql:  sql,
		}
	}

	e.Sql = sql
	return e
}

// Error error 结构体
type Error struct {
	Type EType
	Code int
	Msg  string
	Sql  string //发生 error 时的 sql 语句
}

// Error error 信息
func (e *Error) Error() string {
	if e == nil {
		return "success"
	}

	if e.Sql != "" {
		return fmt.Sprintf("type:%s, code:%d, msg:%s, sql=[%s]", typeDesc(e.Type), e.Code, e.Msg, e.Sql)
	} else {
		return fmt.Sprintf("type:%s, code:%d, msg:%s", typeDesc(e.Type), e.Code, e.Msg)
	}
}

// Format 实现 fmt.Formatter 接口
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		if s.Flag('+') {
			if e.Msg != "" {
				var msg string
				if e.Sql != "" {
					msg = fmt.Sprintf("type:%s, code:%d, msg:%s, sql=[%s]", typeDesc(e.Type), e.Code, e.Msg, e.Sql)
				} else {
					msg = fmt.Sprintf("type:%s, code:%d, msg:%s", typeDesc(e.Type), e.Code, e.Msg)
				}
				_, _ = io.WriteString(s, msg)
			}
			return
		}
		fallthrough
	case 's':
		_, _ = io.WriteString(s, e.Error())
	case 'q':
		_, _ = fmt.Fprintf(s, "%q", e.Error())
	default:
		_, _ = fmt.Fprintf(s, "%%!%c(errs.Error=%s)", verb, e.Error())
	}
}
