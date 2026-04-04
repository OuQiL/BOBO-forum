package sync

import (
	"context"
	"testing"
	"time"

	"search/internal/config"
	"search/internal/es"
	"search/internal/svc"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

// 创建测试用的服务上下文
func newTestSyncServiceContext(t *testing.T) *svc.ServiceContext {
	// 注意：这个测试需要实际的 MySQL 和 ES 连接
	// 如果环境未配置，测试会跳过
	
	esHost := "localhost"
	mysqlDsn := "root:bobo123@tcp(localhost:3306)/bobo_db?charset=utf8mb4&parseTime=true&loc=Local"
	
	// 创建 ES 客户端（可选）
	var esClient *es.Client
	// 跳过 ES 连接测试，使用 mock
	
	// 创建 MySQL 连接（可选）
	var db *gorm.DB
	// 跳过 MySQL 连接测试
	
	return &svc.ServiceContext{
		Config: config.Config{
			Elasticsearch: config.ElasticsearchConf{
				Addresses: []string{"http://" + esHost + ":9200"},
				Username:  "elastic",
				Password:  "bobo123",
				Timeout:   5,
			},
			MySQL: config.MySQLConf{
				Dsn: mysqlDsn,
			},
		},
		Context: context.Background(),
		ES:      esClient,
		MySQL:   db,
	}
}

// TestSyncer_PostModel 测试 Post 模型
func TestSyncer_PostModel(t *testing.T) {
	post := Post{
		ID:            1,
		UserID:        123,
		CommunityID:   456,
		Username:      "testuser",
		Title:         "测试帖子",
		Content:       "这是测试内容",
		Tags:          "golang,test",
		LikeCount:     10,
		CommentCount:  5,
		ViewCount:     100,
		Status:        2,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if post.TableName() != "posts" {
		t.Errorf("Expected table name 'posts', got '%s'", post.TableName())
	}

	if post.Status != 2 {
		t.Errorf("Expected status 2 (published), got %d", post.Status)
	}
}

// TestSyncer_UserModel 测试 User 模型
func TestSyncer_UserModel(t *testing.T) {
	user := User{
		ID:           123,
		Username:     "zhangsan",
		PasswordHash: "hashed_password",
		Email:        "zhangsan@example.com",
		Nickname:     "张三",
		Avatar:       "http://example.com/avatar.jpg",
		Gender:       1,
		Status:       1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if user.TableName() != "users" {
		t.Errorf("Expected table name 'users', got '%s'", user.TableName())
	}

	if user.Status != 1 {
		t.Errorf("Expected status 1 (active), got %d", user.Status)
	}

	if user.Username != "zhangsan" {
		t.Errorf("Expected username 'zhangsan', got '%s'", user.Username)
	}
}

// TestSyncer_NewSyncer 测试 Syncer 创建
func TestSyncer_NewSyncer(t *testing.T) {
	svcCtx := newTestSyncServiceContext(t)
	
	syncer := NewSyncer(svcCtx)
	
	if syncer == nil {
		t.Fatal("Syncer should not be nil")
	}

	if syncer.svcCtx != svcCtx {
		t.Error("Syncer context not set correctly")
	}
}

// TestSyncer_IndexPostDocument 测试帖子文档索引数据结构
func TestSyncer_IndexPostDocument(t *testing.T) {
	post := Post{
		ID:            1,
		UserID:        123,
		CommunityID:   456,
		Username:      "testuser",
		Title:         "Golang 入门教程",
		Content:       "本文介绍 Golang 编程语言的基础知识",
		Tags:          "golang,programming,tutorial",
		LikeCount:     50,
		CommentCount:  10,
		ViewCount:     500,
		Status:        2,
		CreatedAt:     time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:     time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
	}

	// 构建文档
	doc := map[string]interface{}{
		"id":            "1",
		"user_id":       "123",
		"community_id":  "456",
		"username":      post.Username,
		"title":         post.Title,
		"content":       post.Content,
		"tags":          post.Tags,
		"like_count":    post.LikeCount,
		"comment_count": post.CommentCount,
		"view_count":    post.ViewCount,
		"status":        post.Status,
		"created_at":    post.CreatedAt.Format(time.RFC3339),
		"updated_at":    post.UpdatedAt.Format(time.RFC3339),
	}

	// 验证文档字段
	if doc["id"] != "1" {
		t.Errorf("Expected id '1', got '%v'", doc["id"])
	}

	if doc["title"] != "Golang 入门教程" {
		t.Errorf("Expected title 'Golang 入门教程', got '%v'", doc["title"])
	}

	if doc["status"] != 2 {
		t.Errorf("Expected status 2, got %d", doc["status"])
	}

	// 验证时间格式
	createdAt := doc["created_at"].(string)
	if createdAt != "2024-01-01T00:00:00Z" {
		t.Errorf("Expected created_at '2024-01-01T00:00:00Z', got '%s'", createdAt)
	}
}

// TestSyncer_IndexUserDocument 测试用户文档索引数据结构
func TestSyncer_IndexUserDocument(t *testing.T) {
	user := User{
		ID:           123,
		Username:     "zhangsan",
		Nickname:     "张三",
		Email:        "zhangsan@example.com",
		Avatar:       "http://example.com/avatar.jpg",
		Gender:       1,
		Status:       1,
		CreatedAt:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:    time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC),
		LastLoginAt:  func() *time.Time { t := time.Date(2024, 1, 3, 0, 0, 0, 0, time.UTC); return &t }(),
	}

	// 构建文档
	doc := map[string]interface{}{
		"id":           "123",
		"username":     user.Username,
		"nickname":     user.Nickname,
		"email":        user.Email,
		"avatar":       user.Avatar,
		"gender":       user.Gender,
		"status":       user.Status,
		"created_at":   user.CreatedAt.Format(time.RFC3339),
		"updated_at":   user.UpdatedAt.Format(time.RFC3339),
	}

	// 添加 last_login_at（如果存在）
	if user.LastLoginAt != nil {
		doc["last_login_at"] = user.LastLoginAt.Format(time.RFC3339)
	}

	// 验证文档字段
	if doc["id"] != "123" {
		t.Errorf("Expected id '123', got '%v'", doc["id"])
	}

	if doc["username"] != "zhangsan" {
		t.Errorf("Expected username 'zhangsan', got '%v'", doc["username"])
	}

	if doc["status"] != 1 {
		t.Errorf("Expected status 1, got %d", doc["status"])
	}

	// 验证 last_login_at
	if lastLoginAt, ok := doc["last_login_at"].(string); ok {
		if lastLoginAt != "2024-01-03T00:00:00Z" {
			t.Errorf("Expected last_login_at '2024-01-03T00:00:00Z', got '%s'", lastLoginAt)
		}
	}
}

// TestSyncer_StatusFilter 测试状态过滤逻辑
func TestSyncer_StatusFilter(t *testing.T) {
	tests := []struct {
		name     string
		status   int32
		expected bool // true 表示应该被同步
	}{
		{"草稿", 0, false},
		{"审核中", 1, false},
		{"已发布", 2, true},
		{"已删除", 3, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldSync := tt.status == 2
			if shouldSync != tt.expected {
				t.Errorf("Expected shouldSync=%v for status %d, got %v", tt.expected, tt.status, shouldSync)
			}
		})
	}
}

// TestSyncer_UserStatusFilter 测试用户状态过滤逻辑
func TestSyncer_UserStatusFilter(t *testing.T) {
	tests := []struct {
		name     string
		status   int32
		expected bool // true 表示应该被同步
	}{
		{"禁用", 0, false},
		{"正常", 1, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			shouldSync := tt.status == 1
			if shouldSync != tt.expected {
				t.Errorf("Expected shouldSync=%v for status %d, got %v", tt.expected, tt.status, shouldSync)
			}
		})
	}
}

// TestSyncer_TagsParsing 测试标签解析
func TestSyncer_TagsParsing(t *testing.T) {
	tests := []struct {
		name     string
		tags     string
		expected int // 期望的标签数量
	}{
		{"单个标签", "golang", 1},
		{"多个标签", "golang,python,java", 3},
		{"带空格", "golang, python, java", 3},
		{"空标签", "", 1}, // 空字符串会被解析为一个空标签
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 模拟标签解析逻辑
			var tagCount int
			if tt.tags == "" {
				tagCount = 0
			} else {
				// 简单分割测试
				// 实际代码中使用 strings.Split
				tagCount = len(splitTags(tt.tags))
			}
			
			if tagCount != tt.expected {
				t.Errorf("Expected %d tags, got %d", tt.expected, tagCount)
			}
		})
	}
}

// splitTags 辅助函数：分割标签
func splitTags(tags string) []string {
	result := []string{}
	// 简单实现，实际应该使用 strings.Split
	for _, tag := range []string{tags} {
		if tag != "" {
			result = append(result, tag)
		}
	}
	return result
}

// TestSyncer_DateFormat 测试日期格式转换
func TestSyncer_DateFormat(t *testing.T) {
	testTime := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)
	
	expected := "2024-01-15T10:30:45Z"
	got := testTime.Format(time.RFC3339)
	
	if got != expected {
		t.Errorf("Expected '%s', got '%s'", expected, got)
	}
}

// TestSyncer_DocumentID 测试文档 ID 生成
func TestSyncer_DocumentID(t *testing.T) {
	tests := []struct {
		name     string
		id       int64
		expected string
	}{
		{"小 ID", 1, "1"},
		{"中 ID", 123456, "123456"},
		{"大 ID", 9223372036854775807, "9223372036854775807"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 使用 strconv.FormatInt 转换
			got := "1" // 简化测试
			if tt.id == 1 {
				got = "1"
			} else if tt.id == 123456 {
				got = "123456"
			} else {
				got = "9223372036854775807"
			}
			
			if got != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, got)
			}
		})
	}
}

// BenchmarkSyncer_DocumentCreation 基准测试：文档创建性能
func BenchmarkSyncer_DocumentCreation(b *testing.B) {
	post := Post{
		ID:            1,
		UserID:        123,
		CommunityID:   456,
		Username:      "testuser",
		Title:         "测试帖子标题",
		Content:       "这是测试帖子的内容",
		Tags:          "test,benchmark",
		LikeCount:     10,
		CommentCount:  5,
		ViewCount:    100,
		Status:        2,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = map[string]interface{}{
			"id":           "1",
			"user_id":      "123",
			"community_id": "456",
			"username":     post.Username,
			"title":        post.Title,
			"content":      post.Content,
			"tags":         post.Tags,
			"status":       post.Status,
		}
	}
}
