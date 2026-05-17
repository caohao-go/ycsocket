package lock

import (
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

// 本地锁缓存（ristretto 实现）
var (
	lockStore     *ristretto.Cache[string, string]
	lockStoreOnce sync.Once
)

// getLockStore 获取锁缓存单例
func getLockStore() *ristretto.Cache[string, string] {
	lockStoreOnce.Do(func() {
		var err error
		lockStore, err = ristretto.NewCache(&ristretto.Config[string, string]{
			NumCounters: 1e5,     // 10 万个 key 的频率追踪
			MaxCost:     1 << 20, // 1MB 由 Set 的 cost 决定，一共 100w 记录
			BufferItems: 64,
		})
		if err != nil {
			panic("failed to create lock cache: " + err.Error())
		}
	})
	return lockStore
}

// ========================= 锁 =========================

// Lock 加锁，防止操作过快（使用本地 ristretto）
func Lock(key string, ttl int) error {
	if ttl <= 0 {
		ttl = 2
	}
	cache := getLockStore()
	lockKey := "lck_" + key

	// 检查锁是否存在且未过期
	val, found := cache.Get(lockKey)
	if found && len(val) > 0 {
		expireAt, err := strconv.ParseInt(val, 10, 64)
		if err == nil && time.Now().Unix() < expireAt {
			return fmt.Errorf("操作太快，请稍后再试")
		}
		// 已过期，清除旧锁
		cache.Del(lockKey)
	}

	// 设置锁，存储过期时间戳
	cache.SetWithTTL(lockKey, "1", 1, time.Duration(ttl)*time.Second)
	return nil
}

// Unlock 解锁
func Unlock(key string) {
	cache := getLockStore()
	cache.Del("lck_" + key)
}
