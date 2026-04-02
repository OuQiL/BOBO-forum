package logic

import (
	"context"

	"interaction/api/proto"
	"interaction/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCommentsLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetCommentsLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCommentsLogic {
	return &GetCommentsLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetCommentsLogic) GetComments(in *proto.GetCommentsRequest) (*proto.GetCommentsResponse, error) {
	return &proto.GetCommentsResponse{}, nil
}