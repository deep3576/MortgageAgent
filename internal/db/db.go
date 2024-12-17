package db

import (
	"database/sql"
	"errors"

	_ "github.com/go-sql-driver/mysql"

	"MortgageAgent/internal/models"

	"golang.org/x/crypto/bcrypt"
)

func InitDB(dsn string) (*sql.DB, error) {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, err
	}
	if err = db.Ping(); err != nil {
		return nil, err
	}
	return db, nil
}

func MigrateDB(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS users (
		id INT AUTO_INCREMENT PRIMARY KEY,
		email VARCHAR(255) UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		user_type VARCHAR(50) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);`

	_, err := db.Exec(query)
	return err
}

func SeedAdminUser(db *sql.DB) error {
	// Check if admin user already exists
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM users WHERE user_type='admin'").Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		// Create a default admin user
		pwHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		_, err = db.Exec("INSERT INTO users (email, password_hash, user_type) VALUES (?, ?, ?)", "admin@company.com", string(pwHash), "admin")

		if err != nil {
			return err
		}
	}

	return nil
}

// GetUserByEmail fetches a user by email
func GetUserByEmail(db *sql.DB, email string) (*models.User, error) {
	u := &models.User{}
	row := db.QueryRow("SELECT id, email, password_hash, user_type FROM users WHERE email=?", email)
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.UserType)
	if err != nil {
		return nil, err
	}
	return u, nil
}

// CreateUser creates a broker user
func CreateUser(db *sql.DB, email, password string) error {
	// Check if user already exists
	user, _ := GetUserByEmail(db, email)
	if user != nil {
		return errors.New("user already exists")
	}

	pwHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	_, err = db.Exec("INSERT INTO users (email, password_hash, user_type) VALUES (?, ?, ?)", email, pwHash, "broker")
	return err
}
