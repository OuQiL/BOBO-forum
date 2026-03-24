package config

import (
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	JwtAuth AuthConfig
	MySQL   MySQLConfig
}

type AuthConfig struct {
	AccessSecret string
	AccessExpire int64
}

type MySQLConfig struct {
	DataSource string
}
