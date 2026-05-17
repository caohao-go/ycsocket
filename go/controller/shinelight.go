package controller

import (
	"context"
	"fmt"

	"server_golang/model"
	"server_golang/repo/table"
)

// 路由分发（dispatch + 结构体定义）

// ShinelightController 处理主游戏逻辑
// 原 PHP Shinelight 控制器包含 218+ 个接口方法，全部游戏逻辑在此路由。
type ShinelightController struct {
	*BaseController
}

// getAuthUser 通用的用户认证提取方法
func (c *ShinelightController) getAuthUser(ctx context.Context) (int64, string, *table.UserInfo, error) {
	userID := c.Params.GetInt64E("userid")
	token := c.Params.GetStringE("token")
	sourceUID := model.GetUIDByZoneUserID(userID)
	userInfo, err := model.GetUserAndAuth(ctx, sourceUID, token)
	return userID, token, userInfo, err
}

// dispatchShinelight 将 action 分发到对应的处理方法
func dispatchShinelight(ctx context.Context, base *BaseController, action string) *Result {
	c := &ShinelightController{BaseController: base}

	switch action {
	// ==================== 基础信息 ====================
	case "ver":
		return c.VerAction(ctx)
	case "switch":
		return c.SwitchAction(ctx)
	case "gonggao":
		return c.GonggaoAction(ctx)
	case "config":
		return c.ConfigAction(ctx)
	case "redPointInit":
		return c.RedPointInitAction(ctx)

	// ==================== 聊天系统 ====================
	case "chat":
		return c.ChatAction(ctx)
	case "chatHis":
		return c.ChatHisAction(ctx)

	// ==================== 用户信息 ====================
	case "updateUserinfo":
		return c.UpdateUserinfoAction(ctx)
	case "nicknameSame":
		return c.NicknameSameAction(ctx)
	case "userInfo":
		return c.UserInfoAction(ctx)
	case "userSpace":
		return c.UserSpaceAction(ctx)
	case "otherUserSpace":
		return c.OtherUserSpaceAction(ctx)

	// ==================== 排行榜 ====================
	case "rankBest":
		return c.RankBestAction(ctx)
	case "rankList":
		return c.RankListAction(ctx)
	case "heroRankList":
		return c.HeroRankListAction(ctx)

	// ==================== 新人礼包 ====================
	case "getNewGift":
		return c.GetNewGiftAction(ctx)
	case "lingquNew7dayLogin":
		return c.LingquNew7dayLoginAction(ctx)
	case "lingquNewGift":
		return c.LingquNewGiftAction(ctx)

	// ==================== 道具系统 ====================
	case "openItem":
		return c.OpenItemAction(ctx)
	case "heroCombine":
		return c.HeroCombineAction(ctx)
	case "userItem":
		return c.UserItemAction(ctx)
	case "userEquipItem":
		return c.UserEquipItemAction(ctx)

	// ==================== 商店系统 ====================
	case "getShop":
		return c.GetShopAction(ctx)
	case "sellItem":
		return c.SellItemAction(ctx)
	case "shopBuy":
		return c.ShopBuyAction(ctx)
	case "itemBuy":
		return c.ItemBuyAction(ctx)
	case "refreshShop":
		return c.RefreshShopAction(ctx)

	// ==================== 熔炼系统 ====================
	case "ronglianLingqu":
		return c.RonglianLingquAction(ctx)
	case "mergeEquip":
		return c.MergeEquipAction(ctx)
	case "mergeEquipAll":
		return c.MergeEquipAllAction(ctx)
	case "mergeRune":
		return c.MergeRuneAction(ctx)

	// ==================== 符文系统 ====================
	case "userRuneReset":
		return c.UserRuneResetAction(ctx)
	case "saveRune":
		return c.SaveRuneAction(ctx)
	case "userRuneItem":
		return c.UserRuneItemAction(ctx)
	case "dressRune":
		return c.DressRuneAction(ctx)
	case "unloadRune":
		return c.UnloadRuneAction(ctx)

	// ==================== 装备系统 ====================
	case "dressEquip":
		return c.DressEquipAction(ctx)
	case "unloadEquip":
		return c.UnloadEquipAction(ctx)

	// ==================== 英雄系统 ====================
	case "userHero":
		return c.UserHeroAction(ctx)
	case "savePostion":
		return c.SavePostionAction(ctx)
	case "otherUserHeroProp":
		return c.OtherUserHeroPropAction(ctx)
	case "userHeroProp":
		return c.UserHeroPropAction(ctx)
	case "upHeroLv":
		return c.UpHeroLvAction(ctx)
	case "upHeroStageInfo":
		return c.UpHeroStageInfoAction(ctx)
	case "upHeroStage":
		return c.UpHeroStageAction(ctx)
	case "heroStarUpInfo":
		return c.HeroStarUpInfoAction(ctx)
	case "heroStarUpDetailARongheshendian":
		return c.HeroStarUpDetailARongheshendianAction(ctx)
	case "heroStarUpDetail":
		return c.HeroStarUpDetailAction(ctx)
	case "heroStarUp":
		return c.HeroStarUpAction(ctx)

	// ==================== 献祭系统 ====================
	case "sacrificeInfo":
		return c.SacrificeInfoAction(ctx)
	case "sacrifice":
		return c.SacrificeAction(ctx)

	// ==================== 召唤系统 ====================
	case "zhaohuanInfo":
		return c.ZhaohuanInfoAction(ctx)
	case "heroZhao":
		return c.HeroZhaoAction(ctx)

	// ==================== 先知系统 ====================
	case "xianzhiInfo":
		return c.XianzhiInfoAction(ctx)
	case "xianzhiZhao":
		return c.XianzhiZhaoAction(ctx)
	case "heroExchange":
		return c.HeroExchangeAction(ctx)
	case "saveHeroExchange":
		return c.SaveHeroExchangeAction(ctx)

	// ==================== 副本系统 ====================
	case "copyFight":
		return c.CopyFightAction(ctx)
	case "getFightHeros":
		return c.GetFightHerosAction(ctx)
	case "copyRewardStat":
		return c.CopyRewardStatAction(ctx)
	case "lingquCopyReward":
		return c.LingquCopyRewardAction(ctx)
	case "getCountReward":
		return c.GetCountRewardAction(ctx)

	// ==================== 日常副本 ====================
	case "allFunction":
		return c.AllFunctionAction(ctx)
	case "getFunction":
		return c.GetFunctionAction(ctx)
	case "finishFunction":
		return c.FinishFunctionAction(ctx)

	// ==================== 爬塔系统 ====================
	case "climbLingqu":
		return c.ClimbLingquAction(ctx)
	case "climbtowerInfo":
		return c.ClimbtowerInfoAction(ctx)
	case "climbtowerRank":
		return c.ClimbtowerRankAction(ctx)
	case "climbtowerSaodang":
		return c.ClimbtowerSaodangAction(ctx)
	case "climbSaodangInfo":
		return c.ClimbSaodangInfoAction(ctx)
	case "crossClimbtower":
		return c.CrossClimbtowerAction(ctx)
	case "AlreadycrossClimbtower":
		return c.AlreadyCrossClimbtowerAction(ctx)

	// ==================== 无尽模式 ====================
	case "endlessInfo":
		return c.EndlessInfoAction(ctx)
	case "endlessFight":
		return c.EndlessFightAction(ctx)
	case "endlessRank":
		return c.EndlessRankAction(ctx)
	case "endlessHelpList":
		return c.EndlessHelpListAction(ctx)
	case "endlessHelpAdd":
		return c.EndlessHelpAddAction(ctx)
	case "endlessChooseHelp":
		return c.EndlessChooseHelpAction(ctx)
	case "endlessFirstLingqu":
		return c.EndlessFirstLingquAction(ctx)
	case "resetTodayEndless":
		return c.ResetTodayEndlessAction(ctx)

	// ==================== 远征系统 ====================
	case "getExpeditionInfo":
		return c.GetExpeditionInfoAction(ctx)
	case "chooseExpedition":
		return c.ChooseExpeditionAction(ctx)
	case "getExpeditionLayer":
		return c.GetExpeditionLayerAction(ctx)
	case "upExpedition":
		return c.UpExpeditionAction(ctx)
	case "openBaoxiang":
		return c.OpenBaoxiangAction(ctx)
	case "expeditionHelpList":
		return c.ExpeditionHelpListAction(ctx)
	case "expeditionHelpAdd":
		return c.ExpeditionHelpAddAction(ctx)
	case "expeditionChooseHelp":
		return c.ExpeditionChooseHelpAction(ctx)

	// ==================== 电币系统 ====================
	case "getDianCoin":
		return c.GetDianCoinAction(ctx)
	case "buyDianCoin":
		return c.BuyDianCoinAction(ctx)

	// ==================== 每日奖励 ====================
	case "dailyreward":
		return c.DailyrewardAction(ctx)
	case "getdailyreward":
		return c.GetdailyrewardAction(ctx)
	case "getvipdailyreward":
		return c.GetvipdailyrewardAction(ctx)

	// ==================== 任务系统 ====================
	case "taskInfo":
		return c.TaskInfoAction(ctx)
	case "weekTaskLingqu":
		return c.WeekTaskLingquAction(ctx)
	case "taskScoreRank":
		return c.TaskScoreRankAction(ctx)
	case "lingquIGift":
		return c.LingquIGiftAction(ctx)

	// ==================== 日常任务 ====================
	case "task":
		return c.TaskAction(ctx)
	case "taskLingqu":
		return c.TaskLingquAction(ctx)
	case "taskActiveLingqu":
		return c.TaskActiveLingquAction(ctx)

	// ==================== 成就任务 ====================
	case "achieveTask":
		return c.AchieveTaskAction(ctx)
	case "achieveTaskLingqu":
		return c.AchieveTaskLingquAction(ctx)

	// ==================== 星河神殿 ====================
	case "templateInfo":
		return c.templateInfoAction(ctx)
	case "templateDetail":
		return c.templateDetailAction(ctx)
	case "upTemplate":
		return c.UpTemplateAction(ctx)

	// ==================== 探宝系统 ====================
	case "tanbaoInfo":
		return c.TanbaoInfoAction(ctx)
	case "tanbao":
		return c.TanbaoAction(ctx)
	case "tanbaoLingqu":
		return c.TanbaoLingquAction(ctx)
	case "tanbaoRefresh":
		return c.TanbaoRefreshAction(ctx)

	// ==================== 远航系统 ====================
	case "voyageInfo":
		return c.VoyageInfoAction(ctx)
	case "beatVoyage":
		return c.BeatVoyageAction(ctx)
	case "voyageIngHeros":
		return c.VoyageIngHerosAction(ctx)
	case "lingquVoyage":
		return c.LingquVoyageAction(ctx)
	case "accelerateVoyage":
		return c.AccelerateVoyageAction(ctx)
	case "refreshVoyage":
		return c.RefreshVoyageAction(ctx)

	// ==================== 挂机系统 ====================
	case "onHook":
		return c.OnHookAction(ctx)
	case "hookLingqu":
		return c.HookLingquAction(ctx)
	case "fastHookCount":
		return c.FastHookCountAction(ctx)
	case "fastHookLingqu":
		return c.FastHookLingquAction(ctx)

	// ==================== 称号系统 ====================
	case "chooseRoleTitle":
		return c.ChooseRoleTitleAction(ctx)

	// ==================== PK 系统 ====================
	case "getpkOp":
		return c.GetpkOpAction(ctx)
	case "pkOp":
		return c.PkOpAction(ctx)
	case "pkRet":
		return c.PkRetAction(ctx)
	case "pkList":
		return c.PkListAction(ctx)
	case "pkRank":
		return c.PkRankAction(ctx)
	case "myPkRank":
		return c.MyPkRankAction(ctx)

	// ==================== 公会系统 ====================
	case "createGuild":
		return c.CreateGuildAction(ctx)
	case "guildInfo":
		return c.GuildInfoAction(ctx)
	case "guildList":
		return c.GuildListAction(ctx)
	case "applyGuild":
		return c.ApplyGuildAction(ctx)
	case "quitGuild":
		return c.QuitGuildAction(ctx)
	case "guildApplyList":
		return c.GuildApplyListAction(ctx)
	case "guildApplyHandle":
		return c.GuildApplyHandleAction(ctx)
	case "getGuildUsers":
		return c.GetGuildUsersAction(ctx)
	case "guildTanhe":
		return c.GuildTanheAction(ctx)
	case "guildOp":
		return c.GuildOpAction(ctx)
	case "guildCopyAddAtk":
		return c.GuildCopyAddAtkAction(ctx)
	case "getContribution":
		return c.GetContributionAction(ctx)
	case "contributeLingqu":
		return c.ContributeLingquAction(ctx)
	case "setGuildLimit", "setGuildDeclar":
		return c.GuildEditAction(ctx)
	case "getGuildSkill":
		return c.GetGuildSkillAction(ctx)
	case "activeGuildSkill":
		return c.ActiveGuildSkillAction(ctx)
	case "guildRankList":
		return c.GuildRankListAction(ctx)
	case "contribute":
		return c.ContributeAction(ctx)
	case "guildCopy":
		return c.GuildCopyAction(ctx)
	case "upGuildCopy":
		return c.UpGuildCopyAction(ctx)
	case "saodangGuildCopy":
		return c.SaodangGuildCopyAction(ctx)

	// ==================== 公会战 ====================
	case "getGuildFight":
		return c.GetGuildFightAction(ctx)
	case "fightGuild":
		return c.FightGuildAction(ctx)
	case "fightRank":
		return c.FightRankAction(ctx)
	case "guildCopyRank":
		return c.GuildCopyRankAction(ctx)
	case "guildPosFightLog":
		return c.GuildPosFightLogAction(ctx)
	case "guildFightLog":
		return c.GuildFightLogAction(ctx)
	case "guildDuizhanList":
		return c.GuildDuizhanListAction(ctx)
	case "guildLog":
		return c.GuildLogAction(ctx)

	// ==================== 邮件系统 ====================
	case "getUserMail":
		return c.GetUserMailAction(ctx)
	case "readMail":
		return c.ReadMailAction(ctx)
	case "delMail":
		return c.DelMailAction(ctx)
	case "lingquMailItems":
		return c.LingquMailItemsAction(ctx)

	// ==================== 好友系统 ====================
	case "friendsList":
		return c.FriendsListAction(ctx)
	case "recFriendList":
		return c.RecFriendListAction(ctx)
	case "applyFriendsList":
		return c.ApplyFriendsListAction(ctx)
	case "handleApply":
		return c.HandleApplyAction(ctx)
	case "sendLover":
		return c.SendLoverAction(ctx)
	case "loverReceiveList":
		return c.LoverReceiveListAction(ctx)
	case "lingquLover":
		return c.LingquLoverAction(ctx)
	case "getFriendByName":
		return c.GetFriendByNameAction(ctx)
	case "addFriend":
		return c.AddFriendAction(ctx)
	case "delFriend":
		return c.DelFriendAction(ctx)

	// ==================== VIP/充值系统 ====================
	case "vipBuyInfo":
		return c.VipBuyInfoAction(ctx)
	case "vipBuyzhizunLingqu":
		return c.VipBuyzhizunLingquAction(ctx)
	case "vipBuyLingqu":
		return c.VipBuyLingquAction(ctx)
	case "leijiChongzhi":
		return c.LeijiChongzhiAction(ctx)
	case "leijiChongzhiLingqu":
		return c.LeijiChongzhiLingquAction(ctx)
	case "jiTian":
		return c.JiTianAction(ctx)
	case "jiTianLingqu":
		return c.JiTianLingquAction(ctx)
	case "Yueka":
		return c.YuekaAction(ctx)
	case "Yuekalingqu":
		return c.YuekalingquAction(ctx)
	case "chongzhiList":
		return c.ChongzhiListAction(ctx)
	case "shouchong":
		return c.ShouchongAction(ctx)
	case "shouchonglingqu":
		return c.ShouchonglingquAction(ctx)
	case "dayShouchong":
		return c.DayShouchongAction(ctx)
	case "dayShouchonglingqu":
		return c.DayShouchonglingquAction(ctx)
	case "jijin":
		return c.JijinAction(ctx)
	case "jijinLingqu":
		return c.JijinLingquAction(ctx)
	case "yueduLibao":
		return c.YueduLibaoAction(ctx)
	case "dayLibao":
		return c.DayLibaoAction(ctx)
	case "weekLibao":
		return c.WeekLibaoAction(ctx)
	case "qianggou":
		return c.QianggouAction(ctx)
	case "yiyuanlibao":
		return c.YiyuanlibaoAction(ctx)
	case "anychonglibao":
		return c.AnychonglibaoAction(ctx)
	case "anychonglingqu":
		return c.AnychonglingquAction(ctx)
	case "tequanShop":
		return c.TequanShopAction(ctx)
	case "tequanBuy":
		return c.TequanBuyAction(ctx)
	case "monJijin":
		return c.MonJijinAction(ctx)
	case "monJijinLing":
		return c.MonJijinLingAction(ctx)

	// ==================== 分享/邀请/视频 ====================
	case "shareinfo":
		return c.ShareinfoAction(ctx)
	case "sharelingqu":
		return c.SharelingquAction(ctx)
	case "inviteinfo":
		return c.InviteinfoAction(ctx)
	case "invite":
		return c.InviteAction(ctx)
	case "invitelingqu":
		return c.InvitelingquAction(ctx)
	case "vedioinfo":
		return c.VedioinfoAction(ctx)
	case "vedio":
		return c.VedioAction(ctx)
	case "vediolingqu":
		return c.VediolingquAction(ctx)

	// ==================== 礼包码/十连抽/引导 ====================
	case "useLibaoCode":
		return c.UseLibaoCodeAction(ctx)
	case "lingqu10chou":
		return c.Lingqu10chouAction(ctx)
	case "begin10chou":
		return c.Begin10chouAction(ctx)
	case "getCurrentGuide":
		return c.GetCurrentGuideAction(ctx)
	case "guideTaskLingqu":
		return c.GuideTaskLingquAction(ctx)

	default:
		return c.ResponseError(3, fmt.Sprintf("shinelight/%s not implemented yet", action))
	}
}
