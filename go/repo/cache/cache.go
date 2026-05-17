// Package cache 封装本地内存缓存（替代 RedisUserinfo），支持 per-key TTL
package cache

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"server_golang/common/json"
)

var store *ristretto.Cache[string, string]

func init() {
	var err error
	store, err = ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: 1e6,     // 100 万个 key 的频率追踪
		MaxCost:     1 << 23, // 8MB 由 Set 的 cost 决定，一共 800w记录
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
}

// Get 获取缓存
func Get(key string) (string, bool) {
	return store.Get(key)
}

// SetWithTTL 设置缓存，指定 TTL
func SetWithTTL(key string, value interface{}, ttl int) {
	store.SetWithTTL(key, json.Marshal(value), 1, time.Duration(ttl)*time.Second)
}

// Del 删除缓存
func Del(key string) {
	store.Del(key)
}
