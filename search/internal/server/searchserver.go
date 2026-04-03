package server

import (
	"context"
	"search/api/proto"
	"search/internal/logic"
	"search/internal/svc"
)

type SearchServer struct {
	proto.UnimplementedSearchServer
	svcCtx *svc.ServiceContext
}

func NewSearchServer(svcCtx *svc.ServiceContext) *SearchServer {
	return &SearchServer{
		svcCtx: svcCtx,
	}
}

func (s *SearchServer) Search(ctx context.Context, req *proto.SearchRequest) (*proto.SearchResponse, error) {
	l := logic.NewSearchLogic(ctx, s.svcCtx)
	return l.Search(req)
}

func (s *SearchServer) SyncData(ctx context.Context, req *proto.SyncRequest) (*proto.SyncResponse, error) {
	l := logic.NewSyncLogic(ctx, s.svcCtx)
	return l.SyncData(req)
}

func (s *SearchServer) HealthCheck(ctx context.Context, req *proto.HealthCheckRequest) (*proto.HealthCheckResponse, error) {
	l := logic.NewHealthCheckLogic(ctx, s.svcCtx)
	return l.HealthCheck(req)
}
