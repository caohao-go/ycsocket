package controller

import (
	"context"

	"server_golang/common/types"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/cache"
)

// 聊天系统

func (c *ShinelightController) ChatAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	chatType := c.Params.GetIntE("type")
	content := c.Params.GetStringE("content")
	if content == "" {
		return c.ResponseError(13342339, "内容不能为空")
	}

	// 敏感词过滤（与 PHP chatAction 一致：str_replace($sensitive_words, "*", $content)）
	content = logic.FilterSensitiveWords(content)

	result := types.Map{
		"userid": userID, "type": chatType,
		"nickname":   c.Params.GetStringE("nickname"),
		"avatar_url": c.Params.GetStringE("avatar_url"),
		"gender":     c.Params.GetStringE("gender"),
		"vip_level":  c.Params.GetStringE("vip_level"),
		"lv":         c.Params.GetStringE("lv"),
		"content":    content,
	}

	if chatType == 0 {
		// 世界聊天（与 PHP chatAction type==0 一致）
		// 引导任务（与 PHP 一致：type==0 时触发）
		model.GuideTaskHandle(ctx, userID, 49, 1)

		// 存储世界聊天历史（与 PHP 一致：rpush("chat_his_world", serialize($result))，最多50条）
		cache.PushChatHistory(cache.ChatHistoryWorldKey(), result)

		return c.ResponseSuccessToAll(result)
	} else if chatType == 2 {
		// 公会聊天（与 PHP chatAction type==2 一致）
		guildID := model.GetUsersGuildID(ctx, userID)
		if guildID == 0 {
			return c.ResponseError(9382, "未加入行会")
		}

		// 存储公会聊天历史（与 PHP 一致：rpush("chat_his_guild_{$guild_id}", serialize($result))，最多50条）
		cache.PushChatHistory(cache.ChatHistoryGuildKey(guildID), result)

		// 获取公会所有成员ID，发送给公会全员（与 PHP 一致）
		guildMembers := model.GetGuildsUser(ctx, guildID, true, 0)
		uids := make([]int64, 0, len(guildMembers))
		for _, m := range guildMembers {
			uid := m.GetInt64E("user_id")
			if uid > 0 {
				uids = append(uids, uid)
			}
		}
		if len(uids) > 0 {
			return c.ResponseSuccessToUids(uids, result)
		}
		return c.ResponseSuccessToMe(result)
	}

	// 私聊（与 PHP chatAction 一致）
	toUserID := c.Params.GetInt64E("to_userid")
	if toUserID > 0 {
		return c.ResponseSuccessToUids([]int64{userID, toUserID}, result)
	}
	return c.ResponseSuccessToAll(result)
}

func (c *ShinelightController) ChatHisAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	chatType := c.Params.GetIntE("type")

	var list []types.Map

	if chatType == 2 {
		// 公会聊天历史（与 PHP chatHisAction type==2 一致：lrange("chat_his_guild_{$guild_id}", 0, 49)）

		guildID := model.GetUsersGuildID(ctx, userID)
		if guildID > 0 {
			list = cache.GetChatHistory(cache.ChatHistoryGuildKey(guildID))
		}
	} else {
		// 世界聊天历史（与 PHP chatHisAction 一致：lrange("chat_his_world", 0, 49)）
		list = cache.GetChatHistory(cache.ChatHistoryWorldKey())
	}

	if list == nil {
		list = make([]types.Map, 0)
	}

	return c.ResponseSuccessToMe(types.Map{"list": list})
}
