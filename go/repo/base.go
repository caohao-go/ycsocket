package repo

import (
	extorm "git.code.oa.com/pcg-csd/trpc-ext/orm"
	"server_golang/config"
)

// UserDB 返回用户信息库的 OrmClient (shine_user)
func UserDB() extorm.Client {
	return extorm.NewOrmClient(config.MysqlUser)
}

// InfoDB 返回静态配置库的 OrmClient (shine_info)
func InfoDB() extorm.Client {
	return extorm.NewOrmClient(config.MysqlInfo)
}

// WorldDB 返回区服动态数据库的 OrmClient (shine_world_zone1)
func WorldDB() extorm.Client {
	return extorm.NewOrmClient(config.MysqlWorld)
}

// NewStmt 创建新的 SQL 语句构建器
func NewStmt() *extorm.Statement {
	return extorm.NewDbStatement()
}
