package main

import (
	"context"
	"flag"
	"fmt"
	"time"

	"interaction/api/proto"
	"interaction/internal/config"
	"interaction/internal/server"
	"interaction/internal/svc"
	snowid "interaction/pkg/snowid"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/interaction.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	if err := snowid.Init(3); err != nil {
		panic(fmt.Sprintf("snowflake init failed: %v", err))
	}

	ctx := svc.NewServiceContext(c)

	ctx.LikeSyncer.Start()
	defer ctx.LikeSyncer.Stop()

	ctx.KafkaConsumer.StartBatchConsumer(
		context.Background(),
		c.Kafka.BatchSize,
		time.Duration(c.Kafka.BatchTimeout)*time.Second,
		ctx.LikeSyncer.ProcessBatch,
	)

	ctx.LikeChecker.Start()
	defer ctx.LikeChecker.Stop()

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		proto.RegisterInteractionServer(grpcServer, server.NewInteractionServer(ctx))
		reflection.Register(grpcServer)
	})
	defer s.Stop()

	fmt.Printf("Starting interaction-rpc server at %s...\n", c.ListenOn)
	s.Start()
}
