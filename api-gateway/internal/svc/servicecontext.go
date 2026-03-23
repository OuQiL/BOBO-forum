package svc

import (
	"api-gateway/internal/config"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	AuthRpc        zrpc.Client
	PostRpc        zrpc.Client
	SearchRpc      zrpc.Client
	InteractionRpc zrpc.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		AuthRpc:        zrpc.MustNewClient(c.Auth),
		PostRpc:        zrpc.MustNewClient(c.Post),
		SearchRpc:      zrpc.MustNewClient(c.Search),
		InteractionRpc: zrpc.MustNewClient(c.Interaction),
	}
}
