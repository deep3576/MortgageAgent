package handlers

import (
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"MortgageAgent/internal/db"
)

func StartApplication(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/broker", http.StatusFound)
			return
		}

		user := GetUserFromContext(r)
		if user == nil || user.UserType != "broker" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		appType := r.FormValue("application_type")
		if appType != "self" && appType != "someone_else" {
			http.Error(w, "Invalid application type", http.StatusBadRequest)
			return
		}

		appID, err := db.CreateApplication(database, user.ID, appType)
		if err != nil {
			http.Error(w, "Could not create application", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/application-form?id="+strconv.Itoa(appID), http.StatusFound)

	}
}

func ApplicationFormPage(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := GetUserFromContext(r)
		if user == nil || user.UserType != "broker" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			app, err := db.GetApplicationByID(database, id)
			if err != nil || app == nil {
				http.Error(w, "Application not found", http.StatusNotFound)
				return
			}
			if app.BrokerID != user.ID {
				http.Error(w, "Unauthorized to view this application", http.StatusForbidden)
				return
			}

			data := struct {
				ApplicationID string
			}{
				ApplicationID: id,
			}

			tmpl := template.Must(template.ParseFiles("internal/templates/application_form.html"))
			tmpl.Execute(w, data)

		} else if r.Method == http.MethodPost {
			appID := r.FormValue("application_id")
			app, err := db.GetApplicationByID(database, appID)
			if err != nil || app == nil {
				http.Error(w, "Application not found", http.StatusNotFound)
				return
			}
			if app.BrokerID != user.ID {
				http.Error(w, "Unauthorized", http.StatusForbidden)
				return
			}

			// Categories to upload
			categories := []string{
				"Proof_of_income",
				"Identification",
				"Basic_financial_information",
				"Down_payment_confirmation",
				"Property_details",
			}

			// Directory to store uploads (ensure this directory exists and is writable)
			uploadDir := "uploads"
			os.MkdirAll(uploadDir, 0777)

			for _, cat := range categories {
				processFile(cat, r, uploadDir, database, w, appID)
			}
			// for _, cat := range categories {

			// 	fmt.Println("category ::" + cat)

			// 	file, header, err := r.FormFile(cat)

			// 	fmt.Println(header)

			// 	if err == nil && header != nil {
			// 		defer file.Close()

			// 		// Create a unique file path
			// 		filePath := filepath.Join(uploadDir, header.Filename)
			// 		out, err := os.Create(filePath)
			// 		fmt.Println("output " + out.Name())

			// 		if err != nil {
			// 			print(err)
			// 			http.Error(w, "File saving error", http.StatusInternalServerError)
			// 			return
			// 		}
			// 		defer out.Close()
			// 		_, err = io.Copy(out, file)
			// 		if err != nil {
			// 			print(err)
			// 			http.Error(w, "File saving error", http.StatusInternalServerError)
			// 			return
			// 		}

			// 		// Add document record in DB
			// 		err = db.AddDocument(database, app.ID, cat, filePath)
			// 		if err != nil {
			// 			print(err)
			// 			http.Error(w, "Error recording document", http.StatusInternalServerError)
			// 			return
			// 		}
			// 	}
			// }

			// Assign application to admin (round robin)

			adminID, err := db.AssignApplicationToAdmin(database, int64(app.ID))
			if err != nil {
				http.Error(w, "Failed to assign admin", http.StatusInternalServerError)
				return
			}

			fmt.Printf("Application %s assigned to admin %d\n", appID, adminID)

			http.Redirect(w, r, "/broker?submitted=true", http.StatusFound)

		} else {
			http.NotFound(w, r)
		}
	}
}

func processFile(cat string, r *http.Request, uploadDir string, database *sql.DB, w http.ResponseWriter, appID string) {
	fmt.Println("category ::" + cat)

	file, header, err := r.FormFile(cat)

	fmt.Println(header)

	if err == nil && header != nil {
		defer file.Close()

		// Create a unique file path
		filePath := filepath.Join(uploadDir, header.Filename)
		out, err := os.Create(filePath)
		fmt.Println("output " + out.Name())

		if err != nil {
			print(err)
			http.Error(w, "File saving error", http.StatusInternalServerError)
			return
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			print(err)
			http.Error(w, "File saving error", http.StatusInternalServerError)
			return
		}

		// Add document record in DB
		app, err := db.GetApplicationByID(database, appID)
		err = db.AddDocument(database, app.ID, cat, filePath)
		if err != nil {
			print(err)
			http.Error(w, "Error recording document", http.StatusInternalServerError)
			return
		}
	}
}
