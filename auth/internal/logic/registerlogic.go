package logic

import (
	"context"

	"auth/auth"
	"auth/internal/svc"
	"auth/jwt"

	"github.com/zeromicro/go-zero/core/logx"
)

type RegisterLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewRegisterLogic(ctx context.Context, svcCtx *svc.ServiceContext) *RegisterLogic {
	return &RegisterLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *RegisterLogic) Register(in *auth.RegisterRequest) (*auth.LoginResponse, error) {
	result, err := l.svcCtx.DB.Exec("INSERT INTO users(username,password,email)VALUES(?,?,?)", in.Username, in.Password, in.Email)
	// TODO:bcrypt password
	if err != nil {
		return nil, err
	}

	userID, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	token, err := jwt.GenerateToken(userID, in.Username, l.svcCtx.Config.Auth.AccessSecret, l.svcCtx.Config.Auth.AccessExpire)
	if err != nil {
		return nil, err
	}

	return &auth.LoginResponse{
		Token: token,
		UserInfo: &auth.UserInfo{
			Id:       userID,
			Username: in.Username,
			Email:    in.Email,
		},
	}, nil
}
