package repository

import (
	"context"
	"database/sql"

	"auth/internal/model"

	"github.com/zeromicro/go-zero/core/stores/sqlx"
)

type userRepository struct {
	db sqlx.SqlConn
}

func NewUserRepository(db sqlx.SqlConn) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByUsernameAndPassword(ctx context.Context, username, password string) (*model.User, error) {
	var user model.User
	err := r.db.QueryRowCtx(ctx, &user, "SELECT id, username, password_hash, email FROM users WHERE username=? AND password_hash=?", username, password)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *model.User) (int64, error) {
	result, err := r.db.ExecCtx(ctx, "INSERT INTO users(username, password_hash, email) VALUES(?, ?, ?)", user.Username, user.PasswordHash, user.Email)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (r *userRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	var user model.User
	err := r.db.QueryRowCtx(ctx, &user, "SELECT id, username, password_hash, email FROM users WHERE id=?", id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
