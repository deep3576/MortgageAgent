package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"

	"MortgageAgent/internal/db"
)

func LoginPage(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		tmpl := template.Must(template.ParseFiles("internal/templates/login.html"))
		tmpl.Execute(w, nil)
	}
}

func Login(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		email := r.FormValue("email")
		password := r.FormValue("password")

		user, err := db.GetUserByEmail(database, email)
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Set cookie-based session
		// In production, you might use a secure session store
		cookie := &http.Cookie{
			Name:     "session_email",
			Value:    user.Email,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)

		// Redirect based on user type
		if user.UserType == "admin" {
			http.Redirect(w, r, "/admin", http.StatusFound)
		} else {
			http.Redirect(w, r, "/broker", http.StatusFound)
		}
	}
}

func SignUpPage(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		tmpl := template.Must(template.ParseFiles("internal/templates/signup.html"))
		tmpl.Execute(w, nil)
	}
}

func Register(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/signup", http.StatusFound)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")

		err := db.CreateUser(database, email, password)
		if err != nil {
			http.Error(w, "Error creating user: "+err.Error(), http.StatusBadRequest)
			return
		}

		http.Redirect(w, r, "/", http.StatusFound)
	}
}

func Logout() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Clear the session cookie
		cookie := &http.Cookie{
			Name:    "session_email",
			Value:   "",
			Expires: time.Unix(0, 0),
			MaxAge:  -1,
		}
		http.SetCookie(w, cookie)
		http.Redirect(w, r, "/", http.StatusFound)
	}
}
