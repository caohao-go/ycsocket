package config

const EmptyString = "999999999"

const (
	// PK 初始分数
	PkInitScore = 1000

	// PK 排名键前缀
	PrePkRankKey = "pre_rank_pk"

	// 排行榜类型
	RankEndlessLayer     = "endless_layer"
	RankCopy             = "copy"
	RankClimbtower       = "climbtower_layer"
	RankFightPoint       = "fight_point"
	RankExpeditionScore  = "expedition_score"
	RankQuickBattleScore = "quick_battle_score"
	RankGuildHarm        = "guild_harm"
	RankGuildFight       = "guild_fight"

	RankHeroFightPointProp = "rank_hero_fight_point_prop_%d"
	RankHeroFightPoint     = "rank_hero_fight_point"
)

const (
	UserId      = "userid" // 玩家id
	UserAchieve = "user_achieve"
)

// Cache key 前缀常量
const (
	CacheLockGuildCopy          = "lock_guild_copy_%d"
	CacheLockGuildFight         = "lock_guild_fight_%d"
	CacheAddProp                = "pre_add_prop_%d"
	CacheGuildInfo              = "pre_guild_info_%d"
	CacheOrder                  = "pre_order_%d"
	CacheRedisUserInfo          = "pre_redis_user_info_%d"
	CacheUserMail               = "pre_user_mail_%d"
	CacheUserTongguanReward     = "pre_user_tongguan_reward_%d_%d"
	CacheUserTutengPk           = "pre_user_tuteng_pk_%d"
	CacheUserClimbtowerRecord   = "pre_user_climbtower_record_%d"
	CacheUserEndlessHelpHero    = "pre_user_endless_help_%s"
	CacheUserExpeditionHelpHero = "pre_user_expedition_help_%s"
	CacheUserPosition           = "pre_user_position_%d_%d"
)

// Pika 用户固定属性
const (
	AttrNickname          = "nickname"    // 冒险者昵称
	AttrAvatarUrl         = "avatar_url"  // 冒险者头像
	AttrGender            = "gender"      // 冒险者性别
	AttrLv                = "lv"          // 冒险者等级
	AttrVipLevel          = "vip_level"   // vip等级
	AttrCopy              = "copy"        // 当前关卡
	AttrChapter           = "chapter"     // 当前章节
	AttrRoleTitle         = "role_title"  // 使用的称号
	AttrNewGift           = "new_gift"    // 新手礼包
	AttrDay7              = "day7"        // 新手登录7天奖励
	AttrIGift             = "i_gift"      // 初始化奖励
	AttrFightPoint        = "fight_point" // 战斗力
	AttrRegT              = "reg_t"       // 注册时间
	AttrOffTime           = "off_time"    // 上次离线时间
	AttrInitFlag          = "init_flag"   // 初始化标记
	AttrBegin10Chou       = "begin_10_chou"
	AttrLingqu10Chou      = "lingqu_10_chou"
	AttrFinishFunctionNum = "finish_function_num"
	AttrFourStarNum       = "four_star_num"
	AttrFiveStarNum       = "five_star_num"
	AttrSixStarNum        = "six_star_num"
	AttrGuildBossNum      = "guild_boss_num"
	AttrKillGuildBossNum  = "kill_guild_boss_num"
	AttrMergeRuneNum      = "merge_rune_num"
	AttrGuideFinished     = "guide_finished"
	AttrSacrificeNum      = "sacrifice_num"
	AttrVoyageChengNum    = "voyage_cheng_num"
	AttrVoyageHongNum     = "voyage_hong_num"
	AttrXianZhiNum        = "xianzhi_num"
	AttrXingHeNum         = "xinghe_num"
	AttrTanBaoLucky       = "tanbao_lucky_"
)

// Pika 日常
const (
	DailyClimbtowerSaodang       = "climbtower_saodang"
	DailyRewardTimes             = "daily_reward_times"
	DailyTaskActiv               = "daily_task_activ"
	DailyTaskActivLingqu         = "daily_task_activ_lingqu_"
	DailyTaskCount               = "daily_task_count_"
	DailyTaskLingqu              = "daily_task_lingqu_"
	DailyEndlessToday            = "endless_today_"
	DailyEndlessFightHero        = "endless_fight_hero"
	DailyEndlessHelpChoose       = "endless_help_choose"
	DailyEndlessThistime         = "endless_thistime"
	DailyExpeditionBaoxiangOpen  = "expedition_baoxiang_open_"
	DailyExpeditionHelpChooseAll = "expedition_help_choose_all_"
	DailyExpeditionHerosHp       = "expedition_heros_hp_"
	DailyExpeditionOpHeros       = "expedition_op_heros"
	DailyExpeditionOpHerosHp     = "expedition_op_heros_hp_"
	DailyExpeditionToday         = "expedition_today_"
	DailyFastUsedCnt             = "fast_used_cnt"
	DailyFunctionTimes           = "function_times_"
	DailyActiveLingquStatus      = "active_lingqu_status_"
	DailyFreePkTimes             = "free_pk_times"
	DailyFastClimbCnt            = "fast_climb_cnt_"
	DailyGuildContribute         = "guild_contribute_"
	DailyGuildCopyCount          = "guild_copy_count"
	DailyMonCardAmt              = "mon_card_amt_"
	DailyMonCardLingqu           = "mon_card_lingqu_"
	DailyNew7dayGift             = "new_7day_gift"
	DailyPkUserIds               = "dalily_pk_user_ids"
	DailyTodayPoint              = "today_point"
	DailyRefreshTimes            = "refresh_times_"
	DailyShareTime               = "share_time"
	DailyUserTanbao              = "user_tanbao_"
	DailyVedioTime               = "vedio_time"
	DailyVipContentsDay          = "vip_contents_day"
	DailyVipZhizunLingqu         = "vip_zhizun_lingqu"
	DailyVoyageFreeCnt           = "voyage_free_cnt"
)

// Pika 日/周常
const (
	DWActiveGetNum  = "active_get_num_%d_%s"
	DWTaskFinishNum = "task_finish_num_%d_%s_%s"
	DWTaskLingqu    = "task_lingqu_%d_%s_%s"
	DWVipContentW   = "vip_content_w_%s_%d"
)

// Pika 帮会日常
const (
	KeyGuildContributeActive = "guild_contribute_active_%d_%s"
	KeyGuildTodayGongxian    = "guild_today_gongxian_%d_%s"
)

// Pika key 前缀
const (
	KeyDailyReward              = "daily_reward_%s_%d"
	KeyClimbLingquReward        = "climb_lingqu_reward_%d"
	KeyFirstQuit                = "first_quit_%d"
	KeyGuildApplyUid            = "guild_apply_uid_%d"
	KeyGuildCopyAddAtk          = "guild_copy_add_atk%d"
	KeyGuildFightInfo           = "guild_fight_info_%d_%d"
	KeyGuildFightLog            = "guild_fight_log_%d_%d"
	KeyMyGuildFightLog          = "my_guild_fight_log_%d_%d"
	KeyGuildGongxian            = "guild_gongxian_%d"
	KeyGuildOwnerLoginTime      = "guild_owner_login_time_%d"
	KeyGuildStars               = "guild_stars_%d_%d"
	KeyPosGuildLog              = "pos_guild_log_%d_%d_%d"
	KeyPreGuildFightAssigned    = "pre_guild_fight_assigned_%d_%s"
	KeyPreGuildFightChangci     = "pre_guild_fight_changci_%d"
	KeyPreGuildFightChangciIncr = "pre_guild_fight_changci_incr_%d_%s"
	KeyPreGuildFightReset       = "pre_guild_fight_reset_%d_%s"
	KeyPreGuildFightStatus      = "pre_guild_fight_status_%d"
	KeyHeroExchange             = "hero_exchange_%d_%s"
	KeyLastCopyHarmBlood        = "last_copy_harm_blood_%d_%d_%s"
	KeyLovers                   = "lovers_%s_%d"
	KeyMonCardExpire            = "mon_card_expire_%d_%d"
	KeyNewGiftTime              = "new_gift_time_%d"
	KeyNewOnHook                = "new_on_hook_%d"
	KeyOnHookTimeH              = "on_hook_time_h_%d"
	KeyOnHookTimeM              = "on_hook_time_m_%d"
	KeyPreDalilyCollectionItem  = "pre_dalily_collection_item_%d_%d_%s"
	KeyPreH5payBack             = "pre_h5pay_back_%d"
	KeyPreMppayBack             = "pre_mppay_back_%d"
	KeyPreShinePay              = "pre_shine_pay_%d"
	KeyPreRandCollectionItem    = "pre_rand_collection_item_%d_%d"
	KeyPreRedisUserPower        = "pre_redis_user_power_%d_%d"
	KeyPreUserGuide             = "pre_user_guide_%d"
	KeyQuitWaitTime             = "quit_wait_time_%d"
	KeyReceiveLovers            = "receive_lovers_%s_%d"
	KeyDianCoin                 = "dian_coin_%d"
	KeyRewardStatus             = "reward_status_%d_%d"
	KeyRewardTimes              = "reward_times_%d_%d"
	KeyRuneExchange             = "rune_exchange_%d_%v"
	KeyShareTimeRefresh         = "share_time_refresh%d%s"
	KeyShinelightTanbaoHistory  = "shinelight_tanbao_history_%d"
	KeyShopBuy                  = "shop_buy_%d_%d%s"
	KeyShopTotalBuy             = "shop_total_buy_%d%s"
	KeyTanbaoLuckLingqu         = "tanbao_luck_lingqu_%d_%d"
	KeyVedioStatus              = "vedio_status_%d%s"
	KeyVipContentM              = "vip_content_m_%d"
	KeyVipContentsLimitM        = "vip_contents_limit_m_%d_%d"
	KeyVipContentsLimitW        = "vip_contents_limit_w_%d_%d"
	KeyVoyageHero               = "voyage_hero_%d"
	KeyVoyageList               = "voyage_list%d"
	KeyFightHero                = "fight_hero_%s_%d"
	KeyTemplateLv               = "temp_lv"
	KeyTemplateInfo             = "temp_info"
	KeyTemplateInfoTimeout      = "temp_info_timeout_%d"
)
