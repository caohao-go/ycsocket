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

type DBType uint8

const ( // 支持的服务类型
	DBTypeElastic    DBType = 1  // elastic search
	DBTypeMongo      DBType = 2  // mongo 暂未支持
	DBTypeRedis      DBType = 3  // redis
	DBTypeMySQL      DBType = 10 // mysql
	DBTypePostgreSQL DBType = 11 // postgresql
	DBTypeClickHouse DBType = 12 // clickhouse
	DBTypeOracle     DBType = 13 // oracle 暂未支持
	DBTypeDB2        DBType = 14 // DB2 暂未支持
	DBTypeSQLite     DBType = 15 // sqlite 暂未支持
	DBTypeRPC        DBType = 40 // rpc 协议，暂未支持，spring cloud 协议可以选 grpc、thrift、tars、dubbo 协议
	DBTypeHTTP       DBType = 50 // http 请求
	DBTypeFunction   DBType = 60 // 函数逻辑
)

var DBTypeMap = map[string]DBType{
	"elastic":    DBTypeElastic,
	"redis":      DBTypeRedis,
	"mysql":      DBTypeMySQL,
	"postgresql": DBTypePostgreSQL,
	"clickhouse": DBTypeClickHouse,
	"oracle":     DBTypeOracle,
	"db2":        DBTypeDB2,
	"sqlite":     DBTypeSQLite,
	"mongo":      DBTypeMongo,
	"rpc":        DBTypeRPC,
	"http":       DBTypeHTTP,
	"function":   DBTypeFunction,
}

var DBTypeDesc = map[DBType]string{
	DBTypeElastic:    "elastic",
	DBTypeRedis:      "redis",
	DBTypeMySQL:      "mysql",
	DBTypePostgreSQL: "postgresql",
	DBTypeClickHouse: "clickhouse",
	DBTypeOracle:     "oracle",
	DBTypeDB2:        "db2",
	DBTypeSQLite:     "sqlite",
	DBTypeMongo:      "mongo",
	DBTypeRPC:        "rpc",
	DBTypeHTTP:       "http",
	DBTypeFunction:   "function",
}
