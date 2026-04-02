package svc

import (
	"interaction/internal/cache"
	"interaction/internal/config"
	"interaction/internal/kafka"
	"interaction/internal/repository"
	"interaction/internal/sync"
	"time"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config             config.Config
	DB                 sqlx.SqlConn
	Redis              *redis.Redis
	LikeRepo           repository.LikeRepository
	LikeCache          *cache.LikeCache
	KafkaProducer      *kafka.Producer
	KafkaConsumer      *kafka.Consumer
	LikeSyncer         *sync.KafkaLikeCountSyncer
	LikeChecker        *sync.LikeConsistencyChecker
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewMysql(c.MySQL.DataSource)
	redisClient := redis.MustNewRedis(c.RedisConf)
	
	likeCache := cache.NewLikeCache(redisClient)
	
	producer := kafka.NewProducer(kafka.ProducerConfig{
		Brokers:      c.Kafka.Brokers,
		Topic:        c.Kafka.Topic,
		BatchSize:    c.Kafka.BatchSize,
		BatchTimeout: time.Duration(c.Kafka.BatchTimeout) * time.Second,
	})
	
	consumer := kafka.NewConsumer(kafka.ConsumerConfig{
		Brokers:   c.Kafka.Brokers,
		Topic:     c.Kafka.Topic,
		GroupId:   c.Kafka.GroupId,
		BatchSize: c.Kafka.BatchSize,
		MaxWait:   time.Duration(c.Kafka.BatchTimeout) * time.Second,
	})
	
	likeSyncer := sync.NewKafkaLikeCountSyncer(redisClient, db, likeCache, c.Kafka.BatchSize)
	
	likeChecker := sync.NewLikeConsistencyChecker(
		redisClient,
		db,
		likeCache,
		sync.ConsistencyCheckerConfig{
			CheckInterval:  time.Duration(c.ConsistencyCheck.IntervalMinutes) * time.Minute,
			BatchSize:      c.ConsistencyCheck.BatchSize,
			AutoFix:        c.ConsistencyCheck.AutoFix,
			EnableAlarm:    c.ConsistencyCheck.EnableAlarm,
			AlarmThreshold: c.ConsistencyCheck.AlarmThreshold,
		},
	)
	
	return &ServiceContext{
		Config:        c,
		DB:            db,
		Redis:         redisClient,
		LikeRepo:      repository.NewLikeRepository(db),
		LikeCache:     likeCache,
		KafkaProducer: producer,
		KafkaConsumer: consumer,
		LikeSyncer:    likeSyncer,
		LikeChecker:   likeChecker,
	}
}