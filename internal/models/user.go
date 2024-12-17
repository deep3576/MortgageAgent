package models

type User struct {
	ID           int
	FirstName    string
	LastName     string
	Email        string
	PasswordHash string
	Phone        string
	PostalCode   string
	UserType     string
}
