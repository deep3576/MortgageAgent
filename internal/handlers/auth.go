package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"html/template"
	"log"
	"net/http"
	"time"

	"net/smtp"
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

type SignupPageData struct {
	ErrorMessage string
}

type ForgotPasswordData struct {
	ErrorMessage   string
	SuccessMessage string
}

type ResetPasswordData struct {
	ErrorMessage   string
	SuccessMessage string
	Token          string
}

func SignUpPage(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.NotFound(w, r)
			return
		}
		data := SignupPageData{ErrorMessage: ""}
		tmpl := template.Must(template.ParseFiles("internal/templates/signup.html"))
		tmpl.Execute(w, data)
	}
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
			http.Redirect(w, r, "/admin-dashboard", http.StatusFound)
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

		// Basic validation for email/password etc. done previously...
		// Check if user exists
		user, _ := db.GetUserByEmail(database, email)
		if user != nil {
			// User already exists, show error on same page
			data := SignupPageData{ErrorMessage: "User already exists. Please try a different email."}
			tmpl := template.Must(template.ParseFiles("internal/templates/signup.html"))
			tmpl.Execute(w, data)
			return
		}

		// If user doesn't exist, create the user
		err := db.CreateUser(database, firstName, lastName, email, phone, postalCode, password)
		if err != nil {
			// Some other error occurred while creating user
			data := SignupPageData{ErrorMessage: "Error creating user: " + err.Error()}
			tmpl := template.Must(template.ParseFiles("internal/templates/signup.html"))
			tmpl.Execute(w, data)
			return
		}

		// If successful, redirect to signup-success or login
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

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func ForgotPasswordPage(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			data := ForgotPasswordData{}
			tmpl := template.Must(template.ParseFiles("internal/templates/forgot_password.html"))
			tmpl.Execute(w, data)
		} else if r.Method == http.MethodPost {
			firstName := r.FormValue("first_name")
			lastName := r.FormValue("last_name")
			email := strings.TrimSpace(r.FormValue("email"))

			user, err := db.GetUserByEmail(database, email)
			if err != nil || user == nil {
				data := ForgotPasswordData{ErrorMessage: "No user found with the provided details."}
				tmpl := template.Must(template.ParseFiles("internal/templates/forgot_password.html"))
				tmpl.Execute(w, data)
				return
			}

			if user.FirstName != firstName || user.LastName != lastName {
				data := ForgotPasswordData{ErrorMessage: "Provided details do not match any user."}
				tmpl := template.Must(template.ParseFiles("internal/templates/forgot_password.html"))
				tmpl.Execute(w, data)
				return
			}

			token, err := generateResetToken()
			if err != nil {
				log.Println("Error generating reset token:", err)
				data := ForgotPasswordData{ErrorMessage: "Internal server error. Please try again later."}
				tmpl := template.Must(template.ParseFiles("internal/templates/forgot_password.html"))
				tmpl.Execute(w, data)
				return
			}

			err = db.SetResetToken(database, user.Email, token, time.Now().Add(1*time.Hour))
			if err != nil {
				log.Println("Error setting reset token:", err)
				data := ForgotPasswordData{ErrorMessage: "Internal server error. Please try again later."}
				tmpl := template.Must(template.ParseFiles("internal/templates/forgot_password.html"))
				tmpl.Execute(w, data)
				return
			}

			resetLink := "http://localhost:8080/reset-password?token=" + token
			emailBody := "Click the link below to reset your password:\n\n" + resetLink

			err = SendEmail(user.Email, "Password Reset", emailBody)
			if err != nil {
				data := ForgotPasswordData{ErrorMessage: "Failed to send email. Please try again later."}
				tmpl := template.Must(template.ParseFiles("internal/templates/forgot_password.html"))
				tmpl.Execute(w, data)
				return
			}

			data := ForgotPasswordData{SuccessMessage: "A password reset link has been sent to " + user.Email}
			tmpl := template.Must(template.ParseFiles("internal/templates/forgot_password.html"))
			tmpl.Execute(w, data)
		} else {
			http.NotFound(w, r)
		}
	}
}

func ResetPasswordPage(database *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			token := r.URL.Query().Get("token")
			if token == "" {
				http.NotFound(w, r)
				return
			}

			user, err := db.GetUserByResetToken(database, token)
			data := ResetPasswordData{Token: token}
			if err != nil || user == nil {
				data.ErrorMessage = "Invalid or expired reset token."
			}

			tmpl := template.Must(template.ParseFiles("internal/templates/reset_password.html"))
			tmpl.Execute(w, data)

		} else if r.Method == http.MethodPost {
			token := r.FormValue("token")
			newPassword := r.FormValue("new_password")
			confirmPassword := r.FormValue("confirm_password")

			data := ResetPasswordData{Token: token}
			if newPassword != confirmPassword {
				data.ErrorMessage = "Passwords do not match."
				tmpl := template.Must(template.ParseFiles("internal/templates/reset_password.html"))
				tmpl.Execute(w, data)
				return
			}

			user, err := db.GetUserByResetToken(database, token)
			if err != nil || user == nil {
				data.ErrorMessage = "Invalid or expired reset token."
				tmpl := template.Must(template.ParseFiles("internal/templates/reset_password.html"))
				tmpl.Execute(w, data)
				return
			}

			pwHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
			if err != nil {
				data.ErrorMessage = "Internal error. Try again."
				tmpl := template.Must(template.ParseFiles("internal/templates/reset_password.html"))
				tmpl.Execute(w, data)
				return
			}

			err = db.UpdateUserPassword(database, user.ID, string(pwHash))
			if err != nil {
				data.ErrorMessage = "Internal error. Try again."
				tmpl := template.Must(template.ParseFiles("internal/templates/reset_password.html"))
				tmpl.Execute(w, data)
				return
			}

			data.SuccessMessage = "Your password has been successfully reset!"
			tmpl := template.Must(template.ParseFiles("internal/templates/reset_password.html"))
			tmpl.Execute(w, data)

		} else {
			http.NotFound(w, r)
		}
	}
}

func SendEmail(to, subject, body string) error {
	// SMTP server configuration.
	smtpHost := "smtp.gmail.com"
	smtpPort := "587"

	// Authentication - use your credentials.
	auth := smtp.PlainAuth("", "kingsmansoftwaresolution@gmail.com", "hhsy drzs yajh csej", smtpHost)

	from := "kingsmansoftwaresolutions@gmail.com"
	message := []byte("Subject: " + subject + "\r\n" +
		"From: " + from + "\r\n" +
		"To: " + to + "\r\n" +
		"Content-Type: text/plain; charset=UTF-8\r\n" +
		"\r\n" +
		body + "\r\n")

	// The 'to' parameter in SendMail is a list of recipients.
	recipients := []string{to}

	print("Subject :" + subject + "To" + to + "body" + body)

	// Connect to the SMTP server and send the email.
	err := smtp.SendMail(smtpHost+":"+smtpPort, auth, from, recipients, message)

	if err != nil {
		print(" ")
		log.Printf("smtp error: %s", err)
		return err
	}
	return nil
}
