package content

import (
	"context"
	"fmt"
	"sync"

	"git.code.oa.com/pcg-csd/trpc-ext/redis"
	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/repo"
)

// userContents 内存中缓存的用户 content 数据
// key: userID, value: map[contentType]contentValue(JSON 字符串)
var userContents = map[int64]types.Map{}
var contentMutex = sync.RWMutex{}

// initUserContent 将用户 content 从 Pika 加载到内存
func initUserContent(uid int64) {
	ctx := context.Background()

	contentMutex.RLock()
	if _, ok := userContents[uid]; ok {
		contentMutex.RUnlock()
		return
	}
	contentMutex.RUnlock()

	// 从 Pika 加载到内存
	data, err := repo.RedisHGetAll(ctx, pikaKey(uid))
	if err != nil && !redis.IsNil(err) {
		panic(err)
	}

	contentMutex.Lock()
	if len(data) == 0 {
		userContents[uid] = types.Map{}
	} else {
		userContents[uid] = data
	}
	contentMutex.Unlock()
}

// ========== JSON Map 类型 content 操作 ==========

// GetMap 获取用户 JSON content，返回解析后的 types.Map
func GetMap(uid int64, contentType string) types.Map {
	initUserContent(uid)

	contentMutex.RLock()
	raw, exists := userContents[uid][contentType]
	contentMutex.RUnlock()

	if !exists || raw == nil || raw == "" {
		return types.Map{}
	}

	return types.ToMapE(raw)
}

// SetMap 更新用户 JSON content（Map 类型），同步写内存，异步写 Pika
func SetMap(uid int64, contentType string, content types.Map) {
	initUserContent(uid)

	if content == nil {
		content = types.Map{}
	}
	val := json.Marshal(content)

	contentMutex.Lock()
	userContents[uid][contentType] = val
	contentMutex.Unlock()

	go func(val string) { // 异步更新 Pika
		ctx := context.Background()
		err := repo.RedisHSet(ctx, pikaKey(uid), contentType, val)
		if err != nil {
			log.Errorf(ctx, -1, "content.SetMap failed for uid=%d type=%s, err=%v", uid, contentType, err)
		}
	}(val)
}

// ========== JSON Array 类型 content 操作 ==========

// GetArray 获取用户 JSON content，返回解析后的数组
func GetArray(uid int64, contentType string) []interface{} {
	initUserContent(uid)

	contentMutex.RLock()
	raw, exists := userContents[uid][contentType]
	contentMutex.RUnlock()

	if !exists || raw == nil || raw == "" {
		return []interface{}{}
	}

	return types.ToArrayE(raw)
}

// SetArray 更新用户数组类型 content，同步写内存，异步写 Pika
func SetArray(uid int64, contentType string, content []interface{}) {
	initUserContent(uid)

	if content == nil {
		content = []interface{}{}
	}
	val := json.Marshal(content)

	contentMutex.Lock()
	userContents[uid][contentType] = val
	contentMutex.Unlock()

	go func(val string) { // 异步更新 Pika
		ctx := context.Background()
		err := repo.RedisHSet(ctx, pikaKey(uid), contentType, val)
		if err != nil {
			log.Errorf(ctx, -1, "content.SetArray failed for uid=%d type=%s, err=%v", uid, contentType, err)
		}
	}(val)
}

// ========== Int 类型 content 操作 ==========

// GetInt 获取用户整数 content
func GetInt(uid int64, contentType string) int {
	initUserContent(uid)

	contentMutex.RLock()
	raw, exists := userContents[uid][contentType]
	contentMutex.RUnlock()

	if !exists || raw == nil || raw == "" {
		return 0
	}

	return types.ToIntE(raw)
}

// SetInt 更新用户整数 content，同步写内存，异步写 Pika
func SetInt(uid int64, contentType string, num int) {
	initUserContent(uid)

	contentMutex.Lock()
	userContents[uid][contentType] = num
	contentMutex.Unlock()

	go func(num int) { // 异步更新 Pika
		ctx := context.Background()
		err := repo.RedisHSet(ctx, pikaKey(uid), contentType, num)
		if err != nil {
			log.Errorf(ctx, -1, "content.SetInt failed for uid=%d type=%s, err=%v", uid, contentType, err)
		}
	}(num)
}

// IncrInt 自增用户整数 content，同步写内存，异步写 Pika
func IncrInt(uid int64, contentType string, incr int) int {
	initUserContent(uid)

	contentMutex.Lock()
	num := userContents[uid].GetIntE(contentType) + incr
	userContents[uid][contentType] = num
	contentMutex.Unlock()

	go func(incr int) { // 异步更新 Pika
		ctx := context.Background()
		_, err := repo.RedisHIncrBy(ctx, pikaKey(uid), contentType, int64(incr))
		if err != nil {
			log.Errorf(ctx, -1, "content.IncrInt failed for uid=%d type=%s, incr=%d, err=%v", uid, contentType, incr, err)
		}
	}(incr)

	return num
}

// GetMoreInt 批量获取多个用户的整数 content（用于排行等场景）
// 注意：此方法直接从 Pika 读取，不经过内存缓存
func GetMoreInt(uids []int64, contentType string) map[int64]int {
	ret := make(map[int64]int, len(uids))
	for _, uid := range uids {
		ret[uid] = GetInt(uid, contentType)
	}
	return ret
}

// pikaKey 返回 Pika 中存储 user_content 的 Hash key
func pikaKey(uid int64) string {
	return fmt.Sprintf("content_%d", uid)
}
