package logic

import (
	"context"
	"fmt"
	"time"

	"interaction/api/proto"
	"interaction/internal/kafka"
	"interaction/internal/repository"
	"interaction/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type UnlikeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewUnlikeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *UnlikeLogic {
	return &UnlikeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *UnlikeLogic) Unlike(in *proto.UnlikeRequest) (*proto.UnlikeResponse, error) {
	l.Infof("Unlike request: user_id=%d, target_id=%d, target_type=%d", in.UserId, in.TargetId, in.TargetType)

	alreadyUnliked := false
	err := l.svcCtx.LikeRepo.Transact(l.ctx, func(ctx context.Context, repo repository.LikeRepository) error {
		existingLike, err := repo.FindByTargetAndUser(ctx, in.TargetId, in.UserId)
		if err != nil {
			return fmt.Errorf("find like failed: %w", err)
		}

		if existingLike == nil {
			l.Infof("User %d has not liked target %d", in.UserId, in.TargetId)
			alreadyUnliked = true
			return nil
		}

		if existingLike.Status == 0 {
			l.Infof("User %d already unliked target %d", in.UserId, in.TargetId)
			alreadyUnliked = true
			return nil
		}

		now := time.Now()
		existingLike.Status = 0
		existingLike.Unliketime = &now

		err = repo.Update(ctx, existingLike)
		if err != nil {
			return fmt.Errorf("update like failed: %w", err)
		}
		l.Infof("Unlike success in DB: user_id=%d, target_id=%d", in.UserId, in.TargetId)
		return nil
	})

	if err != nil {
		l.Errorf("Unlike transaction failed: %v", err)
		return nil, err
	}

	if alreadyUnliked {
		return &proto.UnlikeResponse{Success: true}, nil
	}

	_, err = l.svcCtx.LikeCache.DecrLikeCount(l.ctx, in.TargetType, in.TargetId)
	if err != nil {
		l.Errorf("Failed to decrement like count in cache: %v", err)
	}

	err = l.svcCtx.LikeCache.RemoveLikeRelation(l.ctx, in.TargetType, in.TargetId, in.UserId)
	if err != nil {
		l.Errorf("Failed to remove like relation in cache: %v", err)
	}

	msg := &kafka.LikeMessage{
		TargetType: in.TargetType,
		TargetId:   in.TargetId,
		UserId:     in.UserId,
		Action:     kafka.ActionUnlike,
		Timestamp:  time.Now().Unix(),
	}

	err = l.svcCtx.KafkaProducer.SendMessage(l.ctx, msg)
	if err != nil {
		l.Errorf("Failed to send unlike message to kafka: %v", err)
	}

	l.Infof("Unlike message sent to kafka: user_id=%d, target_id=%d", in.UserId, in.TargetId)

	return &proto.UnlikeResponse{Success: true}, nil
}