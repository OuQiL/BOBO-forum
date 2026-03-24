package svc

import (
	"api-gateway/internal/config"
	"log"

	authclient "auth/pkg/client/auth"

	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config config.Config

	AuthRpc        authclient.Auth
	PostRpc        zrpc.Client
	SearchRpc      zrpc.Client
	InteractionRpc zrpc.Client
}

func NewServiceContext(c config.Config) *ServiceContext {
	sc := &ServiceContext{
		Config: c,
	}

	sc.AuthRpc = newAuthClient(c.Auth)
	sc.PostRpc = newRpcClient(c.Post, "post-rpc")
	sc.SearchRpc = newRpcClient(c.Search, "search-rpc")
	sc.InteractionRpc = newRpcClient(c.Interaction, "interaction-rpc")

	return sc
}

func newAuthClient(conf zrpc.RpcClientConf) authclient.Auth {
	client, err := zrpc.NewClient(conf)
	if err != nil {
		log.Printf("Warning: failed to connect to auth-rpc: %v (will retry on demand)", err)
		return nil
	}
	return authclient.NewAuth(client)
}

func newRpcClient(conf zrpc.RpcClientConf, name string) zrpc.Client {
	client, err := zrpc.NewClient(conf)
	if err != nil {
		log.Printf("Warning: failed to connect to %s: %v (will retry on demand)", name, err)
	}
	return client
}
