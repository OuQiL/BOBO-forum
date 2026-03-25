package logic

import (
	"context"

	"api-gateway/internal/contextkey"
)

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	userID, ok := ctx.Value(contextkey.GetUserIDKey()).(int64)
	return userID, ok
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(contextkey.GetUsernameKey()).(string)
	return username, ok
}
