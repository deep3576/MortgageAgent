package handlers

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"time"

	"MortgageAgent/internal/db"

	"MortgageAgent/internal/models"
)

type ApplicationWithDocs struct {
	ID              int
	BrokerID        int
	ApplicationType string
	Documents       []DocumentInfo
}

type DocumentInfo struct {
	Category string
	FilePath string
}

func convertDocs(docs []db.Document) []DocumentInfo {
	var res []DocumentInfo
	for _, d := range docs {
		res = append(res, DocumentInfo{
			Category: d.Category,
			FilePath: d.FilePath,
		})
	}
	return res
}

// Additional handler to view a single application if needed:
// internal/handlers/admin.go

type ViewApplicationData struct {
	ID              int
	BrokerID        int
	ApplicationType string
	CreatedAt       time.Time
	Documents       []models.DocumentInfo
}

// internal/handlers/admin.go

func ViewApplication(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the current admin user from the context
		user := GetUserFromContext(r)
		if user == nil || user.UserType != "admin" {
			log.Printf("Unauthorized access attempt by user: %v\n", user)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Get application ID from query parameters
		idStr := r.URL.Query().Get("id")
		if idStr == "" {
			log.Println("Missing application ID in request")
			http.Error(w, "Missing application ID", http.StatusBadRequest)
			return
		}

		appID, err := strconv.Atoi(idStr)
		if err != nil {
			log.Printf("Invalid application ID format: %s, error: %v\n", idStr, err)
			http.Error(w, "Invalid application ID", http.StatusBadRequest)
			return
		}

		log.Printf("Admin ID %d is viewing application ID %d\n", user.ID, appID)

		// Fetch application details from the database
		app, err := db.GetApplicationByID(database, idStr)
		if err != nil {
			log.Printf("Error fetching application ID %d: %v\n", appID, err)
			http.Error(w, "Application not found", http.StatusNotFound)
			return
		}

		if app.AssignedAdminID == nil || *app.AssignedAdminID != user.ID {
			log.Printf("Admin ID %d not authorized to view application ID %d\n", user.ID, app.ID)
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		// Fetch documents associated with the application
		documents, err := db.GetDocumentsForApplication(database, app.ID)
		if err != nil {
			log.Printf("Error fetching documents for application ID %d: %v\n", app.ID, err)
			http.Error(w, "Error fetching documents", http.StatusInternalServerError)
			return
		}

		// Map documents to the view data
		var documentInfos []models.DocumentInfo
		for _, doc := range documents {
			documentInfos = append(documentInfos, models.DocumentInfo{
				Category: doc.Category,
				FilePath: doc.FilePath,
			})
		}

		// Prepare data for the template
		data := ViewApplicationData{
			ID:              app.ID,
			BrokerID:        app.BrokerID,
			ApplicationType: app.ApplicationType,
			CreatedAt:       app.CreatedAt,
			Documents:       documentInfos,
		}

		// Render the view_application template
		tmpl, err := template.ParseFiles("internal/templates/view_application.html")
		if err != nil {
			log.Printf("Error parsing template: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		err = tmpl.Execute(w, data)
		if err != nil {
			log.Printf("Error executing template: %v\n", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		log.Printf("Successfully rendered view for application ID %d\n", app.ID)
	}
}

type AdminDashboardData struct {
	ErrorMessage string
	Applications []models.ApplicationWithDocuments
}

// internal/handlers/admin.go

func AdminDashboard(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Retrieve the current admin user from the context
		user := GetUserFromContext(r)
		if user == nil || user.UserType != "admin" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Fetch applications assigned to this admin
		applications, err := db.GetApplicationsForAdmin(database, user.ID)
		if err != nil {
			log.Println("Error fetching applications:", err)
			data := AdminDashboardData{
				ErrorMessage: "Error fetching applications. Please try again later.",
				Applications: nil,
			}
			tmpl := template.Must(template.ParseFiles("internal/templates/admin_dashboard.html"))
			tmpl.Execute(w, data)
			return
		}

		// Check if any applications are assigned
		if len(applications) == 0 {
			log.Printf("No applications assigned to admin ID %d\n", user.ID)
		} else {
			log.Printf("Admin ID %d has %d applications\n", user.ID, len(applications))
		}

		// Prepare data for the template
		data := AdminDashboardData{
			ErrorMessage: "",
			Applications: applications,
		}

		// Render the admin dashboard template
		tmpl := template.Must(template.ParseFiles("internal/templates/admin_dashboard.html"))
		err = tmpl.Execute(w, data)
		if err != nil {
			log.Println("Error rendering template:", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
	}
}
