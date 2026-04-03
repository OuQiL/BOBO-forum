package main

import (
	"flag"
	"fmt"

	"search/internal/config"
	"search/internal/server"
	"search/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"search/api/proto"
)

var configFile = flag.String("f", "etc/search.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	// 创建服务上下文
	ctx, err := svc.NewServiceContext(c)
	if err != nil {
		logx.Errorf("Failed to create service context: %v", err)
		return
	}

	// 创建 RPC 服务器
	server := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proto.RegisterSearchServer(grpcServer, server.NewSearchServer(ctx))
	})

	defer server.Stop()

	fmt.Printf("Starting search-rpc at %s...\n", c.ListenOn)
	server.Start()
}
