package main

import (
	"flag"
	"fmt"

	"post/api/proto"
	"post/internal/config"
	"post/internal/server"
	"post/internal/svc"
	snowid "post/pkg/snowid"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/post.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if err := snowid.Init(1); err != nil {
		panic(fmt.Sprintf("failed to initialize snowflake: %v", err))
	}

	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proto.RegisterPostServer(grpcServer, server.NewPostServer(ctx))
		reflection.Register(grpcServer)
	})
	defer s.Stop()

	fmt.Printf("Starting post-rpc server at %s...\n", c.ListenOn)
	s.Start()
}
