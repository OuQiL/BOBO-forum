package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"post/internal/model"
	snowid "post/pkg/snowid"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type postRepository struct {
	db          sqlx.SqlConn
	redis       *redis.Redis
	localCache  *HotPostLocalCache
	hotPostSvc  *HotPostService
	cacheLoader *CacheLoader
}

func NewPostRepository(db sqlx.SqlConn, rds *redis.Redis, localCache *HotPostLocalCache, hotPostSvc *HotPostService) PostRepository {
	return &postRepository{
		db:          db,
		redis:       rds,
		localCache:  localCache,
		hotPostSvc:  hotPostSvc,
		cacheLoader: NewCacheLoader(rds, localCache),
	}
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

	postID, err := snowid.NextID()
	if err != nil {
		return 0, fmt.Errorf("failed to generate snowflake ID: %w", err)
	}
	post.ID = postID

	createdAt := time.Unix(post.CreatedAt, 0)
	updatedAt := time.Unix(post.UpdatedAt, 0)

	result, err := r.db.ExecCtx(ctx, "INSERT INTO posts(id, user_id, community_id, username, title, content, tags, status, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		post.ID, post.UserID, post.CommunityID, post.Username, post.Title, post.Content, string(tagsJSON), post.Status, createdAt, updatedAt)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}
	if affected == 0 {
		return 0, fmt.Errorf("no rows affected")
	}

	if err := AddWhenCreate(ctx, r.redis, postID, post.UserID); err != nil {
		return postID, err
	}

	return postID, nil
}
func (r *postRepository) GetPostDetail(ctx context.Context, userID int64, id int64) (*model.Post, error) {
	post, err := r.cacheLoader.GetOrLoad(ctx, id, func(ctx context.Context) (*model.Post, error) {
		return r.findByIDFromDB(ctx, id)
	})
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, nil
	}

	post.ViewCount = r.getViewCountFromRedis(ctx, id, post.ViewCount)

	if userID > 0 && userID != post.UserID {
		AddRead(ctx, r.redis, id, userID)
		post.ViewCount = r.getViewCountFromRedis(ctx, id, post.ViewCount)
	}

	return post, nil
}

func (r *postRepository) findByIDFromDB(ctx context.Context, id int64) (*model.Post, error) {
	var post model.Post
	err := r.db.QueryRowCtx(ctx, &post, "SELECT id, user_id, community_id, username, title, content, tags, like_count, comment_count, view_count, status, UNIX_TIMESTAMP(created_at) as created_at, UNIX_TIMESTAMP(updated_at) as updated_at FROM posts WHERE id=? AND status != 3", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &post, nil
}
func (r *postRepository) FindByID(ctx context.Context, id int64) (*model.Post, error) {
	var post model.Post
	err := r.db.QueryRowCtx(ctx, &post, "SELECT id, user_id, community_id, username, title, content, tags, like_count, comment_count, view_count, status, UNIX_TIMESTAMP(created_at) as created_at, UNIX_TIMESTAMP(updated_at) as updated_at FROM posts WHERE id=? AND status != 3", id)
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
	if err != nil {
		return err
	}

	return r.hotPostSvc.DeletePostCache(ctx, post.ID)
}

func (r *postRepository) Delete(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecCtx(ctx, "UPDATE posts SET status = 3 WHERE id=? AND user_id=?", id, userID)
	if err != nil {
		return err
	}

	return r.hotPostSvc.DeletePostCache(ctx, id)
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
	query := "SELECT id, user_id, community_id, username, title, content, tags, like_count, comment_count, view_count, status, UNIX_TIMESTAMP(created_at) as created_at, UNIX_TIMESTAMP(updated_at) as updated_at FROM posts WHERE status = 2 AND created_at > DATE_SUB(NOW(), INTERVAL ? DAY) ORDER BY created_at DESC"

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
