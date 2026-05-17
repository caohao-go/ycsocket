// 合成数据模块
package logic

import (
	"context"
	"fmt"
	"math/rand"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/util"
	"server_golang/repo/info"
)

// MergeData 合成配置
type MergeData struct {
	FinalID     int `json:"final_id"`
	ConsumeCoin int `json:"consume_coin"`
}

// RuneMergeData 符文合成配置
type RuneMergeData struct {
	FinalID     int          `json:"final_id"`
	ConsumeCoin int          `json:"consume_coin"`
	MergeNum    int          `json:"merge_num"`
	SuccessRate int          `json:"success_rate"`
	FailGet     util.TypeNum `json:"fail_get"`
}

// 合成全局数据
var (
	// 基础合成 original_id => MergeData
	MergeDatas map[int]MergeData
	// 符文合成 original_id => merge_num => RuneMergeData
	RuneMergeDatas map[int]map[int]RuneMergeData
)

// InitMerge 初始化合成数据
func InitMerge(ctx context.Context) {
	MergeDatas = make(map[int]MergeData)
	RuneMergeDatas = make(map[int]map[int]RuneMergeData)

	// 基础合成
	rows, err := info.GetAllMergeConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("init merge_config error: %v", err))
	}

	for _, val := range rows {
		consumeCoin := util.ToTypeNums(string(val.Consume))
		MergeDatas[val.OriginalId] = MergeData{
			FinalID:     val.FinalId,
			ConsumeCoin: consumeCoin[0].Num,
		}
	}

	// 符文合成
	runeRows, err := info.GetAllRuneMerge(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init rune_merge error: %v", err)
	}
	for _, val := range runeRows {
		failGets := util.ToTypeNums(val.FailGet)
		var failGet util.TypeNum
		if len(failGets) > 0 {
			failGet = failGets[0]
		}

		if RuneMergeDatas[val.OriginalId] == nil {
			RuneMergeDatas[val.OriginalId] = make(map[int]RuneMergeData)
		}

		RuneMergeDatas[val.OriginalId][val.MergeNum] = RuneMergeData{
			FinalID:     val.FinalId,
			ConsumeCoin: val.MergeConsume,
			MergeNum:    val.MergeNum,
			SuccessRate: val.SuccessRate,
			FailGet:     failGet,
		}
	}
}

// MergeRateSuccess 合成成功判定
func MergeRateSuccess(rate int) bool {
	r := rand.Intn(100) + 1
	return r <= rate
}
