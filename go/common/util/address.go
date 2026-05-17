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

package util

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"

	"server_golang/common/consts"
	"server_golang/common/errs"
	"server_golang/common/types"
	"server_golang/common/url"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
)

type DBAddress struct {
	Type    consts.DBType `json:"type,omitempty"`    // 数据库类型 1-elastic 2-mongo 3-redis 10-mysql 11-postgresql 12-clickhouse 13-oracle 14-DB2 15-sqlite
	Version string        `json:"version,omitempty"` // 数据库版本，比如elastic v6，v7
	Network string        `json:"network,omitempty"` // network TCP/UDP
	Address string        `json:"address,omitempty"` // address

	Conn *DBConnInfo `json:"conn,omitempty"` // connect info

	WriteTimeout int `json:"write_timeout,omitempty"` // 写超时（毫秒）
	ReadTimeout  int `json:"read_timeout,omitempty"`  // 读超时（毫秒）

	WarnTimeout int  `json:"warn_timeout,omitempty"` // 告警超时（ms），如果请求耗时超过这个时间，就会打 warning 日志
	OmitError   int8 `json:"omit_error,omitempty"`   // 是否忽略 error 日志，0-否 1-是
	Debug       int8 `json:"debug,omitempty"`        // 是否开启 debug 日志，正常的数据库请求也会被打印到日志，0-否 1-是，会造成海量日志，慎重开启
}

type DBConnInfo struct {
	Schema   string `json:"schema,omitempty"`   // schema
	Target   string `json:"target,omitempty"`   // target
	DB       string `json:"db,omitempty"`       // db name
	Password string `json:"password,omitempty"` // password
	Params   string `json:"params,omitempty"`   // params
	DSN      string `json:"dsn,omitempty"`      // dsn
	DSN2     string `json:"dsn2,omitempty"`     // dsn2
	DSN3     string `json:"dsn3,omitempty"`     // dsn3
}

var (
	DBConnMap     = map[consts.DBType]map[string]*DBConnInfo{}
	DBConnMapLock = new(sync.RWMutex)
)

// ParseConnFromAddress 解析数据库网络信息
func ParseConnFromAddress(addr *DBAddress) (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = fmt.Errorf("parse address failed: %v", e)
		}
	}()

	if addr == nil {
		return errors.New("address is nil")
	}

	if addr.Address == "" {
		return errors.New("address is empty")
	}

	DBConnMapLock.RLock()
	dbConnMap, ok := DBConnMap[addr.Type]

	if ok {
		conn, ok := dbConnMap[addr.Address]
		DBConnMapLock.RUnlock()

		if ok {
			addr.Conn = conn
			return
		}
	} else {
		DBConnMapLock.RUnlock()

		DBConnMapLock.Lock()
		DBConnMap[addr.Type] = map[string]*DBConnInfo{}
		DBConnMapLock.Unlock()
	}

	addr.Conn, err = getConnFromAddress(addr.Type, addr.Address)
	if err != nil {
		return err
	}

	DBConnMapLock.Lock()
	DBConnMap[addr.Type][addr.Address] = addr.Conn
	DBConnMapLock.Unlock()

	return nil
}

func getConnFromAddress(typ consts.DBType, address string) (*DBConnInfo, error) {
	conn := DBConnInfo{}
	conn.Schema = "ip"

	switch typ {
	case consts.DBTypeRedis, consts.DBTypeElastic:
		_, conn.Schema, conn.Target = types.CutString(address, "://")

		target, _, params, e := ParseTarget(conn.Target)
		if e != nil {
			return nil, e
		}

		if target == "" {
			return nil, errors.New("target is empty")
		}

		conn.DSN = target
		conn.Target = target
		conn.DB, _ = params["db"]
		conn.Password, _ = params["password"]

		delete(params, "password")
		conn.Params = url.ParamEncode(params)
	case consts.DBTypeMySQL, consts.DBTypePostgreSQL, consts.DBTypeClickHouse, consts.DBTypeOracle, consts.DBTypeDB2:
		var network = "tcp"

		found, password, target := types.CutString(address, "@tcp(")
		if !found {
			found, password, target = types.CutString(address, "@udp(")
			network = "udp"
		}

		if !found {
			return nil, errors.New("address need include @tcp or @udp")
		}

		_, conn.Schema, conn.Password = types.CutString(password, "://")
		if conn.Password == "" {
			return nil, errors.New("password is empty")
		}

		_, conn.Target, conn.DB = types.CutString(target, ")/")
		if conn.Target == "" {
			return nil, errors.New("target is empty")
		}

		db, _, params, _ := ParseTarget(conn.DB)
		if db == "" {
			return nil, errors.New("db is empty")
		}

		conn.DB = db

		if typ == consts.DBTypeClickHouse {
			conn.DSN = fmt.Sprintf("dsn://%s@%s(%s)/%s?%s",
				conn.Password, network, conn.Target, conn.DB, url.ParamEncode(params))
		} else {
			conn.DSN = fmt.Sprintf("%s@%s(%s)/%s?%s",
				conn.Password, network, conn.Target, conn.DB, url.ParamEncode(params))
		}
	default:
		_, conn.Schema, conn.Target = types.CutString(address, "://")
		target, params, _, e := ParseTarget(conn.Target)
		if e != nil {
			return nil, e
		}

		if target == "" {
			return nil, errors.New("target is empty")
		}

		conn.Target = target
		conn.Params = params
		conn.DSN = address
	}

	return &conn, nil
}

// ParseTarget address should be like: target?param1=value1&param2=value2&param3=value3 ...
func ParseTarget(address string) (target, params string, paramsMap map[string]string, err error) {
	if address == "" {
		return
	}

	found, s1, s2 := types.CutString(address, "?")
	if found {
		target = s1
		params = s2
	} else {
		target = s2
	}

	paramsMap, err = url.ParseQuery(params)
	return
}

var localIP string
var localIPOnce = new(sync.Once)

// GetLocalIP 获取本机 ip 地址
func GetLocalIP() string {
	localIPOnce.Do(
		func() {
			addrs, err := net.InterfaceAddrs()
			if err != nil {
				log.Errorf(context.Background(), errs.ErrSystem, "get local address error")
				return
			}

			for _, address := range addrs {
				if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
					if ipNet.IP.To4() != nil {
						localIP = ipNet.IP.String()
					}
				}
			}

			return
		})

	return localIP
}

// GetIpFromAddr 从地址中获取 ip
func GetIpFromAddr(addr net.Addr) string {
	ipPort := addr.String()
	cutIpPort := strings.Index(ipPort, ":")

	switch cutIpPort {
	case -1:
		return ipPort
	case 0:
		return ipPort[1:]
	default:
		return ipPort[0:cutIpPort]
	}
}

// GetIDFromContainerName 从容器名中获取 id
func GetIDFromContainerName(s string) int {
	runes := []rune(s)
	for i := len(runes) - 1; i >= 0; i-- {
		if runes[i] < '0' || runes[i] > '9' {
			n, _ := strconv.Atoi(string(runes[i+1:]))
			return n
		}
	}

	n, _ := strconv.Atoi(string(runes))
	return n
}

// ParseHostPort 解析地址
func ParseHostPort(address string) string {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return address
	}

	ip := net.ParseIP(host)
	if ip != nil {
		return address
	}

	// host 不是 ip
	return net.JoinHostPort(getIP(host), port)
}

// ip records the local nic name->nicIP mapping.
var ip = &netInterfaceIP{}

// nicIP defines the parameters used to record the ip address (ipv4 & ipv6) of the nic.
type nicIP struct {
	nic  string
	ipv4 []string
	ipv6 []string
}

func getIP(nic string) string {
	return ip.getIPByNic(nic)
}

type netInterfaceIP struct {
	once sync.Once
	ips  map[string]*nicIP
}

func (p *netInterfaceIP) enumAllIP() map[string]*nicIP {
	p.once.Do(func() {
		p.ips = make(map[string]*nicIP)
		interfaces, err := net.Interfaces()
		if err != nil {
			return
		}
		for _, i := range interfaces {
			p.addInterface(i)
		}
	})
	return p.ips
}

func (p *netInterfaceIP) addInterface(i net.Interface) {
	addrs, err := i.Addrs()
	if err != nil {
		return
	}
	for _, v := range addrs {
		ipNet, ok := v.(*net.IPNet)
		if !ok {
			continue
		}
		if ipNet.IP.To4() != nil {
			p.addIPv4(i.Name, ipNet.IP.String())
		} else if ipNet.IP.To16() != nil {
			p.addIPv6(i.Name, ipNet.IP.String())
		}
	}
}

func (p *netInterfaceIP) addIPv4(nic string, ip4 string) {
	ips := p.getNicIP(nic)
	ips.ipv4 = append(ips.ipv4, ip4)
}

func (p *netInterfaceIP) addIPv6(nic string, ip6 string) {
	ips := p.getNicIP(nic)
	ips.ipv6 = append(ips.ipv6, ip6)
}

func (p *netInterfaceIP) getNicIP(nic string) *nicIP {
	if _, ok := p.ips[nic]; !ok {
		p.ips[nic] = &nicIP{nic: nic}
	}
	return p.ips[nic]
}

func (p *netInterfaceIP) getIPByNic(nic string) string {
	p.enumAllIP()
	if len(p.ips) <= 0 {
		return ""
	}
	if _, ok := p.ips[nic]; !ok {
		return ""
	}
	ip := p.ips[nic]
	if len(ip.ipv4) > 0 {
		return ip.ipv4[0]
	}
	if len(ip.ipv6) > 0 {
		return ip.ipv6[0]
	}
	return ""
}
