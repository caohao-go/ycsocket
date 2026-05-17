// ItemModel 道具系统核心模型
// 翻译自 PHP ItemModel.php，负责所有道具的增删查、货币校验、开箱逻辑
package item

import (
	"context"
	"fmt"
	"math/rand"
	"sync"

	"git.code.oa.com/pcg-csd/trpc-ext/redis"
	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/repo"
	daoInfo "server_golang/repo/info"
)

// 道具类型常量
const (
	ErrorNotSoMuch      = "NOT_SO_MUCH"
	ItemCoin            = 1
	ItemZuan            = 2
	ItemExp             = 3
	ItemActive          = 4
	ItemPk              = 5
	ItemVoyage          = 6
	ItemHeroExp         = 7
	ItemGuildContribute = 8
	ItemGuildExp        = 9
	ItemFriendPoint     = 10
)

const KeyUserItems = "items_%d_%d"

// UserItems 全局内存存储：userID → itemType → itemID → []*Stack
var UserItems = map[int64]map[int]map[int][]*Stack{}

// loadedUserTypes 记录已加载的 (userID, itemType) 组合
var loadedUserTypes = map[int64]map[int]bool{}

var itemMutex = sync.RWMutex{}

// initUserItemsByType 将用户指定类型的道具从 pika 加载到内存
func initUserItemsByType(userID int64, itemType int) {
	itemMutex.RLock()
	if m, ok := loadedUserTypes[userID]; ok && m[itemType] {
		itemMutex.RUnlock()
		return
	}
	itemMutex.RUnlock()

	ctx := context.Background()
	redisKey := fmt.Sprintf(KeyUserItems, itemType, userID)
	all, err := repo.RedisHGetAll(ctx, redisKey)
	if err != nil && !redis.IsNil(err) {
		panic(fmt.Errorf("initUserItemsByType failed for userID=%d type=%d, err=%v", userID, itemType, err))
	}

	itemMutex.Lock()
	defer itemMutex.Unlock()

	// double check
	if m, ok := loadedUserTypes[userID]; ok && m[itemType] {
		return
	}

	if UserItems[userID] == nil {
		UserItems[userID] = map[int]map[int][]*Stack{}
	}
	if UserItems[userID][itemType] == nil {
		UserItems[userID][itemType] = map[int][]*Stack{}
	}

	for itemIDStr, val := range all {
		itemID := types.ToIntE(itemIDStr)
		var stacks []*Stack
		if json.Unmarshal(val, &stacks) == nil {
			UserItems[userID][itemType][itemID] = stacks
		}
	}

	if loadedUserTypes[userID] == nil {
		loadedUserTypes[userID] = map[int]bool{}
	}
	loadedUserTypes[userID][itemType] = true
}

type Stack struct {
	Id   int64         `json:"id"`
	Num  int           `json:"num"`
	Prop []util.FuProp `json:"prop,omitempty"`
}

// Info 道具信息结构
type Info struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Type      int            `json:"type"`
	TypeSub   int            `json:"type_sub"`
	Stack     int            `json:"stack"`
	Color     int            `json:"color"`
	Star      int            `json:"star"`
	NameColor int            `json:"name_color"`
	HP        int            `json:"hp"`
	Atk       int            `json:"atk"`
	Def       int            `json:"def"`
	Speed     int            `json:"speed"`
	Price     []util.TypeNum `json:"price"`
	Open      []int          `json:"open"`
	Remark    interface{}    `json:"remark"`
	OpenRaw   string         `json:"-"` // 原始 open JSON 字符串
}

// Table 全局道具配置表（从 DB 加载）
var Table map[int]*Info

// Init 从数据库初始化道具配置表
func Init(ctx context.Context) {
	Table = make(map[int]*Info)
	rows, err := daoInfo.GetAllItems(ctx)
	if err != nil {
		panic(fmt.Errorf("init items error: %v", err))
	}
	if len(rows) == 0 {
		panic(fmt.Errorf("init items error: empty result"))
	}

	for _, data := range rows {
		info := &Info{
			ID:        data.Id,
			Name:      data.Name,
			Type:      data.Type,
			TypeSub:   data.TypeSub,
			Stack:     data.Stack,
			Color:     0, // items 表无 color 字段
			NameColor: types.ToIntE(data.NameColor),
			HP:        data.Hp,
			Atk:       data.Atk,
			Def:       data.Def,
			Speed:     data.Speed,
		}

		// 解析价格
		info.OpenRaw = data.Open
		info.Price = util.ToTypeNums(data.Price)

		// 解析 open（碎片合成的英雄ID列表）
		// 注意：items 表中 type=5（碎片）的 open 字段是二维 JSON 数组 [[h1,h2,...], [w1,w2,...]]
		// 第一层是 hero_id 列表，第二层是权重（可选）。PHP 版取的是 $open[0]
		openStr := data.Open
		if openStr != "" {
			var openArr [][]int
			if json.Unmarshal(openStr, &openArr) == nil && len(openArr) > 0 && len(openArr[0]) > 0 {
				for _, v := range openArr[0] { // 只取第一层（hero_id 列表）
					info.Open = append(info.Open, v)
				}
			} else {
				// 兼容非碎片类型的 open（一维数组格式）
				var openFlat []int
				if json.Unmarshal(openStr, &openFlat) == nil {
					for _, v := range openFlat {
						info.Open = append(info.Open, v)
					}
				}
			}
		}

		// remark 解析（碎片的星级等信息）
		if data.Remark != "" {
			var remark interface{}
			json.Unmarshal(data.Remark, &remark)
			info.Remark = remark
			if remarkMap, err := types.ToMap(remark, ""); err == nil {
				if star, ok := remarkMap["star"]; ok {
					info.Star = types.ToIntE(star)
				}
			}
		}

		Table[data.Id] = info
	}
	log.Infof(context.Background(), "道具配置表加载完成，共 %d 条", len(Table))

	initSequence(ctx)
}

// Name 获取道具名称
func Name(itemId int) string {
	if info, ok := Table[itemId]; ok {
		return info.Name
	}
	return ""
}

// Exp 获取用户经验
func Exp(userID int64) int {
	return Total(userID, ItemExp)
}

// Coin 获取用户金币
func Coin(userID int64) int {
	return Total(userID, ItemCoin)
}

// Zuan 获取用户钻石
func Zuan(userID int64) int {
	return Total(userID, ItemZuan)
}

// Total 获取用户某道具总数量（从内存读取）
func Total(userID int64, itemID int) int {
	info := Table[itemID]
	if info == nil {
		return 0
	}
	initUserItemsByType(userID, info.Type)

	itemMutex.RLock()
	stacks := UserItems[userID][info.Type][itemID]
	total := 0
	for _, v := range stacks {
		total += v.Num
	}
	itemMutex.RUnlock()

	return total
}

// GetItemByID 根据道具ID和唯一ID获取道具（从内存读取）
func GetItemByID(userID int64, itemID int, id int64) *Stack {
	info := Table[itemID]
	if info == nil {
		return nil
	}
	initUserItemsByType(userID, info.Type)

	itemMutex.RLock()
	stacks := UserItems[userID][info.Type][itemID]
	for _, v := range stacks {
		if v.Id == id {
			itemMutex.RUnlock()
			return v
		}
	}
	itemMutex.RUnlock()
	return nil
}

// GetListByType 获取用户指定类型的所有道具（从内存读取）
func GetListByType(userID int64, itemType int) []types.Map {
	initUserItemsByType(userID, itemType)

	itemMutex.RLock()
	itemsByType := UserItems[userID][itemType]
	if len(itemsByType) == 0 {
		itemMutex.RUnlock()
		return nil
	}

	var result []types.Map
	for itemID, stacks := range itemsByType {
		itemInfo := Table[itemID]
		for _, v := range stacks {
			ret := types.Map{
				"item_id": itemID,
				"num":     v.Num,
				"id":      v.Id,
			}
			if itemInfo != nil {
				ret["color"] = itemInfo.NameColor
			}
			if v.Prop != nil {
				ret["prop"] = v.Prop
			}
			result = append(result, ret)
		}
	}
	itemMutex.RUnlock()
	return result
}

// AddCoin 增加用户金币
func AddCoin(userID int64, num int) {
	Add(userID, ItemCoin, num, nil)
}

// AddZuan 增加用户钻石
func AddZuan(userID int64, num int) {
	Add(userID, ItemZuan, num, nil)
}

// Add 发放道具（核心方法）
// runeProp: 特殊属性（符文用），id 将原有装备卸载下来重新塞回背包，id 不变
func Add(userID int64, itemID, num int, runeProp []util.FuProp, id ...int64) ([]util.FuProp, error) {
	info := Table[itemID]
	if info == nil {
		return nil, fmt.Errorf("item_id error: %d", itemID)
	}

	maxStack := info.Stack
	if maxStack <= 0 {
		return nil, fmt.Errorf("stack is zero [%d]", itemID)
	}

	initUserItemsByType(userID, info.Type)

	itemMutex.Lock()
	stacks := UserItems[userID][info.Type][itemID]

	newProp := runeProp

	if info.Type == 1 {
		if len(stacks) == 0 {
			// 新用户首次添加货币类道具，创建初始格子
			stacks = append(stacks, &Stack{
				Id:  Sequence(),
				Num: num,
			})
		} else {
			stacks[0].Num += num
		}
	} else if maxStack == 1 {
		// 不可叠加：每个占一格
		for i := 0; i < num; i++ {
			var idTmp int64

			if len(id) > 0 { // 一般是原有装备卸下来
				idTmp = id[0]
			} else {
				idTmp = Sequence()
			}

			newStack := Stack{
				Id:  idTmp,
				Num: 1,
			}
			if len(runeProp) > 0 {
				newStack.Prop = runeProp
			} else if info.Type == 4 {
				// 符文属性随机
				newStack.Prop = InitRandRuneProp(itemID)
				newProp = newStack.Prop
			}
			stacks = append(stacks, &newStack)
		}
	} else {
		// 可叠加：先填充已有格子
		remaining := num
		for i := 0; i < len(stacks) && remaining > 0; i++ {
			hasNum := stacks[i].Num
			if hasNum < maxStack {
				canAdd := maxStack - hasNum
				if remaining <= canAdd {
					stacks[i].Num = hasNum + remaining
					remaining = 0
				} else {
					stacks[i].Num = maxStack
					remaining -= canAdd
				}
			}
		}

		// 剩余的创建新格子
		if remaining > 0 {
			fullSlots := remaining / maxStack
			for i := 0; i < fullSlots; i++ {
				stacks = append(stacks, &Stack{
					Id:  Sequence(),
					Num: maxStack,
				})
			}
			remainder := remaining % maxStack
			if remainder > 0 {
				stacks = append(stacks, &Stack{
					Id:  Sequence(),
					Num: remainder,
				})
			}
		}
	}

	// 同步更新内存
	UserItems[userID][info.Type][itemID] = stacks
	itemMutex.Unlock()

	// 异步写 pika
	redisKey := fmt.Sprintf(KeyUserItems, info.Type, userID)
	go func() {
		ctx := context.Background()
		err := repo.RedisHSet(ctx, redisKey, itemID, stacks)
		if err != nil {
			log.Errorf(ctx, -1, "Add async pika failed for userID=%d itemID=%d, err=%v", userID, itemID, err)
		}
	}()

	return newProp, nil
}

// Sub 扣除用户道具
// ids: 是否需要删除指定 id 的数据（可选）
func Sub(userID int64, itemID, num int, ids ...int64) string {
	info := Table[itemID]
	if info == nil {
		return ErrorNotSoMuch
	}
	initUserItemsByType(userID, info.Type)

	itemMutex.Lock()
	stacks := UserItems[userID][info.Type][itemID]
	if len(stacks) == 0 {
		itemMutex.Unlock()
		return ErrorNotSoMuch
	}

	// 计算总数
	totalNum := 0
	for _, val := range stacks {
		totalNum += val.Num
	}

	if num > totalNum {
		itemMutex.Unlock()
		return ErrorNotSoMuch
	}

	redisKey := fmt.Sprintf(KeyUserItems, info.Type, userID)

	if num == totalNum {
		// 全部扣完，直接删除
		delete(UserItems[userID][info.Type], itemID)
		itemMutex.Unlock()

		go func() {
			ctx := context.Background()
			repo.RedisHDel(ctx, redisKey, itemID)
		}()
		return ""
	}

	if info.Type == 1 {
		stacks[0].Num -= num
	} else {
		// 部分扣除
		idsMap := make(map[int64]bool)
		for _, id := range ids {
			idsMap[id] = true
		}

		remaining := num
		for i := 0; i < len(stacks) && remaining > 0; i++ {
			if len(idsMap) > 0 && !idsMap[stacks[i].Id] {
				continue
			}
			valNum := stacks[i].Num
			if valNum > remaining {
				stacks[i].Num = valNum - remaining
				remaining = 0
			} else if valNum == remaining {
				stacks = append(stacks[:i], stacks[i+1:]...)
				remaining = 0
			} else {
				remaining -= valNum
				stacks = append(stacks[:i], stacks[i+1:]...)
				i--
			}
		}
	}

	UserItems[userID][info.Type][itemID] = stacks
	itemMutex.Unlock()

	// 异步写 pika
	go func() {
		ctx := context.Background()
		err := repo.RedisHSet(ctx, redisKey, itemID, stacks)
		if err != nil {
			log.Errorf(ctx, -1, "Sub async pika failed for userID=%d itemID=%d, err=%v", userID, itemID, err)
		}
	}()

	return ""
}

// UpdatePropByID 更新道具属性（符文用）
func UpdatePropByID(userID int64, itemID int, id int64, prop []util.FuProp) {
	info := Table[itemID]
	if info == nil {
		return
	}
	initUserItemsByType(userID, info.Type)

	itemMutex.Lock()
	stacks := UserItems[userID][info.Type][itemID]
	if len(stacks) == 0 {
		itemMutex.Unlock()
		return
	}

	for i, v := range stacks {
		if v.Id == id {
			stacks[i].Prop = prop
		}
	}

	UserItems[userID][info.Type][itemID] = stacks
	itemMutex.Unlock()

	// 异步写 pika
	redisKey := fmt.Sprintf(KeyUserItems, info.Type, userID)
	go func() {
		ctx := context.Background()
		err := repo.RedisHSet(ctx, redisKey, itemID, stacks)
		if err != nil {
			log.Errorf(ctx, -1, "UpdatePropByID async pika failed for userID=%d itemID=%d, err=%v", userID, itemID, err)
		}
	}()
}

// NotEnough 校验 item 是否足够
func NotEnough(userID int64, itemId, needNum int) bool {
	userMoney := Total(userID, itemId)
	return userMoney < needNum
}

// Open 开箱逻辑
func Open(userID int64, itemID int, id int64, num int) ([]util.TypeNum, bool) {
	info := Table[itemID]
	if info == nil {
		return nil, false
	}
	initUserItemsByType(userID, info.Type)

	itemMutex.Lock()
	stacks := UserItems[userID][info.Type][itemID]
	if len(stacks) == 0 {
		itemMutex.Unlock()
		return nil, false
	}

	// 复制一份操作
	data := make([]*Stack, len(stacks))
	for i, s := range stacks {
		cp := *s
		data[i] = &cp
	}

	var openItems = []util.TypeNum{}
	ok := false

	for i, val := range data {
		if val.Id == id || id == 0 {
			valNum := val.Num
			if num > valNum {
				itemMutex.Unlock()
				return nil, false
			}
			if num == valNum {
				data = append(data[:i], data[i+1:]...)
			} else {
				data[i].Num = valNum - num
			}

			items := getOpenItems(itemID)
			if items == nil {
				itemMutex.Unlock()
				return nil, false
			}

			for _, openItem := range items {
				openItem.Num = openItem.Num * num
				openItems = append(openItems, openItem)
			}

			ok = true
			break
		}
	}

	if ok {
		UserItems[userID][info.Type][itemID] = data
	}
	itemMutex.Unlock()

	if ok {
		// 异步写 pika
		redisKey := fmt.Sprintf(KeyUserItems, info.Type, userID)
		go func() {
			ctx := context.Background()
			err := repo.RedisHSet(ctx, redisKey, itemID, data)
			if err != nil {
				log.Errorf(ctx, -1, "Open async pika failed for userID=%d itemID=%d, err=%v", userID, itemID, err)
			}
		}()

		// 发放开箱物品（这里会递归调用 Add，Add 已经是内存操作了）
		for _, openItem := range openItems {
			Add(userID, openItem.Type, openItem.Num, nil)
		}
	}
	return openItems, ok
}

// OpenAdd 直接开箱并发放（不从背包扣除）
func OpenAdd(userID int64, itemID, num int) bool {
	items := getOpenItems(itemID)
	if items == nil {
		return false
	}

	for _, openItem := range items {
		oNum := openItem.Num * num
		Add(userID, openItem.Type, oNum, nil)
	}
	return true
}

// GetOpenHeroItem 获取英雄卡的开箱内容（用于召唤系统）
// 返回 hero_id 和 star，与 PHP getOpenItems 对 type==6 的逻辑一致
func GetOpenHeroItem(itemID int) (heroID int, star int, ok bool) {
	items := getOpenItems(itemID)
	if len(items) == 0 {
		return 0, 0, false
	}
	if items[0].HeroId == 0 {
		return 0, 0, false
	}
	return items[0].HeroId, items[0].Star, true
}

// 获取开箱内容
func getOpenItems(itemID int) []util.TypeNum {
	info := Table[itemID]
	if info == nil {
		return nil
	}

	var open = util.ToTypeNums(info.OpenRaw)
	if len(open) == 0 {
		return nil
	}

	if info.Type == 6 {
		// 英雄卡
		if len(open) > 0 {
			open[0].HeroId = open[0].Type
			open[0].Star = open[0].Num
			return open
		}
		return nil
	}

	if len(open) == 1 {
		return open
	}

	// 检查是否有权重
	if open[0].Probablity > 0 {
		key := getKeyByWeight(open)
		return []util.TypeNum{open[key]}
	}
	return open
}

// 物品编号

var Seq int64
var SeqMutex sync.Mutex

func initSequence(ctx context.Context) {
	tmp, err := repo.RedisIncrBy(ctx, "useritem_sequence", 10000)
	if err != nil {
		panic(fmt.Errorf("sequence error: %v", err))
	}
	Seq = tmp
}

// Sequence 获取自增序列号（用于道具唯一ID / 还用于生成临时怪物英雄ID）
func Sequence() int64 {
	var ret int64

	SeqMutex.Lock()
	Seq = Seq + 1 + int64(rand.Intn(3))
	ret = Seq
	SeqMutex.Unlock()

	go func() {
		err := repo.RedisSet(context.Background(), "useritem_sequence", ret)
		if err != nil {
			log.Errorf(context.Background(), 0, "redis set useritem_sequence error: %v", err)
		}
	}()

	return ret
}

// 按权重随机取 key
func getKeyByWeight(items []util.TypeNum) int {
	totalWeight := 0
	for _, item := range items {
		if item.Probablity > 0 {
			totalWeight += item.Probablity
		}
	}
	if totalWeight <= 0 {
		return 0
	}

	r := rand.Intn(totalWeight)
	cumWeight := 0
	for i, item := range items {
		cumWeight += item.Probablity
		if r < cumWeight {
			return i
		}
	}
	return 0
}
