// 公会系统逻辑模块 - 公会等级、技能、任务、活跃度、公会副本、公会战等
package logic

import (
	"context"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/repo/info"
	"server_golang/repo/table"
)

// GuildRecordReward 公会战绩奖励
type GuildRecordReward struct {
	ID     int            `json:"id"`
	Rank   util.MinMax    `json:"rank"`
	Reward []util.TypeNum `json:"reward"`
}

// GuildCopyReward 对应 guild_copy_reward 表
type GuildCopyReward struct {
	Id                 int            `orm:"id,int,omitempty" json:"id"`
	RewardType         int            `orm:"reward_type,int" json:"reward_type"`
	Rank               util.MinMax    `orm:"rank,varchar" json:"rank"`
	Reward             []util.TypeNum `orm:"reward,varchar" json:"reward"`
	ItemPerRangeNumber []util.TypeNum `orm:"item_per_range_number,varchar" json:"item_per_range_number"`
}

// GuildActiveAttr 对应 guild_active_attr 表
type GuildActiveAttr struct {
	ActiveLv       int            `orm:"active_lv,int" json:"active_lv"`
	ActiveNum      int            `orm:"active_num,int" json:"active_num"`
	TotalActiveNum int            `orm:"total_active_num,int" json:"total_active_num"`
	ActiveAttr     []NumPair      `orm:"active_attr,varchar" json:"active_attr"`
	ActiveReward   []util.TypeNum `orm:"active_reward,varchar" json:"active_reward"`
}

// 公会全局数据
var (
	// 公会等级数据 guild_lv => data
	GuildLvDatas map[int]*table.GuildMemberLimit
	// 公会技能数据 profession_type => attr_key => data
	GuildSkillDatas map[int]map[int]*table.GuildSkillAttr
	// 公会技能升级消耗 profession_type => lv => attr_key => consume
	GuildSkillConsumeDatas map[int]map[int]map[int][]util.TypeNum
	// 公会任务数据 task_id => data
	GuildTaskDatas map[int]table.GuildTask
	// 公会活跃度属性数据 active_lv => data
	GuildActiveAttrDatas map[int]GuildActiveAttr
	// 公会副本奖励数据 reward_type => []data
	GuildCopyRewardDatas map[int][]GuildCopyReward
	// 公会副本排行奖励索引 rank => index
	GuildCopyRankRewardDatas map[int]int
	// 公会战绩计算数据 [combat_rank][star] => {star, zhanji}
	GuildRecordCalculateDatas map[int]map[int]map[string]int
	// 公会战绩奖励数据
	GuildRecordRewardDatas []GuildRecordReward
	// 公会Boss配置 chapter => data
	GuildBossConfigDatas map[int]*table.GuildBossConfig
	// 公会战场次
	GuildFightChangci int
)

// InitGuild 初始化公会数据，从MySQL加载到内存
func InitGuild(ctx context.Context) {
	if GuildFightChangci <= 0 {
		GuildFightChangci = 1
		log.Infof(ctx, "guild_fight_changci 未设置，使用默认值 1")
	}

	// 公会等级数据
	GuildLvDatas = make(map[int]*table.GuildMemberLimit)
	gmlRows, err := info.GetAllGuildMemberLimit(ctx)
	if err != nil || len(gmlRows) == 0 {
		log.Errorf(ctx, 0, "init guild error")
	} else {
		for _, data := range gmlRows {
			GuildLvDatas[data.GuildLv] = data
		}
	}

	// 公会技能属性数据
	GuildSkillDatas = make(map[int]map[int]*table.GuildSkillAttr)
	rows, err := info.GetAllGuildSkillAttr(ctx)
	if err != nil || len(rows) == 0 {
		log.Errorf(ctx, 0, "init guild_skill_attr error")
	} else {
		for _, v := range rows {
			professionType := v.ProfessionType
			attrKey := v.AttrKey
			if GuildSkillDatas[professionType] == nil {
				GuildSkillDatas[professionType] = make(map[int]*table.GuildSkillAttr)
			}
			GuildSkillDatas[professionType][attrKey] = v
		}
	}

	// 技能升级消费表
	GuildSkillConsumeDatas = make(map[int]map[int]map[int][]util.TypeNum)
	skillConsumeRows, err := info.GetAllGuildskillConsume(ctx)
	if err != nil || len(skillConsumeRows) == 0 {
		log.Errorf(ctx, 0, "init guild_skill_consume_datas error")
	} else {
		for _, v := range skillConsumeRows {
			professionType := v.ProfessionalType
			lv := v.Lv
			attrKey := v.AttrKey
			gongxianConsume := util.ToTypeNums(v.GongxianConsume)

			if GuildSkillConsumeDatas[professionType] == nil {
				GuildSkillConsumeDatas[professionType] = make(map[int]map[int][]util.TypeNum)
			}
			if GuildSkillConsumeDatas[professionType][lv] == nil {
				GuildSkillConsumeDatas[professionType][lv] = make(map[int][]util.TypeNum)
			}
			GuildSkillConsumeDatas[professionType][lv][attrKey] = gongxianConsume
		}
	}

	// 公会任务数据
	GuildTaskDatas = make(map[int]table.GuildTask)
	taskRows, err := info.GetAllGuildTask(ctx)
	if err != nil || len(taskRows) == 0 {
		log.Errorf(ctx, 0, "init guild_task error")
	} else {
		for _, data := range taskRows {
			taskID := types.ToIntE(data.TaskId)
			GuildTaskDatas[taskID] = *data
		}
	}

	// 公会活跃度属性数据
	GuildActiveAttrDatas = make(map[int]GuildActiveAttr)
	activeAttrRows, err := info.GetAllGuildActiveAttr(ctx)
	if err != nil || len(activeAttrRows) == 0 {
		log.Errorf(ctx, 0, "init guild_active_attr error")
	} else {
		totalActiveNum := 0
		for _, v := range activeAttrRows {
			activeLv := v.ActiveLv
			totalActiveNum += v.ActiveNum

			activeAttr := util.ToTypeNums(v.ActiveAttr)
			activeReward := util.ToTypeNums(v.ActiveReward)

			GuildActiveAttrDatas[activeLv] = GuildActiveAttr{
				ActiveLv:       activeLv,
				ActiveNum:      v.ActiveNum,
				TotalActiveNum: totalActiveNum,
				ActiveAttr: []NumPair{
					{Type: "hp", Num: activeAttr[0].Num},
					{Type: "atk", Num: activeAttr[1].Num},
				},
				ActiveReward: activeReward,
			}
		}
	}

	// 公会副本奖励数据
	GuildCopyRewardDatas = make(map[int][]GuildCopyReward)
	copyRewardRows, err := info.GetAllGuildCopyReward(ctx)
	if err != nil || len(copyRewardRows) == 0 {
		log.Errorf(ctx, 0, "init guild_copy_reward error")
	} else {
		for _, v := range copyRewardRows {
			rewardType := v.RewardType
			GuildCopyRewardDatas[rewardType] = append(GuildCopyRewardDatas[rewardType], GuildCopyReward{
				Id:                 v.Id,
				RewardType:         rewardType,
				Rank:               util.ToMinMax(v.Rank),
				Reward:             util.ToTypeNums(v.Reward),
				ItemPerRangeNumber: util.ToTypeNums(v.ItemPerRangeNumber),
			})
		}
	}

	// 排行奖励索引
	GuildCopyRankRewardDatas = make(map[int]int)
	rankRewardRows, err := info.GetGuildCopyRewardByRewardType(ctx, 4)
	if err != nil || len(rankRewardRows) == 0 {
		log.Errorf(ctx, 0, "init guild_copy_rank_reward error")
	} else {
		j := 0
		for _, v := range rankRewardRows {
			var rankMin, rankMax int
			rankVal := util.ToMinMax(v.Rank)
			rankMin = int(rankVal.Min)
			rankMax = int(rankVal.Max)
			for i := rankMin; i <= rankMax; i++ {
				GuildCopyRankRewardDatas[i] = j
			}
			j++
		}
	}

	// 公会战绩计算数据
	GuildRecordCalculateDatas = make(map[int]map[int]map[string]int)
	recordCalcRows, err := info.GetAllRecordCalculate(ctx)
	if err != nil || len(recordCalcRows) == 0 {
		log.Errorf(ctx, 0, "init record_calculate error")
	} else {
		for _, v := range recordCalcRows {
			combatRank := v.CombatRank - 1
			star := v.Star - 1
			zhanji := v.RecordNum
			if GuildRecordCalculateDatas[combatRank] == nil {
				GuildRecordCalculateDatas[combatRank] = make(map[int]map[string]int)
			}
			GuildRecordCalculateDatas[combatRank][star] = map[string]int{"star": star + 1, "zhanji": zhanji}
		}
	}

	// 公会战绩奖励数据
	GuildRecordRewardDatas = make([]GuildRecordReward, 0)
	recordRewardRows, err := info.GetAllRecordReward(ctx)
	if err != nil || len(recordRewardRows) == 0 {
		log.Errorf(ctx, 0, "init record_reward error")
		return
	}
	for _, data := range recordRewardRows {
		GuildRecordRewardDatas = append(GuildRecordRewardDatas, GuildRecordReward{
			ID:     data.Id,
			Rank:   util.ToMinMax(data.Rank),
			Reward: util.ToTypeNums(data.Reward),
		})
	}

	// 公会Boss配置
	GuildBossConfigDatas = make(map[int]*table.GuildBossConfig)
	bossConfigRows, err := info.GetAllGuildBossConfig(ctx)
	if err != nil || len(bossConfigRows) == 0 {
		log.Errorf(ctx, 0, "init guild_boss_config error")
	} else {
		for _, data := range bossConfigRows {
			GuildBossConfigDatas[data.Chapter] = data
		}
	}
}

// SkillAttrType2Prop 技能属性转加成属性
// attr_type: 1生命上限，2攻击，3暴击率，4暴伤，5速度，6伤害加成，7控制，8抗暴，10免伤，11抗控,12防御,13法术伤害加成，14物理伤害加成
// prop_type: 1生命上限，2攻击，3防御，4速度，5暴击率，6爆伤，7抗爆，8免伤，9伤害加成，10法术伤害加成，11物理伤害加成，13控制，14抗控
func SkillAttrType2Prop(attrType int) int {
	switch attrType {
	case 3:
		return 5
	case 4:
		return 6
	case 5:
		return 4
	case 6:
		return 9
	case 7:
		return 13
	case 8:
		return 7
	case 10:
		return 8
	case 11:
		return 14
	case 12:
		return 3
	case 13:
		return 10
	case 14:
		return 11
	default:
		return attrType
	}
}

// GetCopyChapterHP 获取公会副本章节Boss血量
func GetCopyChapterHP(chapter int) int {
	if cfg, ok := GuildBossConfigDatas[chapter]; ok {
		return int(cfg.BossHp)
	}
	return 0
}

// GetGuildCopyRewards 获取公会副本奖励
func GetGuildCopyRewards(rewardType int, lv int) []*GuildCopyReward {
	rewards := make([]*GuildCopyReward, 0)
	for _, v := range GuildCopyRewardDatas[rewardType] {
		// 复制一份避免修改原数据
		copied := v.Clone()
		rewardList := copied.Reward
		rewardTmp := make([]util.TypeNum, len(rewardList))
		copy(rewardTmp, rewardList)

		if itemPer := copied.ItemPerRangeNumber; len(itemPer) >= 2 && len(rewardTmp) >= 2 {
			rewardTmp[0].Num += itemPer[0].Num * lv
			rewardTmp[1].Num += itemPer[1].Num * lv
		}

		copied.Reward = rewardTmp
		rewards = append(rewards, copied)
	}
	return rewards
}

// GetCopyRankRewards 获取公会副本排行奖励
func GetCopyRankRewards(rank int, lv int) []util.TypeNum {
	index, ok := GuildCopyRankRewardDatas[rank]
	if !ok {
		return nil
	}
	rewardDatas := GuildCopyRewardDatas[4]
	if index >= len(rewardDatas) {
		return nil
	}
	rewards := rewardDatas[index]
	rewardList := rewards.Reward
	rewardTmp := make([]util.TypeNum, len(rewardList))
	copy(rewardTmp, rewardList)

	if itemPer := rewards.ItemPerRangeNumber; len(itemPer) >= 2 && len(rewardTmp) >= 2 {
		rewardTmp[0].Num += itemPer[0].Num * lv
		rewardTmp[1].Num += itemPer[1].Num * lv
	}

	return rewardTmp
}

// GetActiveRewards 获取公会活跃度奖励列表
func GetActiveRewards() []types.Map {
	ret := make([]types.Map, 0)
	for activeLv, data := range GuildActiveAttrDatas {
		ret = append(ret, types.Map{
			"lv":     activeLv,
			"reward": data.ActiveReward,
		})
	}
	return ret
}

// GetActiveLv 根据公会活跃值获取等级信息
func GetActiveLv(guildActive int) types.Map {
	currentLv := 0
	for activeLv, data := range GuildActiveAttrDatas {
		if guildActive < data.TotalActiveNum {
			currentLv = activeLv - 1
			break
		}
	}
	nextLv := currentLv + 1

	ret := types.Map{
		"current_lv":   currentLv,
		"current_data": GuildActiveAttrDatas[currentLv],
		"next_lv":      nextLv,
		"next_data":    GuildActiveAttrDatas[nextLv],
	}
	return ret
}

// GetGuildFightChangci 获取当前公会战场次（对外接口）
func GetGuildFightChangci() int {
	return GuildFightChangci
}

// GetGuildRecordCalDatas 获取战绩计算数据（按位置返回一维数组）
func GetGuildRecordCalDatas() []int {
	// 返回 zhanji 值的列表，按位置顺序
	result := make([]int, 0)
	for combatRank := 0; combatRank < len(GuildRecordCalculateDatas); combatRank++ {
		for star := 0; star < len(GuildRecordCalculateDatas[combatRank]); star++ {
			if data, ok := GuildRecordCalculateDatas[combatRank][star]; ok {
				result = append(result, data["zhanji"])
			}
		}
	}
	if len(result) == 0 {
		// 默认值
		for i := 0; i < 20; i++ {
			result = append(result, (i+1)*10)
		}
	}
	return result
}

func (g *GuildCopyReward) Clone() *GuildCopyReward {
	rewards := make([]util.TypeNum, len(g.Reward))
	copy(rewards, g.Reward)

	itemPerRangeNumber := make([]util.TypeNum, len(g.ItemPerRangeNumber))
	copy(itemPerRangeNumber, g.ItemPerRangeNumber)
	return &GuildCopyReward{
		Id:                 g.Id,
		RewardType:         g.RewardType,
		Rank:               util.MinMax{Min: g.Rank.Min, Max: g.Rank.Max},
		Reward:             rewards,
		ItemPerRangeNumber: itemPerRangeNumber,
	}
}
