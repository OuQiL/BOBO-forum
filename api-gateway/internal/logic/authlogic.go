package logic

import (
	"context"

	"api-gateway/internal/svc"
	"api-gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type AuthLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewAuthLogic(ctx context.Context, svcCtx *svc.ServiceContext) *AuthLogic {
	return &AuthLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *AuthLogic) Register(req *types.RegisterRequest) (resp interface{}, err error) {
	// TODO: 调用auth.rpc服务进行用户注册
	// conn := l.svcCtx.AuthRpc
	// client := auth.NewAuthClient(conn)
	// return client.Register(l.ctx, &auth.RegisterRequest{...})

	return map[string]interface{}{
		"message": "注册成功",
		"user":    req.Username,
	}, nil
}

func (l *AuthLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	// TODO: 调用auth.rpc服务进行用户登录
	// TODO: 生成JWT token
	// TODO: 返回用户信息

	return &types.LoginResponse{
		Token: "demo-token-placeholder",
		UserInfo: struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		}{
			ID:       1,
			Username: req.Username,
			Email:    "demo@example.com",
		},
	}, nil
}
