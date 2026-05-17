// copys 包实现了副本系统
package logic

import (
	"context"
	"fmt"
	"math/rand"

	"server_golang/common/json"
	"server_golang/common/util"
	"server_golang/repo/info"
)

type CopyInfo struct {
	Lv      int             `json:"lv"`
	OpenLv  int             `json:"open_lv"`
	Time    []util.TypeNum  `json:"time"`
	Monster []*HeroBaseInfo `json:"monster"`
}

type CopyBossReward struct {
	Id         int            `orm:"id,int,omitempty" json:"id"`
	Reward     []util.TypeNum `orm:"reward,varchar" json:"reward"`
	RandNum    []int          `orm:"rand_num,varchar" json:"rand_num"`
	RandReward []util.TypeNum `orm:"rand_reward,varchar" json:"rand_reward"`
}

type RandomCountConfig struct {
	Id           int         `json:"id"`
	Probability  int         `json:"probability"`
	CollectionId int         `json:"collection_id"`
	CopyRange    util.MinMax `json:"copy_range"`
	Num          util.MinMax `json:"num"`
	TimeRange    util.MinMax `json:"time_range"`
}

// CopyData 副本数据
var (
	CopyDataMap        = make(map[int]*CopyInfo)
	UserLvDataMap      = make(map[int]int)
	CopyRewardMap      = make(map[int]map[string][]util.TypeNum)
	CopyRandRewards    = []RandomCountConfig{}
	CopyBossRewardData = make(map[int]*CopyBossReward)
)

// InitCopy 初始化副本数据
func InitCopy(ctx context.Context) {
	// 等级经验表
	lvRows, err := info.GetAllLvInfo(ctx)
	if err != nil {
		panic(fmt.Errorf("init copy error: %v", err))
	}

	for _, row := range lvRows {
		UserLvDataMap[row.Lv] = row.Exp
	}

	// 副本配置
	copyRows, err := info.GetAllCopy(ctx)
	if err != nil {
		panic(fmt.Errorf("init copy error: %v", err))
	}
	for _, row := range copyRows {
		lv := row.Lv
		data := CopyInfo{Lv: lv, OpenLv: row.OpenLv}
		data.Time = util.ToTypeNums(row.Time)

		// 解析 monster JSON 字符串为怪物对象数组
		monsterArr := util.ToTypeNums(row.Monster)
		opHeros := make([]*HeroBaseInfo, 0)
		for i, v := range monsterArr {
			heroID := v.Type
			pos := v.Num

			opHeros = append(opHeros, &HeroBaseInfo{
				Id:     i + 1,
				UserId: 1,
				HeroId: heroID,
				Star:   row.MonsterStar,
				Stage:  row.MonsterStage,
				Lv:     row.MonsterLv,
				Pos:    pos,
			})
		}
		data.Monster = opHeros

		CopyDataMap[lv] = &data
	}

	// 挂机奖励
	onhookRows, err := info.GetAllOnhookReward(ctx)
	if err != nil {
		panic(fmt.Errorf("init copy error: %v", err))
	}
	for _, row := range onhookRows {
		id := row.Id
		if CopyRewardMap[id] == nil {
			CopyRewardMap[id] = make(map[string][]util.TypeNum)
		}

		CopyRewardMap[id]["rewards_min"] = util.ToTypeNums(row.OnhookMin)
		CopyRewardMap[id]["rewards_hour"] = util.ToTypeNums(row.OnhookHour)
	}

	// Boss 奖励表
	bossRows, err := info.GetAllCopyBossReward(ctx)
	if err != nil {
		panic(fmt.Errorf("init copy error: %v", err))
	}
	for _, row := range bossRows {
		randNum := []int{}
		json.Unmarshal(row.RandNum, &randNum)
		CopyBossRewardData[row.Id] = &CopyBossReward{
			Id:         row.Id,
			Reward:     util.ToTypeNums(row.Reward),
			RandNum:    randNum,
			RandReward: util.ToTypeNums(row.RandReward),
		}
	}

	// 随机掉落配置（random_count_config）
	randRows, err := info.GetAllRandomCountConfig(ctx)
	if err != nil {
		panic(fmt.Errorf("init copy error: %v", err))
	}
	for _, row := range randRows {
		item := RandomCountConfig{
			Id:           row.Id,
			Probability:  row.Probability,
			CollectionId: row.CollectionId,
			CopyRange:    util.ToMinMax(row.CopyRange),
			TimeRange:    util.ToMinMax(row.TimeRange),
			Num:          util.ToMinMax(row.Num),
		}
		CopyRandRewards = append(CopyRandRewards, item)
	}
}

// GetLvUpdateExp 获取升级所需经验
func GetLvUpdateExp(lv int) int {
	return UserLvDataMap[lv+1]
}

// GetOpenLv 根据副本进度获取开放等级
func GetOpenLv(copy int) int {
	if data, ok := CopyDataMap[copy]; ok {
		return data.OpenLv
	}
	return 0
}

// GetCopyRewards 获取挂机奖励（与 PHP Copys::getRewards 一致，含 VIP 加成）
func GetCopyRewards(copy int, vipLv int) map[string][]util.TypeNum {
	if copy > 650 {
		copy = 650
	}
	data := make(map[string][]util.TypeNum)

	if vipLv == 0 {
		// 无 VIP 加成，直接返回原始数据
		if rd, ok := CopyRewardMap[copy]; ok {
			data["min"] = rd["rewards_min"]
		}
	} else {
		// VIP 加成（与 PHP 一致）
		vipExpUp := float64(GetVipInfoLv(vipLv, "offline_exp")) / 100.0        // 人物经验加成
		vipCoinUp := float64(GetVipInfoLv(vipLv, "offline_coin")) / 100.0      // 金币加成
		vipExpHero := float64(GetVipInfoLv(vipLv, "offline_hero_exp")) / 100.0 // 英雄经验加成

		if rd, ok := CopyRewardMap[copy]; ok {
			rewards := make([]util.TypeNum, 0, len(rd["rewards_min"]))
			for _, minReward := range rd["rewards_min"] {
				t := minReward.Type
				num := minReward.Num
				switch t {
				case 3: // 人物经验
					num = int(float64(num) * (1 + vipExpUp))
				case 1: // 金币
					num = int(float64(num) * (1 + vipCoinUp))
				case 7: // 英雄经验
					num = int(float64(num) * (1 + vipExpHero))
				}
				rewards = append(rewards, util.TypeNum{Type: t, Num: num})
			}
			data["min"] = rewards
		}
	}

	if rd, ok := CopyRewardMap[copy]; ok {
		data["hour"] = rd["rewards_hour"]
	}
	if data["min"] == nil {
		data["min"] = []util.TypeNum{}
	}
	if data["hour"] == nil {
		data["hour"] = []util.TypeNum{}
	}
	return data
}

// GetCopyBossReward 获取 Boss 奖励
func GetCopyBossReward(copyLv int) []util.TypeNum {
	bossData := CopyBossRewardData[copyLv].Clone()
	if bossData == nil {
		return []util.TypeNum{}
	}

	// 固定奖励
	rewards := bossData.Reward

	// 随机奖励
	randNum := 0
	if len(bossData.RandNum) >= 2 {
		minN := bossData.RandNum[0]
		maxN := bossData.RandNum[1]
		if maxN > minN {
			randNum = minN + rand.Intn(maxN-minN+1)
		} else {
			randNum = minN
		}
	}

	if randNum > 0 {
		var randRewards = bossData.RandReward
		if randNum >= len(randRewards) {
			rewards = append(rewards, randRewards...)
		} else if len(randRewards) > 0 {
			perm := rand.Perm(len(randRewards))
			for i := 0; i < randNum && i < len(perm); i++ {
				rewards = append(rewards, randRewards[perm[i]])
			}
		}
	}

	return rewards
}

// GetRandRewards 获取挂机随机掉落奖励（与 PHP Copys::getRandRewards 一致）
func GetRandRewards(copy int, onHookHour int, onHookMin int) []util.TypeNum {
	data := make([]util.TypeNum, 0)
	timeRangeHookTemp := float64(onHookHour) + float64(onHookMin)/60.0
	timeRangeHook := timeRangeHookTemp
	if timeRangeHook >= 12 {
		timeRangeHook = 12
	}
	randProbli := rand.Intn(100) + 1

	for _, v := range CopyRandRewards {
		copyMin := int(v.CopyRange.Min)
		copyMax := int(v.CopyRange.Max)
		timeMin := v.TimeRange.Min
		timeMax := v.TimeRange.Max

		if copy >= copyMin && copy < copyMax {
			if timeRangeHook > timeMin && timeRangeHook <= timeMax {
				probability := v.Probability
				if randProbli <= probability {
					numMin := int(v.Num.Min)
					numMax := int(v.Num.Max)
					rewardCount := numMin
					if numMax > numMin {
						rewardCount = numMin + rand.Intn(numMax-numMin+1)
					}
					for i := 0; i < rewardCount; i++ {
						items := GetRandCollectionItem(v.CollectionId, 1)
						if len(items) > 0 {
							itemsID := items[0].ItemsId
							number := items[0].Number
							if itemsID > 0 {
								data = append(data, util.TypeNum{Type: itemsID, Num: number})
							}
						}
					}
					break
				}
				// 概率不命中直接返回空
				return data
			}
		}
	}

	return data
}

func (c *CopyBossReward) Clone() *CopyBossReward {
	rewards := make([]util.TypeNum, len(c.Reward))
	for k, v := range c.Reward {
		rewards[k] = v
	}
	randReward := make([]util.TypeNum, len(c.RandReward))
	for k, v := range c.RandReward {
		randReward[k] = v
	}
	randNum := make([]int, len(c.RandNum))
	for k, v := range c.RandNum {
		randNum[k] = v
	}
	return &CopyBossReward{
		Id:         c.Id,
		Reward:     rewards,
		RandNum:    randNum,
		RandReward: randReward,
	}
}
