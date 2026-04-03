package config

import "github.com/zeromicro/go-zero/zrpc"

type Config struct {
	zrpc.RpcServerConf

	Elasticsearch ElasticsearchConf
	MySQL         MySQLConf
}

type ElasticsearchConf struct {
	Addresses []string
	Username  string
	Password  string
	Timeout   int
}

type MySQLConf struct {
	Dsn          string
	MaxIdleConns int
	MaxOpenConns int
}
