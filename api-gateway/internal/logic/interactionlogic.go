package logic

import (
	"context"

	"api-gateway/internal/svc"
	"api-gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type InteractionLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewInteractionLogic(ctx context.Context, svcCtx *svc.ServiceContext) *InteractionLogic {
	return &InteractionLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *InteractionLogic) Like(req *types.LikeRequest) (resp interface{}, err error) {
	userID, ok := GetUserIDFromContext(l.ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	username, _ := GetUsernameFromContext(l.ctx)
	l.Infof("User %d (%s) liking post: %d", userID, username, req.PostID)

	return map[string]interface{}{
		"message": "点赞成功",
		"post_id": req.PostID,
	}, nil
}

func (l *InteractionLogic) Unlike(req *types.LikeRequest) (resp interface{}, err error) {
	userID, ok := GetUserIDFromContext(l.ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	l.Infof("User %d unliking post: %d", userID, req.PostID)

	return map[string]interface{}{
		"message": "取消点赞成功",
		"post_id": req.PostID,
	}, nil
}

func (l *InteractionLogic) Comment(req *types.CommentRequest) (resp interface{}, err error) {
	userID, ok := GetUserIDFromContext(l.ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	username, _ := GetUsernameFromContext(l.ctx)
	l.Infof("User %d (%s) commenting on post: %d", userID, username, req.PostID)

	return map[string]interface{}{
		"message":    "评论成功",
		"post_id":    req.PostID,
		"content":    req.Content,
		"comment_id": 1,
	}, nil
}

func (l *InteractionLogic) Follow(req *types.FollowRequest) (resp interface{}, err error) {
	userID, ok := GetUserIDFromContext(l.ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	username, _ := GetUsernameFromContext(l.ctx)
	l.Infof("User %d (%s) following user: %d", userID, username, req.UserID)

	return map[string]interface{}{
		"message": "关注成功",
		"user_id": req.UserID,
	}, nil
}

func (l *InteractionLogic) Unfollow(req *types.FollowRequest) (resp interface{}, err error) {
	userID, ok := GetUserIDFromContext(l.ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	l.Infof("User %d unfollowing user: %d", userID, req.UserID)

	return map[string]interface{}{
		"message": "取消关注成功",
		"user_id": req.UserID,
	}, nil
}

func (l *InteractionLogic) GetFollowers(userID int64, page, size int) (resp interface{}, err error) {
	return map[string]interface{}{
		"total": 0,
		"list":  []interface{}{},
	}, nil
}

func (l *InteractionLogic) GetFollowing(userID int64, page, size int) (resp interface{}, err error) {
	return map[string]interface{}{
		"total": 0,
		"list":  []interface{}{},
	}, nil
}
