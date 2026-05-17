// 游戏服务器入口包。
// 使用 tRPC-Go 框架，基于 tnet WebSocket transport，
// MySQL 采用 trpc-ext/orm，Redis 采用 trpc-ext/redis。
package main

import (
	"context"
	"io"
	"runtime"
	"strings"
	"sync"
	"time"

	_ "git.code.oa.com/pcg-csd/trpc-ext/orm"
	_ "git.code.oa.com/pcg-csd/trpc-ext/redis/trpc"
	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"git.code.oa.com/trpc-go/trpc-go"
	"git.woa.com/trpc-go/tnet"
	"git.woa.com/trpc-go/tnet/extensions/websocket"
	twebsocket "git.woa.com/trpc-go/trpc-tnet-transport/websocket"
	"server_golang/auth"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/config"
	"server_golang/connector"
	"server_golang/controller"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// GameWSService 实现 twebsocket.Service 接口，用于游戏服务器 WebSocket 服务。
type GameWSService struct {
	sync.RWMutex
	router *controller.Router
}

// Accept 在升级为 WebSocket 连接之前调用。
// 用于安全检查、来源校验等。
func (s *GameWSService) Accept(ctx context.Context) (context.Context, error) {
	// 可选：检查 HTTP 请求头、查询参数进行鉴权
	g, _ := websocket.UpgraderFromContext(ctx)
	if g != nil {
		g.OnHeader = func(key, value []byte) error {
			// 保存 X-Real-Ip 供后续使用
			if string(key) == "X-Real-Ip" {
				ctx = context.WithValue(ctx, clientIPKey{}, string(value))
			}
			if string(key) == "X-Forwarded-For" {
				if ctx.Value(clientIPKey{}) == nil {
					ctx = context.WithValue(ctx, clientIPKey{}, string(value))
				}
			}
			return nil
		}
	}
	return ctx, nil
}

// Connected 在 WebSocket 握手成功后调用。
func (s *GameWSService) Connected(ctx context.Context, conn websocket.Conn) error {
	log.Infof(context.Background(), "WebSocket connected: %v", conn.RemoteAddr())
	return nil
}

// Read 在连接上收到数据时调用。
func (s *GameWSService) Read(ctx context.Context, conn websocket.Conn) error {
	opcode, data, err := conn.ReadMessage()
	if err != nil {
		if err == io.EOF {
			return nil
		}
		return err
	}

	if opcode != websocket.Text {
		return nil
	}

	message := string(data)
	if message == "" {
		return nil
	}

	log.Infof(context.Background(), "Read message: %s", message)

	// 从上下文或连接获取客户端IP
	clientIP := ""
	if v, ok := ctx.Value(clientIPKey{}).(string); ok {
		clientIP = v
	}
	if clientIP == "" {
		clientIP = conn.RemoteAddr().String()
	}

	// 处理关闭连接命令
	if strings.HasPrefix(message, "close") {
		parts := strings.Split(message, "_")
		if len(parts) > 1 {
			uid := types.ToInt64E(parts[1])
			if uid > 0 {
				connector.Manager.Close(uid)
			}
		}
		conn.WriteMessage(websocket.Text, []byte("closed"))
		return nil
	}

	// 处理心跳
	if strings.HasPrefix(message, "heartbeat") {
		parts := strings.Split(message, "_")
		if len(parts) > 1 {
			uid := types.ToInt64E(parts[1])
			if uid == 0 {
				conn.WriteMessage(websocket.Text, []byte("heartbeat no uid"))
				return nil
			}
			connector.Manager.SetConnectExpire(uid)
			conn.WriteMessage(websocket.Text, types.ToBytes(time.Now().Unix()))
			return nil
		}
	}

	// 解析JSON输入
	var input types.Map
	if err := json.Api.Unmarshal(data, &input); err != nil || input["c"] == nil || input["m"] == nil {
		errResp := json.MarshalToBytes(types.Map{
			"code":        "1",
			"description": "input error",
			"data":        message,
		})
		conn.WriteMessage(websocket.Text, errResp)
		return nil
	}

	uid := input.GetInt64E("userid")
	if uid > 0 {
		connector.Manager.SetWsConn(uid, conn)
	}

	// 签名验证
	code, msg := auth.Verify(input)
	if code != 0 {
		errResp := json.MarshalToBytes(types.Map{"code": code, "msg": msg})
		conn.WriteMessage(websocket.Text, errResp)
		return nil
	}

	// 分发到控制器
	func() {
		defer func() {
			if r := recover(); r != nil {
				buf := make([]byte, 4096)
				n := runtime.Stack(buf, false)
				log.Errorf(context.Background(), 0, "Catch Exception: %v\nStack:\n%s", r, string(buf[:n]))
				errResp := json.MarshalToBytes(types.Map{
					"c":     input.GetStringE("c"),
					"m":     input.GetStringE("m"),
					"reqid": input.GetStringE("reqid"),
					"code":  99,
					"msg":   "系统异常"})
				conn.WriteMessage(websocket.Text, errResp)
			}
		}()

		result := s.router.Dispatch(context.Background(), input, clientIP)

		// 处理登录响应
		c := input.GetStringE("c")
		m := input.GetStringE("m")
		if c == "user" && (m == "login" || m == "accountLogin" || m == "register") && uid == 0 {
			loginOut := json.ToMap(result.Msg)
			newUID := loginOut.GetInt64E("userid")
			if newUID > 0 {
				uid = newUID
				connector.Manager.SetWsConn(uid, conn)
			}
		}

		// 发送结果
		sendWsResult(conn, uid, result)
	}()

	return nil
}

// Push 不会被框架自动调用，需要用户自定义推送逻辑。
func (s *GameWSService) Push(ctx context.Context, data []byte) error {
	return nil
}

// Stop 在 WebSocket 连接关闭时调用。
func (s *GameWSService) Stop(ctx context.Context, conn websocket.Conn) {
	log.Infof(context.Background(), "WebSocket disconnected: %v", conn.RemoteAddr())
}

func sendWsResult(conn websocket.Conn, uid int64, result *controller.Result) {
	if result == nil {
		return
	}

	log.Infof(context.Background(), "sendWsResult uid: %d, result: %v", uid, result)

	switch v := result.SendUser.(type) {
	case string:
		if v == "me" {
			conn.WriteMessage(websocket.Text, []byte(result.Msg))
		} else if v == "all" {
			connector.Manager.SendAll(result.Msg)
		}
	case []int64:
		if len(v) > 0 {
			connector.Manager.SendFds(v, result.Msg)
		}
	}
}

type clientIPKey struct{}

func main() {
	tnet.SetNumPollers(4)

	// 创建 tRPC 服务器
	s := trpc.NewServer()

	config.Load("./config/config.yaml")

	// 初始化连接管理器（从配置读取 zone_id）
	zoneID := 1
	if len(config.Cfg.Zone.ZoneInfo) > 0 {
		zoneID = config.Cfg.Zone.ZoneInfo[0].ZoneID
	}

	connector.Init(zoneID)

	// 启动连接超时检查器
	connector.Manager.StartExpireChecker()

	// 注册 WebSocket 服务
	svc := &GameWSService{
		router: controller.NewRouter(),
	}

	twebsocket.RegisterNamedWebsocketService(
		"trpc.shinelight.gameserver.WsHandler",
		s,
		svc,
	)

	log.Infof(context.Background(), "游戏服务器启动中...")

	// 初始化游戏配置数据（从 db_dabaojian 加载静态配置表到内存）
	initCtx := context.Background()
	item.Init(initCtx)
	logic.InitHeroinfo(initCtx)
	logic.InitHeroattr(initCtx)
	logic.InitSkill(initCtx)
	logic.InitCopy(initCtx)
	logic.InitCombination(initCtx)
	model.InitGuildFightChangci(initCtx)
	logic.InitGuild(initCtx)
	logic.InitEndless(initCtx)
	logic.InitClimbtower(initCtx)
	logic.InitMonsters(initCtx)
	item.InitEquipment(initCtx)
	logic.InitMerge(initCtx)
	logic.InitItemsCollection(initCtx)
	logic.InitVipConfig(initCtx)
	logic.InitFunctionConfig(initCtx)
	logic.InitVoyage(initCtx)
	logic.InitSacrifice(initCtx)
	logic.InitDaily(initCtx)
	logic.InitTitle(initCtx)
	logic.InitCheckpointReward(initCtx)
	logic.InitTanbao(initCtx)
	logic.InitExpedition(initCtx)
	logic.InitPositions(initCtx)
	logic.InitChapter(initCtx)
	logic.InitZoneinfo(initCtx)
	logic.InitFriends(initCtx)
	model.InitTemplate(initCtx)
	logic.InitAchieve(initCtx)
	logic.InitTaskWeekly(initCtx)
	logic.InitTaskGuide(initCtx)
	logic.InitTaskConfig(initCtx)
	logic.InitSensitiveWords(initCtx)

	log.Infof(context.Background(), "所有配置表加载完成")

	// 注册开服初始化定时任务（首次启动运行一次，Redis SETNX 锁防重复）
	//crontab.RegisterOpenZoneTimer(s.Service("trpc.shinelight.gameserver.OpenZoneTimer"), zoneID)

	// just for test 测试
	//item.AddCoin(10066001, 100000000)
	//item.AddZuan(10066001, 20000)
	//item.Add(10066001, item.ItemHeroExp, 100000000, nil)
	//model.AddExp(10066001, 100000000)
	//item.Add(10066001, 20201, 20000, nil)
	//item.Add(10066001, 40201, 20, nil)

	// 启动服务
	if err := s.Serve(); err != nil {
		log.Fatalf(context.Background(), "Server error: %v", err)
	}
}
