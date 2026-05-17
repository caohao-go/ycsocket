package heroattr

import (
	"context"
	"fmt"
	"sync"

	"git.code.oa.com/pcg-csd/trpc-ext/redis"
	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/logic"
	"server_golang/repo"
)

var heroAttrs = map[int64]map[int]*logic.Hero{}
var heroAttrMutex = sync.RWMutex{}

// initHeroAttr 将用户固定属性从 pika 加载到内存
func initHeroAttr(uid int64) {
	ctx := context.Background()

	heroAttrMutex.RLock()
	if _, ok := heroAttrs[uid]; ok {
		heroAttrMutex.RUnlock()
		return
	}
	heroAttrMutex.RUnlock()

	// 从 pika 加载到内存
	data, err := repo.RedisHGetAll(ctx, fmt.Sprintf("heroattr_%d", uid))
	if err != nil && !redis.IsNil(err) {
		panic(fmt.Errorf("initHeroAttr failed for uid=%d  err=%v", uid, err))
	}

	heroAttrMutex.Lock()
	heroAttrs[uid] = map[int]*logic.Hero{}
	if len(data) > 0 {
		for k, v := range data {
			tmp := logic.Hero{}
			json.Unmarshal(v, &tmp)
			heroAttrs[uid][types.ToIntE(k)] = &tmp
		}
	}
	heroAttrMutex.Unlock()

	return
}

// Set 更新英雄属性，直接更新内存，再异步更新 pika
func Set(uid int64, id int, val *logic.Hero) {
	initHeroAttr(uid)

	heroAttrMutex.Lock()
	heroAttrs[uid][id] = val
	heroAttrMutex.Unlock()

	go func(tmp *logic.Hero) { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHSet(ctx, fmt.Sprintf("heroattr_%d", uid), id, json.Marshal(tmp))
		if err != nil {
			log.Errorf(ctx, -1, "Set failed for uid=%d field=%d, val=%v, err=%v", uid, id, tmp, err)
		}
	}(val)
}

// Get 获取用户属性，直接从内存获取
func Get(uid int64, id int) (*logic.Hero, bool) {
	initHeroAttr(uid)

	heroAttrMutex.RLock()
	ret, exists := heroAttrs[uid][id]
	heroAttrMutex.RUnlock()

	if ret == nil {
		return nil, exists
	}
	return ret.Clone(), exists
}
