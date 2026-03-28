package svc

import (
	"database/sql"
	"time"

	"post/internal/config"
	"post/internal/repository"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config      config.Config
	DB          sqlx.SqlConn
	Redis       *redis.Redis
	PostRepo    repository.PostRepository
	CommentRepo repository.CommentRepository
	HotPostSvc  *repository.HotPostService
	BloomFilter *repository.PostBloomFilter
}

func NewServiceContext(c config.Config) *ServiceContext {
	mysqlCfg, err := mysql.ParseDSN(c.MySQL.DataSource)
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("mysql", mysqlCfg.FormatDSN())
	if err != nil {
		panic(err)
	}

	db.SetMaxOpenConns(c.MySQL.MaxOpenConns)
	db.SetMaxIdleConns(c.MySQL.MaxIdleConns)
	db.SetConnMaxLifetime(time.Duration(c.MySQL.ConnMaxLifetime) * time.Second)
	db.SetConnMaxIdleTime(time.Duration(c.MySQL.ConnMaxIdleTime) * time.Second)

	sqlConn := sqlx.NewSqlConnFromDB(db)
	redisClient := redis.MustNewRedis(c.Redis)

	hotPostCache := repository.NewHotPostLocalCache(
		c.Cache.HotPostMaxEntries,
		time.Duration(c.Cache.HotPostCacheTTL)*time.Second,
		c.Cache.HotPostThreshold,
	)

	hotPostSvc := repository.NewHotPostService(redisClient, hotPostCache)

	bloomFilter := repository.NewPostBloomFilter(
		c.Bloom.ExpectedItems,
		c.Bloom.FalsePositiveRate,
		redisClient,
	)

	postRepo := repository.NewPostRepository(sqlConn, redisClient, hotPostCache, hotPostSvc)

	return &ServiceContext{
		Config:      c,
		DB:          sqlConn,
		Redis:       redisClient,
		PostRepo:    postRepo,
		CommentRepo: repository.NewCommentRepository(sqlConn),
		HotPostSvc:  hotPostSvc,
		BloomFilter: bloomFilter,
	}
}
