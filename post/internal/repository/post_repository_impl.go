package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"post/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type postRepository struct {
	db sqlx.SqlConn
}

func NewPostRepository(db sqlx.SqlConn) PostRepository {
	return &postRepository{db: db}
}

func (r *postRepository) Create(ctx context.Context, post *model.Post) (int64, error) {
	tagsJSON, err := json.Marshal(post.Tags)
	if err != nil {
		return 0, err
	}

	result, err := r.db.ExecCtx(ctx, "INSERT INTO posts(user_id, community_id, username, title, content, tags, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?, ?)",
		post.UserID, post.CommunityID, post.Username, post.Title, post.Content, string(tagsJSON), post.CreatedAt, post.UpdatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *postRepository) FindByID(ctx context.Context, id int64) (*model.Post, error) {
	var post model.Post
	err := r.db.QueryRowCtx(ctx, &post, "SELECT id, user_id, community_id, username, title, content, tags, like_count, comment_count, view_count, created_at, updated_at FROM posts WHERE id=?", id)
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

	query := "SELECT id, user_id, community_id, username, title, content, tags, like_count, comment_count, view_count, created_at, updated_at FROM posts"
	countQuery := "SELECT COUNT(*) FROM posts"
	var whereClauses []string
	var args []interface{}

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
	_, err := r.db.ExecCtx(ctx, "DELETE FROM posts WHERE id=? AND user_id=?", id, userID)
	return err
}

func (r *postRepository) IncrementViewCount(ctx context.Context, postID int64) error {
	_, err := r.db.ExecCtx(ctx, "UPDATE posts SET view_count = view_count + 1 WHERE id=?", postID)
	return err
}

func (r *postRepository) IncrementLikeCount(ctx context.Context, postID int64) error {
	_, err := r.db.ExecCtx(ctx, "UPDATE posts SET like_count = like_count + 1 WHERE id=?", postID)
	return err
}

func (r *postRepository) DecrementLikeCount(ctx context.Context, postID int64) error {
	_, err := r.db.ExecCtx(ctx, "UPDATE posts SET like_count = GREATEST(like_count - 1, 0) WHERE id=?", postID)
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
