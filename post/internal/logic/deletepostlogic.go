package logic

import (
	"context"

	"post/api/proto"
	"post/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type DeletePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewDeletePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *DeletePostLogic {
	return &DeletePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *DeletePostLogic) DeletePost(in *proto.DeletePostRequest) (*proto.DeletePostResponse, error) {
	err := l.svcCtx.PostRepo.Delete(l.ctx, in.PostId, in.UserId)
	if err != nil {
		return &proto.DeletePostResponse{
			Success: false,
		}, err
	}

	return &proto.DeletePostResponse{
		Success: true,
	}, nil
}
