package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf

	Auth        zrpc.RpcClientConf
	Post        zrpc.RpcClientConf
	Search      zrpc.RpcClientConf
	Interaction zrpc.RpcClientConf

	JWT JWTConfig
}

type JWTConfig struct {
	Secret      string
	Expire      int64
	PrevSecret  string
	PrevExpire  int64
}
