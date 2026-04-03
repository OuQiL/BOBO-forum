package es

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// IndexMapping 定义索引映射
var IndexMapping = map[string]map[string]interface{}{
	"posts": {
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":            map[string]interface{}{"type": "keyword"},
				"user_id":       map[string]interface{}{"type": "keyword"},
				"community_id":  map[string]interface{}{"type": "keyword"},
				"username":      map[string]interface{}{"type": "keyword"},
				"title": map[string]interface{}{
					"type":      "text",
					"analyzer":  "ik_smart",
					"search_analyzer": "ik_smart",
				},
				"content": map[string]interface{}{
					"type":      "text",
					"analyzer":  "ik_smart",
					"search_analyzer": "ik_smart",
				},
				"tags": map[string]interface{}{
					"type":      "text",
					"analyzer":  "ik_smart",
					"search_analyzer": "ik_smart",
				},
				"like_count":     map[string]interface{}{"type": "integer"},
				"comment_count":  map[string]interface{}{"type": "integer"},
				"view_count":     map[string]interface{}{"type": "integer"},
				"status":         map[string]interface{}{"type": "integer"},
				"created_at":     map[string]interface{}{"type": "date"},
				"updated_at":     map[string]interface{}{"type": "date"},
			},
		},
		"settings": map[string]interface{}{
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"ik_smart": map[string]interface{}{
						"type": "ik_smart",
					},
				},
			},
		},
	},
	"users": {
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				"id":           map[string]interface{}{"type": "keyword"},
				"username": map[string]interface{}{
					"type":      "text",
					"analyzer":  "ik_smart",
					"search_analyzer": "ik_smart",
				},
				"nickname": map[string]interface{}{
					"type":      "text",
					"analyzer":  "ik_smart",
					"search_analyzer": "ik_smart",
				},
				"email":        map[string]interface{}{"type": "keyword"},
				"avatar":       map[string]interface{}{"type": "keyword"},
				"gender":       map[string]interface{}{"type": "byte"},
				"status":       map[string]interface{}{"type": "integer"},
				"created_at":   map[string]interface{}{"type": "date"},
				"updated_at":   map[string]interface{}{"type": "date"},
				"last_login_at": map[string]interface{}{"type": "date"},
			},
		},
		"settings": map[string]interface{}{
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					"ik_smart": map[string]interface{}{
						"type": "ik_smart",
					},
				},
			},
		},
	},
}

// CreateIndex 创建索引
func (c *Client) CreateIndex(ctx context.Context, indexName string) error {
	// 检查索引是否已存在
	exists, err := c.IndexExists(ctx, indexName)
	if err != nil {
		return err
	}
	if exists {
		fmt.Printf("Index %s already exists\n", indexName)
		return nil
	}

	// 创建索引
	mapping := IndexMapping[indexName]
	body, err := json.Marshal(mapping)
	if err != nil {
		return fmt.Errorf("failed to marshal mapping: %w", err)
	}

	res, err := c.ES.Indices.Create(
		indexName,
		c.ES.Indices.Create.WithBody(bytes.NewReader(body)),
		c.ES.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to create index: %s", string(bodyBytes))
	}

	fmt.Printf("Index %s created successfully\n", indexName)
	return nil
}

// IndexExists 检查索引是否存在
func (c *Client) IndexExists(ctx context.Context, indexName string) (bool, error) {
	res, err := c.ES.Indices.Exists([]string{indexName}, c.ES.Indices.Exists.WithContext(ctx))
	if err != nil {
		return false, err
	}
	defer res.Body.Close()

	return !res.IsError(), nil
}

// DeleteIndex 删除索引
func (c *Client) DeleteIndex(ctx context.Context, indexName string) error {
	res, err := c.ES.Indices.Delete([]string{indexName}, c.ES.Indices.Delete.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("failed to delete index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		bodyBytes, _ := io.ReadAll(res.Body)
		return fmt.Errorf("failed to delete index: %s", string(bodyBytes))
	}

	fmt.Printf("Index %s deleted successfully\n", indexName)
	return nil
}

// CreateAllIndices 创建所有索引
func (c *Client) CreateAllIndices(ctx context.Context) error {
	for indexName := range IndexMapping {
		if err := c.CreateIndex(ctx, indexName); err != nil {
			return fmt.Errorf("failed to create index %s: %w", indexName, err)
		}
	}
	return nil
}
