package main

import (
	"log"
	"net/http"

	"MortgageAgent/internal/db"
	"MortgageAgent/internal/handlers"
	//"github.com/gorilla/mux"
)

// func main() {
// 	dsn := "app.db"
// 	database, err := db.InitDB(dsn)
// 	if err != nil {
// 		log.Fatal("Failed to connect to database:", err)
// 	}

// 	err = db.MigrateDB(database)
// 	if err != nil {
// 		log.Fatal("Failed to migrate database:", err)
// 	}

// 	err = db.SeedAdminUser(database)
// 	if err != nil {
// 		log.Fatal("Failed to seed admin user:", err)
// 	}

// 	mux := http.NewServeMux()

// 	// Serve static files (CSS, JS, images)
// 	fileServer := http.FileServer(http.Dir("internal/static"))
// 	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
// 	mux.Handle("/serve-document", handlers.AuthMiddleware(handlers.ServeDocument(database), database, "admin"))

// 	mux.HandleFunc("/", handlers.LoginPage(database))
// 	mux.HandleFunc("/login", handlers.Login(database))
// 	mux.HandleFunc("/signup", handlers.SignUpPage(database))
// 	mux.HandleFunc("/signup-success", handlers.SignUpSuccessPage())
// 	mux.HandleFunc("/register", handlers.Register(database))
// 	mux.Handle("/broker", handlers.AuthMiddleware(handlers.BrokerLanding(), database, "broker"))
// 	mux.Handle("/application", handlers.AuthMiddleware(handlers.StartApplication(database), database, "broker"))
// 	mux.Handle("/application-form", handlers.AuthMiddleware(handlers.ApplicationFormPage(database), database, "broker"))

// 	mux.HandleFunc("/forgot-password", handlers.ForgotPasswordPage(database))
// 	mux.HandleFunc("/reset-password", handlers.ResetPasswordPage(database))

// 	mux.Handle("/admin", handlers.AuthMiddleware(handlers.AdminLanding(), database, "admin"))
// 	mux.Handle("/admin-dashboard", handlers.AuthMiddleware(handlers.AdminDashboard(database), database, "admin"))
// 	mux.Handle("/view-application", handlers.AuthMiddleware(handlers.ViewApplication(database), database, "admin"))
// 	// cmd/main.go

// 	mux.Handle("/logout", handlers.Logout())

// 	log.Println("Server running on :8080")
// 	http.ListenAndServe(":8080", mux)
// }

func main() {
	// Initialize DB
	database, err := db.InitDB("app.db") // Adjust DSN as needed
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Migrate DB
	err = db.MigrateDB(database)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Seed Admin User
	err = db.SeedAdminUser(database)
	if err != nil {
		log.Fatal("Failed to seed admin user:", err)
	}

	mux := http.NewServeMux()

	// Serve static files
	fileServer := http.FileServer(http.Dir("internal/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))

	// Serve uploaded documents securely
	mux.Handle("/serve-document", handlers.AuthMiddleware(handlers.ServeDocument(database), database, "admin"))

	// Routes without middleware
	mux.HandleFunc("/", handlers.LoginPage(database))
	mux.HandleFunc("/login", handlers.Login(database))
	mux.HandleFunc("/signup", handlers.SignUpPage(database))
	mux.HandleFunc("/register", handlers.Register(database))
	mux.HandleFunc("/signup-success", handlers.SignUpSuccessPage())

	// Routes with middleware
	mux.Handle("/broker", handlers.AuthMiddleware(handlers.BrokerLanding(), database, "broker"))
	mux.Handle("/admin-dashboard", handlers.AuthMiddleware(handlers.AdminDashboard(database), database, "admin"))
	mux.Handle("/logout", handlers.Logout())

	// Forgot/Reset Password
	mux.HandleFunc("/forgot-password", handlers.ForgotPasswordPage(database))
	mux.HandleFunc("/reset-password", handlers.ResetPasswordPage(database))

	// Application Routes
	mux.Handle("/application", handlers.AuthMiddleware(handlers.StartApplication(database), database, "broker"))
	mux.Handle("/application-form", handlers.AuthMiddleware(handlers.ApplicationFormPage(database), database, "broker"))

	// Admin Specific Routes
	mux.Handle("/view-application", handlers.AuthMiddleware(handlers.ViewApplication(database), database, "admin"))

	log.Println("Server running on :8080")
	err = http.ListenAndServe(":8080", mux)
	if err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}
