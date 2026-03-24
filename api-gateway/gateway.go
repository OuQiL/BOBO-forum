package main

import (
	"flag"
	"fmt"

	"api-gateway/internal/config"
	"api-gateway/internal/handler"
	"api-gateway/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandler(server, ctx)

	fmt.Printf("Starting api-gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
