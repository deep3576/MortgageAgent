package db

import (
	"database/sql"
	"errors"
	"strings"

	_ "github.com/go-sql-driver/mysql"

	"MortgageAgent/internal/models"

	_ "modernc.org/sqlite"
	"golang.org/x/crypto/bcrypt"
)

func InitDB(dsn string) (*sql.DB, error) {
	// For SQLite, DSN is typically just a file name
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func MigrateDB(db *sql.DB) error {
	query := `
    CREATE TABLE IF NOT EXISTS users (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        first_name TEXT,
        last_name TEXT,
        email TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL,
        phone TEXT,
        postal_code TEXT,
        user_type TEXT NOT NULL,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );`
	_, err := db.Exec(query)
	return err
}

func SeedAdminUser(db *sql.DB) error {
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE user_type='admin'").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		pwHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = db.Exec("INSERT INTO users (first_name, last_name, email, password_hash, phone, postal_code, user_type) VALUES (?, ?, ?, ?,?,?, ?)",
			"Admin", "User", "admin@company.com", string(pwHash), "555-666-7777", "N3H0C3", "admin")
		return err
	}

	return nil
}

// GetUserByEmail fetches a user by email
func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	email = strings.TrimSpace(email)

	if err := db.Ping(); err != nil {
		println("Debug: DB ping failed:", err)
	}

	u := &models.User{}
	row := db.QueryRow("SELECT id, first_name, last_name, email, password_hash, phone, postal_code, user_type FROM users WHERE email=?", email)
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.Phone, &u.PostalCode, &u.UserType)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// CreateUser creates a broker user
func CreateUser(db *sql.DB, firstName, lastName, email, phone, postalCode, password string) error {
	// Check if user already exists
	_, err := GetUserByEmail(db, email)
	if err == nil {
		return errors.New("user already exists")
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (first_name, last_name, email, password_hash, phone, postal_code, user_type) VALUES (?, ?, ?, ?, ?, ?, ?)",
		firstName, lastName, email, pwHash, phone, postalCode, "broker")
	return err
}
