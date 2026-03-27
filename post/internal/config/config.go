package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MySQL MySQLConfig
	Redis redis.RedisConf
	Cache LocalCacheConf
}

type MySQLConfig struct {
	DataSource      string
	MaxOpenConns    int `json:",default=100"`
	MaxIdleConns    int `json:",default=10"`
	ConnMaxLifetime int `json:",default=3600"`
	ConnMaxIdleTime int `json:",default=600"`
}

type LocalCacheConf struct {
	HotPostCacheTTL     int `json:",default=300"`
	HotPostMaxEntries   int `json:",default=1000"`
	HotPostThreshold    int `json:",default=1000"`
}
