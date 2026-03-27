package server

import (
	"context"

	"post/api/proto"
	"post/internal/logic"
	"post/internal/svc"
)

type PostServer struct {
	svcCtx *svc.ServiceContext
	proto.UnimplementedPostServer
}

func NewPostServer(svcCtx *svc.ServiceContext) *PostServer {
	return &PostServer{
		svcCtx: svcCtx,
	}
}

func (s *PostServer) CreatePost(ctx context.Context, in *proto.CreatePostRequest) (*proto.CreatePostResponse, error) {
	l := logic.NewCreatePostLogic(ctx, s.svcCtx)
	return l.CreatePost(in)
}

func (s *PostServer) GetPost(ctx context.Context, in *proto.GetPostRequest) (*proto.GetPostResponse, error) {
	l := logic.NewGetPostLogic(ctx, s.svcCtx)
	return l.GetPost(in)
}

func (s *PostServer) ListPosts(ctx context.Context, in *proto.ListPostsRequest) (*proto.ListPostsResponse, error) {
	l := logic.NewListPostsLogic(ctx, s.svcCtx)
	return l.ListPosts(in)
}

func (s *PostServer) ListUserPosts(ctx context.Context, in *proto.ListUserPostsRequest) (*proto.ListUserPostsResponse, error) {
	l := logic.NewListUserPostsLogic(ctx, s.svcCtx)
	return l.ListUserPosts(in)
}

func (s *PostServer) UpdatePost(ctx context.Context, in *proto.UpdatePostRequest) (*proto.UpdatePostResponse, error) {
	l := logic.NewUpdatePostLogic(ctx, s.svcCtx)
	return l.UpdatePost(in)
}

func (s *PostServer) DeletePost(ctx context.Context, in *proto.DeletePostRequest) (*proto.DeletePostResponse, error) {
	l := logic.NewDeletePostLogic(ctx, s.svcCtx)
	return l.DeletePost(in)
}

func (s *PostServer) CreateComment(ctx context.Context, in *proto.CreateCommentRequest) (*proto.CreateCommentResponse, error) {
	l := logic.NewCreateCommentLogic(ctx, s.svcCtx)
	return l.CreateComment(in)
}

func (s *PostServer) DeleteComment(ctx context.Context, in *proto.DeleteCommentRequest) (*proto.DeleteCommentResponse, error) {
	l := logic.NewDeleteCommentLogic(ctx, s.svcCtx)
	return l.DeleteComment(in)
}

func (s *PostServer) IncrementViewCount(ctx context.Context, in *proto.IncrementViewCountRequest) (*proto.IncrementViewCountResponse, error) {
	l := logic.NewIncrementViewCountLogic(ctx, s.svcCtx)
	return l.IncrementViewCount(in)
}
