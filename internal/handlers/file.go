// internal/handlers/file.go

package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"MortgageAgent/internal/db"
)

// internal/handlers/file.go

func ServeDocument(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user == nil || user.UserType != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			http.NotFound(w, r)
			return
		}

		log.Printf("Admin ID %d requesting document with path: %s\n", user.ID, filePath)

		// Fetch the document details from the database to verify access
		document, err := db.GetDocumentByPath(database, filePath)
		if err != nil || document == nil {
			log.Printf("Document not found for path: %s, error: %v\n", filePath, err)
			http.NotFound(w, r)
			return
		}

		// Verify that the document belongs to an application assigned to this admin
		app, err := db.GetApplicationByID(database, strconv.Itoa(document.ApplicationID))
		if err != nil || app == nil {
			log.Printf("Application not found for ID: %d, error: %v\n", document.ApplicationID, err)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		if app.AssignedAdminID == nil || *app.AssignedAdminID != user.ID {
			log.Printf("Admin ID %d not authorized to access application ID %d\n", user.ID, app.ID)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Construct the full file path
		fullPath := filepath.Join("", document.FilePath)
		log.Printf("Serving file: %s to admin ID %d\n", fullPath, user.ID)

		// Check if the file exists
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			log.Printf("File does not exist: %s\n", fullPath)
			http.NotFound(w, r)
			return
		}

		// Serve the file
		http.ServeFile(w, r, fullPath)
	}
}
