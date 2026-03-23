package logic

import (
	"context"

	"auth/auth"
	"auth/internal/svc"
	"auth/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *auth.LoginRequest) (*auth.LoginResponse, error) {
	type user struct {
		ID    int64
		Email string
	}

	var u user
	err := l.svcCtx.DB.QueryRow(&u, "SELECT id, email FROM users WHERE username=? AND password=?", in.Username, in.Password)
	if err != nil {
		return nil, err
	}

	token, err := jwt.GenerateToken(u.ID, in.Username, l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResponse{
		Token: token,
		UserInfo: &auth.UserInfo{
			Id:       u.ID,
			Username: in.Username,
			Email:    u.Email,
		},
	}, nil
}
