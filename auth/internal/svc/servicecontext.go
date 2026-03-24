package svc

import (
	"auth/internal/config"
	"auth/internal/repository"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type ServiceContext struct {
	Config  config.Config
	DB      sqlx.SqlConn
	UserRepo repository.UserRepository
}

func NewServiceContext(c config.Config) *ServiceContext {
	db := sqlx.NewMysql(c.MySQL.DataSource)
	return &ServiceContext{
		Config:   c,
		DB:       db,
		UserRepo: repository.NewUserRepository(db),
	}
}
