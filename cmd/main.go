package main

import (
	"log"
	"net/http"

	"MortgageAgent/internal/db"
	"MortgageAgent/internal/handlers"
)

func main() {
	dsn := "app.db"
	database, err := db.InitDB(dsn)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	err = db.MigrateDB(database)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	err = db.SeedAdminUser(database)
	if err != nil {
		log.Fatal("Failed to seed admin user:", err)
	}

	mux := http.NewServeMux()

	// Serve static files (CSS, JS, images)
	fileServer := http.FileServer(http.Dir("internal/static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fileServer))
	mux.HandleFunc("/", handlers.LoginPage(database))
	mux.HandleFunc("/login", handlers.Login(database))
	mux.HandleFunc("/signup", handlers.SignUpPage(database))
	mux.HandleFunc("/signup-success", handlers.SignUpSuccessPage())
	mux.HandleFunc("/register", handlers.Register(database))
	mux.Handle("/broker", handlers.AuthMiddleware(handlers.BrokerLanding(), database, "broker"))
	mux.Handle("/admin", handlers.AuthMiddleware(handlers.AdminLanding(), database, "admin"))

	mux.Handle("/logout", handlers.Logout())

	log.Println("Server running on :8080")
	http.ListenAndServe(":8080", mux)
}
