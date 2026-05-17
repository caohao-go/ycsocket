package userhero

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
	"git.code.oa.com/pcg-csd/trpc-ext/util/log"
	"server_golang/common/types"
	"server_golang/repo"
	"server_golang/repo/table"
)

// heros 以 UserHero.Id 为 key 存储所有英雄对象
var heros = map[int]table.UserHero{}

// userHeroIds 以 user_id 为 key，存储该用户拥有的所有英雄 Id 列表
var userHeroIds = map[int64][]int{}

var heroMutex = sync.RWMutex{}

// InitUserHeros 将用户英雄数据从数据库加载到内存
func InitUserHeros(userID int64) {
	heroMutex.RLock()
	if _, ok := userHeroIds[userID]; ok {
		heroMutex.RUnlock()
		return
	}
	heroMutex.RUnlock()

	ctx := context.Background()
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_hero").AndEqual("user_id", userID)

	dest := []*table.UserHero{}
	err := client.FindAll(ctx, stmt, &dest)
	if err != nil {
		panic(fmt.Errorf("initUserHeros failed for userID=%d, err=%v", userID, err))
	}

	heroMutex.Lock()
	ids := make([]int, 0, len(dest))
	for _, h := range dest {
		heros[h.Id] = *h
		ids = append(ids, h.Id)
	}
	userHeroIds[userID] = ids
	heroMutex.Unlock()
}

// GetUserHeroList 获取用户英雄列表
func GetUserHeroList(userID int64) []*table.UserHero {
	InitUserHeros(userID)

	heroMutex.RLock()
	ids := userHeroIds[userID]
	result := make([]*table.UserHero, 0, len(ids))
	for _, id := range ids {
		h := heros[id]
		result = append(result, &h)
	}
	heroMutex.RUnlock()
	return result
}

// GetUserHeroById 根据 ID 获取单个英雄
// 先从内存取，有则返回；没有则到数据库按 id 查找，
// 找不到返回 nil，找到了拿到 user_id 后调用 InitUserHeros 初始化，再从内存返回
func GetUserHeroById(id int) *table.UserHero {
	// 先尝试从内存获取
	heroMutex.RLock()
	h, ok := heros[id]
	heroMutex.RUnlock()
	if ok {
		return &h
	}

	// 内存没有，从数据库按 id 查找
	ctx := context.Background()
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_hero").AndEqual("id", id)

	dest := &table.UserHero{}
	err := client.FindOne(ctx, stmt, dest)
	if err != nil || dest.Id == 0 {
		return nil
	}

	// 找到了，拿到 user_id，调用 InitUserHeros 加载该用户所有英雄到内存
	InitUserHeros(dest.UserId)

	// 从内存返回
	heroMutex.RLock()
	h, ok = heros[id]
	heroMutex.RUnlock()
	if !ok {
		return nil
	}
	return &h
}

// GetUserHeroByIds 根据多个 ID 获取英雄列表
// 先从内存取，记录内存没有的，然后从数据库批量查找这些 id 对应的 user_id，
// 去重后逐个调用 InitUserHeros 初始化，最后重新从内存取所有 ids 对应的英雄
func GetUserHeroByIds(ids []int) []*table.UserHero {
	if len(ids) == 0 {
		return nil
	}

	// 先从内存取，记录缺失的 id
	heroMutex.RLock()
	missingIds := make([]int, 0)
	for _, id := range ids {
		if _, ok := heros[id]; !ok {
			missingIds = append(missingIds, id)
		}
	}
	heroMutex.RUnlock()

	// 如果有缺失的，从数据库批量查找
	if len(missingIds) > 0 {
		ctx := context.Background()
		client := repo.WorldDB()
		stmt := repo.NewStmt()
		stmt.SetTableName("user_hero").AndIn("id", missingIds)

		dest := []*table.UserHero{}
		err := client.FindAll(ctx, stmt, &dest)
		if err == nil && len(dest) > 0 {
			// 去重
			uidSet := map[int64]bool{}
			uniqueUserIds := make([]int64, 0, len(dest))
			for _, tmp := range dest {
				if _, exists := uidSet[tmp.UserId]; !exists {
					uidSet[tmp.UserId] = true
					uniqueUserIds = append(uniqueUserIds, tmp.UserId)
				}
			}

			// 一次查询所有未加载用户的英雄数据
			allStmt := repo.NewStmt()
			allStmt.SetTableName("user_hero").AndIn("user_id", uniqueUserIds)

			allHeros := []*table.UserHero{}
			err2 := client.FindAll(ctx, allStmt, &allHeros)
			if err2 == nil {
				// 按 user_id 分组
				grouped := map[int64][]*table.UserHero{}
				for _, h := range allHeros {
					grouped[h.UserId] = append(grouped[h.UserId], h)
				}

				// 批量写入内存
				heroMutex.Lock()
				for uid, heroList := range grouped {
					if _, ok := userHeroIds[uid]; ok {
						continue
					}
					heroIds := make([]int, 0, len(heroList))
					for _, h := range heroList {
						heros[h.Id] = *h
						heroIds = append(heroIds, h.Id)
					}
					userHeroIds[uid] = heroIds
				}
				heroMutex.Unlock()
			}
		}
	}

	// 重新从内存取所有 ids 对应的英雄
	heroMutex.RLock()
	result := make([]*table.UserHero, 0, len(ids))
	for _, id := range ids {
		if h, ok := heros[id]; ok {
			result = append(result, &h)
		}
	}
	heroMutex.RUnlock()
	return result
}

// GetUserCountHero 获取用户英雄数量
func GetUserCountHero(userID int64) int {
	InitUserHeros(userID)

	heroMutex.RLock()
	count := len(userHeroIds[userID])
	heroMutex.RUnlock()
	return count
}

// InsertUserHero 插入英雄：同步更新内存，异步写数据库
func InsertUserHero(userID int64, data *table.UserHero) (int, error) {
	// 先同步写数据库获取自增 ID
	ctx := context.Background()
	client := repo.WorldDB()
	stmt := repo.NewStmt()
	stmt.SetTableName("user_hero").InsertStruct(data)
	id, err := client.Insert(ctx, stmt)
	if userID > 10000 && id < 100000000 {
		newId := 100000001 + rand.Intn(10)
		_, err = client.Exec(ctx, fmt.Sprintf("update user_hero set id=%d where id=%d", newId, id))
		id = int64(newId)
	}
	if err != nil {
		log.Errorf(ctx, -1, "InsertUserHero DB failed for userID=%d heroID=%d, err=%v", userID, data.HeroId, err)
		return 0, err
	}

	if data.Lv <= 0 {
		data.Lv = 1
	}

	data.Id = int(id)

	// 同步更新内存
	heroMutex.Lock()
	data.Updatetime = time.Now()
	heros[data.Id] = *data
	userHeroIds[userID] = append(userHeroIds[userID], data.Id)
	heroMutex.Unlock()

	return data.Id, nil
}

// UpdateUserHeroById 更新英雄字段：同步更新内存，异步写数据库
func UpdateUserHeroById(userID int64, id int, updateData types.Map) error {
	InitUserHeros(userID)

	// 同步更新内存
	heroMutex.Lock()
	h, ok := heros[id]
	if !ok {
		heroMutex.Unlock()
		return fmt.Errorf("hero id=%d not found in memory", id)
	}
	applyUpdateToHero(&h, updateData)
	heros[id] = h
	heroMutex.Unlock()

	// 异步写数据库
	go func() {
		ctx := context.Background()
		client := repo.WorldDB()
		stmt := repo.NewStmt()
		stmt.SetTableName("user_hero").
			UpdateMap(extorm.SetMap(updateData)).
			AndEqual("id", id)
		_, err := client.Update(ctx, stmt)
		if err != nil {
			log.Errorf(ctx, -1, "UpdateUserHeroById DB failed for id=%d, err=%v", id, err)
		}
	}()

	return nil
}

// IncrUserHeroLvStarStage 增量更新英雄 lv/star/stage：同步更新内存，异步写数据库
func IncrUserHeroLvStarStage(userID int64, id, lv, star, stage int) error {
	if lv == 0 && star == 0 && stage == 0 {
		return nil
	}

	InitUserHeros(userID)

	// 同步更新内存
	heroMutex.Lock()
	h, ok := heros[id]
	if !ok {
		heroMutex.Unlock()
		return fmt.Errorf("hero id=%d not found in memory", id)
	}
	h.Lv += lv
	h.Star += star
	h.Stage += stage
	h.Updatetime = time.Now()
	heros[id] = h
	heroMutex.Unlock()

	// 异步写数据库
	go func() {
		ctx := context.Background()
		client := repo.WorldDB()
		sql := fmt.Sprintf("UPDATE user_hero SET lv=lv+(%d), star=star+(%d), stage=stage+(%d) WHERE id=?", lv, star, stage)
		_, err := client.Exec(ctx, sql, id)
		if err != nil {
			log.Errorf(ctx, -1, "IncrUserHeroLvStarStage DB failed for id=%d, err=%v", id, err)
		}
	}()

	return nil
}

// DeleteUserHeroByIds 删除英雄：同步更新内存，异步写数据库
func DeleteUserHeroByIds(userID int64, ids []int) error {
	InitUserHeros(userID)

	// 同步更新内存
	heroMutex.Lock()
	for _, id := range ids {
		delete(heros, id)
	}
	// 从 userHeroIds 中移除
	remaining := make([]int, 0)
	for _, existID := range userHeroIds[userID] {
		found := false
		for _, delID := range ids {
			if existID == delID {
				found = true
				break
			}
		}
		if !found {
			remaining = append(remaining, existID)
		}
	}
	userHeroIds[userID] = remaining
	heroMutex.Unlock()

	// 异步写数据库
	go func() {
		ctx := context.Background()
		client := repo.WorldDB()
		stmt := repo.NewStmt()
		stmt.SetTableName("user_hero").AndIn("id", ids)
		_, err := client.Delete(ctx, stmt)
		if err != nil {
			log.Errorf(ctx, -1, "DeleteUserHeroByIds DB failed for userID=%d ids=%v, err=%v", userID, ids, err)
		}
	}()

	return nil
}

// 将 updateData 中的字段应用到 hero 结构体
func applyUpdateToHero(h *table.UserHero, updateData types.Map) {
	if v, ok := updateData["hero_id"]; ok {
		h.HeroId = types.ToIntE(v)
	}
	if v, ok := updateData["star"]; ok {
		h.Star = types.ToIntE(v)
	}
	if v, ok := updateData["stage"]; ok {
		h.Stage = types.ToIntE(v)
	}
	if v, ok := updateData["lv"]; ok {
		h.Lv = types.ToIntE(v)
	}
	if v, ok := updateData["fit"]; ok {
		h.Fit = types.ToString(v)
	}
	if v, ok := updateData["fu"]; ok {
		h.Fu = types.ToString(v)
	}
	h.Updatetime = time.Now()
}
