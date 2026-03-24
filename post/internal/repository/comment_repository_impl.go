package repository

import (
	"context"

	"post/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type commentRepository struct {
	db sqlx.SqlConn
}

func NewCommentRepository(db sqlx.SqlConn) CommentRepository {
	return &commentRepository{db: db}
}

func (r *commentRepository) Create(ctx context.Context, comment *model.Comment) (int64, error) {
	result, err := r.db.ExecCtx(ctx, "INSERT INTO comments(post_id, user_id, username, content, parent_id, created_at, updated_at) VALUES(?, ?, ?, ?, ?, ?, ?)",
		comment.PostID, comment.UserID, comment.Username, comment.Content, comment.ParentID, comment.CreatedAt, comment.UpdatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *commentRepository) FindByID(ctx context.Context, id int64) (*model.Comment, error) {
	var comment model.Comment
	err := r.db.QueryRowCtx(ctx, &comment, "SELECT id, post_id, user_id, username, content, parent_id, created_at, updated_at FROM comments WHERE id=?", id)
	if err != nil {
		return nil, err
	}
	return &comment, nil
}

func (r *commentRepository) ListByPostID(ctx context.Context, postID int64) ([]*model.Comment, error) {
	var comments []*model.Comment
	err := r.db.QueryRowsCtx(ctx, &comments, "SELECT id, post_id, user_id, username, content, parent_id, created_at, updated_at FROM comments WHERE post_id=? ORDER BY created_at ASC", postID)
	if err != nil {
		return nil, err
	}
	return comments, nil
}

func (r *commentRepository) Delete(ctx context.Context, id, userID int64) error {
	_, err := r.db.ExecCtx(ctx, "DELETE FROM comments WHERE id=? AND user_id=?", id, userID)
	return err
}
