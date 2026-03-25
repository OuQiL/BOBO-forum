package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"api-gateway/internal/contextkey"
	"api-gateway/internal/svc"

	"github.com/golang-jwt/jwt/v4"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var (
	ErrMissingToken       = errors.New("missing authorization token")
	ErrInvalidTokenFormat = errors.New("invalid authorization header format")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token has expired")
)

type JWTClaims struct {
	UserID   int64  `json:"user_id"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func JWTAuthMiddleware(svcCtx *svc.ServiceContext) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				httpx.ErrorCtx(r.Context(), w, ErrMissingToken)
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				httpx.ErrorCtx(r.Context(), w, ErrInvalidTokenFormat)
				return
			}

			tokenString := parts[1]
			claims, err := ParseToken(tokenString, svcCtx.Config.JWT.Secret)
			if err != nil {
				if strings.Contains(err.Error(), "token is expired") {
					if svcCtx.Config.JWT.PrevSecret != "" {
						claims, err = ParseToken(tokenString, svcCtx.Config.JWT.PrevSecret)
						if err != nil {
							httpx.ErrorCtx(r.Context(), w, ErrTokenExpired)
							return
						}
					} else {
						httpx.ErrorCtx(r.Context(), w, ErrTokenExpired)
						return
					}
				} else {
					httpx.ErrorCtx(r.Context(), w, ErrInvalidToken)
					return
				}
			}

			ctx := r.Context()
			ctx = context.WithValue(ctx, contextkey.GetUserIDKey(), claims.UserID)
			ctx = context.WithValue(ctx, contextkey.GetUsernameKey(), claims.Username)

			r = r.WithContext(ctx)

			next(w, r)
		}
	}
}

func GenerateToken(userID int64, username string, secret string, expireSeconds int64) (string, error) {
	now := time.Now()
	expire := now.Add(time.Duration(expireSeconds) * time.Second)

	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expire),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString string, secret string) (*JWTClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*JWTClaims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(contextkey.GetUserIDKey()).(int64)
	return userID, ok
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(contextkey.GetUsernameKey()).(string)
	return username, ok
}
