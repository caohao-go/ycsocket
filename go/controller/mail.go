package controller

import (
	"context"
	"strings"

	"server_golang/common/json"
	"server_golang/common/types"
	"server_golang/common/util"
	"server_golang/model"
)

// 邮件系统

func (c *ShinelightController) GetUserMailAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	mails := model.GetUserMail(ctx, userID)
	if mails == nil {
		mails = make([]types.Map, 0)
	}
	return c.ResponseSuccessToMe(types.Map{"mails": mails})
}

func (c *ShinelightController) ReadMailAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	mailID := c.Params.GetInt64E("id")
	model.ReadUserMail(ctx, userID, mailID)
	return c.ResponseSuccessToMe(types.Map{})
}

// DelMailAction 删除邮件（与 PHP delMailAction 一致：支持逗号分隔多 ID）
func (c *ShinelightController) DelMailAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}
	idsStr := c.Params.GetStringE("ids")
	var ids []int64
	if idsStr != "" {
		for _, idStr := range strings.Split(idsStr, ",") {
			id := types.ToInt64E(strings.TrimSpace(idStr))
			if id > 0 {
				ids = append(ids, id)
			}
		}
	}
	if len(ids) == 0 {
		return c.ResponseSuccessToMe(types.Map{})
	}
	model.DelUserMail(ctx, userID, ids)
	return c.ResponseSuccessToMe(types.Map{})
}

// LingquMailItemsAction 领取邮件附件（与 PHP lingquMailItemsAction 一致）
func (c *ShinelightController) LingquMailItemsAction(ctx context.Context) *Result {
	userID, _, _, err := c.getAuthUser(ctx)
	if err != nil {
		return c.ResponseError(99900031, err.Error())
	}

	// 解析 ids 参数（逗号分隔）
	idsStr := c.Params.GetStringE("ids")
	var ids []int64
	if idsStr != "" {
		for _, idStr := range strings.Split(idsStr, ",") {
			id := types.ToInt64E(strings.TrimSpace(idStr))
			if id > 0 {
				ids = append(ids, id)
			}
		}
	}

	if len(ids) == 0 {
		return c.ResponseError(99900032, "ids is empty")
	}

	// 1. 获取邮件列表（与 PHP 一致：getMailsByIds）
	data := model.GetMailsByIDs(ctx, userID, ids)

	// 2. 遍历邮件，找出未领取的附件（与 PHP 一致：item_get_flag==1）
	var lingquIDs []int64
	var lingquItems []util.TypeNum
	for _, val := range data {
		if val.ItemGetFlag == 1 {
			// 解析 add_items
			addItems := []util.TypeNum{}
			json.Unmarshal(val.AddItems, &addItems)
			lingquItems = util.Merge(lingquItems, addItems)
			lingquIDs = append(lingquIDs, int64(val.Id))
		}
	}

	// 3. 标记为已领取（与 PHP 一致：lingquMail）
	if len(lingquIDs) > 0 {
		model.LingquMail(ctx, userID, lingquIDs)
	}

	// 4. 发放奖励（与 PHP 一致：giveRewards）
	if len(lingquItems) > 0 {
		model.GiveReward(userID, lingquItems...)
	}

	return c.ResponseSuccessToMe(types.Map{"lingqu": lingquItems})
}

// ======================== 好友系统 ========================
