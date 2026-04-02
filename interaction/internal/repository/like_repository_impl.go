package repository

import (
	"context"
	"database/sql"
	"fmt"

	"interaction/internal/model"
	snowid "interaction/pkg/snowid"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type likeRepository struct {
	db sqlx.SqlConn
}

type likeTxRepository struct {
	session sqlx.Session
}

func NewLikeRepository(db sqlx.SqlConn) LikeRepository {
	return &likeRepository{db: db}
}

func (r *likeRepository) Transact(ctx context.Context, fn func(ctx context.Context, repo LikeRepository) error) error {
	return r.db.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		txRepo := &likeTxRepository{session: session}
		return fn(ctx, txRepo)
	})
}

func (r *likeRepository) Create(ctx context.Context, like *model.Like) error {
	id, err := snowid.NextID()
	if err != nil {
		return fmt.Errorf("generate snowflake id failed: %w", err)
	}
	like.Id = id

	query := `INSERT INTO likes (id, type, target_id, user_id, status, liketime) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.db.ExecCtx(ctx, query, like.Id, like.Type, like.TargetId, like.UserId, like.Status, like.Liketime)
	if err != nil {
		return fmt.Errorf("insert like failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected failed: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("expected 1 row affected, got %d", rows)
	}

	return nil
}

func (r *likeRepository) Update(ctx context.Context, like *model.Like) error {
	query := `UPDATE likes SET status = ?, liketime = ?, unliketime = ?, updated_at = CURRENT_TIMESTAMP WHERE target_id = ? AND user_id = ?`
	result, err := r.db.ExecCtx(ctx, query, like.Status, like.Liketime, like.Unliketime, like.TargetId, like.UserId)
	if err != nil {
		return fmt.Errorf("update like failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected failed: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("expected 1 row affected, got %d", rows)
	}

	return nil
}

func (r *likeRepository) FindByTargetAndUser(ctx context.Context, targetId, userId int64) (*model.Like, error) {
	query := `SELECT id, type, target_id, user_id, status, liketime, unliketime, created_at, updated_at FROM likes WHERE target_id = ? AND user_id = ?`
	var like model.Like
	err := r.db.QueryRowCtx(ctx, &like, query, targetId, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find like failed: %w", err)
	}
	return &like, nil
}

func (r *likeRepository) GetLikeStatus(ctx context.Context, userId int64, targetIds []int64) (map[int64]bool, error) {
	if len(targetIds) == 0 {
		return make(map[int64]bool), nil
	}

	query := `SELECT target_id, status FROM likes WHERE user_id = ? AND target_id IN (?)`
	args := append([]interface{}{userId}, interfaceSlice(targetIds)...)

	var results []struct {
		TargetId int64 `db:"target_id"`
		Status   int8  `db:"status"`
	}
	err := r.db.QueryRowsCtx(ctx, &results, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query like status failed: %w", err)
	}

	result := make(map[int64]bool)
	for _, res := range results {
		result[res.TargetId] = res.Status == 1
	}

	return result, nil
}

// likeTxRepository methods for transaction
func (r *likeTxRepository) Transact(ctx context.Context, fn func(ctx context.Context, repo LikeRepository) error) error {
	return fmt.Errorf("nested transaction not supported")
}

func (r *likeTxRepository) Create(ctx context.Context, like *model.Like) error {
	id, err := snowid.NextID()
	if err != nil {
		return fmt.Errorf("generate snowflake id failed: %w", err)
	}
	like.Id = id

	query := `INSERT INTO likes (id, type, target_id, user_id, status, liketime) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.session.ExecCtx(ctx, query, like.Id, like.Type, like.TargetId, like.UserId, like.Status, like.Liketime)
	if err != nil {
		return fmt.Errorf("insert like failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected failed: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("expected 1 row affected, got %d", rows)
	}

	return nil
}

func (r *likeTxRepository) Update(ctx context.Context, like *model.Like) error {
	query := `UPDATE likes SET status = ?, liketime = ?, unliketime = ?, updated_at = CURRENT_TIMESTAMP WHERE target_id = ? AND user_id = ?`
	result, err := r.session.ExecCtx(ctx, query, like.Status, like.Liketime, like.Unliketime, like.TargetId, like.UserId)
	if err != nil {
		return fmt.Errorf("update like failed: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("get rows affected failed: %w", err)
	}
	if rows != 1 {
		return fmt.Errorf("expected 1 row affected, got %d", rows)
	}

	return nil
}

func (r *likeTxRepository) FindByTargetAndUser(ctx context.Context, targetId, userId int64) (*model.Like, error) {
	query := `SELECT id, type, target_id, user_id, status, liketime, unliketime, created_at, updated_at FROM likes WHERE target_id = ? AND user_id = ?`
	var like model.Like
	err := r.session.QueryRowCtx(ctx, &like, query, targetId, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find like failed: %w", err)
	}
	return &like, nil
}

func (r *likeTxRepository) GetLikeStatus(ctx context.Context, userId int64, targetIds []int64) (map[int64]bool, error) {
	if len(targetIds) == 0 {
		return make(map[int64]bool), nil
	}

	query := `SELECT target_id, status FROM likes WHERE user_id = ? AND target_id IN (?)`
	args := append([]interface{}{userId}, interfaceSlice(targetIds)...)

	var results []struct {
		TargetId int64 `db:"target_id"`
		Status   int8  `db:"status"`
	}
	err := r.session.QueryRowsCtx(ctx, &results, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query like status failed: %w", err)
	}

	result := make(map[int64]bool)
	for _, res := range results {
		result[res.TargetId] = res.Status == 1
	}

	return result, nil
}

func interfaceSlice(slice []int64) []interface{} {
	result := make([]interface{}, len(slice))
	for i, v := range slice {
		result[i] = v
	}
	return result
}