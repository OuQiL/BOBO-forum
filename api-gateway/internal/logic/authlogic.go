package logic

import (
	"context"

	"api-gateway/internal/middleware"
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
	// TODO: 调用 auth.rpc 服务进行用户注册
	// res, err := l.svcCtx.AuthRpc.Register(l.ctx, &auth.RegisterRequest{...})

	// Demo: 本地生成 token
	token, err := middleware.GenerateToken(
		1,
		req.Username,
		l.svcCtx.Config.JWT.Secret,
		l.svcCtx.Config.JWT.Expire,
	)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message":  "注册成功",
		"user_id":  1,
		"username": req.Username,
		"token":    token,
	}, nil
}

func (l *AuthLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	// TODO: 调用 auth.rpc 服务进行用户登录验证
	// res, err := l.svcCtx.AuthRpc.Login(l.ctx, &auth.LoginRequest{...})

	// Demo: 本地生成 token
	token, err := middleware.GenerateToken(
		1,
		req.Username,
		l.svcCtx.Config.JWT.Secret,
		l.svcCtx.Config.JWT.Expire,
	)
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		Token: token,
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
