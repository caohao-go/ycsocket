// Package config
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
package config

import (
	"os"
	"sync/atomic"
	"time"

	"gopkg.in/yaml.v3"
)

const (
	DefaultIdleTimeout = 60000 // 单位 ms
	MaxCloseWaitTime   = 10 * time.Second
)

// Location 地域信息
type Location struct {
	Region string `yaml:"region"`
	Zone   string `yaml:"zone"`
	Campus string `yaml:"campus"`
}

// ServerConfig 配置
type ServerConfig struct {
	Namespace string    `yaml:"namespace"`  // namespace
	Env       string    `yaml:"env"`        // 环境
	Machine   string    `yaml:"machine"`    // 机器名（容器名）
	Set       string    `yaml:"set"`        // 机器所属set信息
	MachineID int       `yaml:"machine_id"` // 机器编号（容器编号）（主要用于 snowflake 生成全局唯一 id），当 machine_id 未设置的时候从 machine 提取
	LocalIP   string    `yaml:"local_ip"`   // 本地 ip
	Location  *Location `yaml:"location"`   // 地域信息
	Server    struct {
		App     string `yaml:"app"`    // 服务名
		Server  string `yaml:"server"` // 注销名字服务之后的等待时间，让名字服务更新实例列表。 (单位 ms) 默认: 0ms, 最大: 10s.
		Service []struct {
			Name             string `yaml:"name"`                // 服务名
			Type             string `yaml:"type"`                // 类型：web、rpc、http
			Port             uint16 `yaml:"port"`                // 监听端口
			CloseWaitTime    int    `yaml:"close_wait_time"`     // 注销名字服务之后的等待时间，让名字服务更新实例列表。 (单位 ms) 默认: 0ms, 最大: 10s.
			MaxCloseWaitTime int    `yaml:"max_close_wait_time"` // 进程结束之前等待请求完成的最大等待时间。(单位 ms)
			Timeout          int    `yaml:"timeout"`             // 服务超时时间(单位 ms)
			IdleTime         int    `yaml:"idle_time"`           // 连接最大空闲时间，默认为 1 分钟。(单位 ms)
			EventLoopNum     int    `yaml:"event_loop_num"`      // gnet loop 大小，默认取 CPU 核数
			TLSKey           string `yaml:"tls_key"`             // tls key
			TLSCert          string `yaml:"tls_cert"`            // tls cert
			CACert           string `yaml:"ca_cert"`             // ca cert
		}
	}

	Log []*LoggerConfig `yaml:"log"`

	// Register 名字服务
	Register *RegisterConfig `yaml:"register"`

	// GatewayToken 网关认证token
	GatewayToken map[string]string `yaml:"gateway_token"`

	// WebSalt web 端加密盐
	WebSalt string `yaml:"web_salt"`
}

var globalConfig atomic.Value // 服务端配置

// Config returns the common Config.
func Config() *ServerConfig {
	return globalConfig.Load().(*ServerConfig)
}

// LoadConfig 加载配置文件
func LoadConfig(configPath string) (*ServerConfig, error) {
	cfg, err := parseConfigFile(configPath)
	if err != nil {
		return nil, err
	}

	for k := range cfg.Server.Service {
		cfg.Server.Service[k].IdleTime = DefaultIdleTimeout
	}

	globalConfig.Store(cfg)

	return cfg, nil
}

func parseConfigFile(configPath string) (*ServerConfig, error) {
	buf, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	cfg := &ServerConfig{}
	if err := yaml.Unmarshal(buf, cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}
