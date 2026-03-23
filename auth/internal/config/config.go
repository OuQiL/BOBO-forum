package config

import (
	"github.com/zeromicro/go-zero/core/stores/sqlx"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	Auth  AuthConfig
	MySQL sqlx.SqlConf
}

type AuthConfig struct {
	AccessSecret string
	AccessExpire int64
}
