package models

type User struct {
	ID           int
	Email        string
	PasswordHash string
	UserType     string // "admin" or "broker"
}
