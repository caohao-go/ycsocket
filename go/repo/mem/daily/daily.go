package daily

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"git.code.oa.com/pcg-csd/trpc-ext/redis"
	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/repo"
)

var dailyAttrs = map[string]map[int64]types.Map{}
var dailyAttrMutex = sync.RWMutex{}

func init() {
	dailyAttrs[util.DateYmd()] = map[int64]types.Map{}
	dailyAttrs[util.NextDateYmd()] = map[int64]types.Map{}
}

// ClearLastDate 每天早上3点清除昨天的内存数据（pika 会自动过期）
func ClearLastDate(ctx context.Context) {
	dailyAttrMutex.Lock()
	delete(dailyAttrs, util.LastDateYmd())                 // 删除昨天内存数据
	dailyAttrs[util.NextDateYmd()] = map[int64]types.Map{} // 创建明天的内存 map
	dailyAttrMutex.Unlock()
}

// initDailyAttr 将每日用户属性从 pika 加载到内存
func initDailyAttr(uid int64) {
	ctx := context.Background()
	today := util.DateYmd()

	dailyAttrMutex.RLock()
	if _, ok := dailyAttrs[today][uid]; ok {
		dailyAttrMutex.RUnlock()
		return
	}
	dailyAttrMutex.RUnlock()

	// 从pika加载到内存
	data, err := repo.RedisHGetAll(ctx, fmt.Sprintf("daily_%s_%d", today, uid))
	if err != nil && !redis.IsNil(err) {
		panic(err)
	}

	if data == nil || len(data) == 0 {
		// 初始化用户属性
		err = repo.RedisHSet(ctx, fmt.Sprintf("daily_%s_%d", today, uid), config.UserId, uid)

		if err != nil {
			panic(err)
		}

		_ = repo.RedisExpire(ctx, fmt.Sprintf("daily_%s_%d", today, uid), 86400)

		dailyAttrMutex.Lock()
		dailyAttrs[today][uid] = types.Map{config.UserId: uid}
		dailyAttrMutex.Unlock()
	} else {
		dailyAttrMutex.Lock()
		dailyAttrs[today][uid] = data
		dailyAttrMutex.Unlock()
	}

	return
}

// Set 更新每日用户属性，直接更新内存，再异步更新 pika
func Set(uid int64, field string, val interface{}) {
	initDailyAttr(uid)
	today := util.DateYmd()

	dailyAttrMutex.Lock()
	dailyAttrs[today][uid][field] = val
	dailyAttrMutex.Unlock()

	go func(val interface{}) { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHSet(ctx, fmt.Sprintf("daily_%s_%d", today, uid), field, json.Marshal(val))
		if err != nil {
			log.Errorf(ctx, -1, "Set failed for uid=%d field=%s, val=%v, err=%v", uid, field, val, err)
		}
	}(val)
}

// Incr 每日属性+1
func Incr(uid int64, field string, incr int64) int64 {
	initDailyAttr(uid)
	today := util.DateYmd()

	dailyAttrMutex.Lock()
	var num int64
	if _, ok := dailyAttrs[today][uid][field]; !ok {
		num = incr
	} else {
		num = dailyAttrs[today][uid].GetInt64E(field) + incr
	}

	dailyAttrs[today][uid][field] = num
	dailyAttrMutex.Unlock()

	go func(incr int64) { // 异步更新 pika
		ctx := context.Background()
		_, err := repo.RedisHIncrBy(ctx, fmt.Sprintf("daily_%s_%d", today, uid), field, incr)
		if err != nil {
			log.Errorf(ctx, -1, "Incr failed for uid=%d field=%s, val=%d, err=%v", uid, field, incr, err)
		}
	}(incr)

	return num
}

// Get 获取每日用户属性，直接从内存获取
func Get(uid int64, field string) (interface{}, bool) {
	initDailyAttr(uid)

	dailyAttrMutex.RLock()
	val, exists := dailyAttrs[util.DateYmd()][uid][field]
	dailyAttrMutex.RUnlock()
	return val, exists
}

// GetAll 获取每日用户所有属性
func GetAll(uid int64) types.Map {
	initDailyAttr(uid)

	dailyAttrMutex.RLock()
	all, _ := dailyAttrs[util.DateYmd()][uid]
	dailyAttrMutex.RUnlock()
	return types.CopyMap(all)
}

// Del 删除用户属性
func Del(uid int64, field string) {
	initDailyAttr(uid)
	today := util.DateYmd()

	dailyAttrMutex.Lock()
	if _, ok := dailyAttrs[today][uid][field]; ok {
		delete(dailyAttrs[today][uid], field)
	}
	dailyAttrMutex.Unlock()

	go func() { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHDel(ctx, fmt.Sprintf("daily_%s_%d", today, uid), field)
		if err != nil {
			log.Errorf(ctx, -1, "Del failed for uid=%d field=%s, err=%v", uid, field, err)
		}
	}()
	return
}

// IncrByPrefix 根据 prefix 每日属性+1
func IncrByPrefix(uid int64, prefix string, id interface{}, incr int64) int64 {
	return Incr(uid, prefix+types.ToString(id), incr)
}

// GetAllByPrefix 根据 prefix 获取每日用户相应属性
func GetAllByPrefix(uid int64, prefix string) types.Map {
	initDailyAttr(uid)

	dailyAttrMutex.RLock()
	all, _ := dailyAttrs[util.DateYmd()][uid]
	dailyAttrMutex.RUnlock()

	l := len(prefix)

	ret := types.Map{}
	for k, v := range all {
		if strings.Index(k, prefix) == 0 {
			ret[k[l:]] = v
		}
	}

	return ret
}

// HmGetByPrefix 根据 prefix 获取每日用户多个属性
func HmGetByPrefix(uid int64, prefix string, ids interface{}) types.Map {
	initDailyAttr(uid)

	dailyAttrMutex.RLock()
	all, _ := dailyAttrs[util.DateYmd()][uid]
	dailyAttrMutex.RUnlock()

	fieldMap := map[string]string{}

	idStrArr, _ := types.ToStringArray(ids)
	if len(idStrArr) == 0 {
		return types.Map{}
	}

	for _, id := range idStrArr {
		fieldMap[prefix+id] = id
	}

	ret := types.Map{}
	for field, id := range fieldMap {
		if val, ok := all[field]; ok {
			ret[id] = val
		}
	}

	return ret
}

// HmSetByPrefix 根据 prefix 设置多个属性
func HmSetByPrefix(uid int64, prefix string, ids types.Map) {
	initDailyAttr(uid)

	datas := make(types.Map, len(ids))
	for k, v := range ids {
		datas[prefix+k] = v
	}

	today := util.DateYmd()
	dailyAttrMutex.Lock()
	for k, v := range datas {
		dailyAttrs[today][uid][k] = v
	}
	dailyAttrMutex.Unlock()

	go func(datas types.Map) { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHMSet(ctx, fmt.Sprintf("daily_%s_%d", today, uid), datas)
		if err != nil {
			log.Errorf(ctx, -1, "HmSetByPrefix failed for uid=%d field=%v, err=%v", uid, datas, err)
		}
	}(datas)
}

// GetByPrefix 根据 prefix 获取用户属性
func GetByPrefix(uid int64, prefix string, id interface{}) (interface{}, bool) {
	field := prefix + types.ToString(id)
	return Get(uid, field)
}

// SetByPrefix 根据 prefix 设置用户属性
func SetByPrefix(uid int64, prefix string, id interface{}, val interface{}) {
	field := prefix + types.ToString(id)
	Set(uid, field, val)
}

// DelAllByPrefix 根据 prefix 删除用户相应所有属性
func DelAllByPrefix(uid int64, prefix string) {
	initDailyAttr(uid)
	today := util.DateYmd()

	dailyAttrMutex.Lock()
	all, _ := dailyAttrs[today][uid]

	delFields := []interface{}{}
	for k := range all {
		if strings.Index(k, prefix) == 0 {
			delete(dailyAttrs[today][uid], k)
			delFields = append(delFields, k)
		}
	}
	dailyAttrMutex.Unlock()

	go func() { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHDel(ctx, fmt.Sprintf("daily_%s_%d", today, uid), delFields...)
		if err != nil {
			log.Errorf(ctx, -1, "Del failed for uid=%d field=%v, err=%v", uid, delFields, err)
		}
	}()
}

// HDelByPrefix 根据 prefix 删除用户多个属性
func HDelByPrefix(uid int64, prefix string, ids interface{}) {
	initDailyAttr(uid)
	today := util.DateYmd()

	idStrArr, _ := types.ToArray(ids)
	if len(idStrArr) == 0 {
		return
	}

	dailyAttrMutex.Lock()
	for _, id := range idStrArr {
		delete(dailyAttrs[today][uid], prefix+types.ToString(id))
		idStrArr = append(idStrArr, prefix+types.ToString(id))
	}
	dailyAttrMutex.Unlock()

	go func() { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHDel(ctx, fmt.Sprintf("daily_%s_%d", today, uid), idStrArr...)
		if err != nil {
			log.Errorf(ctx, -1, "HDelSetDailyAttrByPrefix failed for uid=%d field=%v, err=%v", uid, idStrArr, err)
		}
	}()
}

// DelByPrefix 根据 prefix 删除用户单个属性
func DelByPrefix(uid int64, prefix string, id interface{}) {
	Del(uid, prefix+types.ToString(id))
}
