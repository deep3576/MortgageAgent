package handlers

import (
	"database/sql"
	"net/http"
	"path/filepath"
	"strconv"

	"MortgageAgent/internal/db"
)

// ServeDocument securely serves uploaded documents to authorized admins
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

		// Fetch the document details from the database to verify access
		document, err := db.GetDocumentByPath(database, filePath)
		if err != nil || document == nil {
			http.NotFound(w, r)
			return
		}

		// Verify that the document belongs to an application assigned to this admin
		app, err := db.GetApplicationByID(database, strconv.Itoa(document.ApplicationID))
		if err != nil || app == nil || (app.AssignedAdminID != nil && *app.AssignedAdminID != user.ID) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Serve the file
		fullPath := filepath.Join("uploads", document.FilePath)
		http.ServeFile(w, r, fullPath)
	}
}
