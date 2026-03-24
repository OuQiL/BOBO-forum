package logic

import (
	"context"

	"post/api/proto"
	"post/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type IncrementViewCountLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewIncrementViewCountLogic(ctx context.Context, svcCtx *svc.ServiceContext) *IncrementViewCountLogic {
	return &IncrementViewCountLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *IncrementViewCountLogic) IncrementViewCount(in *proto.IncrementViewCountRequest) (*proto.IncrementViewCountResponse, error) {
	err := l.svcCtx.PostRepo.IncrementViewCount(l.ctx, in.PostId)
	if err != nil {
		return &proto.IncrementViewCountResponse{
			Success: false,
		}, err
	}

	return &proto.IncrementViewCountResponse{
		Success: true,
	}, nil
}
