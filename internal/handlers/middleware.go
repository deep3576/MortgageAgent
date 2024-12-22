package handlers

import (
	"context"
	"database/sql"
	"net/http"

	"MortgageAgent/internal/db"
	"MortgageAgent/internal/models"
)

type contextKey string

var userContextKey = contextKey("user")

func AuthMiddleware(next http.Handler, database *sql.DB, requiredRole string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_email")
		if err != nil || cookie.Value == "" {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		user, err := db.GetUserByEmail(database, cookie.Value)
		if err != nil || user == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}

		if requiredRole != "" && user.UserType != requiredRole {
			http.Error(w, "Unauthorized Access", http.StatusForbidden)
			return
		}

		// Store user in context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}

// Helper function to retrieve user from context
func GetUserFromContext(r *http.Request) *models.User {
	u, ok := r.Context().Value(userContextKey).(*models.User)
	if !ok {
		return nil
	}
	return u
}

// GetUserFromContext retrieves the user from the request context.
