package controller

import (
	"context"
	"strings"

	"server_golang/common/types"
	"server_golang/config"
	"server_golang/logic"
	"server_golang/model"
	"server_golang/repo/mem/item"
)

// 好友系统

// AddFriendAction 添加好友（对应 PHP addFriendAction）
func (c *ShinelightController) AddFriendAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	friendUID := c.Params.GetInt64E("friend_uid")

	// 不能加自己为好友
	if userID == friendUID {
		return c.ResponseError(95745, "不能加自己为好友")
	}

	// 检查好友用户是否存在
	friendUserInfo := model.GetUserAttr(friendUID)
	if friendUserInfo.GetIntE("init_flag") == 0 {
		return c.ResponseError(9575, "用户不存在")
	}

	// 检查是否已有好友关系（双向查找）
	friendInfo := model.GetUserFriendPair(ctx, userID, friendUID)
	if friendInfo != nil {
		return c.ResponseSuccessToMe(types.Map{})
	}

	// 插入好友申请（status=1 表示待处理）
	model.InsertUserFriendPair(ctx, userID, friendUID, 1)
	return c.ResponseSuccessToMe(types.Map{})
}

// DelFriendAction 删除好友（对应 PHP delFriendAction）
func (c *ShinelightController) DelFriendAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	friendUID := c.Params.GetInt64E("friend_uid")

	model.DelUserFriendPair(ctx, userID, friendUID)
	return c.ResponseSuccessToMe(types.Map{})
}

// FriendsListAction 好友列表（对齐 PHP friendsListAction）
func (c *ShinelightController) FriendsListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 获取好友UID列表（status=0 已确认好友，与 PHP getFriendsList 默认 status=0 一致）
	uids := model.GetFriendsList(ctx, userID, 0)

	// 批量获取好友详情（与 PHP 一致：getMuiltiUsersDetail($uids, ['lv', 'fight_point'])，包含 lv 和 fight_point）
	list := model.GetUsersWithDetail(ctx, uids, 1, config.AttrLv, config.AttrFightPoint)

	// 获取今日送心列表
	loverList := model.GetLoversList(ctx, userID)

	// 组装返回数据，添加 lover 字段（与 PHP 一致：intval($lover_list[$val['user_id']])）
	retList := make([]types.Map, 0, len(list))
	for uid, info := range list {
		uidStr := types.ToString(uid)
		loverVal := 0
		if v, ok := loverList[uidStr]; ok {
			loverVal = types.ToIntE(v)
		}
		info["lover"] = loverVal
		retList = append(retList, info)
	}

	// 获取用户持有的心道具数量（itemID=10）

	lovers := item.Total(userID, 10)

	// 与 PHP 一致：return $this->response_success_to_me(['lovers' => $lovers, 'list' => array_values($list)])
	return c.ResponseSuccessToMe(types.Map{"lovers": lovers, "list": retList})
}

// RecFriendListAction 推荐好友列表（对应 PHP recFriendListAction）
func (c *ShinelightController) RecFriendListAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data := logic.GetRandRecFriends(ctx)
	if data == nil {
		data = []types.Map{}
	}
	return c.ResponseSuccessToMe(types.Map{"list": data})
}

// ApplyFriendsListAction 好友申请列表（对应 PHP applyFriendsListAction）
func (c *ShinelightController) ApplyFriendsListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	data, _ := model.GetApplyFriendsList(ctx, userID)
	if data == nil {
		return c.ResponseSuccessToMe(types.Map{"list": []interface{}{}})
	}
	return c.ResponseSuccessToMe(types.Map{"list": data})
}

// HandleApplyAction 处理好友申请（对应 PHP handleApplyAction）
func (c *ShinelightController) HandleApplyAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	friendUID := c.Params.GetInt64E("friend_uid")
	accept := c.Params.GetIntE("accept") // 0-拒绝 1-接受

	if accept == 0 {
		// 拒绝：删除好友关系
		model.DelUserFriendPair(ctx, userID, friendUID)
	} else {
		// 接受：更新状态为 0（已确认好友）
		model.UpdateUserFriendPair(ctx, userID, friendUID, 0)
	}

	return c.ResponseSuccessToMe(types.Map{})
}

// SendLoverAction 赠送爱心（对应 PHP sendLoverAction）
func (c *ShinelightController) SendLoverAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	friendUIDStr := c.Params.GetStringE("friend_uid")
	friendUIDs := strings.Split(friendUIDStr, ",")

	// 获取今日已送心列表
	loverList := model.GetLoversList(ctx, userID)
	if len(loverList) > 10 {
		return c.ResponseError(9555, "最多送10个")
	}

	for _, uidStr := range friendUIDs {
		uid := types.ToInt64E(uidStr)
		if uid == 0 {
			continue
		}
		// 已经送过的跳过
		if _, ok := loverList[uidStr]; ok {
			continue
		}
		model.SetLovers(ctx, userID, uid)
	}

	return c.ResponseSuccessToMe(types.Map{})
}

// LoverReceiveListAction 可领取爱心列表（对齐 PHP loverReceiveListAction）
func (c *ShinelightController) LoverReceiveListAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 获取收到的爱心列表（与 PHP 一致：RedisProxy::getLoversReceiveList($userId)）
	loverList := model.GetLoversReceiveList(ctx, userID)
	uids := make([]int64, 0, len(loverList))
	for uidStr := range loverList {
		uid := types.ToInt64E(uidStr)
		if uid > 0 {
			uids = append(uids, uid)
		}
	}

	// 批量获取用户详情（与 PHP 一致：getMuiltiUsersDetail($uids, ['lv', 'fight_point'])）
	list := model.GetUsersWithDetail(ctx, uids, 1, config.AttrLv, config.AttrFightPoint)

	// 与 PHP 一致：response_success_to_me(['data' => $list])
	// PHP 返回的 $list 是以 user_id 为 key 的关联数组（object）
	// getMuiltiUsersDetail 返回 $user_info_tmp[zone_user_id] = value 形式
	retList := make([]types.Map, 0, len(list))
	for _, info := range list {
		retList = append(retList, info)
	}

	return c.ResponseSuccessToMe(types.Map{"data": retList})
}

// LingquLoverAction 领取爱心（对应 PHP lingquLoverAction）
func (c *ShinelightController) LingquLoverAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	friendUIDStr := c.Params.GetStringE("friend_uid")
	friendUIDs := strings.Split(friendUIDStr, ",")

	num := 0
	receiveLoverList := model.GetLoversReceiveList(ctx, userID)
	for _, uidStr := range friendUIDs {
		uid := types.ToInt64E(uidStr)
		if uid == 0 {
			continue
		}
		// 没有可领取的跳过
		if _, ok := receiveLoverList[uidStr]; !ok {
			continue
		}
		num += 10
		model.LingquLovers(ctx, userID, uid)
	}

	// 发放爱心道具（itemID=10）
	if num > 0 {

		item.Add(userID, 10, num, nil)
	}
	return c.ResponseSuccessToMe(types.Map{})
}

// GetFriendByNameAction 按昵称搜索好友（对应 PHP getFriendByNameAction）
func (c *ShinelightController) GetFriendByNameAction(ctx context.Context) *Result {
	_, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	nickname := c.Params.GetStringE("nickname")

	// 模糊搜索用户
	data := model.GetUserByName(ctx, nickname)
	if data == nil || len(data) == 0 {
		return c.ResponseSuccessToMe(types.Map{"data": []interface{}{}})
	}

	// 收集 user_id 数组
	userIDs := make([]int64, 0, len(data))
	for _, v := range data {
		userIDs = append(userIDs, v.UserId)
	}

	// 批量获取用户等级信息
	usersGrades := model.GetMultiUserAttr(userIDs)

	ret := make([]types.Map, len(data))
	for k := range data {
		ret[k] = types.Map{
			"user_id":     data[k].UserId,
			"nickname":    usersGrades[data[k].UserId].GetStringE("nickname"),
			"avatar_url":  usersGrades[data[k].UserId].GetStringE("avatar_url"),
			"lv":          usersGrades[data[k].UserId].GetIntE("lv"),
			"fight_point": usersGrades[data[k].UserId].GetIntE("fight_point"),
		}
	}

	return c.ResponseSuccessToMe(types.Map{"data": ret})
}
