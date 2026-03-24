package logic

import (
	"context"

	"post/api/proto"
	"post/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeleteCommentLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeleteCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeleteCommentLogic {
	return &DeleteCommentLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeleteCommentLogic) DeleteComment(in *proto.DeleteCommentRequest) (*proto.DeleteCommentResponse, error) {
	comment, err := l.svcCtx.CommentRepo.FindByID(l.ctx, in.CommentId)
	if err != nil {
		return &proto.DeleteCommentResponse{
			Success: false,
		}, err
	}
	if comment == nil {
		return &proto.DeleteCommentResponse{
			Success: false,
		}, nil
	}

	err = l.svcCtx.CommentRepo.Delete(l.ctx, in.CommentId, in.UserId)
	if err != nil {
		return &proto.DeleteCommentResponse{
			Success: false,
		}, err
	}

	err = l.svcCtx.PostRepo.DecrementCommentCount(l.ctx, comment.PostID)
	if err != nil {
		return &proto.DeleteCommentResponse{
			Success: false,
		}, err
	}

	return &proto.DeleteCommentResponse{
		Success: true,
	}, nil
}
