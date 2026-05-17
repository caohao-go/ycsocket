// VIP配置模块 - 充值奖励、月卡、基金、礼包、特权商店
package logic

import (
	"context"
	"time"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/repo/info"
	"server_golang/repo/table"
)

// ---- 常量/全局变量 ----

// 充值列表
var ChongzhiList = []int{6, 30, 68, 128, 198, 328, 648}

// 首充6元第1-3天
var Shouchong6Day1 = []util.TypeNum{{Type: 502001, Num: 50}, {Type: 21001, Num: 1}}
var Shouchong6Day2 = []util.TypeNum{{Type: 30108, Num: 1}, {Type: 30208, Num: 1}}
var Shouchong6Day3 = []util.TypeNum{{Type: 30308, Num: 1}, {Type: 30408, Num: 1}}

// 首充100元第1-3天
var Shouchong100Day1 = []util.TypeNum{{Type: 30111, Num: 1}, {Type: 30211, Num: 1}}
var Shouchong100Day2 = []util.TypeNum{{Type: 30311, Num: 1}, {Type: 21901, Num: 1}}
var Shouchong100Day3 = []util.TypeNum{{Type: 30411, Num: 1}, {Type: 21901, Num: 2}}

// 每日首充18元
var DayShouchong180Day1 = []util.TypeNum{{Type: 20201, Num: 100}, {Type: 21001, Num: 1}, {Type: 2, Num: 200}, {Type: 1, Num: 1000000}}

// 成长基金奖励
var JijinRewards = map[int][]util.TypeNum{
	10:  {{Type: 2, Num: 168}, {Type: 2, Num: 888}},
	20:  {{Type: 2, Num: 168}, {Type: 2, Num: 888}},
	30:  {{Type: 2, Num: 168}, {Type: 2, Num: 888}},
	40:  {{Type: 2, Num: 168}, {Type: 2, Num: 888}},
	50:  {{Type: 2, Num: 168}, {Type: 2, Num: 888}},
	60:  {{Type: 2, Num: 1288}, {Type: 2, Num: 888}},
	70:  {{Type: 2, Num: 1288}, {Type: 2, Num: 888}},
	80:  {{Type: 2, Num: 1288}, {Type: 2, Num: 888}},
	90:  {{Type: 2, Num: 1288}, {Type: 2, Num: 888}},
	100: {{Type: 2, Num: 1288}, {Type: 2, Num: 888}},
	110: {{Type: 2, Num: 2288}, {Type: 2, Num: 888}},
	120: {{Type: 2, Num: 2288}, {Type: 2, Num: 888}},
	130: {{Type: 2, Num: 2288}, {Type: 2, Num: 888}},
	140: {{Type: 2, Num: 2288}, {Type: 2, Num: 888}},
	150: {{Type: 2, Num: 2288}, {Type: 2, Num: 888}},
}

// 月度超值礼包限额
var YueduLibaoLimit = map[int]int{30: 1, 68: 1, 128: 2, 328: 3, 448: 3, 648: 3, 668: 3}
var YueduLibaoRewards = map[int][]util.TypeNum{
	30:  {{Type: 2, Num: 300}, {Type: 20201, Num: 888}, {Type: 21001, Num: 8}, {Type: 21301, Num: 5}},
	68:  {{Type: 2, Num: 680}, {Type: 21901, Num: 1}, {Type: 21001, Num: 10}, {Type: 21301, Num: 10}},
	128: {{Type: 2, Num: 1280}, {Type: 21901, Num: 3}, {Type: 21001, Num: 15}, {Type: 20701, Num: 2000}},
	328: {{Type: 2, Num: 3280}, {Type: 21901, Num: 7}, {Type: 21001, Num: 24}, {Type: 1, Num: 20000000}},
	448: {{Type: 2, Num: 4480}, {Type: 21901, Num: 9}, {Type: 21001, Num: 31}, {Type: 20701, Num: 30000000}},
	648: {{Type: 2, Num: 6480}, {Type: 21901, Num: 13}, {Type: 21001, Num: 42}, {Type: 50117, Num: 50}},
	668: {{Type: 2, Num: 6480}, {Type: 21901, Num: 13}, {Type: 21001, Num: 42}, {Type: 50118, Num: 50}},
}

// 每周限购礼包
var WeekLibaoLimit = map[int]int{30: 3, 68: 3, 128: 3, 328: 3, 448: 3, 648: 1}
var WeekLibaoRewards = map[int][]util.TypeNum{
	30:  {{Type: 2, Num: 300}, {Type: 50113, Num: 10}, {Type: 21101, Num: 5}, {Type: 21001, Num: 5}},
	68:  {{Type: 2, Num: 680}, {Type: 50113, Num: 25}, {Type: 20701, Num: 500}, {Type: 21001, Num: 8}},
	128: {{Type: 2, Num: 1280}, {Type: 21302, Num: 1}, {Type: 20701, Num: 1200}, {Type: 21001, Num: 15}},
	328: {{Type: 2, Num: 3280}, {Type: 20201, Num: 6000}, {Type: 1, Num: 8000000}, {Type: 21001, Num: 22}},
	448: {{Type: 2, Num: 4480}, {Type: 20701, Num: 10000}, {Type: 1, Num: 10000000}, {Type: 21001, Num: 31}},
	648: {{Type: 2, Num: 6480}, {Type: 40401, Num: 1}, {Type: 50113, Num: 50}, {Type: 21001, Num: 42}},
}

// 每日礼包
var DayLibaoLimit = map[int]int{1: 3, 6: 3, 12: 3}
var DayLibaoRewards = map[int][]util.TypeNum{
	1:  {{Type: 2, Num: 20}, {Type: 21301, Num: 3}, {Type: 1, Num: 100000}, {Type: 7, Num: 100000}},
	6:  {{Type: 2, Num: 120}, {Type: 21001, Num: 1}, {Type: 21101, Num: 6}, {Type: 7, Num: 200000}},
	12: {{Type: 2, Num: 240}, {Type: 21001, Num: 2}, {Type: 21101, Num: 8}, {Type: 7, Num: 300000}},
}

// 签到奖励（31天）
var QiandaoRewards []util.TypeNum

// VIP特权礼包
var TequanRewards = map[int][]util.TypeNum{
	0:  {{Type: 1, Num: 200000}, {Type: 7, Num: 100000}, {Type: 21301, Num: 2}},
	1:  {{Type: 20201, Num: 300}, {Type: 21301, Num: 3}, {Type: 1, Num: 2000000}},
	2:  {{Type: 50111, Num: 30}, {Type: 50112, Num: 30}, {Type: 7, Num: 2000000}},
	3:  {{Type: 21302, Num: 1}, {Type: 21901, Num: 1}, {Type: 1, Num: 4000000}},
	4:  {{Type: 50113, Num: 50}, {Type: 21901, Num: 1}, {Type: 7, Num: 4000000}},
	5:  {{Type: 504005, Num: 50}, {Type: 21901, Num: 1}, {Type: 1, Num: 6000000}},
	6:  {{Type: 502007, Num: 50}, {Type: 21901, Num: 2}, {Type: 40301, Num: 1}, {Type: 7, Num: 6000000}},
	7:  {{Type: 502007, Num: 50}, {Type: 21901, Num: 3}, {Type: 40301, Num: 1}, {Type: 1, Num: 8000000}},
	8:  {{Type: 502014, Num: 50}, {Type: 21901, Num: 5}, {Type: 30112, Num: 1}, {Type: 7, Num: 8000000}},
	9:  {{Type: 501012, Num: 50}, {Type: 21901, Num: 5}, {Type: 30412, Num: 1}, {Type: 1, Num: 10000000}},
	10: {{Type: 502014, Num: 50}, {Type: 40401, Num: 1}, {Type: 30312, Num: 1}, {Type: 7, Num: 10000000}},
	11: {{Type: 501012, Num: 50}, {Type: 50117, Num: 50}, {Type: 40401, Num: 1}, {Type: 30112, Num: 1}, {Type: 1, Num: 15000000}},
	12: {{Type: 50117, Num: 50}, {Type: 50118, Num: 50}, {Type: 502012, Num: 50}, {Type: 40401, Num: 1}, {Type: 20901, Num: 10}, {Type: 7, Num: 15000000}},
	13: {{Type: 50117, Num: 150}, {Type: 504008, Num: 50}, {Type: 40401, Num: 2}, {Type: 20901, Num: 10}, {Type: 1, Num: 20000000}},
}

// VIP特权礼包购买价格
var VipBuy = map[int][]util.TypeNum{
	0:  {{Type: 2, Num: 0}, {Type: 2, Num: 238}},
	1:  {{Type: 2, Num: 238}, {Type: 2, Num: 888}},
	2:  {{Type: 2, Num: 688}, {Type: 2, Num: 2880}},
	3:  {{Type: 2, Num: 988}, {Type: 2, Num: 4288}},
	4:  {{Type: 2, Num: 1888}, {Type: 2, Num: 8888}},
	5:  {{Type: 2, Num: 2288}, {Type: 2, Num: 10880}},
	6:  {{Type: 2, Num: 3888}, {Type: 2, Num: 18880}},
	7:  {{Type: 2, Num: 4188}, {Type: 2, Num: 20880}},
	8:  {{Type: 2, Num: 4688}, {Type: 2, Num: 23880}},
	9:  {{Type: 2, Num: 4888}, {Type: 2, Num: 26880}},
	10: {{Type: 2, Num: 6888}, {Type: 2, Num: 38880}},
	11: {{Type: 2, Num: 9888}, {Type: 2, Num: 58880}},
	12: {{Type: 2, Num: 15880}, {Type: 2, Num: 98880}},
	13: {{Type: 2, Num: 18880}, {Type: 2, Num: 118880}},
}

// 特权商城
var TequanShop = map[int][]util.TypeNum{
	1: {{Type: 20901, Num: 3}, {Type: 20201, Num: 1000}, {Type: 1, Num: 100000}},
	2: {{Type: 22102, Num: 1}, {Type: 21001, Num: 10}, {Type: 20201, Num: 1000}},
	3: {{Type: 22103, Num: 1}, {Type: 21001, Num: 5}, {Type: 20201, Num: 200}},
	4: {{Type: 22104, Num: 1}, {Type: 21901, Num: 1}, {Type: 21001, Num: 10}},
}

// 特权商城价格
var TequanShopShell = map[int][]util.TypeNum{
	3: {{Type: 2, Num: 980}},
	4: {{Type: 2, Num: 1980}},
}

// 月基金128
var Yue128JijinReward = map[int][]util.TypeNum{
	1: {{Type: 2, Num: 400}}, 2: {{Type: 20201, Num: 500}}, 3: {{Type: 21901, Num: 3}},
	4: {{Type: 2, Num: 200}}, 5: {{Type: 2, Num: 200}}, 6: {{Type: 2, Num: 500}},
	7: {{Type: 21302, Num: 1}}, 8: {{Type: 2, Num: 200}}, 9: {{Type: 2, Num: 200}},
	10: {{Type: 2, Num: 200}}, 11: {{Type: 2, Num: 200}}, 12: {{Type: 2, Num: 200}},
	13: {{Type: 2, Num: 500}}, 14: {{Type: 21302, Num: 3}}, 15: {{Type: 2, Num: 500}},
	16: {{Type: 2, Num: 200}}, 17: {{Type: 2, Num: 200}}, 18: {{Type: 2, Num: 200}},
	19: {{Type: 2, Num: 200}}, 20: {{Type: 2, Num: 500}}, 21: {{Type: 2, Num: 200}},
	22: {{Type: 2, Num: 200}}, 23: {{Type: 2, Num: 200}}, 24: {{Type: 2, Num: 200}},
	25: {{Type: 2, Num: 200}}, 26: {{Type: 2, Num: 200}}, 27: {{Type: 2, Num: 500}},
	28: {{Type: 50113, Num: 50}}, 29: {{Type: 2, Num: 1000}}, 30: {{Type: 21901, Num: 3}},
}

// 月基金328
var Yue328JijinReward = map[int][]util.TypeNum{
	1: {{Type: 2, Num: 800}}, 2: {{Type: 20201, Num: 1000}}, 3: {{Type: 21901, Num: 6}},
	4: {{Type: 2, Num: 400}}, 5: {{Type: 2, Num: 400}}, 6: {{Type: 2, Num: 1000}},
	7: {{Type: 21302, Num: 3}}, 8: {{Type: 2, Num: 400}}, 9: {{Type: 2, Num: 400}},
	10: {{Type: 2, Num: 400}}, 11: {{Type: 2, Num: 400}}, 12: {{Type: 2, Num: 400}},
	13: {{Type: 2, Num: 1000}}, 14: {{Type: 21302, Num: 8}}, 15: {{Type: 2, Num: 400}},
	16: {{Type: 2, Num: 400}}, 17: {{Type: 2, Num: 400}}, 18: {{Type: 2, Num: 400}},
	19: {{Type: 2, Num: 400}}, 20: {{Type: 2, Num: 1000}}, 21: {{Type: 2, Num: 400}},
	22: {{Type: 2, Num: 400}}, 23: {{Type: 2, Num: 400}}, 24: {{Type: 2, Num: 400}},
	25: {{Type: 2, Num: 400}}, 26: {{Type: 2, Num: 400}}, 27: {{Type: 2, Num: 1000}},
	28: {{Type: 50113, Num: 150}}, 29: {{Type: 2, Num: 2000}}, 30: {{Type: 21901, Num: 6}},
}

// 礼包抢购
var LibaoQianggouLimit = map[int]int{30: 3, 68: 3, 198: 3, 328: 3, 648: 3}
var LibaoQianggouRewards = map[int][]util.TypeNum{
	30:  {{Type: 2, Num: 300}, {Type: 21001, Num: 6}, {Type: 20201, Num: 500}, {Type: 1, Num: 1000000}},
	68:  {{Type: 2, Num: 680}, {Type: 21001, Num: 8}, {Type: 20201, Num: 1200}, {Type: 7, Num: 3000000}},
	198: {{Type: 2, Num: 1980}, {Type: 21001, Num: 16}, {Type: 20201, Num: 2000}, {Type: 20701, Num: 2000}},
	328: {{Type: 2, Num: 3280}, {Type: 21001, Num: 20}, {Type: 1, Num: 10000000}, {Type: 7, Num: 5000000}},
	648: {{Type: 2, Num: 6480}, {Type: 21901, Num: 12}, {Type: 30113, Num: 1}, {Type: 50113, Num: 50}},
}

// 积天豪礼
var JitianHaoliReward = map[int][]util.TypeNum{
	1:  {{Type: 1, Num: 200000}, {Type: 7, Num: 100000}, {Type: 21901, Num: 1}},
	2:  {{Type: 1, Num: 400000}, {Type: 7, Num: 200000}, {Type: 20701, Num: 200}},
	3:  {{Type: 1, Num: 600000}, {Type: 7, Num: 300000}, {Type: 21001, Num: 3}},
	4:  {{Type: 1, Num: 800000}, {Type: 7, Num: 400000}, {Type: 21301, Num: 5}},
	5:  {{Type: 1, Num: 1000000}, {Type: 7, Num: 500000}, {Type: 21302, Num: 1}},
	6:  {{Type: 1, Num: 1200000}, {Type: 7, Num: 600000}, {Type: 30109, Num: 1}},
	7:  {{Type: 1, Num: 1400000}, {Type: 7, Num: 700000}, {Type: 50113, Num: 50}},
	8:  {{Type: 1, Num: 1600000}, {Type: 7, Num: 800000}, {Type: 30110, Num: 1}},
	9:  {{Type: 1, Num: 1800000}, {Type: 7, Num: 900000}, {Type: 20701, Num: 500}},
	10: {{Type: 1, Num: 2000000}, {Type: 7, Num: 1000000}, {Type: 21901, Num: 1}},
	11: {{Type: 1, Num: 2200000}, {Type: 7, Num: 1100000}, {Type: 21301, Num: 12}},
	12: {{Type: 1, Num: 2400000}, {Type: 7, Num: 1200000}, {Type: 21302, Num: 1}},
	13: {{Type: 1, Num: 2600000}, {Type: 7, Num: 1300000}, {Type: 50113, Num: 50}},
	14: {{Type: 1, Num: 2800000}, {Type: 7, Num: 1400000}, {Type: 21001, Num: 2}},
	15: {{Type: 1, Num: 3000000}, {Type: 7, Num: 1500000}, {Type: 502007, Num: 50}},
}

// 累计充值
var LejiChongCount = []int{50, 100, 328, 648}
var LejiChongzhiRewards = map[int][]util.TypeNum{
	50:  {{Type: 21001, Num: 3}, {Type: 20701, Num: 100}, {Type: 7, Num: 80000}},
	100: {{Type: 21302, Num: 2}, {Type: 20701, Num: 300}, {Type: 7, Num: 150000}},
	328: {{Type: 21901, Num: 1}, {Type: 20701, Num: 1000}, {Type: 50113, Num: 10}},
	648: {{Type: 40301, Num: 1}, {Type: 20701, Num: 3000}, {Type: 50113, Num: 50}},
}

// 至尊月卡专属奖励
var ZhizunZhuanshuRewards = map[int][]util.TypeNum{
	0:  {{Type: 2, Num: 1}, {Type: 21301, Num: 1}},
	1:  {{Type: 2, Num: 20}, {Type: 21301, Num: 1}},
	2:  {{Type: 2, Num: 40}, {Type: 21301, Num: 1}},
	3:  {{Type: 2, Num: 60}, {Type: 21301, Num: 1}},
	4:  {{Type: 2, Num: 80}, {Type: 21301, Num: 1}},
	5:  {{Type: 2, Num: 100}, {Type: 21301, Num: 2}},
	6:  {{Type: 2, Num: 110}, {Type: 50113, Num: 1}},
	7:  {{Type: 2, Num: 120}, {Type: 50113, Num: 1}},
	8:  {{Type: 2, Num: 130}, {Type: 50113, Num: 1}},
	9:  {{Type: 2, Num: 140}, {Type: 50113, Num: 1}},
	10: {{Type: 2, Num: 150}, {Type: 50113, Num: 2}},
	11: {{Type: 2, Num: 160}, {Type: 50113, Num: 2}},
	12: {{Type: 2, Num: 170}, {Type: 50113, Num: 2}},
	13: {{Type: 2, Num: 180}, {Type: 50113, Num: 2}},
}

// 一元礼包
var YiyuanLibaoRewards = []util.TypeNum{{Type: 21001, Num: 10}}

// 任意充值礼包
var AnychongRewards = []util.TypeNum{{Type: 21001, Num: 10}}

// 开局10连抽英雄
var Begin10Chou = [][2]int{{1201, 5}, {1205, 4}, {2202, 4}, {3401, 3}, {4405, 3}, {2506, 3}, {1213, 3}, {4504, 3}, {2103, 3}, {2104, 3}}

// 分享奖励
var ShareRewards = map[int][]util.TypeNum{
	0: {{Type: 2, Num: 20}, {Type: 1, Num: 100000}},
	1: {{Type: 2, Num: 20}, {Type: 20201, Num: 50}},
}

// 邀请奖励
var InviteRewards = map[int][]util.TypeNum{
	1: {{Type: 2, Num: 100}, {Type: 20201, Num: 100}, {Type: 1, Num: 100000}, {Type: 7, Num: 100000}},
	3: {{Type: 2, Num: 250}, {Type: 20201, Num: 500}, {Type: 1, Num: 350000}, {Type: 7, Num: 350000}},
	5: {{Type: 2, Num: 500}, {Type: 20201, Num: 1000}, {Type: 1, Num: 600000}, {Type: 7, Num: 600000}},
}

// 看视频奖励
var VedioRewards = map[int][]util.TypeNum{
	1: {{Type: 2, Num: 20}, {Type: 1, Num: 100000}},
	3: {{Type: 2, Num: 20}, {Type: 1, Num: 150000}, {Type: 21001, Num: 1}},
	5: {{Type: 2, Num: 20}, {Type: 1, Num: 150000}, {Type: 20201, Num: 50}, {Type: 50113, Num: 10}},
	7: {{Type: 2, Num: 50}, {Type: 1, Num: 150000}, {Type: 20201, Num: 100}, {Type: 50113, Num: 10}},
}

type LibaoRewards struct {
	Id      int            `json:"id"`
	Code    string         `json:"code"`
	Rewards []util.TypeNum `json:"rewards"`
}

// ---- 动态数据 ----
var (
	// VIP配置 vip_lv => key => value(int)
	VipConfigs map[int]map[string]int
	// VIP原始数据 vip_lv => data
	VipDatas map[int]*table.VipConfig
	// 礼包码奖励数据 id => data
	LibaoRewardsData map[int]LibaoRewards
	// 按周算周期
	WeekZhouqi []map[string]int64
	// 按月算周期
	MonthZhouqi []map[string]int64
)

func initQiandaoRewards() {
	QiandaoRewards = make([]util.TypeNum, 31)
	for i := 0; i < 31; i++ {
		QiandaoRewards[i] = util.TypeNum{Type: 2, Num: 100}
	}
}

// InitVipConfig 初始化VIP配置
func InitVipConfig(ctx context.Context) {
	initQiandaoRewards()

	VipConfigs = make(map[int]map[string]int)
	VipDatas = make(map[int]*table.VipConfig)

	vcRows, err := info.GetAllVipConfig(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init vip_config error: %v", err)
	}
	for _, vc := range vcRows {
		// 用 ObjectToMap 构建 map[string]int（保留动态 key 访问能力）
		tmpMap := types.ObjectToMap(vc)
		intMap := make(map[string]int)
		for k, v := range tmpMap {
			intMap[k] = types.ToIntE(v)
		}
		VipConfigs[vc.VipLv] = intMap
		VipDatas[vc.VipLv] = vc
	}

	// 礼包码奖励
	LibaoRewardsData = make(map[int]LibaoRewards)
	libaoRows, err := info.GetAllLibaoRewards(ctx)
	if err != nil {
		log.Errorf(ctx, 0, "init libao_rewards error: %v", err)
	}
	for _, row := range libaoRows {
		data := LibaoRewards{
			Id:      row.Id,
			Code:    row.Code,
			Rewards: util.ToTypeNums(row.Rewards),
		}
		LibaoRewardsData[row.Id] = data
	}

	// 以周为周期
	zoneInfo := ZoneInfo()
	if len(zoneInfo) == 0 {
		return
	}

	// 获取区服开始时间（假设ShineZoneID从配置获取）
	zoneIdx := GetShineZoneID() - 1
	if zoneIdx < 0 || zoneIdx >= len(zoneInfo) {
		return
	}

	zoneTimeStr := types.ToString(zoneInfo[zoneIdx]["time"])
	zoneStartTime, err := time.Parse("2006-01-02 15:04:05", zoneTimeStr)
	if err != nil {
		zoneStartTime, err = time.Parse("2006-01-02", zoneTimeStr)
		if err != nil {
			log.Errorf(ctx, 0, "parse zone time error: %v", err)
			return
		}
	}
	zoneStartDate := time.Date(zoneStartTime.Year(), zoneStartTime.Month(), zoneStartTime.Day(), 0, 0, 0, 0, zoneStartTime.Location())
	zoneStartTimestamp := zoneStartDate.Unix()

	// 按周
	WeekZhouqi = make([]map[string]int64, 0, 64)
	ts := zoneStartTimestamp
	for i := 0; i < 64; i++ {
		start := ts
		ts = ts + 7*86400
		WeekZhouqi = append(WeekZhouqi, map[string]int64{"start": start, "end": ts})
	}

	// 按月
	MonthZhouqi = make([]map[string]int64, 0, 32)
	year := zoneStartDate.Year()
	month := int(zoneStartDate.Month())
	day := zoneStartDate.Day()
	for i := 0; i < 32; i++ {
		start := time.Date(year, time.Month(month), day, 0, 0, 0, 0, zoneStartDate.Location()).Unix()
		month++
		if month == 13 {
			year++
			month = 1
		}
		end := time.Date(year, time.Month(month), day, 0, 0, 0, 0, zoneStartDate.Location()).Unix()
		MonthZhouqi = append(MonthZhouqi, map[string]int64{"start": start, "end": end})
	}
}

// CurrentWeekZhouqi 获取当前周期索引和结束时间（按周）
func CurrentWeekZhouqi() (int, int64) {
	now := time.Now().Unix()
	for i, zhouqi := range WeekZhouqi {
		if now >= zhouqi["start"] && now < zhouqi["end"] {
			return i, zhouqi["end"]
		}
	}
	return 0, 0
}

// CurrentMonthZhouqi 获取当前周期索引和结束时间（按月）
func CurrentMonthZhouqi() (int, int64) {
	now := time.Now().Unix()
	for i, zhouqi := range MonthZhouqi {
		if now >= zhouqi["start"] && now < zhouqi["end"] {
			return i, zhouqi["end"]
		}
	}
	return 0, 0
}

// GetVipInfo 获取VIP配置项
func GetVipInfo(userGrade types.Map, configType string) int {
	vipLevel := types.ToIntE(userGrade["vip_level"])
	if cfg, ok := VipConfigs[vipLevel]; ok {
		return cfg[configType]
	}
	return 0
}

// GetVipInfoLv 根据VIP等级获取配置项
func GetVipInfoLv(vipLv int, configType string) int {
	if cfg, ok := VipConfigs[vipLv]; ok {
		return cfg[configType]
	}
	return 0
}

// GetVipLevelByCostMoney 根据充值金额获取VIP等级
func GetVipLevelByCostMoney(costMoney int) int {
	if costMoney < 6 {
		return 0
	} else if costMoney < 30 {
		return 1
	} else if costMoney < 100 {
		return 2
	} else if costMoney < 200 {
		return 3
	} else if costMoney < 500 {
		return 4
	} else if costMoney < 1000 {
		return 5
	} else if costMoney < 1500 {
		return 6
	} else if costMoney < 2000 {
		return 7
	} else if costMoney < 3000 {
		return 8
	} else if costMoney < 5000 {
		return 9
	} else if costMoney < 10000 {
		return 10
	} else if costMoney < 15000 {
		return 11
	} else if costMoney < 30000 {
		return 12
	}
	return 13
}
