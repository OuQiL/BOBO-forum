package logic

import "errors"

var ErrUnauthorized = errors.New("unauthorized: missing or invalid token")
