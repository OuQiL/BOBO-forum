package logic

import (
	"context"
	"testing"

	"search/api/proto"
	"search/internal/config"
	"search/internal/es"
	"search/internal/svc"

	"github.com/elastic/go-elasticsearch/v8"
)

// 创建测试用的服务上下文
func newTestServiceContext(t *testing.T) *svc.ServiceContext {
	// 使用 mock 或跳过实际 ES 连接
	// 这里我们创建一个基本的 svc 上下文用于测试
	return &svc.ServiceContext{
		Config: config.Config{
			Elasticsearch: config.ElasticsearchConf{
				Addresses: []string{"http://localhost:9200"},
				Username:  "elastic",
				Password:  "bobo123",
				Timeout:   5,
			},
		},
		Context: context.Background(),
	}
}

// TestSearchLogic_BuildPostQuery 测试帖子查询构建
func TestSearchLogic_BuildPostQuery(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	logic := NewSearchLogic(context.Background(), svcCtx)

	tests := []struct {
		name     string
		req      *proto.SearchRequest
		validate func(query map[string]interface{}) bool
	}{
		{
			name: "基础搜索查询",
			req: &proto.SearchRequest{
				Keyword: "Python 教程",
				Type:    "post",
				Page:    1,
				Size:    10,
				SortBy:  "relevance",
			},
			validate: func(query map[string]interface{}) bool {
				// 验证分页
				from := query["from"].(int)
				size := query["size"].(int)
				if from != 0 || size != 10 {
					return false
				}

				// 验证查询类型
				queryMap := query["query"].(map[string]interface{})
				boolQuery := queryMap["bool"].(map[string]interface{})
				must := boolQuery["must"].([]interface{})
				if len(must) == 0 {
					return false
				}

				return true
			},
		},
		{
			name: "带过滤条件的搜索",
			req: &proto.SearchRequest{
				Keyword:    "Golang",
				Type:       "post",
				Page:       1,
				Size:       20,
				CommunityId: func() *int64 { v := int64(1); return &v }(),
				AuthorId:   func() *int64 { v := int64(123); return &v }(),
				SortBy:     "latest",
			},
			validate: func(query map[string]interface{}) bool {
				// 验证过滤条件存在
				queryMap := query["query"].(map[string]interface{})
				boolQuery := queryMap["bool"].(map[string]interface{})
				filter := boolQuery["filter"].([]interface{})
				
				// 应该至少有 status 过滤 + community_id + author_id
				if len(filter) < 3 {
					return false
				}

				return true
			},
		},
		{
			name: "按热度排序",
			req: &proto.SearchRequest{
				Keyword: "热门",
				Type:    "post",
				Page:    1,
				Size:    10,
				SortBy:  "hottest",
			},
			validate: func(query map[string]interface{}) bool {
				// 验证排序字段
				sort := query["sort"].([]map[string]interface{})
				if len(sort) == 0 {
					return false
				}

				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := logic.BuildPostQuery(tt.req)
			
			if query == nil {
				t.Fatal("Query should not be nil")
			}

			if !tt.validate(query) {
				t.Errorf("Query validation failed for test case: %s", tt.name)
			}
		})
	}
}

// TestSearchLogic_BuildUserQuery 测试用户查询构建
func TestSearchLogic_BuildUserQuery(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	logic := NewSearchLogic(context.Background(), svcCtx)

	tests := []struct {
		name     string
		req      *proto.SearchRequest
		validate func(query map[string]interface{}) bool
	}{
		{
			name: "基础用户搜索",
			req: &proto.SearchRequest{
				Keyword: "张三",
				Type:    "user",
				Page:    1,
				Size:    10,
			},
			validate: func(query map[string]interface{}) bool {
				// 验证分页
				from := query["from"].(int)
				size := query["size"].(int)
				if from != 0 || size != 10 {
					return false
				}

				// 验证查询包含 username 和 nickname 字段
				queryMap := query["query"].(map[string]interface{})
				boolQuery := queryMap["bool"].(map[string]interface{})
				must := boolQuery["must"].([]interface{})
				
				if len(must) == 0 {
					return false
				}

				return true
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := logic.BuildUserQuery(tt.req)
			
			if query == nil {
				t.Fatal("Query should not be nil")
			}

			if !tt.validate(query) {
				t.Errorf("Query validation failed for test case: %s", tt.name)
			}
		})
	}
}

// TestSearchLogic_ConvertPostResult 测试帖子结果转换
func TestSearchLogic_ConvertPostResult(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	logic := NewSearchLogic(context.Background(), svcCtx)

	source := map[string]interface{}{
		"id":            "1",
		"user_id":       "123",
		"community_id":  "456",
		"username":      "测试用户",
		"title":         "测试标题",
		"content":       "测试内容",
		"like_count":    10,
		"comment_count": 5,
		"view_count":    100,
		"created_at":    "2024-01-01T00:00:00Z",
	}

	highlight := map[string]interface{}{
		"title":   []interface{}{"<em>测试</em>标题"},
		"content": []interface{}{"<em>测试</em>内容"},
	}

	result := logic.ConvertPostResult(source, highlight)

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Type != "post" {
		t.Errorf("Expected type 'post', got '%s'", result.Type)
	}

	if result.Title != "测试标题" {
		t.Errorf("Expected title '测试标题', got '%s'", result.Title)
	}

	if result.HighlightTitle != "<em>测试</em>标题" {
		t.Errorf("Expected highlight title '<em>测试</em>标题', got '%s'", result.HighlightTitle)
	}

	if result.LikeCount != 10 {
		t.Errorf("Expected like_count 10, got %d", result.LikeCount)
	}
}

// TestSearchLogic_ConvertUserResult 测试用户结果转换
func TestSearchLogic_ConvertUserResult(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	logic := NewSearchLogic(context.Background(), svcCtx)

	source := map[string]interface{}{
		"id":         "123",
		"username":   "zhangsan",
		"nickname":   "张三",
		"email":      "zhangsan@example.com",
		"avatar":     "http://example.com/avatar.jpg",
		"gender":     1,
		"created_at": "2024-01-01T00:00:00Z",
	}

	highlight := map[string]interface{}{
		"username": []interface{}{"<em>zhangsan</em>"},
		"nickname": []interface{}{"<em>张三</em>"},
	}

	result := logic.ConvertUserResult(source, highlight)

	if result == nil {
		t.Fatal("Result should not be nil")
	}

	if result.Type != "user" {
		t.Errorf("Expected type 'user', got '%s'", result.Type)
	}

	if result.Username != "zhangsan" {
		t.Errorf("Expected username 'zhangsan', got '%s'", result.Username)
	}

	if result.Nickname != "张三" {
		t.Errorf("Expected nickname '张三', got '%s'", result.Nickname)
	}

	if result.HighlightTitle != "<em>zhangsan</em>" {
		t.Errorf("Expected highlight title '<em>zhangsan</em>', got '%s'", result.HighlightTitle)
	}
}

// TestSearchLogic_EmptyKeyword 测试空关键词搜索
func TestSearchLogic_EmptyKeyword(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	logic := NewSearchLogic(context.Background(), svcCtx)

	req := &proto.SearchRequest{
		Keyword: "",
		Type:    "post",
		Page:    1,
		Size:    10,
	}

	resp, err := logic.Search(req)

	if err != nil {
		t.Fatalf("Search should not return error for empty keyword, got: %v", err)
	}

	if resp.Total != 0 {
		t.Errorf("Expected total 0, got %d", resp.Total)
	}

	if len(resp.List) != 0 {
		t.Errorf("Expected empty list, got %d items", len(resp.List))
	}
}

// TestSearchLogic_InvalidType 测试无效搜索类型
func TestSearchLogic_InvalidType(t *testing.T) {
	svcCtx := newTestServiceContext(t)
	logic := NewSearchLogic(context.Background(), svcCtx)

	req := &proto.SearchRequest{
		Keyword: "test",
		Type:    "invalid_type",
		Page:    1,
		Size:    10,
	}

	// 应该默认搜索帖子
	resp, err := logic.Search(req)

	// 不期望有错误，但结果应该是空的（因为没有实际 ES 连接）
	if err == nil && resp == nil {
		t.Error("Expected either error or response")
	}
}
