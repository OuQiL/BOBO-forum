package logic

import (
	"context"

	"auth/api/proto"
	"auth/internal/model"
	"auth/internal/pkg/jwt"
	"auth/internal/svc"

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

func (l *RegisterLogic) Register(in *proto.RegisterRequest) (*proto.LoginResponse, error) {
	user := &model.User{
		Username:     in.Username,
		PasswordHash: in.Password,
		Email:        in.Email,
	}

	userID, err := l.svcCtx.UserRepo.Create(l.ctx, user)
	if err != nil {
		return nil, err
	}

	token, err := jwt.GenerateToken(userID, in.Username, l.svcCtx.Config.JwtAuth.AccessSecret, l.svcCtx.Config.JwtAuth.AccessExpire)
	if err != nil {
		return nil, err
	}

	return &proto.LoginResponse{
		Token: token,
		UserInfo: &proto.UserInfo{
			Id:       userID,
			Username: in.Username,
			Email:    in.Email,
		},
	}, nil
}
