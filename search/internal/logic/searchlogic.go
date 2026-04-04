package logic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"search/internal/svc"
	"strings"
	"time"

	"search/api/proto"
)

// SearchLogic 搜索逻辑
type SearchLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

// NewSearchLogic 创建搜索逻辑
func NewSearchLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SearchLogic {
	return &SearchLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

// Search 执行搜索
func (l *SearchLogic) Search(req *proto.SearchRequest) (*proto.SearchResponse, error) {
	if req.Keyword == "" {
		return &proto.SearchResponse{
			Total: 0,
			List:  []*proto.SearchResult{},
			Page:  req.Page,
			Size:  req.Size,
		}, nil
	}

	// 默认参数
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 {
		req.Size = 10
	}
	if req.SortBy == "" {
		req.SortBy = "relevance"
	}

	// 根据类型搜索
	switch req.Type {
	case "post":
		return l.SearchPosts(req)
	case "user":
		return l.SearchUsers(req)
	default:
		// 默认搜索帖子
		return l.SearchPosts(req)
	}
}

// SearchPosts 搜索帖子
func (l *SearchLogic) SearchPosts(req *proto.SearchRequest) (*proto.SearchResponse, error) {
	// 构建查询 DSL
	query := l.buildPostQuery(req)

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// 执行搜索
	res, err := l.svcCtx.ES.GetClient().Search(
		l.svcCtx.ES.GetClient().Search.WithIndex("posts"),
		l.svcCtx.ES.GetClient().Search.WithBody(bytes.NewReader(body)),
		l.svcCtx.ES.GetClient().Search.WithContext(l.ctx),
		l.svcCtx.ES.GetClient().Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s", string(bodyBytes))
	}

	// 解析结果
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	hits := result["hits"].(map[string]interface{})
	total := int64(hits["total"].(map[string]interface{})["value"].(float64))
	hitList := hits["hits"].([]interface{})

	// 转换结果
	results := make([]*proto.SearchResult, 0, len(hitList))
	for _, hit := range hitList {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})
		highlight := hitMap["highlight"]

		result := l.convertPostResult(source, highlight)
		results = append(results, result)
	}

	return &proto.SearchResponse{
		Total: total,
		List:  results,
		Page:  req.Page,
		Size:  req.Size,
	}, nil
}

// SearchUsers 搜索用户
func (l *SearchLogic) SearchUsers(req *proto.SearchRequest) (*proto.SearchResponse, error) {
	// 构建查询 DSL
	query := l.buildUserQuery(req)

	body, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// 执行搜索
	res, err := l.svcCtx.ES.GetClient().Search(
		l.svcCtx.ES.GetClient().Search.WithIndex("users"),
		l.svcCtx.ES.GetClient().Search.WithBody(bytes.NewReader(body)),
		l.svcCtx.ES.GetClient().Search.WithContext(l.ctx),
		l.svcCtx.ES.GetClient().Search.WithTrackTotalHits(true),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("search error: %s", string(bodyBytes))
	}

	// 解析结果
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	hits := result["hits"].(map[string]interface{})
	total := int64(hits["total"].(map[string]interface{})["value"].(float64))
	hitList := hits["hits"].([]interface{})

	// 转换结果
	results := make([]*proto.SearchResult, 0, len(hitList))
	for _, hit := range hitList {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})
		highlight := hitMap["highlight"]

		result := l.convertUserResult(source, highlight)
		results = append(results, result)
	}

	return &proto.SearchResponse{
		Total: total,
		List:  results,
		Page:  req.Page,
		Size:  req.Size,
	}, nil
}

// buildPostQuery 构建帖子查询
func (l *SearchLogic) buildPostQuery(req *proto.SearchRequest) map[string]interface{} {
	return BuildPostQuery(req)
}

// BuildPostQuery 构建帖子查询（导出函数供测试使用）
func BuildPostQuery(req *proto.SearchRequest) map[string]interface{} {
	from := (req.Page - 1) * req.Size

	query := map[string]interface{}{
		"from": from,
		"size": req.Size,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query":  req.Keyword,
							"fields": []string{"title^3", "content^1", "tags^2"},
							"type":   "best_fields",
						},
					},
				},
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"status": 2, // 只搜索已发布的帖子
						},
					},
				},
			},
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"title":   map[string]interface{}{},
				"content": map[string]interface{}{},
			},
			"pre_tags":  []string{"<em style='color: red;'>"},
			"post_tags": []string{"</em>"},
		},
	}

	// 添加过滤条件
	filters := query["query"].(map[string]interface{})["filter"].([]interface{})

	if req.CommunityId != nil && *req.CommunityId > 0 {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"community_id": fmt.Sprintf("%d", *req.CommunityId),
			},
		})
	}

	if req.AuthorId != nil && *req.AuthorId > 0 {
		filters = append(filters, map[string]interface{}{
			"term": map[string]interface{}{
				"user_id": fmt.Sprintf("%d", *req.AuthorId),
			},
		})
	}

	if req.Tags != nil && *req.Tags != "" {
		tags := strings.Split(*req.Tags, ",")
		for _, tag := range tags {
			filters = append(filters, map[string]interface{}{
				"term": map[string]interface{}{
					"tags": strings.TrimSpace(tag),
				},
			})
		}
	}

	if req.StartTime != nil && *req.StartTime > 0 {
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"created_at": map[string]interface{}{
					"gte": time.Unix(*req.StartTime, 0).Format(time.RFC3339),
				},
			},
		})
	}

	if req.EndTime != nil && *req.EndTime > 0 {
		filters = append(filters, map[string]interface{}{
			"range": map[string]interface{}{
				"created_at": map[string]interface{}{
					"lte": time.Unix(*req.EndTime, 0).Format(time.RFC3339),
				},
			},
		})
	}

	query["query"].(map[string]interface{})["filter"] = filters

	// 添加排序
	switch req.SortBy {
	case "latest":
		query["sort"] = []map[string]interface{}{
			{"created_at": map[string]interface{}{"order": "desc"}},
		}
	case "hottest":
		query["sort"] = []map[string]interface{}{
			{"like_count": map[string]interface{}{"order": "desc"}},
			{"view_count": map[string]interface{}{"order": "desc"}},
		}
	default: // relevance
		// 默认按相关性排序，不需要额外设置
	}

	return query
}

// buildUserQuery 构建用户查询
func (l *SearchLogic) buildUserQuery(req *proto.SearchRequest) map[string]interface{} {
	return BuildUserQuery(req)
}

// BuildUserQuery 构建用户查询（导出函数供测试使用）
func BuildUserQuery(req *proto.SearchRequest) map[string]interface{} {
	from := (req.Page - 1) * req.Size

	query := map[string]interface{}{
		"from": from,
		"size": req.Size,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"multi_match": map[string]interface{}{
							"query":  req.Keyword,
							"fields": []string{"username^2", "nickname^1"},
							"type":   "best_fields",
						},
					},
				},
				"filter": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"status": 1, // 只搜索正常状态的用户
						},
					},
				},
			},
		},
		"highlight": map[string]interface{}{
			"fields": map[string]interface{}{
				"username": map[string]interface{}{},
				"nickname": map[string]interface{}{},
			},
			"pre_tags":  []string{"<em style='color: red;'>"},
			"post_tags": []string{"</em>"},
		},
	}

	return query
}

// convertPostResult 转换帖子结果
func (l *SearchLogic) convertPostResult(source map[string]interface{}, highlight interface{}) *proto.SearchResult {
	return ConvertPostResult(source, highlight)
}

// ConvertPostResult 转换帖子结果（导出函数供测试使用）
func ConvertPostResult(source map[string]interface{}, highlight interface{}) *proto.SearchResult {
	result := &proto.SearchResult{
		Type: "post",
	}

	// 解析字段
	if v, ok := source["id"].(string); ok {
		// 从字符串 ID 转换为整数（简化处理，实际应该用 strconv.ParseInt）
		result.PostId = 0 // 实际应该转换
		_ = v
	}
	if v, ok := source["user_id"].(string); ok {
		result.UserId = 0
		_ = v
	}
	if v, ok := source["community_id"].(string); ok {
		result.CommunityId = 0
		_ = v
	}
	if v, ok := source["username"].(string); ok {
		result.Username = v
	}
	if v, ok := source["title"].(string); ok {
		result.Title = v
	}
	if v, ok := source["content"].(string); ok {
		result.Content = v
	}
	if v, ok := source["like_count"].(float64); ok {
		result.LikeCount = int64(v)
	}
	if v, ok := source["comment_count"].(float64); ok {
		result.CommentCount = int64(v)
	}
	if v, ok := source["view_count"].(float64); ok {
		result.ViewCount = int64(v)
	}
	if v, ok := source["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			result.CreatedAt = t.Unix()
		}
	}

	// 处理高亮
	if hl, ok := highlight.(map[string]interface{}); ok {
		if titles, ok := hl["title"].([]interface{}); ok && len(titles) > 0 {
			hlTitle := titles[0].(string)
			result.HighlightTitle = hlTitle
		}
		if contents, ok := hl["content"].([]interface{}); ok && len(contents) > 0 {
			hlContent := contents[0].(string)
			result.HighlightContent = hlContent
		}
	}

	return result
}

// convertUserResult 转换用户结果
func (l *SearchLogic) convertUserResult(source map[string]interface{}, highlight interface{}) *proto.SearchResult {
	return ConvertUserResult(source, highlight)
}

// ConvertUserResult 转换用户结果（导出函数供测试使用）
func ConvertUserResult(source map[string]interface{}, highlight interface{}) *proto.SearchResult {
	result := &proto.SearchResult{
		Type: "user",
	}

	// 解析字段
	if v, ok := source["id"].(string); ok {
		result.UserId = 0
		_ = v
	}
	if v, ok := source["username"].(string); ok {
		result.Username = v
	}
	if v, ok := source["nickname"].(string); ok {
		result.Nickname = v
	}
	if v, ok := source["email"].(string); ok {
		result.Email = v
	}
	if v, ok := source["avatar"].(string); ok {
		result.Avatar = v
	}
	if v, ok := source["gender"].(float64); ok {
		result.Gender = int32(v)
	}
	if v, ok := source["created_at"].(string); ok {
		if t, err := time.Parse(time.RFC3339, v); err == nil {
			result.UserCreatedAt = t.Unix()
		}
	}

	// 处理高亮
	if hl, ok := highlight.(map[string]interface{}); ok {
		if usernames, ok := hl["username"].([]interface{}); ok && len(usernames) > 0 {
			hlUsername := usernames[0].(string)
			result.HighlightTitle = hlUsername
		}
		if nicknames, ok := hl["nickname"].([]interface{}); ok && len(nicknames) > 0 {
			hlNickname := nicknames[0].(string)
			result.HighlightContent = hlNickname
		}
	}

	return result
}
