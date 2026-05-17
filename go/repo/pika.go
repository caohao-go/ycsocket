package repo

import (
	"context"

	extredis "git.code.oa.com/pcg-csd/trpc-ext/redis"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
)

// pika 返回 trpc_go.yaml 中配置的指定服务名的 pika 客户端
func pika(name ...string) extredis.RedisHelper {
	var str = config.Pika
	if len(name) > 0 {
		str = name[0]
	}
	return extredis.NewHelperWithDefaultCodec(str)
}

// ---- pika 便捷辅助方法 ----

func RedisGet(ctx context.Context, key string) (string, error) {
	client := pika()
	if client == nil {
		return "", nil
	}
	var result string
	err := client.Get(ctx, key, &result)
	if extredis.IsNil(err) {
		return "", nil
	}
	return result, err
}

func RedisSet(ctx context.Context, key string, value interface{}, ttl ...int) error {
	client := pika()
	if client == nil {
		return nil
	}
	if len(ttl) > 0 && ttl[0] > 0 {
		return client.SetEx(ctx, key, json.Marshal(value), ttl[0])
	}

	return client.Set(ctx, key, json.Marshal(value))
}

func RedisDel(ctx context.Context, keys ...string) error {
	client := pika()
	if client == nil {
		return nil
	}
	_, err := client.Del(ctx, keys...)
	return err
}

func RedisHGet(ctx context.Context, key string, field interface{}) (string, error) {
	client := pika()
	if client == nil {
		return "", nil
	}

	var result string
	err := client.HGet(ctx, key, json.Marshal(field), &result)
	if extredis.IsNil(err) {
		return "", nil
	}
	return result, err
}

func RedisHSet(ctx context.Context, key string, values ...interface{}) error {
	client := pika()
	if client == nil {
		return nil
	}
	if len(values) >= 2 {
		return client.HSet(ctx, key,
			json.Marshal(values[0]), json.Marshal(values[1]), values[2:]...)
	}
	return nil
}

func RedisHGetAll(ctx context.Context, key string) (types.Map, error) {
	client := pika()
	if client == nil {
		return nil, nil
	}
	result, err := client.HGetAll(ctx, key)
	if extredis.IsNil(err) {
		return nil, nil
	}

	if result == nil {
		return nil, nil
	}

	ret := make(types.Map, len(result))
	for k, v := range result {
		ret[k] = v
	}
	return ret, err
}

func RedisHDel(ctx context.Context, key string, fields ...interface{}) error {
	client := pika()
	if client == nil {
		return nil
	}
	args := make([]interface{}, len(fields))
	for i, f := range fields {
		args[i] = json.Marshal(f)
	}
	_, err := client.HDel(ctx, key, args...)
	return err
}

func RedisExpire(ctx context.Context, key string, ttl int) error {
	client := pika()
	if client == nil {
		return nil
	}
	return client.Expire(ctx, key, ttl)
}

func RedisZAdd(ctx context.Context, key string, score float64, member interface{}) error {
	client := pika()
	if client == nil {
		return nil
	}
	// ZAdd 参数顺序: member, score, member, score ...
	_, err := client.ZAdd(ctx, key, json.Marshal(member), score)
	return err
}

func RedisIncr(ctx context.Context, key string) (int64, error) {
	client := pika()
	if client == nil {
		return 0, nil
	}
	return client.Incr(ctx, key)
}

func RedisRPop(ctx context.Context, key string) (string, error) {
	client := pika()
	if client == nil {
		return "", nil
	}
	var result string
	isEmpty, err := client.RPop(ctx, key, &result)
	if isEmpty || extredis.IsNil(err) {
		return "", nil
	}
	return result, err
}

func RedisLPush(ctx context.Context, key string, values ...interface{}) error {
	client := pika()
	if client == nil {
		return nil
	}
	_, err := client.LPush(ctx, key, values...)
	return err
}

func RedisHMGet(ctx context.Context, key string, fields ...string) ([]interface{}, error) {
	client := pika()
	if client == nil {
		return nil, nil
	}
	// HmGet 返回 map[string]string，转换为 []interface{} 保持上层兼容
	mapResult, err := client.HmGet(ctx, key, fields)
	if err != nil {
		if extredis.IsNil(err) {
			return make([]interface{}, len(fields)), nil
		}
		return nil, err
	}
	result := make([]interface{}, len(fields))
	for i, f := range fields {
		if v, ok := mapResult[f]; ok {
			result[i] = v
		}
	}
	return result, nil
}

func RedisHMSet(ctx context.Context, key string, fieldVal types.Map) error {
	client := pika()
	if client == nil {
		return nil
	}

	return client.HmSet(ctx, key, fieldVal)
}

func RedisIncrBy(ctx context.Context, key string, value int64) (int64, error) {
	client := pika()
	if client == nil {
		return 0, nil
	}
	return client.IncrBy(ctx, key, int(value))
}

func RedisHIncrBy(ctx context.Context, key string, field interface{}, incr int64) (int64, error) {
	client := pika()
	if client == nil {
		return 0, nil
	}
	return client.HIncrBy(ctx, key, json.Marshal(field), int(incr))
}

func RedisHKeys(ctx context.Context, key string) ([]string, error) {
	client := pika()
	if client == nil {
		return nil, nil
	}
	return client.HKeys(ctx, key)
}

func RedisHExists(ctx context.Context, key string, field interface{}) (bool, error) {
	client := pika()
	if client == nil {
		return false, nil
	}
	return client.HExists(ctx, key, json.Marshal(field))
}

func RedisZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	client := pika()
	if client == nil {
		return nil, nil
	}
	var members []string
	err := client.ZRevRange(ctx, key, int(start), int(stop), &members)
	if extredis.IsNil(err) {
		return nil, nil
	}
	return members, err
}

func RedisZCard(ctx context.Context, key string) (int64, error) {
	client := pika()
	if client == nil {
		return 0, nil
	}
	n, err := client.ZCard(ctx, key)
	return int64(n), err
}

// ScoreMember 有序的分数-成员对，保持 Redis ZSet 返回的排序
type ScoreMember struct {
	Member string
	Score  float64
}

func RedisZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ScoreMember, error) {
	client := pika()
	if client == nil {
		return nil, nil
	}
	var members []string
	scores, err := client.ZRevRangeWithScore(ctx, key, int(start), int(stop), &members)
	if err != nil {
		if extredis.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}
	result := make([]ScoreMember, 0, len(members))
	for i, m := range members {
		if i < len(scores) {
			result = append(result, ScoreMember{Member: m, Score: scores[i]})
		}
	}
	return result, nil
}

func RedisZRevRank(ctx context.Context, key string, member interface{}) (int64, error) {
	client := pika()
	if client == nil {
		return -1, nil
	}
	rank, err := client.ZRevRank(ctx, key, json.Marshal(member))
	if extredis.IsNil(err) {
		return -1, nil
	}
	return int64(rank), err
}

func RedisZScore(ctx context.Context, key string, member interface{}) (float64, error) {
	client := pika()
	if client == nil {
		return 0, nil
	}
	score, err := client.ZScore(ctx, key, json.Marshal(member))
	if extredis.IsNil(err) {
		return 0, nil
	}
	return score, err
}

func RedisZRem(ctx context.Context, key string, members ...interface{}) error {
	client := pika()
	if client == nil {
		return nil
	}
	_, err := client.ZRem(ctx, key, members...)
	return err
}

func RedisZIncrBy(ctx context.Context, key string, member interface{}, increment float64) (float64, error) {
	client := pika()
	if client == nil {
		return 0, nil
	}
	// ZIncrBy(ctx, key, member, incr)
	return client.ZIncrBy(ctx, key, json.Marshal(member), increment)
}

func RedisRPush(ctx context.Context, key string, values ...interface{}) error {
	client := pika()
	if client == nil {
		return nil
	}
	_, err := client.RPush(ctx, key, values...)
	return err
}

func RedisLLen(ctx context.Context, key string) (int64, error) {
	client := pika()
	if client == nil {
		return 0, nil
	}
	n, err := client.LLen(ctx, key)
	return int64(n), err
}

func RedisLPop(ctx context.Context, key string) (string, error) {
	client := pika()
	if client == nil {
		return "", nil
	}
	var result string
	isEmpty, err := client.LPop(ctx, key, &result)
	if isEmpty || extredis.IsNil(err) {
		return "", nil
	}
	return result, err
}

func RedisLRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	client := pika()
	if client == nil {
		return nil, nil
	}
	// Helper 没有 LRange，使用 Do 直接执行
	reply, err := client.Do(ctx, "LRANGE", key, start, stop)
	if err != nil {
		if extredis.IsNil(err) {
			return nil, nil
		}
		return nil, err
	}
	return extredis.Strings(reply, nil)
}
