package repository

import (
	"context"

	"interaction/internal/model"
)

type LikeRepository interface {
	Create(ctx context.Context, like *model.Like) error
	Update(ctx context.Context, like *model.Like) error
	FindByTargetAndUser(ctx context.Context, targetId, userId int64) (*model.Like, error)
	GetLikeStatus(ctx context.Context, userId int64, targetIds []int64) (map[int64]bool, error)
	Transact(ctx context.Context, fn func(ctx context.Context, repo LikeRepository) error) error
}