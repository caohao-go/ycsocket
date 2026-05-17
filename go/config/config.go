package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// MySQL 服务名常量（对应 trpc_go.yaml client.service.name）
const (
	MysqlUser  = "trpc.mysql.user"  // 用户信息库 shine_user（账号/序列/订单/区服信息）
	MysqlInfo  = "trpc.mysql.info"  // 静态配置库 shine_info（物品/英雄/技能/副本等配置表）
	MysqlWorld = "trpc.mysql.world" // 区服动态数据库 shine_world_1（玩家运行时数据表）
)

// Pika 服务名常量（对应 trpc_go.yaml client.service.name）
const (
	Pika = "trpc.redis.pika"
)

type Config struct {
	Zone      ZoneConfig             `yaml:"zone"`
	WxApp     map[string]WxAppConfig `yaml:"wx_app"`
	Constants ConstantsConfig        `yaml:"constants"`
	Filter    FilterConfig           `yaml:"filter"`
	GameConf  GameConfConfig         `yaml:"game_conf"`
}

type ZoneInfo struct {
	ZoneID   int    `yaml:"zone_id"`
	ZoneName string `yaml:"zone_name"`
	Hequ     int    `yaml:"hequ"`
	Hot      string `yaml:"hot"`
	Status   int    `yaml:"status"`
	Socket   string `yaml:"socket"`
	Port     int    `yaml:"port"`
	Time     string `yaml:"time"`
}

type ZoneConfig struct {
	ZoneInfo      []ZoneInfo  `yaml:"zone_info"`
	SourceZone    map[int]int `yaml:"source_zone"`
	RecommendZone []int       `yaml:"recommend_zone"`
}

type WxAppConfig struct {
	AppID  string `yaml:"app_id"`
	Secret string `yaml:"secret"`
}

type ConstantsConfig struct {
	WeixinOpenIDURL         string `yaml:"weixin_open_id_url"`
	WeixinSnsapiUserInfoURL string `yaml:"weixin_snsapi_user_info_url"`
	TokenGenerateKey        string `yaml:"token_generate_key"`
	RankPerPage             int    `yaml:"rank_per_page"`
}

var Cfg *Config

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	Cfg = cfg
	return cfg, nil
}

func GetDBConfigForZone(zoneID int) string {
	if zoneID <= 0 {
		return "shine_light_1"
	}
	return fmt.Sprintf("shine_light_%d", zoneID)
}

func GetRedisConfigForZone(zoneID int, name string) string {
	if zoneID <= 0 {
		zoneID = 1
	}
	return fmt.Sprintf("%s_%d", name, zoneID)
}

// FilterConfig 对应 PHP: Loader::config("filter")
// 结构: filter[appid][version] = FilterVersionConfig
// 以及 filter["filter_area"] = 默认排除地域列表
type FilterConfig struct {
	// AppVersions 按 appid -> version 索引的过滤配置
	AppVersions map[string]map[string]FilterVersionConfig `yaml:"app_versions"`
	// FilterArea 默认排除地域列表（无 appid/version 时的全局过滤）
	FilterArea []string `yaml:"filter_area"`
}

// FilterVersionConfig 某个 appid+version 下的过滤规则
type FilterVersionConfig struct {
	Fanxiang     bool        `yaml:"fanxiang"`      // 是否反向（全局反向标记）
	IncludeUsers []string    `yaml:"include_users"` // 反向包含用户列表
	FilterTimes  [][2]string `yaml:"filter_times"`  // 排除时间段 [[start, end], ...]
	FilterUsers  []string    `yaml:"filter_users"`  // 排除用户列表
	FilterArea   []string    `yaml:"filter_area"`   // 排除地域列表
}

// MsgSecCheck 微信内容安全检查配置（对应 PHP nicknameSameAction 中硬编码的 appid/secret）
const (
	MsgSecCheckAppID  = "wxfa8b612abfc14f0b"
	MsgSecCheckSecret = "ea9d962ed1c9fea860bf0cf431d353db"
)

// GameConfConfig 对应 PHP: Loader::config("gameconf")
// 结构: gameconf[appid][version] = 具体配置 map
type GameConfConfig map[string]map[string]map[string]any
