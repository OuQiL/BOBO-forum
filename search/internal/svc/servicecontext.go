package svc

import (
	"context"
	"fmt"
	"search/internal/config"
	"search/internal/es"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config config.Config

	ES      *es.Client
	MySQL   *gorm.DB
	Context context.Context
}

func NewServiceContext(c config.Config) (*ServiceContext, error) {
	ctx := context.Background()

	// 初始化 ES 客户端
	esClient, err := es.NewClient(c.Elasticsearch)
	if err != nil {
		return nil, fmt.Errorf("failed to create es client: %w", err)
	}

	// 测试 ES 连接
	healthy, msg, err := esClient.HealthCheck(ctx)
	if err != nil {
		fmt.Printf("Warning: ES health check failed: %s\n", msg)
	} else {
		fmt.Printf("ES Health Check: %s\n", msg)
	}

	// 创建索引
	if err := esClient.CreateAllIndices(ctx); err != nil {
		fmt.Printf("Warning: failed to create indices: %v\n", err)
	}

	// 初始化 MySQL 连接
	db, err := gorm.Open(mysql.Open(c.MySQL.Dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mysql: %w", err)
	}

	// 设置连接池
	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get sql db: %w", err)
	}
	sqlDB.SetMaxIdleConns(c.MySQL.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MySQL.MaxOpenConns)

	fmt.Println("MySQL connected successfully")

	return &ServiceContext{
		Config:  c,
		ES:      esClient,
		MySQL:   db,
		Context: ctx,
	}, nil
}
