package svc

import (
	"context"
	"database/sql"
	"time"

	"post/internal/config"
	"post/internal/repository"

	"github.com/go-sql-driver/mysql"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config        config.Config
	DB            sqlx.SqlConn
	Redis         *redis.Redis
	PostRepo      repository.PostRepository
	CommentRepo   repository.CommentRepository
	HotPostSvc    *repository.HotPostService
	BloomFilter   *repository.PostBloomFilter
	ViewCountSync *repository.ViewCountSyncer
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
	redisClient := redis.MustNewRedis(c.RedisConf)

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

	viewCountSyncer := repository.NewViewCountSyncer(redisClient, sqlConn, hotPostCache)

	ctx := &ServiceContext{
		Config:        c,
		DB:            sqlConn,
		Redis:         redisClient,
		PostRepo:      postRepo,
		CommentRepo:   repository.NewCommentRepository(sqlConn),
		HotPostSvc:    hotPostSvc,
		BloomFilter:   bloomFilter,
		ViewCountSync: viewCountSyncer,
	}

	go ctx.warmupCache(context.Background())

	viewCountSyncer.Start()

	return ctx
}

func (s *ServiceContext) WarmupCache(ctx context.Context) error {
	logx.Info("Starting cache warmup...")

	posts, err := s.PostRepo.GetRecentPosts(ctx, 7)
	if err != nil {
		logx.Errorf("Failed to get recent posts for warmup: %v", err)
		return err
	}

	warmupCount := 0
	for _, post := range posts {
		if s.HotPostSvc.ShouldCachePost(post.ViewCount) {
			s.HotPostSvc.CachePost(post)
			warmupCount++
		}
	}

	logx.Infof("Cache warmup completed: %d posts loaded", warmupCount)
	return nil
}

func (s *ServiceContext) warmupCache(ctx context.Context) {
	time.Sleep(2 * time.Second)
	s.WarmupCache(ctx)
}

func (s *ServiceContext) Stop() {
	if s.ViewCountSync != nil {
		s.ViewCountSync.Stop()
	}
}
