package logic

import (
	"context"
	"time"

	"post/api/proto"
	"post/internal/model"
	"post/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type CreateCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewCreateCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *CreateCommentLogic {
	return &CreateCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *CreateCommentLogic) CreateComment(in *proto.CreateCommentRequest) (*proto.CreateCommentResponse, error) {
	now := time.Now().Unix()
	comment := &model.Comment{
		PostID:    in.PostId,
		UserID:    in.UserId,
		Content:   in.Content,
		ParentID:  in.ParentId,
		CreatedAt: now,
		UpdatedAt: now,
	}

	id, err := l.svcCtx.CommentRepo.Create(l.ctx, comment)
	if err != nil {
		return nil, err
	}

	comment.ID = id

	err = l.svcCtx.PostRepo.IncrementCommentCount(l.ctx, in.PostId)
	if err != nil {
		return nil, err
	}

	return &proto.CreateCommentResponse{
		Comment: &proto.CommentInfo{
			Id:        comment.ID,
			PostId:    comment.PostID,
			UserId:    comment.UserID,
			Content:   comment.Content,
			ParentId:  comment.ParentID,
			CreatedAt: comment.CreatedAt,
			UpdatedAt: comment.UpdatedAt,
		},
	}, nil
}
