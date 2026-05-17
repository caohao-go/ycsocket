package connector

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"git.woa.com/trpc-go/tnet/extensions/websocket"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo/mem/attr"
)

const ConnectExpireTime = 300 // 超时时间（秒）

// ConnManager 使用 tnet websocket.Conn 管理 WebSocket 连接
type ConnManager struct {
	mu             sync.RWMutex
	connMap        map[int64]websocket.Conn
	heartbeatTimes map[int64]int64
	zoneID         int
}

var Manager *ConnManager

func Init(zoneID int) {
	Manager = &ConnManager{
		connMap:        make(map[int64]websocket.Conn),
		heartbeatTimes: make(map[int64]int64),
		zoneID:         zoneID,
	}
}

func (m *ConnManager) IsOnline(uid int64) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, ok := m.connMap[uid]
	return ok
}

// SetWsConn 将 tnet websocket.Conn 绑定到用户ID
func (m *ConnManager) SetWsConn(uid int64, conn websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if oldConn, ok := m.connMap[uid]; ok {
		if oldConn != conn {
			data := types.Map{
				"code":        "2",
				"description": "input error",
				"data":        "同一账号不能同时登陆",
			}
			msg := json.MarshalToBytes(data)
			oldConn.WriteMessage(websocket.Text, msg)
			oldConn.Close()
		}
	}

	m.connMap[uid] = conn
	m.heartbeatTimes[uid] = time.Now().Unix() + ConnectExpireTime
}

func (m *ConnManager) SetConnectExpire(uid int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.heartbeatTimes[uid] = time.Now().Unix() + ConnectExpireTime
}

func (m *ConnManager) Close(uid int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if conn, ok := m.connMap[uid]; ok {
		conn.Close()
	}
	delete(m.connMap, uid)
	delete(m.heartbeatTimes, uid)
}

func (m *ConnManager) Send(uid int64, msg string) bool {
	m.mu.RLock()
	conn, ok := m.connMap[uid]
	m.mu.RUnlock()

	if !ok || conn == nil {
		return false
	}

	err := conn.WriteMessage(websocket.Text, []byte(msg))
	return err == nil
}

func (m *ConnManager) SendAll(msg string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := []byte(msg)
	for uid, conn := range m.connMap {
		if uid%1000 != int64(m.zoneID) {
			continue
		}
		conn.WriteMessage(websocket.Text, data)
	}
}

func (m *ConnManager) SendFds(uids []int64, msg string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := []byte(msg)
	for _, uid := range uids {
		if conn, ok := m.connMap[uid]; ok {
			conn.WriteMessage(websocket.Text, data)
		}
	}
}

func (m *ConnManager) ConnectExpire() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now().Unix()
	for uid, quitTime := range m.heartbeatTimes {
		if now >= quitTime {
			if conn, ok := m.connMap[uid]; ok {
				conn.Close()
				delete(m.connMap, uid)

				attr.Set(uid, config.AttrOffTime, time.Now().Format("2006-01-02 15:04:05"))
			}
			delete(m.heartbeatTimes, uid)
		}
	}
}

func (m *ConnManager) StartExpireChecker() {
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for range ticker.C {
			func() {
				defer func() {
					if r := recover(); r != nil {
						log.Errorf(context.Background(), 0, "ConnectExpire panic: %v", r)
					}
				}()
				m.ConnectExpire()
			}()
		}
	}()
}

// GetOnlineCount 返回当前在线用户数量
func (m *ConnManager) GetOnlineCount() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.connMap)
}

// BroadcastToZone 向当前区所有用户广播消息
func (m *ConnManager) BroadcastToZone(msg string) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	data := []byte(msg)
	for _, conn := range m.connMap {
		conn.WriteMessage(websocket.Text, data)
	}
	log.Infof(context.Background(), "Broadcast to %d users", len(m.connMap))
}

func init() {
	_ = fmt.Sprintf // 避免未使用导入报错
}
