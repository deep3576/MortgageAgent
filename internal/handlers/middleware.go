package handlers

import (
	"database/sql"
	"net/http"

	"MortgageAgent/internal/db"
)

func AuthMiddleware(next http.Handler, database *sql.DB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_email")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		// Verify user still exists
		user, err := db.GetUserByEmail(database, cookie.Value)
		if err != nil || user == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
