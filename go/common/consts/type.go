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

import (
	"reflect"

	"server_golang/common/types"
)

type RetType uint8

const (
	RedisRetTypeNil         RetType = 1 // 无返回
	RedisRetTypeString      RetType = 2 // 字符串
	RedisRetTypeBool        RetType = 3 // bool
	RedisRetTypeInt64       RetType = 4 // int64
	RedisRetTypeFloat64     RetType = 5 // float64
	RedisRetTypeStrings     RetType = 6 // 字符串数组
	RedisRetTypeMapString   RetType = 7 // map[string]string
	RedisRetTypeMemberScore RetType = 8 // 有序成员
)

func GetRedisRetType(op string, withScores, countExists bool) RetType {
	switch op {
	case OpGet, OpGetSet, OpHGet:
		return RedisRetTypeString
	case OpLPop, OpRPop, OpSRandMember, OpSPop:
		if countExists {
			return RedisRetTypeStrings
		} else {
			return RedisRetTypeString
		}
	case OpSet, OpExists, OpSetNX, OpSetBit, OpGetBit, OpHSetNx, OpHExists, OpSIsMember:
		return RedisRetTypeBool
	case OpTTL, OpDel, OpIncr, OpDecr, OpIncrBy, OpBitCount, OpHIncrBy, OpHDel, OpHLen,
		OpHStrLen, OpLPush, OpRPush, OpLLen, OpSAdd, OpSRem, OpSCard, OpSMove, OpZAdd,
		OpZRem, OpZRemRangeByScore, OpZRemRangeByRank, OpZCard, OpZRank, OpZRevRank, OpZCount:
		return RedisRetTypeInt64
	case OpHIncrByFloat, OpZIncrBy, OpZScore:
		return RedisRetTypeFloat64
	case OpMGet, OpHKeys, OpHVals, OpSMembers:
		return RedisRetTypeStrings
	case OpHGetAll, OpHMGet:
		return RedisRetTypeMapString
	case OpZRange, OpZRangeByScore, OpZRevRange, OpZRevRangeByScore:
		if withScores {
			return RedisRetTypeMemberScore
		} else {
			return RedisRetTypeStrings
		}
	case OpZPopMin, OpZPopMax:
		return RedisRetTypeMemberScore
	}

	return RedisRetTypeNil
}

func GetDataType(v interface{}) types.Type {
	switch v.(type) {
	case []byte, *[]byte:
		return types.TypeBytes
	case int, []int, *int, *[]int:
		return types.TypeInt
	case int8, []int8, *int8, *[]int8:
		return types.TypeInt8
	case int16, []int16, *int16, *[]int16:
		return types.TypeInt16
	case int32, []int32, *int32, *[]int32:
		return types.TypeInt32
	case int64, []int64, *int64, *[]int64:
		return types.TypeInt64
	case uint, []uint, *uint, *[]uint:
		return types.TypeUint
	case uint8, *uint8:
		return types.TypeUint8
	case uint16, []uint16, *uint16, *[]uint16:
		return types.TypeUint16
	case uint32, []uint32, *uint32, *[]uint32:
		return types.TypeUint32
	case uint64, []uint64, *uint64, *[]uint64:
		return types.TypeUint64
	default:
		if types.IsTime(reflect.TypeOf(v)) {
			return types.TypeTime
		}
		return 0
	}
}

// RedisParamInfo redis 请求参数信息
type RedisParamInfo struct {
	Name    string // 参数名称
	Cnt     int    // 参数个数
	JustVal bool   // 仅用到值，不需要Name
}

func FindRedisParam(paramInfos []*RedisParamInfo, name string) (*RedisParamInfo, bool) {
	for _, arg := range paramInfos {
		if arg.Name == name {
			return arg, true
		}
	}

	return nil, false
}

var (
	// SetParams [NX | XX] [GET] [EX seconds | PX milliseconds | EXAT unix-time-seconds | PXAT unix-time-milliseconds | KEEPTTL]
	SetParams = []*RedisParamInfo{
		{"NX", 1, false},      // NX   * seconds -- Set the specified expire time, in seconds (a positive integer).
		{"XX", 1, false},      // XX  * milliseconds -- Set the specified expire time, in milliseconds (a positive integer).
		{"GET", 1, false},     // GET  * Return the old string stored at key, or nil if key did not exist. An error is returned and SET aborted if the value stored at key is not a string.
		{"EX", 2, false},      // EX seconds   * seconds -- Set the specified expire time, in seconds (a positive integer).
		{"PX", 2, false},      // PX milliseconds   * milliseconds -- Set the specified expire time, in milliseconds (a positive integer).
		{"EXAT", 2, false},    // EXAT unix-time-seconds * timestamp-seconds -- Set the specified Unix time at which the key will expire, in seconds (a positive integer).
		{"PXAT", 2, false},    // PXAT unix-time-milliseconds  *timestamp-milliseconds -- Set the specified Unix time at which the key will expire, in milliseconds (a positive integer).
		{"KEEPTTL", 1, false}, // KEEPTTL   * Retain the time to live associated with the key.
	}

	// BitCountParams [start end [BYTE | BIT]]
	BitCountParams = []*RedisParamInfo{
		{"start", 2, true},
		{"end", 2, true},
		{"BYTE", 1, false},
		{"BIT", 1, false},
	}

	// ZRangeParams start stop [BYSCORE | BYLEX] [REV] [LIMIT offset count] [WITHSCORES]
	ZRangeParams = []*RedisParamInfo{
		{"start", 2, true},
		{"stop", 2, true},
		{"BYSCORE", 1, false},
		{"BYLEX", 1, false},
		{"REV", 1, false},
		{"LIMIT", 3, false},
		{"WITHSCORES", 1, false},
	}

	// ZRangeByScoreParams min max [WITHSCORES] [LIMIT offset count]
	ZRangeByScoreParams = []*RedisParamInfo{
		{"min", 2, true},
		{"max", 2, true},
		{"WITHSCORES", 1, false},
		{"LIMIT", 3, false},
	}

	// ZRevRangeParams start stop [WITHSCORES]
	ZRevRangeParams = []*RedisParamInfo{
		{"start", 2, true},
		{"stop", 2, true},
		{"WITHSCORES", 1, false},
	}

	// ZRevRangeByScoreParams max min [WITHSCORES] [LIMIT offset count]
	ZRevRangeByScoreParams = []*RedisParamInfo{
		{"max", 2, true},
		{"min", 2, true},
		{"WITHSCORES", 1, false},
		{"LIMIT", 3, false},
	}
)
