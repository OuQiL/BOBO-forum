package logic

import (
	"context"
	"fmt"
	"time"

	"interaction/api/proto"
	"interaction/internal/kafka"
	"interaction/internal/model"
	"interaction/internal/repository"
	"interaction/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikeLogic {
	return &LikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LikeLogic) Like(in *proto.LikeRequest) (*proto.LikeResponse, error) {
	l.Infof("Like request: user_id=%d, target_id=%d, target_type=%d", in.UserId, in.TargetId, in.TargetType)

	alreadyLiked := false
	err := l.svcCtx.LikeRepo.Transact(l.ctx, func(ctx context.Context, repo repository.LikeRepository) error {
		existingLike, err := repo.FindByTargetAndUser(ctx, in.TargetId, in.UserId)
		if err != nil {
			return fmt.Errorf("find like failed: %w", err)
		}

		if existingLike != nil {
			if existingLike.Status == 1 {
				l.Infof("User %d already liked target %d", in.UserId, in.TargetId)
				alreadyLiked = true
				return nil
			}

			now := time.Now()
			existingLike.Status = 1
			existingLike.Liketime = now
			existingLike.Unliketime = nil

			err = repo.Update(ctx, existingLike)
			if err != nil {
				return fmt.Errorf("update like failed: %w", err)
			}
			l.Infof("Re-like success: user_id=%d, target_id=%d", in.UserId, in.TargetId)
			return nil
		}

		newLike := &model.Like{
			Type:     int8(in.TargetType),
			TargetId: in.TargetId,
			UserId:   in.UserId,
			Status:   1,
			Liketime: time.Now(),
		}

		err = repo.Create(ctx, newLike)
		if err != nil {
			return fmt.Errorf("create like failed: %w", err)
		}
		l.Infof("New like success: user_id=%d, target_id=%d, like_id=%d", in.UserId, in.TargetId, newLike.Id)
		return nil
	})

	if err != nil {
		l.Errorf("Like transaction failed: %v", err)
		return nil, err
	}

	if alreadyLiked {
		return &proto.LikeResponse{Success: true}, nil
	}

	_, err = l.svcCtx.LikeCache.IncrLikeCount(l.ctx, in.TargetType, in.TargetId)
	if err != nil {
		l.Errorf("Failed to increment like count in cache: %v", err)
	}

	err = l.svcCtx.LikeCache.AddLikeRelation(l.ctx, in.TargetType, in.TargetId, in.UserId)
	if err != nil {
		l.Errorf("Failed to add like relation in cache: %v", err)
	}

	msg := &kafka.LikeMessage{
		TargetType: in.TargetType,
		TargetId:   in.TargetId,
		UserId:     in.UserId,
		Action:     kafka.ActionLike,
		Timestamp:  time.Now().Unix(),
	}

	err = l.svcCtx.KafkaProducer.SendMessage(l.ctx, msg)
	if err != nil {
		l.Errorf("Failed to send like message to kafka: %v", err)
	}

	l.Infof("Like message sent to kafka: user_id=%d, target_id=%d", in.UserId, in.TargetId)

	return &proto.LikeResponse{Success: true}, nil
}