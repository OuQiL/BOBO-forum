package repository

import (
	"context"

	"auth/internal/model"
)

type UserRepository interface {
	FindByUsernameAndPassword(ctx context.Context, username, password string) (*model.User, error)
	Create(ctx context.Context, user *model.User) (int64, error)
	FindByID(ctx context.Context, id int64) (*model.User, error)
}
