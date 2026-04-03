package contextkey

type contextKey string

const (
	UserIDKey   contextKey = "userID"
	UsernameKey contextKey = "username"
)

func GetUserIDKey() contextKey {
	return UserIDKey
}

func GetUsernameKey() contextKey {
	return UsernameKey
}
