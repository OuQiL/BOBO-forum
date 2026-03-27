package repository

import (
	"context"

	"post/internal/model"
)

type PostRepository interface {
	Create(ctx context.Context, post *model.Post) (int64, error)
	FindByID(ctx context.Context, id int64) (*model.Post, error)
	GetPostDetail(ctx context.Context, userID int64, id int64) (*model.Post, error)
	List(ctx context.Context, page, pageSize int64, communityID int64, tag string) ([]*model.Post, int64, error)
	ListByUserID(ctx context.Context, userID, page, pageSize int64) ([]*model.Post, int64, error)
	GetRecentPosts(ctx context.Context, days int) ([]*model.Post, error)
	Update(ctx context.Context, post *model.Post) error
	Delete(ctx context.Context, id, userID int64) error
	IncrementViewCount(ctx context.Context, postID int64) error
	IncrementCommentCount(ctx context.Context, postID int64) error
	DecrementCommentCount(ctx context.Context, postID int64) error
}
