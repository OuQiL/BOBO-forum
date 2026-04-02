package logic

import (
	"context"
	"fmt"

	"interaction/api/proto"
	"interaction/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetLikeStatusLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetLikeStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetLikeStatusLogic {
	return &GetLikeStatusLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetLikeStatusLogic) GetLikeStatus(in *proto.GetLikeStatusRequest) (*proto.GetLikeStatusResponse, error) {
	l.Infof("GetLikeStatus request: user_id=%d, target_ids=%v, target_type=%d", in.UserId, in.TargetIds, in.TargetType)

	statusMap, err := l.svcCtx.LikeCache.BatchGetLikeStatus(l.ctx, in.TargetType, in.UserId, in.TargetIds)
	if err != nil {
		l.Errorf("Get like status from cache failed: %v, fallback to DB", err)
		
		statusMap, err = l.svcCtx.LikeRepo.GetLikeStatus(l.ctx, in.UserId, in.TargetIds)
		if err != nil {
			l.Errorf("Get like status from DB failed: %v", err)
			return nil, fmt.Errorf("get like status failed: %w", err)
		}
	}

	var statuses []*proto.LikeStatus
	for _, targetId := range in.TargetIds {
		isLiked := statusMap[targetId]
		statuses = append(statuses, &proto.LikeStatus{
			TargetId: targetId,
			IsLiked:  isLiked,
		})
	}

	l.Infof("GetLikeStatus success: user_id=%d, results_count=%d", in.UserId, len(statuses))
	return &proto.GetLikeStatusResponse{Statuses: statuses}, nil
}