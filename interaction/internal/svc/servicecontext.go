package svc

import (
	"interaction/internal/config"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

type ServiceContext struct {
	Config config.Config
	DB     sqlx.SqlConn
	Redis  *redis.Redis
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewMysql(c.MySQL.DataSource)
	redisClient := redis.MustNewRedis(c.RedisConf)
	return &ServiceContext{
		Config: c,
		DB:     db,
		Redis:  redisClient,
	}
}