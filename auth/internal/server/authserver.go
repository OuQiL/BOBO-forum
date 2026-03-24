package server

import (
	"context"

	"auth/api/proto"
	"auth/internal/logic"
	"auth/internal/svc"
)

type AuthServer struct {
	svcCtx *svc.ServiceContext
	proto.UnimplementedAuthServer
}

func NewAuthServer(svcCtx *svc.ServiceContext) *AuthServer {
	return &AuthServer{
		svcCtx: svcCtx,
	}
}

func (s *AuthServer) Register(ctx context.Context, in *proto.RegisterRequest) (*proto.LoginResponse, error) {
	l := logic.NewRegisterLogic(ctx, s.svcCtx)
	return l.Register(in)
}

func (s *AuthServer) Login(ctx context.Context, in *proto.LoginRequest) (*proto.LoginResponse, error) {
	l := logic.NewLoginLogic(ctx, s.svcCtx)
	return l.Login(in)
}
