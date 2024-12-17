package handlers

import (
	"database/sql"
	"html/template"
	"net/http"
	"time"

	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"MortgageAgent/internal/db"
)

var emailRegex = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

// Canadian postal code pattern: letter-digit-letter space digit-letter-digit
// Ref: https://en.wikipedia.org/wiki/Postal_codes_in_Canada
// This regex is somewhat simplified:
// [ABCEGHJ-NPRSTVXY]\d[ABCEGHJ-NPRSTV-Z]\s?\d[ABCEGHJ-NPRSTV-Z]\d
var postalCodeRegex = regexp.MustCompile(`^[ABCEGHJ-NPRSTVXY]\d[ABCEGHJ-NPRSTV-Z]\s?\d[ABCEGHJ-NPRSTV-Z]\d$`)

// ValidateEmail checks if the email looks valid
func ValidateEmail(email string) bool {
	return emailRegex.MatchString(strings.ToLower(email))
}

// ValidateCanadianPostalCode checks if the postal code matches Canadian format
func ValidateCanadianPostalCode(code string) bool {
	return postalCodeRegex.MatchString(strings.ToUpper(strings.ReplaceAll(code, " ", "")))
}

type LoginPageData struct {
	ErrorMessage string
}

func LoginPage(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" || r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		data := LoginPageData{ErrorMessage: ""}
		tmpl := template.Must(template.ParseFiles("internal/templates/login.html"))
		tmpl.Execute(w, data)
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
			// User not found or DB error. Show error on same page.
			renderLoginWithError(w, "Invalid credentials.")
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
		if err != nil {
			// Password mismatch. Show error on same page.
			renderLoginWithError(w, "Invalid credentials. Please try again.")
			return
		}

		// Successful login
		cookie := &http.Cookie{
			Name:     "session_email",
			Value:    user.Email,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		}
		http.SetCookie(w, cookie)

		if user.UserType == "admin" {
			http.Redirect(w, r, "/admin", http.StatusFound)
		} else {
			http.Redirect(w, r, "/broker", http.StatusFound)
		}
	}
}

func renderLoginWithError(w http.ResponseWriter, errorMsg string) {
	data := LoginPageData{ErrorMessage: errorMsg}
	tmpl := template.Must(template.ParseFiles("internal/templates/login.html"))
	tmpl.Execute(w, data)
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

		firstName := r.FormValue("first_name")
		lastName := r.FormValue("last_name")
		email := r.FormValue("email")
		phone := r.FormValue("phone")
		postalCode := r.FormValue("postal_code")
		password := r.FormValue("password")

		// Basic validations
		if firstName == "" || lastName == "" || email == "" || password == "" {
			http.Error(w, "Missing required fields", http.StatusBadRequest)
			return
		}

		if !ValidateEmail(email) {
			http.Error(w, "Invalid email format", http.StatusBadRequest)
			return
		}

		if postalCode != "" && !ValidateCanadianPostalCode(postalCode) {
			http.Error(w, "Invalid Canadian postal code format", http.StatusBadRequest)
			return
		}

		err := db.CreateUser(database, firstName, lastName, email, phone, postalCode, password)
		if err != nil {
			http.Error(w, "Error creating user: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Redirect to signup success page
		http.Redirect(w, r, "/signup-success", http.StatusFound)
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
