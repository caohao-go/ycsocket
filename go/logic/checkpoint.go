// 关卡奖励数据模块
package logic

import (
	"context"
	"fmt"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/repo/info"
)

var (
	// 章节全局数据 id => data
	ChapterDatas map[int]types.Map
	// 关卡奖励数据 checkpoint_type => checkpoint_num => {rewards: []TypeNum}
	CheckpointRewardDatas map[int]map[int][]util.TypeNum
	// checkpoint_type => checkpoint_num => online_time
	CheckpointOnlineTime map[int]map[int]int
)

// InitChapter 初始化章节数据
func InitChapter(ctx context.Context) {
	ChapterDatas = make(map[int]types.Map)

	rows, err := info.GetAllChapter(ctx)
	if err != nil || len(rows) == 0 {
		log.Errorf(ctx, 0, "init chapter error")
		return
	}
	for _, data := range rows {
		var copyData []int
		_ = json.Unmarshal(data.Copy, &copyData)
		ChapterDatas[data.Id] = types.Map{
			"id":     data.Id,
			"name":   data.Name,
			"copy":   copyData,
			"map_id": data.MapId,
		}
	}
}

// InitCheckpointReward 初始化关卡奖励数据
func InitCheckpointReward(ctx context.Context) {
	CheckpointRewardDatas = make(map[int]map[int][]util.TypeNum)
	CheckpointOnlineTime = make(map[int]map[int]int)

	rows, err := info.GetAllCheckpointReward(ctx)
	if err != nil {
		panic(fmt.Errorf("init checkpoint_reward error: %v", err))
	}

	for _, cr := range rows {
		checkpointType := cr.CheckpointType
		checkpointNum := cr.CheckpointNum

		if CheckpointOnlineTime[checkpointType] == nil {
			CheckpointOnlineTime[checkpointType] = make(map[int]int)
		}
		CheckpointOnlineTime[checkpointType][checkpointNum] = cr.OnlineTime

		if CheckpointRewardDatas[checkpointType] == nil {
			CheckpointRewardDatas[checkpointType] = make(map[int][]util.TypeNum)
		}
		CheckpointRewardDatas[checkpointType][checkpointNum] = util.ToTypeNums(string(cr.Reward))
	}
}
