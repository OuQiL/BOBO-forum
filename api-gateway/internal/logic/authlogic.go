package logic

import (
	"context"

	"api-gateway/internal/svc"
	"api-gateway/internal/types"

	"github.com/zeromicro/go-zero/core/logx"

	auth "auth/pkg/client/auth"
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
	res, err := l.svcCtx.AuthRpc.Register(l.ctx, &auth.RegisterRequest{
		Username: req.Username,
		Password: req.Password,
		Email:    req.Email,
	})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"message": "注册成功",
		"user_id": res.UserInfo.Id,
		"token":   res.Token,
	}, nil
}

func (l *AuthLogic) Login(req *types.LoginRequest) (resp *types.LoginResponse, err error) {
	res, err := l.svcCtx.AuthRpc.Login(l.ctx, &auth.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	})
	if err != nil {
		return nil, err
	}

	return &types.LoginResponse{
		Token: res.Token,
		UserInfo: struct {
			ID       int64  `json:"id"`
			Username string `json:"username"`
			Email    string `json:"email"`
		}{
			ID:       res.UserInfo.Id,
			Username: res.UserInfo.Username,
			Email:    res.UserInfo.Email,
		},
	}, nil
}
