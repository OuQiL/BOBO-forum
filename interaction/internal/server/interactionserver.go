package server

import (
	"context"

	"interaction/api/proto"
	"interaction/internal/svc"
)

type InteractionServer struct {
	svcCtx *svc.ServiceContext
	proto.UnimplementedInteractionServer
}

func NewInteractionServer(svcCtx *svc.ServiceContext) *InteractionServer {
	return &InteractionServer{
		svcCtx: svcCtx,
	}
}

func (s *InteractionServer) Like(ctx context.Context, in *proto.LikeRequest) (*proto.LikeResponse, error) {
	l := logic.NewLikeLogic(ctx, s.svcCtx)
	return l.Like(in)
}

func (s *InteractionServer) Unlike(ctx context.Context, in *proto.UnlikeRequest) (*proto.UnlikeResponse, error) {
	l := logic.NewUnlikeLogic(ctx, s.svcCtx)
	return l.Unlike(in)
}

func (s *InteractionServer) CreateComment(ctx context.Context, in *proto.CreateCommentRequest) (*proto.CreateCommentResponse, error) {
	l := logic.NewCreateCommentLogic(ctx, s.svcCtx)
	return l.CreateComment(in)
}

func (s *InteractionServer) DeleteComment(ctx context.Context, in *proto.DeleteCommentRequest) (*proto.DeleteCommentResponse, error) {
	l := logic.NewDeleteCommentLogic(ctx, s.svcCtx)
	return l.DeleteComment(in)
}

func (s *InteractionServer) GetComments(ctx context.Context, in *proto.GetCommentsRequest) (*proto.GetCommentsResponse, error) {
	l := logic.NewGetCommentsLogic(ctx, s.svcCtx)
	return l.GetComments(in)
}

func (s *InteractionServer) GetLikeStatus(ctx context.Context, in *proto.GetLikeStatusRequest) (*proto.GetLikeStatusResponse, error) {
	l := logic.NewGetLikeStatusLogic(ctx, s.svcCtx)
	return l.GetLikeStatus(in)
}