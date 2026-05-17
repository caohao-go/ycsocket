// 好友数据模块
package logic

import (
	"context"
	"math/rand"
	"time"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"

	"server_golang/common/types"
	"server_golang/repo/world"
)

// 好友全局数据
var (
	FriendsDatas          map[int64]types.Map
	FriendsLastUpdateTime int64
)

// InitFriends 初始化好友推荐数据
func InitFriends(ctx context.Context) {
	FriendsDatas = make(map[int64]types.Map)

	userIds, err := world.GetFriendsRecommendUserIds(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init friends error: %v", err)
		return
	}

	for _, uid := range userIds {
		FriendsDatas[uid] = types.Map{
			"user_id":     uid,
			"fight_point": rand.Intn(5555555-11111) + 11111,
		}
	}

	FriendsLastUpdateTime = time.Now().Unix()
}

// GetRandRecFriends 获取随机推荐好友
func GetRandRecFriends(ctx context.Context) []types.Map {
	if time.Now().Unix()-FriendsLastUpdateTime > 5 {
		FriendsLastUpdateTime = time.Now().Unix()
		go InitFriends(ctx)
	}

	if len(FriendsDatas) == 0 {
		return nil
	}

	keys := make([]int64, 0, len(FriendsDatas))
	for k := range FriendsDatas {
		keys = append(keys, k)
	}

	count := 5
	if len(keys) < count {
		count = len(keys)
	}

	// 随机选取
	rand.Shuffle(len(keys), func(i, j int) { keys[i], keys[j] = keys[j], keys[i] })

	ret := make([]types.Map, 0, count)
	for i := 0; i < count; i++ {
		ret = append(ret, FriendsDatas[keys[i]])
	}
	return ret
}
