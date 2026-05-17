// 副本奖励模块
package logic

import (
	"context"
	"time"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/util"
	"server_golang/repo/info"
)

// 副本奖励常量
const (
	RewardCopyIDPKDaily = 10001 // 每日排名奖励
	RewardCopyIDPKRank  = 10002 // 赛季排名奖励
	RewardCopyIDPKCount = 10003 // 次数奖励
	RewardCopyIDEndless = 20100 // 无尽试炼奖励
)

// CopyReward 对应 copy_reward 表
type CopyReward struct {
	Id        int            `orm:"id,int,omitempty" json:"id"`
	CopyId    int            `orm:"copy_id,int" json:"copy_id"`
	Seq       int            `orm:"seq,int" json:"seq"`
	RankCount util.MinMax    `orm:"rank_count,varchar" json:"rank_count"`
	Reward    []util.TypeNum `orm:"reward,varchar" json:"reward"`
	Refresh   int            `orm:"refresh,int" json:"refresh"`
}

// 副本奖励全局数据
var (
	CopyrewardDatas      map[int]map[int]*CopyReward // copy_id => seq => data
	CopyrewardDataExpire int64
)

// GetRewardByCopyID 根据copy_id获取奖励
func GetRewardByCopyID(ctx context.Context, copyID int) map[int]*CopyReward {
	if CopyrewardDatas == nil || time.Now().Unix() > CopyrewardDataExpire {
		getAllCopyRewardData(ctx)
	}
	return CopyrewardDatas[copyID]
}

func getAllCopyRewardData(ctx context.Context) {
	rows, err := info.GetAllCopyReward(ctx)
	if err != nil || len(rows) == 0 {
		log.Errorf(ctx, 0, "init copy_reward error")
		return
	}

	rets := make(map[int]map[int]*CopyReward)
	for _, row := range rows {
		data := CopyReward{}
		data.CopyId = row.CopyId
		data.Seq = row.Seq
		data.Refresh = row.Refresh
		data.RankCount = util.ToMinMax(row.RankCount)
		data.Reward = util.ToTypeNums(row.Reward)

		if rets[row.CopyId] == nil {
			rets[row.CopyId] = make(map[int]*CopyReward)
		}
		rets[row.CopyId][row.Seq] = &data
	}

	CopyrewardDatas = rets
	CopyrewardDataExpire = time.Now().Unix() + 3600
}
