package repository

import (
	"context"

	"post/internal/model"
)

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) (int64, error)
	FindByID(ctx context.Context, id int64) (*model.Post, error)
	List(ctx context.Context, page, pageSize int64, communityID int64, tag string) ([]*model.Post, int64, error)
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id, userID int64) error
	IncrementViewCount(ctx context.Context, postID int64) error
	IncrementLikeCount(ctx context.Context, postID int64) error
	DecrementLikeCount(ctx context.Context, postID int64) error
	IncrementCommentCount(ctx context.Context, postID int64) error
	DecrementCommentCount(ctx context.Context, postID int64) error
}
