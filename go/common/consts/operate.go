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

const ( // 操作符
	OPEqual          = "="   // 等于
	OPBetween        = "()"  // 在某个区间
	OPNotBetween     = "!()" // 不在某个区间
	OPGt             = ">"   // 大于
	OPGte            = ">="  // 大于等于
	OPLt             = "<"   // 小于
	OPLte            = "<="  // 小于等于
	OPNot            = "!"   // 去反
	OPLike           = "~"   // like语句，（或 es 的部分匹配）
	OPNotLike        = "!~"  // not like 语句，（或 es 的部分匹配排除）
	OPMatchPhrase    = "?"   // es 短语匹配 match_phrase
	OPNotMatchPhrase = "!?"  // es 短语匹配排除 must_not match_phrase
	OPMatch          = "*"   // es 全文搜索 match 语句
	OPNotMatch       = "!*"  // es 全文搜索排除 must_not match
)

const ( // 连接词
	AND = "AND"
	OR  = "OR"
	NOT = "NOT"
)

const ( // 连接词类型
	NotOP     = 0
	AndOrNot  = 1
	EsKeyword = 2
)

const (
	OpTypeRead   = 1 // 读数据
	OpTypeAdd    = 2 // 新增数据
	OpTypeMod    = 3 // 修改数据
	OpTypeDel    = 4 // 删除数据
	OpTypeCreate = 5 // 创建表
	OpTypeDrop   = 6 // 删除表
	OpTypeTrans  = 7 // 事务节点
	OpTypeAuth   = 8 // 暂未支持
)

const (
	OpAuth        = "auth" // 暂未支持
	OpInsert      = "insert"
	OpReplace     = "replace"
	OpUpdate      = "update"
	OpDelete      = "delete"
	OpFind        = "find"
	OpFindAll     = "find_all"
	OpCount       = "count"
	OpTransaction = "transaction"

	// ddl
	OpCreate = "create"
	OpDrop   = "drop"

	// redis 操作
	OpExpire = "expire"
	OpTTL    = "ttl"
	OpExists = "exists"
	OpDel    = "del"

	OpSet      = "set"
	OpSetEx    = "setex"
	OpSetNX    = "setnx"
	OpMSet     = "mset"
	OpGet      = "get"
	OpMGet     = "mget"
	OpGetSet   = "getset"
	OpIncr     = "incr"
	OpDecr     = "decr"
	OpIncrBy   = "incrby"
	OpSetBit   = "setbit"
	OpGetBit   = "getbit"
	OpBitCount = "bitcount"

	OpHSet         = "hset"
	OpHSetNx       = "hsetnx"
	OpHmSet        = "hmset"
	OpHIncrBy      = "hincrby"
	OpHIncrByFloat = "hincrbyfloat"
	OpHDel         = "hdel"
	OpHGet         = "hget"
	OpHMGet        = "hmget"
	OpHGetAll      = "hgetall"
	OpHKeys        = "hkeys"
	OpHVals        = "hvals"
	OpHExists      = "hexists"
	OpHLen         = "hlen"
	OpHStrLen      = "hstrlen"

	OpLPush = "lpush"
	OpRPush = "rpush"
	OpLPop  = "lpop"
	OpRPop  = "rpop"
	OpLLen  = "llen"

	OpSAdd        = "sadd"
	OpSMove       = "smove"
	OpSPop        = "spop"
	OpSRem        = "srem"
	OpSCard       = "scard"
	OpSMembers    = "smembers"
	OpSIsMember   = "sismember"
	OpSRandMember = "srandmember"

	OpZAdd             = "zadd"
	OpZRem             = "zrem"
	OpZRemRangeByScore = "zremrangebyscore"
	OpZRemRangeByRank  = "zremrangebyrank"
	OpZIncrBy          = "zincrby"
	OpZPopMin          = "zpopmin"
	OpZPopMax          = "zpopmax"
	OpZCard            = "zcard"
	OpZScore           = "zscore"
	OpZRank            = "zrank"
	OpZRevRank         = "zrevrank"
	OpZCount           = "zcount"
	OpZRange           = "zrange"
	OpZRangeByScore    = "zrangebyscore"
	OpZRevRange        = "zrevrange"
	OpZRevRangeByScore = "zrevrangebyscore"
)

// OpType 操作类型
func OpType(op string) int8 {
	switch op {
	case OpAuth:
		return OpTypeAuth
	case OpFind, OpFindAll:
		return OpTypeRead
	case OpInsert:
		return OpTypeAdd
	case OpReplace, OpUpdate:
		return OpTypeMod
	case OpDelete:
		return OpTypeDel
	case OpCreate:
		return OpTypeCreate
	case OpDrop:
		return OpTypeDrop
	case OpSet, OpSetEx, OpSetNX, OpMSet, OpGetSet, OpSetBit,
		OpHSet, OpHSetNx, OpHmSet, OpSAdd, OpZAdd, OpLPush, OpRPush:
		return OpTypeAdd
	case OpExpire, OpIncr, OpDecr, OpIncrBy, OpHIncrBy, OpHIncrByFloat, OpSMove, OpZIncrBy:
		return OpTypeMod
	case OpDel, OpHDel, OpLPop, OpRPop, OpSPop, OpSRem, OpZRem,
		OpZRemRangeByScore, OpZRemRangeByRank, OpZPopMin, OpZPopMax:
		return OpTypeDel
	case OpTransaction:
		return OpTypeTrans
	default:
		return OpTypeRead
	}
}
