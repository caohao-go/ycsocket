// rank_cache.go 排行榜列表缓存（替代 Redis Hash 缓存 KeyPre1）
// 独立 ristretto 实例，key 拼接为 "hashKey\x00field"，支持 per-key TTL 和 LFU 淘汰
package cache

import (
	"sync"
	"time"

	"github.com/dgraph-io/ristretto/v2"
)

var rankStore *ristretto.Cache[string, string]

// rankFields 记录每个 hash key 当前有哪些 field，用于整 key 清除
var rankFields sync.Map // key -> *sync.Map{field -> struct{}}

func init() {
	var err error
	rankStore, err = ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: 1e5,     // 10 万个 key 的频率追踪
		MaxCost:     1 << 21, // 2MB 由 Set 的 cost 决定，一共 200w记录
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
}

// rankCacheKey 拼接 hash key + field 为扁平 key
func rankCacheKey(key, field string) string {
	return key + "_" + field
}

// RankGet 获取排行榜缓存
func RankGet(key, field string) (string, bool) {
	return rankStore.Get(rankCacheKey(key, field))
}

// RankSet 设置排行榜缓存，指定 TTL
func RankSet(key, field, value string, ttl int) {
	rankStore.SetWithTTL(rankCacheKey(key, field), value, 1, time.Duration(ttl)*time.Second)
	// 记录 field 以便整 key 清除
	actual, _ := rankFields.LoadOrStore(key, &sync.Map{})
	actual.(*sync.Map).Store(field, struct{}{})
}

// RankDelAll 清除指定 key 下所有 field 的缓存
func RankDelAll(key string) {
	if actual, ok := rankFields.LoadAndDelete(key); ok {
		actual.(*sync.Map).Range(func(f, _ any) bool {
			rankStore.Del(rankCacheKey(key, f.(string)))
			return true
		})
	}
}
