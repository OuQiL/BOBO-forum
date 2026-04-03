package sync

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"search/internal/es"
	"search/internal/svc"

	"gorm.io/gorm"
)

// Post 帖子模型
type Post struct {
	ID            int64     `gorm:"column:id;primaryKey"`
	UserID        int64     `gorm:"column:user_id"`
	CommunityID   int64     `gorm:"column:community_id"`
	Username      string    `gorm:"column:username"`
	Title         string    `gorm:"column:title"`
	Content       string    `gorm:"column:content"`
	Tags          string    `gorm:"column:tags"`
	LikeCount     int64     `gorm:"column:like_count"`
	CommentCount  int64     `gorm:"column:comment_count"`
	ViewCount     int64     `gorm:"column:view_count"`
	Status        int32     `gorm:"column:status"`
	CreatedAt     time.Time `gorm:"column:created_at"`
	UpdatedAt     time.Time `gorm:"column:updated_at"`
}

func (Post) TableName() string {
	return "posts"
}

// User 用户模型
type User struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	Username    string     `gorm:"column:username"`
	PasswordHash string    `gorm:"column:password_hash"`
	Email       string     `gorm:"column:email"`
	Nickname    string     `gorm:"column:nickname"`
	Avatar      string     `gorm:"column:avatar"`
	Gender      int32      `gorm:"column:gender"`
	Status      int32      `gorm:"column:status"`
	CreatedAt   time.Time  `gorm:"column:created_at"`
	UpdatedAt   time.Time  `gorm:"column:updated_at"`
	LastLoginAt *time.Time `gorm:"column:last_login_at"`
}

func (User) TableName() string {
	return "users"
}

// Syncer 数据同步器
type Syncer struct {
	svcCtx *svc.ServiceContext
}

// NewSyncer 创建同步器
func NewSyncer(svcCtx *svc.ServiceContext) *Syncer {
	return &Syncer{
		svcCtx: svcCtx,
	}
}

// SyncAllPosts 全量同步帖子数据
func (s *Syncer) SyncAllPosts(ctx context.Context) (int, error) {
	var posts []Post
	result := s.svcCtx.MySQL.Find(&posts)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to query posts: %w", result.Error)
	}

	count := 0
	for _, post := range posts {
		// 只同步已发布的帖子
		if post.Status != 2 {
			continue
		}

		if err := s.IndexPost(ctx, &post); err != nil {
			fmt.Printf("Failed to index post %d: %v\n", post.ID, err)
			continue
		}
		count++
	}

	fmt.Printf("Synced %d posts to Elasticsearch\n", count)
	return count, nil
}

// SyncAllUsers 全量同步用户数据
func (s *Syncer) SyncAllUsers(ctx context.Context) (int, error) {
	var users []User
	result := s.svcCtx.MySQL.Find(&users)
	if result.Error != nil {
		return 0, fmt.Errorf("failed to query users: %w", result.Error)
	}

	count := 0
	for _, user := range users {
		// 只同步正常状态的用户
		if user.Status != 1 {
			continue
		}

		if err := s.IndexUser(ctx, &user); err != nil {
			fmt.Printf("Failed to index user %d: %v\n", user.ID, err)
			continue
		}
		count++
	}

	fmt.Printf("Synced %d users to Elasticsearch\n", count)
	return count, nil
}

// IndexPost 索引单个帖子
func (s *Syncer) IndexPost(ctx context.Context, post *Post) error {
	doc := map[string]interface{}{
		"id":            strconv.FormatInt(post.ID, 10),
		"user_id":       strconv.FormatInt(post.UserID, 10),
		"community_id":  strconv.FormatInt(post.CommunityID, 10),
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

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal post: %w", err)
	}

	res, err := s.svcCtx.ES.GetClient().Index(
		"posts",
		bytes.NewReader(body),
		s.svcCtx.ES.GetClient().Index.WithDocumentID(strconv.FormatInt(post.ID, 10)),
		s.svcCtx.ES.GetClient().Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to index post: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index error: %s", res.String())
	}

	return nil
}

// IndexUser 索引单个用户
func (s *Syncer) IndexUser(ctx context.Context, user *User) error {
	doc := map[string]interface{}{
		"id":           strconv.FormatInt(user.ID, 10),
		"username":     user.Username,
		"nickname":     user.Nickname,
		"email":        user.Email,
		"avatar":       user.Avatar,
		"gender":       user.Gender,
		"status":       user.Status,
		"created_at":   user.CreatedAt.Format(time.RFC3339),
		"updated_at":   user.UpdatedAt.Format(time.RFC3339),
	}

	if user.LastLoginAt != nil {
		doc["last_login_at"] = user.LastLoginAt.Format(time.RFC3339)
	}

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	res, err := s.svcCtx.ES.GetClient().Index(
		"users",
		bytes.NewReader(body),
		s.svcCtx.ES.GetClient().Index.WithDocumentID(strconv.FormatInt(user.ID, 10)),
		s.svcCtx.ES.GetClient().Index.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to index user: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("index error: %s", res.String())
	}

	return nil
}

// DeletePost 删除帖子索引
func (s *Syncer) DeletePost(ctx context.Context, postID int64) error {
	res, err := s.svcCtx.ES.GetClient().Delete(
		"posts",
		strconv.FormatInt(postID, 10),
		s.svcCtx.ES.GetClient().Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete error: %s", res.String())
	}

	return nil
}

// DeleteUser 删除用户索引
func (s *Syncer) DeleteUser(ctx context.Context, userID int64) error {
	res, err := s.svcCtx.ES.GetClient().Delete(
		"users",
		strconv.FormatInt(userID, 10),
		s.svcCtx.ES.GetClient().Delete.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() && res.StatusCode != 404 {
		return fmt.Errorf("delete error: %s", res.String())
	}

	return nil
}

// SyncAll 全量同步所有数据
func (s *Syncer) SyncAll(ctx context.Context) (int, error) {
	postCount, err := s.SyncAllPosts(ctx)
	if err != nil {
		return postCount, err
	}

	userCount, err := s.SyncAllUsers(ctx)
	if err != nil {
		return postCount + userCount, err
	}

	return postCount + userCount, nil
}
