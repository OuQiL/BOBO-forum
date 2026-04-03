package es

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"search/internal/config"
)

// TestHealthCheck 测试 ES 连接健康检查
func TestHealthCheck(t *testing.T) {
	// 从环境变量读取配置或默认值
	addresses := []string{"http://localhost:9200"}
	if esHost := os.Getenv("ES_HOST"); esHost != "" {
		addresses = []string{fmt.Sprintf("http://%s:9200", esHost)}
	}

	username := "elastic"
	if u := os.Getenv("ES_USERNAME"); u != "" {
		username = u
	}

	password := "bobo123"
	if p := os.Getenv("ES_PASSWORD"); p != "" {
		password = p
	}

	conf := config.ElasticsearchConf{
		Addresses: addresses,
		Username:  username,
		Password:  password,
		Timeout:   5,
	}

	// 创建客户端
	client, err := NewClient(conf)
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	// 创建测试上下文
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 执行健康检查
	healthy, message, err := client.HealthCheck(ctx)
	if err != nil {
		t.Errorf("Health check failed: %v", err)
	}

	fmt.Printf("Health Check Result:\n")
	fmt.Printf("  Healthy: %v\n", healthy)
	fmt.Printf("  Message: %s\n", message)

	if !healthy {
		t.Fatal("Elasticsearch is not healthy")
	}

	// 测试 Ping
	err = client.Ping(ctx)
	if err != nil {
		t.Errorf("Ping failed: %v", err)
	} else {
		fmt.Println("  Ping: Success")
	}
}

// TestCreateIndices 测试创建索引
func TestCreateIndices(t *testing.T) {
	addresses := []string{"http://localhost:9200"}
	if esHost := os.Getenv("ES_HOST"); esHost != "" {
		addresses = []string{fmt.Sprintf("http://%s:9200", esHost)}
	}

	conf := config.ElasticsearchConf{
		Addresses: addresses,
		Username:  "elastic",
		Password:  "bobo123",
		Timeout:   5,
	}

	client, err := NewClient(conf)
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 创建所有索引
	err = client.CreateAllIndices(ctx)
	if err != nil {
		t.Errorf("Failed to create indices: %v", err)
	} else {
		fmt.Println("All indices created successfully")
	}

	// 验证索引是否存在
	for indexName := range IndexMapping {
		exists, err := client.IndexExists(ctx, indexName)
		if err != nil {
			t.Errorf("Failed to check index %s: %v", indexName, err)
		} else if !exists {
			t.Errorf("Index %s does not exist", indexName)
		} else {
			fmt.Printf("  Index %s: Exists ✓\n", indexName)
		}
	}
}

// TestConnectionOnly 简单连接测试（不依赖认证）
func TestConnectionOnly(t *testing.T) {
	// 最简单的测试：只检查 ES 是否可达
	addresses := []string{"http://localhost:9200"}
	if esHost := os.Getenv("ES_HOST"); esHost != "" {
		addresses = []string{fmt.Sprintf("http://%s:9200", esHost)}
	}

	conf := config.ElasticsearchConf{
		Addresses: addresses,
		Timeout:   5,
	}

	client, err := NewClient(conf)
	if err != nil {
		t.Fatalf("Failed to create ES client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 使用 Info API 检查连接
	res, err := client.ES.Info(client.ES.Info.WithContext(ctx))
	if err != nil {
		t.Fatalf("ES Info failed: %v", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		t.Fatalf("ES Info returned error: %s", res.String())
	}

	fmt.Println("✓ Elasticsearch connection successful!")
	fmt.Printf("  Cluster: %s\n", res.String())
}
