package week

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"git.code.oa.com/pcg-csd/trpc-ext/redis"
	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/config"
	"server_golang/repo"
)

var weekAttrs = map[string]map[int64]types.Map{}
var weekAttrMutex = sync.RWMutex{}

func init() {
	weekAttrs[util.DateW()] = map[int64]types.Map{}
	weekAttrs[util.NextDateW()] = map[int64]types.Map{}
}

// ClearLastWeek 每周一早上3点清除上周的数据
func ClearLastWeek(ctx context.Context) {
	weekAttrMutex.Lock()
	delete(weekAttrs, util.LastDateW())                 // 删除上周内存数据
	weekAttrs[util.NextDateW()] = map[int64]types.Map{} // 创建下周的内存 map
	weekAttrMutex.Unlock()
}

// initWeekAttr 将每周用户属性从 pika 加载到内存
func initWeekAttr(uid int64) {
	ctx := context.Background()
	thisWeek := util.DateW()

	weekAttrMutex.Lock()
	if _, ok := weekAttrs[thisWeek][uid]; ok {
		weekAttrMutex.Unlock()
		return
	}
	weekAttrMutex.Unlock()

	// 从pika加载到内存
	data, err := repo.RedisHGetAll(ctx, fmt.Sprintf("week_%s_%d", thisWeek, uid))
	if err != nil && !redis.IsNil(err) {
		panic(err)
	}

	if data == nil || len(data) == 0 {
		// 初始化用户属性
		err = repo.RedisHSet(ctx, fmt.Sprintf("week_%s_%d", thisWeek, uid), config.UserId, uid)

		if err != nil {
			panic(err)
		}

		_ = repo.RedisExpire(ctx, fmt.Sprintf("week_%s_%d", thisWeek, uid), 86400)

		weekAttrMutex.Lock()
		weekAttrs[thisWeek][uid] = types.Map{config.UserId: uid}
		weekAttrMutex.Unlock()
	} else {
		weekAttrMutex.Lock()
		weekAttrs[thisWeek][uid] = data
		weekAttrMutex.Unlock()
	}

	return
}

// Set 更新每周用户属性，直接更新内存，再异步更新 pika
func Set(uid int64, field string, val interface{}) {
	initWeekAttr(uid)
	thisWeek := util.DateW()

	weekAttrMutex.Lock()
	weekAttrs[thisWeek][uid][field] = val
	weekAttrMutex.Unlock()

	go func() { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHSet(ctx, fmt.Sprintf("week_%s_%d", thisWeek, uid), field, val)
		if err != nil {
			log.Errorf(ctx, -1, "Set failed for uid=%d field=%s, val=%s, err=%v", uid, field, val, err)
		}
	}()
}

// Incr 每周属性+1
func Incr(uid int64, field string, incr int64) int64 {
	initWeekAttr(uid)
	thisWeek := util.DateW()

	weekAttrMutex.Lock()
	var num int64
	if _, ok := weekAttrs[thisWeek][uid][field]; !ok {
		num = incr
	} else {
		num = weekAttrs[thisWeek][uid].GetInt64E(field) + incr
	}

	weekAttrs[thisWeek][uid][field] = num
	weekAttrMutex.Unlock()

	go func() { // 异步更新 pika
		ctx := context.Background()
		_, err := repo.RedisHIncrBy(ctx, fmt.Sprintf("week_%s_%d", thisWeek, uid), field, incr)
		if err != nil {
			log.Errorf(ctx, -1, "IncrWeekAttr failed for uid=%d field=%s, val=%d, err=%v", uid, field, incr, err)
		}
	}()

	return num
}

// Get 获取每周用户属性，直接从内存获取
func Get(uid int64, field string) (interface{}, bool) {
	initWeekAttr(uid)

	weekAttrMutex.Lock()
	val, exists := weekAttrs[util.DateW()][uid][field]
	weekAttrMutex.Unlock()
	return val, exists
}

// GetAll 获取每周用户所有属性
func GetAll(uid int64) types.Map {
	initWeekAttr(uid)

	weekAttrMutex.Lock()
	all, _ := weekAttrs[util.DateW()][uid]
	weekAttrMutex.Unlock()
	return types.CopyMap(all)
}

// Del 删除用户属性
func Del(uid int64, field string) {
	initWeekAttr(uid)
	thisWeek := util.DateW()

	weekAttrMutex.Lock()
	if _, ok := weekAttrs[thisWeek][uid][field]; ok {
		delete(weekAttrs[thisWeek][uid], field)
	}
	weekAttrMutex.Unlock()

	go func() { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHDel(ctx, fmt.Sprintf("week_%s_%d", thisWeek, uid), field)
		if err != nil {
			log.Errorf(ctx, -1, "Del failed for uid=%d field=%s, err=%v", uid, field, err)
		}
	}()
	return
}

// GetAllByPrefix 根据 prefix 获取每周用户相应属性
func GetAllByPrefix(uid int64, prefix string) types.Map {
	initWeekAttr(uid)

	weekAttrMutex.Lock()
	all, _ := weekAttrs[util.DateW()][uid]
	weekAttrMutex.Unlock()

	l := len(prefix)

	ret := types.Map{}
	for k, v := range all {
		if strings.Index(k, prefix) == 0 {
			ret[k[l:]] = v
		}
	}

	return ret
}

// HmGetByPrefix 根据 prefix 获取每周用户多个属性
func HmGetByPrefix(uid int64, prefix string, ids interface{}) types.Map {
	initWeekAttr(uid)

	weekAttrMutex.Lock()
	all, _ := weekAttrs[util.DateW()][uid]
	weekAttrMutex.Unlock()

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
	initWeekAttr(uid)

	datas := make(types.Map, len(ids))
	for k, v := range ids {
		datas[prefix+k] = v
	}

	thisWeek := util.DateW()
	weekAttrMutex.Lock()
	for k, v := range datas {
		weekAttrs[thisWeek][uid][k] = v
	}
	weekAttrMutex.Unlock()

	go func() { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHMSet(ctx, fmt.Sprintf("week_%s_%d", thisWeek, uid), datas)
		if err != nil {
			log.Errorf(ctx, -1, "HmSetWeekAttrByPrefix failed for uid=%d field=%v, err=%v", uid, datas, err)
		}
	}()
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
	initWeekAttr(uid)
	thisWeek := util.DateW()

	weekAttrMutex.Lock()
	all, _ := weekAttrs[thisWeek][uid]

	delFields := []interface{}{}
	for k := range all {
		if strings.Index(k, prefix) == 0 {
			delete(weekAttrs[thisWeek][uid], k)
			delFields = append(delFields, k)
		}
	}
	weekAttrMutex.Unlock()

	go func() { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHDel(ctx, fmt.Sprintf("week_%s_%d", thisWeek, uid), delFields...)
		if err != nil {
			log.Errorf(ctx, -1, "Del failed for uid=%d field=%v, err=%v", uid, delFields, err)
		}
	}()
}

// HDelByPrefix 根据 prefix 删除用户多个属性
func HDelByPrefix(uid int64, prefix string, ids interface{}) {
	initWeekAttr(uid)
	thisWeek := util.DateW()

	idStrArr, _ := types.ToArray(ids)
	if len(idStrArr) == 0 {
		return
	}

	weekAttrMutex.Lock()
	for _, id := range idStrArr {
		delete(weekAttrs[thisWeek][uid], prefix+types.ToString(id))
		idStrArr = append(idStrArr, prefix+types.ToString(id))
	}
	weekAttrMutex.Unlock()

	go func() { // 异步更新 pika
		ctx := context.Background()
		err := repo.RedisHDel(ctx, fmt.Sprintf("week_%s_%d", thisWeek, uid), idStrArr...)
		if err != nil {
			log.Errorf(ctx, -1, "HDelSetWeekAttrByPrefix failed for uid=%d field=%v, err=%v", uid, idStrArr, err)
		}
	}()
}

// DelByPrefix 根据 prefix 删除用户单个属性
func DelByPrefix(uid int64, prefix string, id interface{}) {
	Del(uid, prefix+types.ToString(id))
}
