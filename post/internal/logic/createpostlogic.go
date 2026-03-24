package logic

import (
	"context"
	"encoding/json"
	"time"

	"post/api/proto"
	"post/internal/model"
	"post/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreatePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreatePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreatePostLogic {
	return &CreatePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreatePostLogic) CreatePost(in *proto.CreatePostRequest) (*proto.CreatePostResponse, error) {
	tagsJSON, err := json.Marshal(in.Tags)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	post := &model.Post{
		UserID:      in.UserId,
		CommunityID: in.CommunityId,
		Title:       in.Title,
		Content:     in.Content,
		Tags:        string(tagsJSON),
		LikeCount:   0,
		CommentCount: 0,
		ViewCount:   0,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	id, err := l.svcCtx.PostRepo.Create(l.ctx, post)
	if err != nil {
		return nil, err
	}

	post.ID = id

	return &proto.CreatePostResponse{
		Post: &proto.PostInfo{
			Id:           post.ID,
			UserId:       post.UserID,
			CommunityId:  post.CommunityID,
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
