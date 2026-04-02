package model

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type Like struct {
	Id         int64      `db:"id"`
	Type       int8       `db:"type"`
	TargetId   int64      `db:"target_id"`
	UserId     int64      `db:"user_id"`
	Status     int8       `db:"status"`
	Liketime   time.Time  `db:"liketime"`
	Unliketime *time.Time `db:"unliketime"`
	CreatedAt  time.Time  `db:"created_at"`
	UpdatedAt  time.Time  `db:"updated_at"`
}

type LikeModel struct {
	conn sqlx.SqlConn
}

func NewLikeModel(conn sqlx.SqlConn) *LikeModel {
	return &LikeModel{conn: conn}
}

func (m *LikeModel) Transact(ctx context.Context, fn func(ctx context.Context, session sqlx.Session) error) error {
	return m.conn.TransactCtx(ctx, func(ctx context.Context, session sqlx.Session) error {
		return fn(ctx, session)
	})
}

func (m *LikeModel) Insert(ctx context.Context, session sqlx.Session, data *Like) error {
	query := `INSERT INTO likes (id, type, target_id, user_id, status, liketime) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := session.ExecCtx(ctx, query, data.Id, data.Type, data.TargetId, data.UserId, data.Status, data.Liketime)
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

func (m *LikeModel) Update(ctx context.Context, session sqlx.Session, data *Like) error {
	query := `UPDATE likes SET status = ?, liketime = ?, unliketime = ?, updated_at = CURRENT_TIMESTAMP WHERE target_id = ? AND user_id = ?`
	result, err := session.ExecCtx(ctx, query, data.Status, data.Liketime, data.Unliketime, data.TargetId, data.UserId)
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

func (m *LikeModel) FindByTargetAndUser(ctx context.Context, session sqlx.Session, targetId, userId int64) (*Like, error) {
	query := `SELECT id, type, target_id, user_id, status, liketime, unliketime, created_at, updated_at FROM likes WHERE target_id = ? AND user_id = ?`
	var like Like
	err := session.QueryRowCtx(ctx, &like, query, targetId, userId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("find like failed: %w", err)
	}
	return &like, nil
}

func (m *LikeModel) GetLikeStatus(ctx context.Context, userId int64, targetIds []int64) (map[int64]bool, error) {
	if len(targetIds) == 0 {
		return make(map[int64]bool), nil
	}

	query := `SELECT target_id, status FROM likes WHERE user_id = ? AND target_id IN (?)`
	args := append([]interface{}{userId}, interfaceSlice(targetIds)...)

	var results []struct {
		TargetId int64 `db:"target_id"`
		Status   int8  `db:"status"`
	}
	err := m.conn.QueryRowsCtx(ctx, &results, query, args...)
	if err != nil {
		return nil, fmt.Errorf("query like status failed: %w", err)
	}

	result := make(map[int64]bool)
	for _, r := range results {
		result[r.TargetId] = r.Status == 1
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