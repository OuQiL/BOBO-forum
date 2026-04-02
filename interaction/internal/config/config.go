package config

import (
	"time"
	
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	MySQL              MySQLConfig
	RedisConf          redis.RedisConf
	Kafka              KafkaConfig
	LikeSync           LikeSyncConfig
	ConsistencyCheck   ConsistencyCheckConfig
}

type MySQLConfig struct {
	DataSource      string
	MaxOpenConns    int `json:",default=100"`
	MaxIdleConns    int `json:",default=10"`
	ConnMaxLifetime int `json:",default=3600"`
	ConnMaxIdleTime int `json:",default=600"`
}

type KafkaConfig struct {
	Brokers      []string
	Topic        string
	GroupId      string
	BatchSize    int `json:",default=100"`
	BatchTimeout int `json:",default=5"`
}

type LikeSyncConfig struct {
	IntervalSeconds int `json:",default=5"`
}

type ConsistencyCheckConfig struct {
	IntervalMinutes int  `json:",default=5"`
	BatchSize       int  `json:",default=100"`
	AutoFix         bool `json:",default=true"`
	EnableAlarm     bool `json:",default=true"`
	AlarmThreshold  int  `json:",default=10"`
}