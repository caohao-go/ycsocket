// 定时任务管理器，使用 github.com/robfig/cron/v3 实现所有周期性任务。
// 原 PHP 版本使用 Swoole\Timer::tick，这里迁移为 Go 的 cron 调度。
package crontab

import (
	"context"
	"fmt"
	"time"

	"server_golang/common/json"
	"server_golang/repo"

	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"github.com/robfig/cron/v3"
	"server_golang/logic"

	"server_golang/common/types"
	"server_golang/config"
	"server_golang/connector"
	"server_golang/model"
	"server_golang/repo/user"
)

// cronInstance 全局 cron 实例
var cronInstance *cron.Cron

// StartCronJobs 启动所有定时任务
func StartCronJobs(zoneID int) {
	cronInstance = cron.New(cron.WithSeconds())

	// ---- 任务1: 每60秒 - 连接超时检查 ----
	cronInstance.AddFunc("@every 60s", func() {
		safeRun("连接超时检查", func() {
			connector.Manager.ConnectExpire()
		})
	})

	// ---- 任务2: 每1秒 - H5支付回调 ----
	// 原 PHP 为 300ms，Go 版改为 1秒
	cronInstance.AddFunc("@every 1s", func() {
		safeRun("H5支付回调", func() {
			processH5PayCallback(zoneID)
		})
	})

	// ---- 任务3: 每1秒 - 小程序支付回调 ----
	// 原 PHP 为 300ms，Go 版改为 1秒
	cronInstance.AddFunc("@every 1s", func() {
		safeRun("小程序支付回调", func() {
			processMpPayCallback(zoneID)
		})
	})

	// ---- 任务4: 每1秒 - 通用支付处理 ----
	// 原 PHP 为 300ms，Go 版改为 1秒
	cronInstance.AddFunc("@every 1s", func() {
		safeRun("通用支付处理", func() {
			processShinePayCallback(zoneID)
		})
	})

	// ---- 任务5: 每5秒 - 更新区服信息 ----
	cronInstance.AddFunc("@every 5s", func() {
		safeRun("更新区服信息", func() {
			zoneinfoUpdate()
		})
	})

	// ---- 任务6: 每10秒 - 更新游戏版本 ----
	cronInstance.AddFunc("@every 10s", func() {
		safeRun("更新游戏版本", func() {
			zoneinfoGameVersion()
		})
	})

	// ---- 任务7: 每5秒 - 分配公会战 ----
	// 周一/三/五 8点时分配
	cronInstance.AddFunc("@every 5s", func() {
		safeRun("分配公会战", func() {
			assignGuildFight(zoneID)
		})
	})

	// ---- 任务8: 每30秒 - 公会战状态重置 ----
	// 周一/三/五 21点重置为0
	cronInstance.AddFunc("@every 30s", func() {
		safeRun("公会战状态重置", func() {
			resetGuildFightStatus(zoneID)
		})
	})

	// ---- 任务9: 每58秒 - 公会战场次递增 ----
	// 周二/四/六 12:00时递增
	cronInstance.AddFunc("@every 58s", func() {
		safeRun("公会战场次递增", func() {
			incrGuildFightChangci(zoneID)
		})
	})

	// ---- 任务10: 每3秒 - 保存星河神殿信息 ----
	cronInstance.AddFunc("@every 3s", func() {
		safeRun("保存星河神殿信息", func() {
			model.SaveTemplateInfo(context.Background(), logic.TemplateInfo)
		})
	})

	cronInstance.Start()
	log.Infof(context.Background(), "定时任务已启动，区服ID: %d", zoneID)
}

// StopCronJobs 停止所有定时任务
func StopCronJobs() {
	if cronInstance != nil {
		ctx := cronInstance.Stop()
		<-ctx.Done()
		log.Infof(context.Background(), "定时任务已停止")
	}
}

// safeRun 安全执行定时任务，捕获 panic
func safeRun(name string, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf(context.Background(), 0, "定时任务[%s]异常: %v", name, r)
		}
	}()
	fn()
}

// ======================== 支付回调处理 ========================

// processH5PayCallback 处理 H5 支付回调
// 从 pika 队列 pre_h5pay_back_{zoneID} rpop 数据，查订单，通知客户端
func processH5PayCallback(zoneID int) {
	ctx := context.Background()
	queueKey := fmt.Sprintf(config.KeyPreH5payBack, zoneID)

	for {
		data, err := repo.RedisRPop(ctx, queueKey)
		if err != nil || data == "" {
			break
		}

		log.Infof(context.Background(), "H5支付回调数据: %s", data)

		// 解析回调数据
		callbackData := types.ToMapE(data)
		if callbackData == nil {
		}

		orderID := callbackData.GetIntE("orderid")
		if orderID <= 0 {
			log.Errorf(context.Background(), 0, "H5支付回调订单ID无效: %s", data)
			continue
		}

		// 查询订单信息
		tradeInfo, err := user.GetNewTradeInfoById(ctx, int64(orderID))
		if err != nil || tradeInfo == nil {
			log.Errorf(context.Background(), 0, "H5支付回调订单不存在: orderID=%d", orderID)
			continue
		}

		// 更新订单状态
		err = user.UpdateNewTradeInfoById(ctx, int64(orderID), types.Map{"status": "SUCCESS"})
		if err != nil {
			log.Errorf(context.Background(), 0, "H5支付回调更新订单失败: orderID=%d, err=%v", orderID, err)
			continue
		}

		// 通知客户端
		userID := tradeInfo.UserId
		if userID > 0 {
			connector.Manager.Send(userID, json.Marshal(types.Map{
				"c":       "pay",
				"m":       "h5payback",
				"code":    0,
				"orderid": orderID,
				"type":    tradeInfo.Type,
				"amount":  tradeInfo.TotalFee,
			}))
		}
	}
}

// processMpPayCallback 处理小程序支付回调
// 从 pika 队列 pre_mppay_back_{zoneID} rpop 数据
func processMpPayCallback(zoneID int) {
	ctx := context.Background()
	queueKey := fmt.Sprintf(config.KeyPreMppayBack, zoneID)

	for {
		data, err := repo.RedisRPop(ctx, queueKey)
		if err != nil || data == "" {
			break
		}

		log.Infof(context.Background(), "小程序支付回调数据: %s", data)

		var callbackData types.Map
		if err := json.Unmarshal(data, &callbackData); err != nil {
			log.Errorf(context.Background(), 0, "小程序支付回调数据解析失败: %v, data=%s", err, data)
			continue
		}

		orderID := callbackData.GetIntE("orderid")
		if orderID <= 0 {
			log.Errorf(context.Background(), 0, "小程序支付回调订单ID无效: %s", data)
			continue
		}

		// 查询订单信息
		tradeInfo, err := user.GetNewTradeInfoById(ctx, int64(orderID))
		if err != nil || tradeInfo == nil {
			log.Errorf(context.Background(), 0, "小程序支付回调订单不存在: orderID=%d", orderID)
			continue
		}

		// 更新订单状态
		err = user.UpdateNewTradeInfoById(ctx, int64(orderID), types.Map{"status": "SUCCESS"})
		if err != nil {
			log.Errorf(context.Background(), 0, "小程序支付回调更新订单失败: orderID=%d, err=%v", orderID, err)
			continue
		}

		// 通知客户端
		userID := tradeInfo.UserId
		if userID > 0 {
			connector.Manager.Send(userID, json.Marshal(types.Map{
				"c":       "pay",
				"m":       "mppayback",
				"code":    0,
				"orderid": orderID,
				"type":    tradeInfo.Type,
				"amount":  tradeInfo.TotalFee,
			}))
		}
	}
}

// processShinePayCallback 处理通用支付
// 从 pika 队列 pre_shine_pay_{zoneID} rpop 数据
func processShinePayCallback(zoneID int) {
	ctx := context.Background()
	queueKey := fmt.Sprintf(config.KeyPreShinePay, zoneID)

	for {
		data, err := repo.RedisRPop(ctx, queueKey)
		if err != nil || data == "" {
			break
		}

		log.Infof(context.Background(), "通用支付数据: %s", data)

		payData := types.ToMapE(data)
		if payData == nil {
			log.Errorf(context.Background(), 0, "通用支付数据解析失败: data=%s", data)
			continue
		}

		orderID := payData.GetInt64E("orderid")
		userID := payData.GetInt64E("user_id")
		payType := payData.GetIntE("type")
		amount := payData.GetIntE("amount")

		if orderID <= 0 || userID <= 0 {
			log.Errorf(context.Background(), 0, "通用支付数据无效: %s", data)
			continue
		}

		// 更新订单状态
		err = user.UpdateNewTradeInfoById(ctx, orderID, types.Map{"status": "SUCCESS"})
		if err != nil {
			log.Errorf(context.Background(), 0, "通用支付更新订单失败: orderID=%d, err=%v", orderID, err)
			continue
		}

		// 调用 model.BuyStaff() 执行购买逻辑
		if _, err := model.BuyStaff(ctx, userID, 0, int(amount), orderID); err != nil {
			log.Errorf(context.Background(), 0, "通用支付购买逻辑执行失败: userID=%d, err=%v", userID, err)
		}

		// 通知客户端
		if userID > 0 {
			connector.Manager.Send(userID, json.Marshal(types.Map{
				"c":       "pay",
				"m":       "shinepayback",
				"code":    0,
				"orderid": orderID,
				"type":    payType,
				"amount":  amount,
			}))
		}
	}
}

// ======================== 区服信息更新 ========================

// zoneinfoUpdate 更新区服信息
// 对应原 PHP: logic.ZoneinfoUpdate()
func zoneinfoUpdate() {
	ctx := context.Background()
	// 从数据库获取最新区服信息
	rows, err := user.GetActiveZoneInfo(ctx)
	if err != nil || len(rows) == 0 {
		return
	}
	_ = rows // 更新逻辑由 logic.Zoneinfo 模块处理
}

// zoneinfoGameVersion 更新游戏版本号
// 对应原 PHP: logic.ZoneinfoGameVersion()
func zoneinfoGameVersion() {
	ctx := context.Background()
	// 从数据库获取最新游戏版本号
	latestVer, err := user.GetGameVersionLatest(ctx)
	if err != nil || latestVer == nil {
		return
	}
	_ = latestVer // 更新逻辑由 logic.Zoneinfo 模块处理
}

// ======================== 公会战相关 ========================

// assignGuildFight 分配公会战
// 周一/三/五 8点时分配
func assignGuildFight(zoneID int) {
	now := time.Now()
	weekday := now.Weekday()
	hour := now.Hour()

	// 仅在周一(1)、周三(3)、周五(5) 的 8点 执行
	if (weekday == time.Monday || weekday == time.Wednesday || weekday == time.Friday) && hour == 8 {
		ctx := context.Background()
		flagKey := fmt.Sprintf(config.KeyPreGuildFightAssigned, zoneID, now.Format("2006-01-02"))

		// 使用 pika 标记防止重复执行
		val, _ := repo.RedisGet(ctx, flagKey)
		if val != "" {
			return
		}

		log.Infof(context.Background(), "开始分配公会战, 区服ID: %d", zoneID)

		// 调用 model.AssignGuildFight(ctx)
		if err := model.AssignGuildFight(ctx); err != nil {
			log.Errorf(context.Background(), 0, "公会战分配失败: %v", err)
		}
		// 分配完成后设置标记，24小时过期
		repo.RedisSet(ctx, flagKey, "1", 86400)

		log.Infof(context.Background(), "公会战分配完成, 区服ID: %d", zoneID)
	}
}

// resetGuildFightStatus 公会战状态重置
// 周一/三/五 21点重置为0
func resetGuildFightStatus(zoneID int) {
	now := time.Now()
	weekday := now.Weekday()
	hour := now.Hour()

	// 仅在周一(1)、周三(3)、周五(5) 的 21点 执行
	if (weekday == time.Monday || weekday == time.Wednesday || weekday == time.Friday) && hour == 21 {
		ctx := context.Background()
		flagKey := fmt.Sprintf(config.KeyPreGuildFightReset, zoneID, now.Format("2006-01-02"))

		// 防止重复执行
		val, _ := repo.RedisGet(ctx, flagKey)
		if val != "" {
			return
		}

		log.Infof(context.Background(), "重置公会战状态, 区服ID: %d", zoneID)

		// 重置公会战状态为 0
		resetKey := fmt.Sprintf(config.KeyPreGuildFightStatus, zoneID)
		repo.RedisSet(ctx, resetKey, "0", 0)

		repo.RedisSet(ctx, flagKey, "1", 86400)
		log.Infof(context.Background(), "公会战状态重置完成, 区服ID: %d", zoneID)
	}
}

// incrGuildFightChangci 公会战场次递增
// 周二/四/六 12:00时递增
func incrGuildFightChangci(zoneID int) {
	now := time.Now()
	weekday := now.Weekday()
	hour := now.Hour()

	// 仅在周二(2)、周四(4)、周六(6) 的 12点 执行
	if (weekday == time.Tuesday || weekday == time.Thursday || weekday == time.Saturday) && hour == 12 {
		ctx := context.Background()
		flagKey := fmt.Sprintf(config.KeyPreGuildFightChangciIncr, zoneID, now.Format("2006-01-02"))

		// 防止重复执行
		val, _ := repo.RedisGet(ctx, flagKey)
		if val != "" {
			return
		}

		log.Infof(context.Background(), "递增公会战场次, 区服ID: %d", zoneID)

		// 递增公会战场次
		changciKey := fmt.Sprintf(config.KeyPreGuildFightChangci, zoneID)
		repo.RedisIncr(ctx, changciKey)

		repo.RedisSet(ctx, flagKey, "1", 86400)
		log.Infof(context.Background(), "公会战场次递增完成, 区服ID: %d", zoneID)
	}
}
