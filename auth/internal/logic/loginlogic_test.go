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

type MockUserRepository struct {
	CreateFunc                    func(ctx context.Context, user *model.User) (int64, error)
	FindByUsernameAndPasswordFunc func(ctx context.Context, username, password string) (*model.User, error)
	FindByIDFunc                  func(ctx context.Context, id int64) (*model.User, error)
}

func (m *MockUserRepository) Create(ctx context.Context, user *model.User) (int64, error) {
	if m.CreateFunc != nil {
		return m.CreateFunc(ctx, user)
	}
	return 0, errors.New("CreateFunc not implemented")
}

func (m *MockUserRepository) FindByUsernameAndPassword(ctx context.Context, username, password string) (*model.User, error) {
	if m.FindByUsernameAndPasswordFunc != nil {
		return m.FindByUsernameAndPasswordFunc(ctx, username, password)
	}
	return nil, errors.New("FindByUsernameAndPasswordFunc not implemented")
}

func (m *MockUserRepository) FindByID(ctx context.Context, id int64) (*model.User, error) {
	if m.FindByIDFunc != nil {
		return m.FindByIDFunc(ctx, id)
	}
	return nil, errors.New("FindByIDFunc not implemented")
}

func TestLoginLogic_Login_Success(t *testing.T) {
	mockRepo := &MockUserRepository{
		FindByUsernameAndPasswordFunc: func(ctx context.Context, username, password string) (*model.User, error) {
			if username == "testuser" && password == "testpass" {
				return &model.User{
					ID:           1,
					Username:     "testuser",
					PasswordHash: "testpass",
					Email:        "test@example.com",
				}, nil
			}
			return nil, nil
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

	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&proto.LoginRequest{
		Username: "testuser",
		Password: "testpass",
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.NotEmpty(t, resp.Token)
	assert.Equal(t, int64(1), resp.UserInfo.Id)
	assert.Equal(t, "testuser", resp.UserInfo.Username)
	assert.Equal(t, "test@example.com", resp.UserInfo.Email)
}

func TestLoginLogic_Login_InvalidCredentials(t *testing.T) {
	mockRepo := &MockUserRepository{
		FindByUsernameAndPasswordFunc: func(ctx context.Context, username, password string) (*model.User, error) {
			return nil, nil
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

	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&proto.LoginRequest{
		Username: "wronguser",
		Password: "wrongpass",
	})

	assert.Error(t, err)
	assert.Equal(t, "invalid username or password", err.Error())
	assert.Nil(t, resp)
}

func TestLoginLogic_Login_DatabaseError(t *testing.T) {
	mockRepo := &MockUserRepository{
		FindByUsernameAndPasswordFunc: func(ctx context.Context, username, password string) (*model.User, error) {
			return nil, errors.New("database connection error")
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

	logic := NewLoginLogic(context.Background(), svcCtx)
	resp, err := logic.Login(&proto.LoginRequest{
		Username: "testuser",
		Password: "testpass",
	})

	assert.Error(t, err)
	assert.Equal(t, "database connection error", err.Error())
	assert.Nil(t, resp)
}
