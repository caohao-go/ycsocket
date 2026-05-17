package attr

import (
	"context"
	"fmt"
	"sync"
	"time"

	"git.code.oa.com/pcg-csd/trpc-ext/redis"
	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo"
)

var userAttrs = map[int64]types.Map{}
var userAttrMutex = sync.RWMutex{}

// initUserAttr 将用户固定属性从 pika 加载到内存
func initUserAttr(uid int64) {
	ctx := context.Background()

	userAttrMutex.RLock()
	if _, ok := userAttrs[uid]; ok {
		userAttrMutex.RUnlock()
		return
	}
	userAttrMutex.RUnlock()

	// 从 pika 加载到内存
	data, err := repo.RedisHGetAll(ctx, fmt.Sprintf("userattr_%d", uid))
	if err != nil && !redis.IsNil(err) {
		panic(err)
	}

	if data == nil || len(data) == 0 {
		initData := types.Map{
			config.UserId:        uid,
			config.AttrLv:        "1",
			config.AttrVipLevel:  "0",
			config.AttrCopy:      "0",
			config.AttrChapter:   "0",
			config.AttrRoleTitle: "0",
			config.AttrNewGift:   "0",
			config.AttrDay7:      "0",
			config.AttrIGift:     "0",
			config.AttrInitFlag:  "0",
			config.AttrRegT:      time.Now().Format("2006-01-02 15:04:05"),
			config.AttrOffTime:   time.Now().Format("2006-01-02 15:04:05"),
		}

		// 初始化用户属性
		err = repo.RedisHMSet(ctx, fmt.Sprintf("userattr_%d", uid), initData)

		if err != nil {
			panic(err)
		}

		userAttrMutex.Lock()
		userAttrs[uid] = initData
		userAttrs[uid]["last_update_time"] = time.Now().Unix()
		userAttrMutex.Unlock()
	} else {
		userAttrMutex.Lock()
		userAttrs[uid] = data
		userAttrs[uid]["last_update_time"] = time.Now().Unix()
		userAttrMutex.Unlock()
	}

	return
}

// Set 更新用户属性，直接更新内存，再异步更新 pika
func Set(uid int64, field string, val interface{}) {
	initUserAttr(uid)

	userAttrMutex.Lock()
	userAttrs[uid][field] = val
	userAttrs[uid]["last_update_time"] = time.Now().Unix()
	userAttrMutex.Unlock()

	go func(val interface{}) { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHSet(ctx, fmt.Sprintf("userattr_%d", uid), field, json.Marshal(val))
		if err != nil {
			log.Errorf(ctx, -1, "Set failed for uid=%d field=%s, val=%s, err=%v", uid, field, val, err)
		}
	}(val)
}

// HmSet 更新用户属性，直接更新内存，再异步更新 pika
func HmSet(uid int64, datas types.Map) {
	initUserAttr(uid)

	userAttrMutex.Lock()
	for k, v := range datas {
		userAttrs[uid][k] = v
	}
	userAttrs[uid]["last_update_time"] = time.Now().Unix()
	userAttrMutex.Unlock()

	go func(datas types.Map) { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHMSet(ctx, fmt.Sprintf("userattr_%d", uid), datas)
		if err != nil {
			log.Errorf(ctx, -1, "HmSet failed for uid=%d datas=%v, err=%v", uid, datas, err)
		}
	}(datas)
}

// Incr 用户属性+1
func Incr(uid int64, field string, incr int64) int64 {
	initUserAttr(uid)

	userAttrMutex.Lock()
	var num int64
	if _, ok := userAttrs[uid][field]; !ok {
		num = incr
	} else {
		num = userAttrs[uid].GetInt64E(field) + incr
	}

	userAttrs[uid][field] = num
	userAttrs[uid]["last_update_time"] = time.Now().Unix()
	userAttrMutex.Unlock()

	go func(incr int64) { // 异步更新 pika
		ctx := context.Background()
		_, err := repo.RedisHIncrBy(ctx, fmt.Sprintf("userattr_%d", uid), field, incr)
		if err != nil {
			log.Errorf(ctx, -1, "Incr failed for uid=%d field=%s, incr=%d, err=%v", uid, field, incr, err)
		}
	}(incr)

	return num
}

// Get 获取用户属性，直接从内存获取
func Get(uid int64, field string) (interface{}, bool) {
	initUserAttr(uid)

	userAttrMutex.RLock()
	ret, exists := userAttrs[uid][field]
	userAttrMutex.RUnlock()

	return ret, exists
}

// GetAll 获取用户所有属性
func GetAll(uid int64) types.Map {
	initUserAttr(uid)

	userAttrMutex.RLock()
	val := types.CopyMap(userAttrs[uid])
	userAttrMutex.RUnlock()
	return val
}

// SetByPrefix 根据 prefix 设置用户属性，直接更新内存，再异步更新 pika
func SetByPrefix(uid int64, prefix string, id, val interface{}) {
	field := prefix + types.ToString(id)
	Set(uid, field, val)
}

// IncrByPrefix 根据 prefix 用户属性+1
func IncrByPrefix(uid int64, prefix string, id interface{}, incr int64) int64 {
	field := prefix + types.ToString(id)
	return Incr(uid, field, incr)
}

// GetByPrefix 根据 prefix 获取用户属性
func GetByPrefix(uid int64, prefix string, id interface{}) (interface{}, bool) {
	field := prefix + types.ToString(id)
	return Get(uid, field)
}
