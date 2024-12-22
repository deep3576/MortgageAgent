package db

import (
	"MortgageAgent/internal/models"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	_ "modernc.org/sqlite"
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
        user_type TEXT NOT NULL,reset_token TEXT
		,
		reset_token_expires_at DATETIME,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    );

	CREATE TABLE IF NOT EXISTS applications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    broker_id INTEGER NOT NULL,
    application_type TEXT NOT NULL,           -- "self" or "someone_else"
    assigned_admin_id INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (broker_id) REFERENCES users(id),
    FOREIGN KEY (assigned_admin_id) REFERENCES users(id)
	);

	CREATE TABLE IF NOT EXISTS documents (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    application_id INTEGER NOT NULL,
    category TEXT NOT NULL,
    file_path TEXT NOT NULL,
    uploaded_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (application_id) REFERENCES applications(id)
);

-- Optional: a settings table to keep track of last assigned admin, etc.
	CREATE TABLE IF NOT EXISTS settings (
    key TEXT PRIMARY KEY,
    value TEXT
);
`
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

func SetResetToken(db *sql.DB, email, token string, expires time.Time) error {
	_, err := db.Exec("UPDATE users SET reset_token=?, reset_token_expires_at=? WHERE email=?", token, expires, email)
	return err
}

func GetUserByResetToken(db *sql.DB, token string) (*models.User, error) {
	u := &models.User{}
	row := db.QueryRow("SELECT id, first_name, last_name, email, password_hash, phone, postal_code, user_type, reset_token_expires_at FROM users WHERE reset_token=?", token)
	var expiresAt time.Time
	err := row.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.PasswordHash, &u.Phone, &u.PostalCode, &u.UserType, &expiresAt)
	if err != nil {
		return nil, err
	}
	if time.Now().After(expiresAt) {
		return nil, sql.ErrNoRows // token expired
	}
	return u, nil
}

func UpdateUserPassword(db *sql.DB, userID int, newHash string) error {
	_, err := db.Exec("UPDATE users SET password_hash=?, reset_token=NULL, reset_token_expires_at=NULL WHERE id=?", newHash, userID)
	return err
}

type Document struct {
	ID            int
	ApplicationID int
	Category      string
	FilePath      string
	UploadedAt    string
}

func AddDocument(db *sql.DB, applicationID int, category, filePath string) error {
	print("upload File path for " + category + filePath)
	_, err := db.Exec("INSERT INTO documents (application_id, category, file_path, uploaded_at) VALUES (?, ?, ?, ?)",
		applicationID, category, filePath, time.Now())
	return err
}

// Round-robin assignment logic
// internal/db/db.go

func AssignApplicationToAdmin(db *sql.DB, applicationID int) (int, error) {
	// Fetch all admin IDs ordered by ID
	rows, err := db.Query("SELECT id FROM users WHERE user_type='admin' ORDER BY id ASC")
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	var adminIDs []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return 0, err
		}
		adminIDs = append(adminIDs, id)
	}

	if len(adminIDs) == 0 {
		return 0, errors.New("no admins available for assignment")
	}

	// Fetch the last assigned admin ID from settings
	var lastAssignedAdminID sql.NullInt64
	err = db.QueryRow("SELECT value FROM settings WHERE key='last_assigned_admin_id'").Scan(&lastAssignedAdminID)
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	// Determine the next admin ID
	var nextAdminID int
	if lastAssignedAdminID.Valid {
		lastID := int(lastAssignedAdminID.Int64)
		index := -1
		for i, id := range adminIDs {
			if id == lastID {
				index = i
				break
			}
		}
		if index != -1 && index < len(adminIDs)-1 {
			nextAdminID = adminIDs[index+1]
		} else {
			nextAdminID = adminIDs[0] // Wrap around
		}
	} else {
		nextAdminID = adminIDs[0] // Start with the first admin
	}

	// Assign the application to the next admin
	_, err = db.Exec("UPDATE applications SET assigned_admin_id=? WHERE id=?", nextAdminID, applicationID)
	if err != nil {
		return 0, err
	}

	// Update the last assigned admin ID in settings
	_, err = db.Exec(`
        INSERT INTO settings (key, value) 
        VALUES ('last_assigned_admin_id', ?) 
        ON CONFLICT(key) DO UPDATE SET value=excluded.value
    `, nextAdminID)
	if err != nil {
		return 0, err
	}

	return nextAdminID, nil
}

func GetApplicationByID(db *sql.DB, id string) (*models.Application, error) {
	a := &models.Application{}
	row := db.QueryRow("SELECT id, broker_id, application_type, assigned_admin_id, created_at FROM applications WHERE id=?", id)

	var assignedAdminID sql.NullInt64
	err := row.Scan(&a.ID, &a.BrokerID, &a.ApplicationType, &assignedAdminID, &a.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			// No application found with given ID
			return nil, nil
		}
		return nil, err
	}

	if assignedAdminID.Valid {
		val := int(assignedAdminID.Int64)
		a.AssignedAdminID = &val
	} else {
		a.AssignedAdminID = nil
	}

	return a, nil
}
func CreateApplication(db *sql.DB, brokerID int, appType string) (int, error) {
	res, err := db.Exec("INSERT INTO applications (broker_id, application_type, created_at) VALUES (?, ?, ?)",
		brokerID, appType, time.Now())
	if err != nil {
		return 0, err
	}

	lastID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return int(lastID), nil
}

// GetApplicationsForAdmin fetches all applications assigned to a specific admin.
// internal/db/db.go

func GetApplicationsForAdmin(db *sql.DB, adminID int) ([]models.ApplicationWithDocuments, error) {
	query := `
        SELECT id, broker_id, application_type, created_at
        FROM applications
        WHERE assigned_admin_id = ?
        ORDER BY created_at DESC
    `
	rows, err := db.Query(query, adminID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var applications []models.ApplicationWithDocuments

	for rows.Next() {
		var app models.ApplicationWithDocuments
		err := rows.Scan(&app.ID, &app.BrokerID, &app.ApplicationType, &app.CreatedAt)
		if err != nil {
			return nil, err
		}

		// Fetch associated documents for each application
		docs, err := GetDocumentsForApplication(db, app.ID)
		if err != nil {
			return nil, err
		}

		// Map []Document to []models.DocumentInfo
		var documentInfos []models.DocumentInfo
		for _, d := range docs {
			documentInfos = append(documentInfos, models.DocumentInfo{
				Category: d.Category,
				FilePath: d.FilePath,
			})
		}
		app.Documents = documentInfos

		applications = append(applications, app)
	}

	return applications, nil
}

// GetDocumentsForApplication fetches all documents for a given application.
func GetDocumentsForApplication(db *sql.DB, applicationID int) ([]Document, error) {
	query := `
        SELECT id, application_id, category, file_path, uploaded_at
        FROM documents
        WHERE application_id = ?
    `
	rows, err := db.Query(query, applicationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var documents []Document

	for rows.Next() {
		var doc Document
		err := rows.Scan(&doc.ID, &doc.ApplicationID, &doc.Category, &doc.FilePath, &doc.UploadedAt)
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}

	fmt.Printf("Fetched %d documents for application ID %d\n", len(documents), applicationID)
	return documents, nil
}

// GetDocumentByPath fetches a document by its file path.
// internal/db/db.go

func GetDocumentByPath(db *sql.DB, filePath string) (*Document, error) {
    var doc Document
    query := `
        SELECT id, application_id, category, file_path, uploaded_at
        FROM documents
        WHERE file_path = ?
    `
    row := db.QueryRow(query, filePath)
    err := row.Scan(&doc.ID, &doc.ApplicationID, &doc.Category, &doc.FilePath, &doc.UploadedAt)
    if err != nil {
        return nil, err
    }
    return &doc, nil
}

