// hash.go 内存式 Hash 缓存，模拟 Redis HGET/HSET/DEL 语义
// 底层用 ristretto（TinyLFU 淘汰），key 拼接为 "hashKey_field"
package cache

import (
	"time"

	"github.com/dgraph-io/ristretto/v2"
	"server_golang/common/json"
)

var hashStore *ristretto.Cache[string, string]

func init() {
	var err error
	hashStore, err = ristretto.NewCache(&ristretto.Config[string, string]{
		NumCounters: 1e6,     // 100 万个 key 的频率追踪
		MaxCost:     1 << 21, // 2MB 由 Set 的 cost 决定，一共 200w记录
		BufferItems: 64,
	})
	if err != nil {
		panic(err)
	}
}

// hashCacheKey 拼接 hash key + field 为一个扁平 key
func hashCacheKey(key string, field interface{}) string {
	return key + "_" + json.Marshal(field)
}

// HGet 获取 hash 缓存中指定 field 的值
func HGet(key string, field interface{}) (string, bool) {
	return hashStore.Get(hashCacheKey(key, field))
}

// HSet 设置 hash 缓存中指定 field 的值，带 TTL
func HSet(key string, field, value interface{}, ttl int) {
	hashStore.SetWithTTL(hashCacheKey(key, field), json.Marshal(value), 1, time.Duration(ttl)*time.Second)
}

// HDel 删除 hash 缓存中指定 field
func HDel(key string, field interface{}) {
	hashStore.Del(hashCacheKey(key, field))
}

// addPropFields 职业类型 field 范围（0-10 覆盖所有可能值）
var addPropFields = []string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "10"}

// ClearAddPropCache 清除指定用户的附加属性 hash 缓存（所有职业类型）
func ClearAddPropCache(key string) {
	HDelAll(key, addPropFields)
}

// HDelAll 删除整个 hash key 下所有 field（通过清理已知 field 列表）
// 由于 ristretto 不支持前缀扫描，调用方无法枚举所有 field，
// 因此提供批量删除接口，由调用方传入需要清理的 field 列表。
func HDelAll(key string, fields []string) {
	for _, f := range fields {
		hashStore.Del(hashCacheKey(key, f))
	}
}
