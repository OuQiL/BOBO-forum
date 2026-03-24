package logic

import (
	"context"
	"fmt"

	"post/api/proto"
	"post/internal/svc"

	"github.com/zeromicro/go-zero/core/logx"
)

type LikePostLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLikePostLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LikePostLogic {
	return &LikePostLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LikePostLogic) LikePost(in *proto.LikePostRequest) (*proto.LikePostResponse, error) {
	likeKey := fmt.Sprintf("post:like:%d:%d", in.PostId, in.UserId)
	
	liked, err := l.svcCtx.Redis.Get(likeKey)
	if err != nil {
		return &proto.LikePostResponse{
			Success: false,
			Liked:   false,
		}, err
	}

	if liked == "1" {
		err = l.svcCtx.PostRepo.DecrementLikeCount(l.ctx, in.PostId)
		if err != nil {
			return &proto.LikePostResponse{
				Success: false,
				Liked:   false,
			}, err
		}
		
		err = l.svcCtx.Redis.Del(likeKey)
		if err != nil {
			return &proto.LikePostResponse{
				Success: false,
				Liked:   false,
			}, err
		}

		return &proto.LikePostResponse{
			Success: true,
			Liked:   false,
		}, nil
	} else {
		err = l.svcCtx.PostRepo.IncrementLikeCount(l.ctx, in.PostId)
		if err != nil {
			return &proto.LikePostResponse{
				Success: false,
				Liked:   false,
			}, err
		}
		
		err = l.svcCtx.Redis.Set(likeKey, "1")
		if err != nil {
			return &proto.LikePostResponse{
				Success: false,
				Liked:   false,
			}, err
		}

		return &proto.LikePostResponse{
			Success: true,
			Liked:   true,
		}, nil
	}
}
