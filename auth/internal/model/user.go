package model

type User struct {
	ID           int64
	Username     string
	PasswordHash string
	Email        string
}
