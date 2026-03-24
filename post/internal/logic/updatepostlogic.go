package logic

import (
	"context"
	"encoding/json"
	"time"

	"post/api/proto"
	"post/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UpdatePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUpdatePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UpdatePostLogic {
	return &UpdatePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UpdatePostLogic) UpdatePost(in *proto.UpdatePostRequest) (*proto.UpdatePostResponse, error) {
	post, err := l.svcCtx.PostRepo.FindByID(l.ctx, in.PostId)
	if err != nil {
		return nil, err
	}
	if post == nil {
		return nil, nil
	}

	tagsJSON, err := json.Marshal(in.Tags)
	if err != nil {
		return nil, err
	}

	post.Title = in.Title
	post.Content = in.Content
	post.Tags = string(tagsJSON)
	post.UpdatedAt = time.Now().Unix()

	if err := l.svcCtx.PostRepo.Update(l.ctx, post); err != nil {
		return nil, err
	}

	return &proto.UpdatePostResponse{
		Post: &proto.PostInfo{
			Id:           post.ID,
			UserId:       post.UserID,
			CommunityId:  post.CommunityID,
			Username:     post.Username,
			Title:        post.Title,
			Content:      post.Content,
			Tags:         in.Tags,
			LikeCount:    post.LikeCount,
			CommentCount: post.CommentCount,
			ViewCount:    post.ViewCount,
			CreatedAt:    post.CreatedAt,
			UpdatedAt:    post.UpdatedAt,
		},
	}, nil
}
