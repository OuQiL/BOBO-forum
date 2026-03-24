package repository

import (
	"context"

	"post/internal/model"
)

type CommentRepository interface {
	Create(ctx context.Context, comment *model.Comment) (int64, error)
	FindByID(ctx context.Context, id int64) (*model.Comment, error)
	ListByPostID(ctx context.Context, postID int64) ([]*model.Comment, error)
	Delete(ctx context.Context, id, userID int64) error
}
