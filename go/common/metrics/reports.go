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

// Package metrics reports the statistics of the framework.
package metrics

var (
	ServiceHandleFail           = Counter("ServiceHandleFail")
	ServiceCodecDecodeFail      = Counter("ServiceCodecDecodeFail")
	ServiceCodecEncodeFail      = Counter("ServiceCodecEncodeFail")
	ServiceHandleRPCNameInvalid = Counter("ServiceHandleRpcNameInvalid")
	ServiceCodecMarshalFail     = Counter("ServiceCodecMarshalFail")

	TCPServerTransportHandleFail = Counter("TcpServerTransportHandleFail")
	TCPServerTransportWriteFail  = Counter("TcpServerTransportWriteFail")

	SelectNodeFail   = Counter("SelectNodeFail")
	ClientCodecEmpty = Counter("ClientCodecEmpty")

	ConnectionPoolGetNewConnection = Counter("ConnectionPoolGetNewConnection")
	ConnectionPoolGetConnectionErr = Counter("ConnectionPoolGetConnectionErr")
	ConnectionPoolRemoteErr        = Counter("ConnectionPoolRemoteErr")
	ConnectionPoolRemoteEOF        = Counter("ConnectionPoolRemoteEOF")
	ConnectionPoolIdleTimeout      = Counter("ConnectionPoolIdleTimeout")
	ConnectionPoolLifetimeExceed   = Counter("ConnectionPoolLifetimeExceed")
	ConnectionPoolOverLimit        = Counter("ConnectionPoolOverLimit")
)
