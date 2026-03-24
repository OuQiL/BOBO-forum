package logic

import (
	"context"
	"errors"

	"auth/api/proto"
	"auth/internal/model"
	"auth/internal/pkg/jwt"
	"auth/internal/svc"

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

func (l *LoginLogic) Login(in *proto.LoginRequest) (*proto.LoginResponse, error) {
	user, err := l.svcCtx.UserRepo.FindByUsernameAndPassword(l.ctx, in.Username, in.Password)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	token, err := jwt.GenerateToken(user.ID, user.Username, l.svcCtx.Config.JwtAuth.AccessSecret, l.svcCtx.Config.JwtAuth.AccessExpire)
	if err != nil {
		return nil, err
	}

	return &proto.LoginResponse{
		Token: token,
		UserInfo: &proto.UserInfo{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}

func (l *LoginLogic) LoginWithUser(user *model.User) (*proto.LoginResponse, error) {
	token, err := jwt.GenerateToken(user.ID, user.Username, l.svcCtx.Config.JwtAuth.AccessSecret, l.svcCtx.Config.JwtAuth.AccessExpire)
	if err != nil {
		return nil, err
	}

	return &proto.LoginResponse{
		Token: token,
		UserInfo: &proto.UserInfo{
			Id:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	}, nil
}
