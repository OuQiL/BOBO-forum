package logic

import (
	"context"
	"errors"
	"testing"

	"auth/api/proto"
	"auth/internal/config"
	"auth/internal/model"
	"auth/internal/svc"

	"github.com/stretchr/testify/assert"
)

func TestRegisterLogic_Register_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		CreateFunc: func(ctx context.Context, user *model.User) (int64, error) {
			return 1, nil
		},
	}

	svcCtx := &svc.ServiceContext{
		Config: config.Config{
			JwtAuth: config.AuthConfig{
				AccessSecret: "test-secret-key-for-unit-testing",
				AccessExpire: 3600,
			},
		},
		UserRepo: mockRepo,
	}

	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&proto.RegisterRequest{
		Username: "newuser",
		Password: "newpass",
		Email:    "newuser@example.com",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, int64(1), resp.UserInfo.Id)
	assert.Equal(t, "newuser", resp.UserInfo.Username)
	assert.Equal(t, "newuser@example.com", resp.UserInfo.Email)
}

func TestRegisterLogic_Register_DatabaseError(t *testing.T) {
	mockRepo := &MockUserRepository{
		CreateFunc: func(ctx context.Context, user *model.User) (int64, error) {
			return 0, errors.New("duplicate entry for username")
		},
	}

	svcCtx := &svc.ServiceContext{
		Config: config.Config{
			JwtAuth: config.AuthConfig{
				AccessSecret: "test-secret-key-for-unit-testing",
				AccessExpire: 3600,
			},
		},
		UserRepo: mockRepo,
	}

	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&proto.RegisterRequest{
		Username: "existinguser",
		Password: "testpass",
		Email:    "existing@example.com",
	})

	assert.Error(t, err)
	assert.Equal(t, "duplicate entry for username", err.Error())
	assert.Nil(t, resp)
}

func TestRegisterLogic_Register_ConnectionError(t *testing.T) {
	mockRepo := &MockUserRepository{
		CreateFunc: func(ctx context.Context, user *model.User) (int64, error) {
			return 0, errors.New("database connection failed")
		},
	}

	svcCtx := &svc.ServiceContext{
		Config: config.Config{
			JwtAuth: config.AuthConfig{
				AccessSecret: "test-secret-key-for-unit-testing",
				AccessExpire: 3600,
			},
		},
		UserRepo: mockRepo,
	}

	logic := NewRegisterLogic(context.Background(), svcCtx)
	resp, err := logic.Register(&proto.RegisterRequest{
		Username: "testuser",
		Password: "testpass",
		Email:    "test@example.com",
	})

	assert.Error(t, err)
	assert.Equal(t, "database connection failed", err.Error())
	assert.Nil(t, resp)
}
