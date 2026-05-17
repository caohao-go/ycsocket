// 开服/新区初始化定时任务（基于 trpc-go/timer 框架）
// 对应 PHP: php/crontab/open_zone.php
//
// 使用方式：
//  1. 在 main() 中调用 RegisterOpenZoneTimer(s, zoneID)
//  2. trpc_go.yaml 中配置 timer service（network 含 startAtOnce=1）
//  3. 服务启动时自动执行一次，通过 pika SETNX 锁防止重复运行
package crontab

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/json"
	"server_golang/repo"
	"server_golang/repo/mem/attr"
	"server_golang/repo/table"

	extredis "git.code.oa.com/pcg-csd/trpc-ext/redis"
	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"git.code.oa.com/trpc-go/trpc-database/timer"
	"git.code.oa.com/trpc-go/trpc-go/server"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/repo/info"
	"server_golang/repo/mem/userhero"
	"server_golang/repo/world"
)

const (
	// openZoneLockKey pika 分布式锁 key，用于防止多节点重复执行开服初始化
	openZoneLockKey = "trpc_open_zone_lock_%d"

	// openZoneSchedulerName 注册到 timer 的调度器名称
	openZoneSchedulerName = "openzone"
)

var (
	// zoneIDForOpenZone 全局区服ID，供 timer handler 使用
	zoneIDForOpenZone int
)

// ======================== pika 分布式调度器 ========================
//
// OpenZoneScheduler 基于 pika SETNX 实现分布式互斥调度。
// 同一个服务多个节点同时启动时，只有 SETNX 成功的那个节点会执行开服初始化，
// 执行完毕后锁永久保留，后续启动不再执行。
type OpenZoneScheduler struct{}

// Schedule 尝试抢占执行权：
//   - serviceName: 格式为 "trpc.timer.{进程名}.openzone"
//   - newNode:     当前节点标识，格式 "ip:port_pid_timestamp"
//   - holdTime:    抢占有效期
//
// 返回值：
//   - nowNode: 抢占成功的节点标识
//   - err:     抢占失败返回非 nil，当前节点不执行
func (s *OpenZoneScheduler) Schedule(serviceName string, newNode string, holdTime time.Duration) (nowNode string, err error) {
	ctx := context.Background()
	lockKey := fmt.Sprintf(openZoneLockKey, zoneIDForOpenZone)

	log.Infof(ctx, "[OpenZone] Schedule 尝试获取锁 key=%s node=%s", lockKey, newNode)

	// 使用 SETNX 原子操作：key 不存在则设置成功（返回 1），已存在则失败（返回 0）
	client := extredis.NewHelperWithDefaultCodec(config.Pika)
	if client == nil {
		return "", fmt.Errorf("pika client is nil")
	}

	var result int64
	reply, err := client.Do(ctx, "SETNX", lockKey, newNode)
	if err != nil {
		return "", err
	}
	result, _ = reply.(int64) // redis-go 返回 int64

	if result != 1 {
		// 已被其他节点抢占，读取当前持有者
		var holder string
		client.Get(ctx, lockKey, &holder)
		return holder, fmt.Errorf("lock already held by %s", holder)
	}

	// 设置锁永不过期（开服初始化只需要执行一次）
	log.Infof(ctx, "[OpenZone] Schedule 获取锁成功! node=%s", newNode)
	return newNode, nil
}

// ======================== 开服初始化处理函数 ========================
//
// OpenZoneHandler 是 trpc-go/timer 的处理函数。
// 在服务首次启动时由 timer 框架调用一次（startAtOnce=1），
// 通过 pika 分布式锁保证全局只执行一次。
func OpenZoneHandler(ctx context.Context) error {
	zoneID := zoneIDForOpenZone
	log.Infof(ctx, "========== 开始开服初始化 (zone_id=%d) ==========", zoneID)

	// ---- 1. 初始化机器人 PK 阵位和英雄 ----
	initRobots(ctx, zoneID)

	// ---- 2. 设置公会战场次初始值 ----
	changciKey := fmt.Sprintf(config.KeyPreGuildFightChangci, zoneID)
	repo.RedisSet(ctx, changciKey, "1", 0)
	log.Infof(ctx, "[OpenZone] 公会战场次已设为 1: %s", changciKey)

	log.Infof(ctx, "========== 开服初始化完成 (zone_id=%d) ==========", zoneID)
	return nil
}

// initRobots 初始化机器人数据（user_id=1,2），从爬塔配置表加载阵容
func initRobots(ctx context.Context, zoneID int) {
	type HeroShow struct {
		HeroID int `json:"hero_id"`
		Pos    int `json:"pos"`
		Star   int `json:"star"`
		Stage  int `json:"stage"`
		Lv     int `json:"lv"`
	}

	ctRows, err := info.GetAllClimbtowerHeroOrderByLayer(ctx)
	if err != nil || len(ctRows) == 0 {
		log.Errorf(ctx, 0, "[OpenZone] 加载爬塔数据失败: %v", err)
		return
	}

	climbtowerData := make(map[int]struct {
		Position int
		Heros    []HeroShow
	})

	for _, v := range ctRows {
		data := struct {
			Position int
			Heros    []HeroShow
		}{Position: v.Position}

		fhJSON := string(v.FightHeros)
		var fh [][]interface{}
		json.Unmarshal(fhJSON, &fh)
		for _, h := range fh {
			if len(h) >= 2 {
				data.Heros = append(data.Heros, HeroShow{
					HeroID: types.ToIntE(h[0]),
					Pos:    types.ToIntE(h[1]),
					Star:   v.Star,
					Stage:  v.Stage,
					Lv:     v.Lv,
				})
			}
		}
		climbtowerData[v.Layer] = data
	}

	initScore := 131 * 5 // 初始积分

	for robotUserID := 1; robotUserID <= 2; robotUserID++ {
		log.Infof(ctx, "[OpenZone] 初始化 %d 号机器人...", robotUserID)

		// 检查是否已有阵位数据
		existRows, _ := world.GetUserPositionByUserId(ctx, int64(robotUserID))
		if len(existRows) > 0 {
			log.Infof(ctx, "[OpenZone] 机器人 %d 已有阵位数据，跳过", robotUserID)
			continue
		}

		layer := robotUserID/5 + 1
		layerData, ok := climbtowerData[layer]
		if !ok || len(layerData.Heros) == 0 {
			bestLayer := 0
			for l := range climbtowerData {
				if l <= layer && l > bestLayer {
					bestLayer = l
				}
			}
			if bestLayer == 0 {
				log.Warnf(ctx, "[OpenZone] 爬塔层 %d 无英雄数据，跳过机器人 %d", layer, robotUserID)
				continue
			}
			layerData = climbtowerData[bestLayer]
			log.Infof(ctx, "[OpenZone] 使用爬塔层 %d 数据替代", bestLayer)
		}

		// 创建 user_grade
		attr.HmSet(int64(robotUserID), types.Map{
			config.UserId:        robotUserID,
			config.AttrLv:        15,
			config.AttrVipLevel:  0,
			config.AttrCopy:      0,
			config.AttrChapter:   0,
			config.AttrRoleTitle: 0,
			config.AttrNewGift:   0,
			config.AttrDay7:      0,
			config.AttrIGift:     0,
			config.AttrInitFlag:  1,
			config.AttrRegT:      time.Now().Format("2006-01-02 15:04:05"),
			config.AttrOffTime:   time.Now().Format("2006-01-02 15:04:05"),
		})

		userPosition := &table.UserPosition{
			UserId:   int64(robotUserID),
			PosType:  2,
			Position: layerData.Position,
		}

		j := 1

		for _, hs := range layerData.Heros {
			id, err := userhero.InsertUserHero(int64(robotUserID), &table.UserHero{
				UserId: int64(robotUserID),
				HeroId: hs.HeroID,
				Star:   hs.Star,
				Stage:  hs.Stage,
				Lv:     hs.Lv,
			})

			if err != nil {
				log.Errorf(ctx, 0, "[OpenZone] 插入 hero %d 失败: %v", hs.HeroID, err)
				continue
			}

			switch j {
			case 1:
				userPosition.Pos1Pos = hs.Pos
				userPosition.Pos1Hero = id
			case 2:
				userPosition.Pos2Pos = hs.Pos
				userPosition.Pos2Hero = id
			case 3:
				userPosition.Pos3Pos = hs.Pos
				userPosition.Pos3Hero = id
			case 4:
				userPosition.Pos4Pos = hs.Pos
				userPosition.Pos4Hero = id
			case 5:
				userPosition.Pos5Pos = hs.Pos
				userPosition.Pos5Hero = id
			}
			j++
		}

		if j <= 1 {
			log.Warnf(ctx, "[OpenZone] 机器人 %d 未插入任何英雄，跳过", robotUserID)
			continue
		}

		// 写入防守阵位(pos_type=2) 和攻击阵位(pos_type=1)
		world.ReplaceUserPosition(ctx, userPosition)
		userPosition.PosType = 1
		world.ReplaceUserPosition(ctx, userPosition)

		// 设置 PK 排名初始分
		repo.RedisZAdd(ctx, "pk_rank_keys", float64(initScore), robotUserID)
		initScore -= (3 + robotUserID%5)

		log.Infof(ctx, "[OpenZone] 机器人 %d 初始化完成 (%d 个英雄)", robotUserID, j-1)
	}
}

// ======================== 注册入口 ========================

// RegisterOpenZoneTimer 注册开服初始化定时任务。
// 必须在 trpc.NewServer() 之后、s.Serve() 之前调用。
func RegisterOpenZoneTimer(s server.Service, zoneID int) {
	zoneIDForOpenZone = zoneID

	// 注册自定义 pika 分布式调度器
	timer.RegisterScheduler(openZoneSchedulerName, &OpenZoneScheduler{})

	// 注册定时器处理服务
	timer.RegisterHandlerService(s, OpenZoneHandler)

	log.Infof(context.Background(), "开服初始化定时器已注册 (zone_id=%d, scheduler=openzone)", zoneID)
}
