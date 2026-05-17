// 聊天历史缓存（替代 PHP 的 Redis rpush/lpop 聊天历史存储）
package cache

import (
	"fmt"
	"sync"

	"server_golang/common/json"

	"server_golang/common/types"
)

const maxChatHistory = 50

// ChatHistory 聊天历史存储（内存版，替代 Redis list）
type ChatHistory struct {
	mu   sync.RWMutex
	data map[string][]string // key => []serialized_message（最多50条）
}

var chatHis = &ChatHistory{
	data: make(map[string][]string),
}

// PushChatHistory 追加聊天记录（对应 PHP 的 rpush + 50条上限 lpop）
func PushChatHistory(key string, message interface{}) {
	chatHis.mu.Lock()
	defer chatHis.mu.Unlock()

	chatHis.data[key] = append(chatHis.data[key], json.Marshal(message))

	// 保留最多50条（与 PHP Redis rpush + if llen > 50 then lpop 一致）
	if len(chatHis.data[key]) > maxChatHistory {
		chatHis.data[key] = chatHis.data[key][len(chatHis.data[key])-maxChatHistory:]
	}
}

// GetChatHistory 获取聊天记录（对应 PHP 的 lrange(key, 0, 49)）
func GetChatHistory(key string) []types.Map {
	chatHis.mu.RLock()
	defer chatHis.mu.RUnlock()

	items := chatHis.data[key]
	result := make([]types.Map, 0, len(items))
	for _, item := range items {
		msg := types.ToMapE(item)
		if msg != nil {
			result = append(result, msg)
		}
	}
	return result
}

// ChatHistoryWorldKey 世界聊天历史 key
func ChatHistoryWorldKey() string {
	return "chat_his_world"
}

// ChatHistoryGuildKey 公会聊天历史 key
func ChatHistoryGuildKey(guildID int) string {
	return fmt.Sprintf("chat_his_guild_%d", guildID)
}
