package auth

import (
	"context"

	"auth/api/proto"

	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type (
	LoginRequest    = proto.LoginRequest
	LoginResponse   = proto.LoginResponse
	RegisterRequest = proto.RegisterRequest
	UserInfo        = proto.UserInfo

	Auth interface {
		Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*LoginResponse, error)
		Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error)
	}

	defaultAuth struct {
		cli zrpc.Client
	}
)

func NewAuth(cli zrpc.Client) Auth {
	return &defaultAuth{
		cli: cli,
	}
}

func (m *defaultAuth) Register(ctx context.Context, in *RegisterRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	client := proto.NewAuthClient(m.cli.Conn())
	return client.Register(ctx, in, opts...)
}

func (m *defaultAuth) Login(ctx context.Context, in *LoginRequest, opts ...grpc.CallOption) (*LoginResponse, error) {
	client := proto.NewAuthClient(m.cli.Conn())
	return client.Login(ctx, in, opts...)
}
