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

package snowflake

import (
	"math/rand"
	"sync"
	"time"

	"github.com/martinlindhe/base36"
)

var (
	sf   *SnowFlake
	once = new(sync.Once)
)

func init() {
	sf = &SnowFlake{}
	sf.initialTime = 1680278400000 // 时间戳从 2023-04-01 00:00:00 开始算，64位可以用到 2092 年。（任何人不得修改该参数）
	sf.randMachine = true
}

type SnowFlake struct {
	initialTime uint64 // 初始时间
	lastTime    uint64 // 上次的时间戳(毫秒级)
	randMachine bool   // 当未设置 MachineID 的时候，每次 GenerateID 就会随机产生一个机器 ID，有一定几率发生碰撞。
	machineID   int    // 机器 ID 占10位, 十进制范围是 [ 0, 1023 ]
	lastSn      uint64 // 防止同一秒出现相同的 sn
	sn          uint64 // 序列号占 12 位,十进制范围是 [ 0, 4095 ]
	lock        sync.RWMutex
}

// SetMachineID 设置 machine ID
// 如果未设置 MachineID 的时候，每次 GenerateID 就会随机产生一个机器 ID，有一定几率发生碰撞。
func SetMachineID(machineID int) {
	once.Do( //不允许被插件调用，导致全局的 machine id 异常，数据冲突。
		func() {
			sf.randMachine = false

			// 把机器 id 左移 12 位,让出 12 位空间给序列号使用
			sf.machineID = (int(machineID) % 1000) << 12 // 1000 - 1023 特殊用途
		},
	)
}

// GenerateID 生成 id
func GenerateID() uint64 {
	// 加锁/解锁
	sf.lock.Lock()
	defer sf.lock.Unlock()

	curTimeStamp := uint64(time.Now().UnixNano()) / 1e6

	sf.sn++
	if sf.sn > 4095 {
		sf.sn = 0
	}

	// 这种方案可以在一定概率上服务器的时钟出现回拨（比如闰秒或者NTP同步）
	if curTimeStamp == sf.lastTime {
		if sf.sn == sf.lastSn { //相同时间出现了相同的 sn，这时候将时间往后挪1毫秒（单机瞬时 QPS > 409万的时候会出现）
			time.Sleep(time.Millisecond)
			curTimeStamp = uint64(time.Now().UnixNano()) / 1e6
		}
	} else { //进入下一秒，记录该秒的第一个 sn
		sf.lastSn = sf.sn
	}

	sf.lastTime = curTimeStamp

	// 时间戳占 41 位，00000000 00000000 00000001 11111111 11111111 11111111 11111111 11111111
	timeBin := uint64(curTimeStamp-sf.initialTime) & 0x1FFFFFFFFFF

	// 时间戳左移 22 位，让出 22 位空间给机器码+序列号使用
	timeBin <<= 22

	machineID := sf.machineID
	if sf.randMachine { // 随机机器码
		machineID = rand.Intn(999) << 12
	}

	// 机器码占 10 位，序列号占 12 位
	id := timeBin | uint64(machineID) | sf.sn

	return id
}

// ParseID 解析 id
func ParseID(id uint64) (t time.Time, machineID int, sn uint64) {
	timeBin := id >> 22

	sf.lock.RLock()
	timeUnix := timeBin + sf.initialTime
	sf.lock.RUnlock()

	t = time.Unix(int64(timeUnix/1000), int64(timeUnix%1000)*1e6)

	//00000000 00000000 00000000 00000000 00000000 00111111 11111111 11111111
	machineID = int((id & 0x3FFFFF) >> 12)

	//00000000 00000000 00000000 00000000 00000000 00000000 00001111 11111111
	sn = id & 0xFFF
	return
}

// Generate36ID 生成 36 进制的 ID，包含 0-9, A-Z 所有字符，传输位数少（但是存储位数不如 uint64 只有 8 个字节）
func Generate36ID() string {
	id := GenerateID()
	return base36.Encode(id)
}

// Parse36ID 解析 36 进制的 ID
func Parse36ID(idStr string) (t time.Time, machineID int, sn uint64) {
	id := base36.Decode(idStr)
	return ParseID(id)
}
