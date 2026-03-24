package svc

import (
	"post/internal/config"
	"post/internal/repository"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config      config.Config
	DB          sqlx.SqlConn
	Redis       *redis.Redis
	PostRepo    repository.PostRepository
	CommentRepo repository.CommentRepository
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewMysql(c.MySQL.DataSource)
	redisClient := redis.MustNewRedis(c.Redis)
	return &ServiceContext{
		Config:      c,
		DB:          db,
		Redis:       redisClient,
		PostRepo:    repository.NewPostRepository(db),
		CommentRepo: repository.NewCommentRepository(db),
	}
}
