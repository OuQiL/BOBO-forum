package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"post/internal/model"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type postRepository struct {
	db         sqlx.SqlConn
	redis      *redis.Redis
	localCache *HotPostLocalCache
}

func NewPostRepository(db sqlx.SqlConn, rds *redis.Redis, localCache *HotPostLocalCache) PostRepository {
	return &postRepository{db: db, redis: rds, localCache: localCache}
}

func (r *postRepository) getViewCountFromRedis(ctx context.Context, postID int64, dbViewCount int64) int64 {
	count, err := GetRC(ctx, r.redis, postID)
	if err != nil {
		return dbViewCount
	}
	return count
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) (int64, error) {
	tagsJSON, err := json.Marshal(post.Tags)
	if err != nil {
		return 0, err
	}

	result, err := r.db.ExecCtx(ctx, "INSERT INTO posts(user_id, community_id, username, title, content, tags, status, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)",
		post.UserID, post.CommunityID, post.Username, post.Title, post.Content, string(tagsJSON), post.Status, post.CreatedAt, post.UpdatedAt)
	if err != nil {
		return 0, err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}

	if err := AddWhenCreate(ctx, r.redis, postID, post.UserID); err != nil {
		return postID, err
	}

	return postID, nil
}
func (r *postRepository) GetPostDetail(ctx context.Context, userID int64, id int64) (*model.Post, error) {
	if post, found := r.localCache.Get(id); found {
		post.ViewCount = r.getViewCountFromRedis(ctx, id, post.ViewCount)
		if userID > 0 && userID != post.UserID {
			AddRead(ctx, r.redis, id, userID)
			post.ViewCount = r.getViewCountFromRedis(ctx, id, post.ViewCount)
		}
		return post, nil
	}

	var post model.Post
	err := r.db.QueryRowCtx(ctx, &post, "SELECT id, user_id, community_id, username, title, content, tags, like_count, comment_count, view_count, status, created_at, updated_at FROM posts WHERE id=? AND status != 3", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	post.ViewCount = r.getViewCountFromRedis(ctx, id, post.ViewCount)

	if userID > 0 && userID != post.UserID {
		AddRead(ctx, r.redis, id, userID)
		post.ViewCount = r.getViewCountFromRedis(ctx, id, post.ViewCount)
	}

	if r.localCache.ShouldCache(post.ViewCount) {
		r.localCache.Set(&post)
	}

	return &post, nil
}
func (r *postRepository) FindByID(ctx context.Context, id int64) (*model.Post, error) {
	var post model.Post
	err := r.db.QueryRowCtx(ctx, &post, "SELECT id, user_id, community_id, username, title, content, tags, like_count, comment_count, view_count, status, created_at, updated_at FROM posts WHERE id=? AND status != 3", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}

func (r *postRepository) List(ctx context.Context, page, pageSize int64, communityID int64, tag string) ([]*model.Post, int64, error) {
	offset := (page - 1) * pageSize
	var posts []*model.Post
	var total int64

	query := "SELECT id, user_id, community_id, username, title, content, tags, like_count, comment_count, view_count, status, created_at, updated_at FROM posts"
	countQuery := "SELECT COUNT(*) FROM posts"
	var whereClauses []string
	var args []interface{}

	whereClauses = append(whereClauses, "status = 2")

	if communityID > 0 {
		whereClauses = append(whereClauses, "community_id = ?")
		args = append(args, communityID)
	}

	if tag != "" {
		whereClauses = append(whereClauses, "JSON_CONTAINS(tags, ?)")
		args = append(args, fmt.Sprintf(`"%s"`, tag))
	}

	if len(whereClauses) > 0 {
		whereClause := " WHERE " + strings.Join(whereClauses, " AND ")
		query += whereClause
		countQuery += whereClause
	}

	query += " ORDER BY created_at DESC LIMIT ? OFFSET ?"
	args = append(args, pageSize, offset)

	err := r.db.QueryRowCtx(ctx, &total, countQuery, args[:len(args)-2]...)
	if err != nil {
		return nil, 0, err
	}

	err = r.db.QueryRowsCtx(ctx, &posts, query, args...)
	if err != nil {
		return nil, 0, err
	}

	for _, post := range posts {
		post.ViewCount = r.getViewCountFromRedis(ctx, post.ID, post.ViewCount)
	}

	return posts, total, nil
}

func (r *postRepository) ListByUserID(ctx context.Context, userID, page, pageSize int64) ([]*model.Post, int64, error) {
	offset := (page - 1) * pageSize
	var posts []*model.Post
	var total int64

	query := "SELECT id, user_id, community_id, username, title, content, tags, like_count, comment_count, view_count, status, created_at, updated_at FROM posts WHERE user_id = ? AND status = 2 ORDER BY created_at DESC LIMIT ? OFFSET ?"
	countQuery := "SELECT COUNT(*) FROM posts WHERE user_id = ? AND status = 2"

	err := r.db.QueryRowCtx(ctx, &total, countQuery, userID)
	if err != nil {
		return nil, 0, err
	}

	err = r.db.QueryRowsCtx(ctx, &posts, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	for _, post := range posts {
		post.ViewCount = r.getViewCountFromRedis(ctx, post.ID, post.ViewCount)
	}

	return posts, total, nil
}

func (r *postRepository) Update(ctx context.Context, post *model.Post) error {
	tagsJSON, err := json.Marshal(post.Tags)
	if err != nil {
		return err
	}

	_, err = r.db.ExecCtx(ctx, "UPDATE posts SET title=?, content=?, tags=?, updated_at=? WHERE id=? AND user_id=?",
		post.Title, post.Content, string(tagsJSON), post.UpdatedAt, post.ID, post.UserID)
	return err
}

func (r *postRepository) Delete(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecCtx(ctx, "UPDATE posts SET status = 3 WHERE id=? AND user_id=?", id, userID)
	if err != nil {
		return err
	}
	r.localCache.Delete(id)
	return nil
}

func (r *postRepository) IncrementViewCount(ctx context.Context, postID int64) error {
	_, err := r.db.ExecCtx(ctx, "UPDATE posts SET view_count = view_count + 1 WHERE id=?", postID)
	return err
}

func (r *postRepository) IncrementCommentCount(ctx context.Context, postID int64) error {
	_, err := r.db.ExecCtx(ctx, "UPDATE posts SET comment_count = comment_count + 1 WHERE id=?", postID)
	return err
}

func (r *postRepository) DecrementCommentCount(ctx context.Context, postID int64) error {
	_, err := r.db.ExecCtx(ctx, "UPDATE posts SET comment_count = GREATEST(comment_count - 1, 0) WHERE id=?", postID)
	return err
}

func (r *postRepository) GetRecentPosts(ctx context.Context, days int) ([]*model.Post, error) {
	query := "SELECT id, user_id, community_id, username, title, content, tags, like_count, comment_count, view_count, status, created_at, updated_at FROM posts WHERE status = 2 AND created_at > DATE_SUB(NOW(), INTERVAL ? DAY) ORDER BY created_at DESC"

	var posts []*model.Post
	err := r.db.QueryRowsCtx(ctx, &posts, query, days)
	if err != nil {
		return nil, err
	}

	for _, post := range posts {
		post.ViewCount = r.getViewCountFromRedis(ctx, post.ID, post.ViewCount)
	}

	return posts, nil
}
