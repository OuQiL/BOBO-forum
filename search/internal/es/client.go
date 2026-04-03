package es

import (
	"context"
	"fmt"
	"log"
	"time"

	"search/internal/config"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// Client Elasticsearch 客户端封装
type Client struct {
	ES *elasticsearch.Client
}

// NewClient 创建 ES 客户端
func NewClient(conf config.ElasticsearchConf) (*Client, error) {
	cfg := elasticsearch.Config{
		Addresses: conf.Addresses,
		Username:  conf.Username,
		Password:  conf.Password,
		Timeout:   time.Duration(conf.Timeout) * time.Second,
	}

	es, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	return &Client{ES: es}, nil
}

// HealthCheck 检查 ES 连接状态
func (c *Client) HealthCheck(ctx context.Context) (bool, string, error) {
	// 使用 Info API 检查连接
	res, err := c.ES.Info(c.ES.Info.WithContext(ctx))
	if err != nil {
		return false, fmt.Sprintf("Elasticsearch connection failed: %v", err), err
	}
	defer res.Body.Close()

	if res.IsError() {
		return false, fmt.Sprintf("Elasticsearch error: %s", res.String()), fmt.Errorf("es error: %s", res.String())
	}

	// 获取集群健康状态
	healthRes, err := c.ES.Cluster.Health(c.ES.Cluster.Health.WithContext(ctx))
	if err != nil {
		return false, fmt.Sprintf("Failed to get cluster health: %v", err), err
	}
	defer healthRes.Body.Close()

	if healthRes.IsError() {
		return false, fmt.Sprintf("Cluster health error: %s", healthRes.String()), fmt.Errorf("cluster health error")
	}

	return true, "Elasticsearch connected successfully", nil
}

// Ping 简单的 Ping 测试
func (c *Client) Ping(ctx context.Context) error {
	res, err := c.ES.Ping(c.ES.Ping.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("ping error: %s", res.String())
	}

	return nil
}

// GetClient 返回原始 ES 客户端
func (c *Client) GetClient() *elasticsearch.Client {
	return c.ES
}

// Perform 执行 ES 请求
func (c *Client) Perform(req *esapi.Request) (*esapi.Response, error) {
	return req.Do(ctx.Background(), c.ES)
}
