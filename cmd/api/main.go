package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/NisargKumarGharde/nexaaudit/internal/db"
	"github.com/NisargKumarGharde/nexaaudit/internal/handlers"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "NexaAudit API is running securely!")
}

func main() {
	// 1. Initialize Database Connection
	dbPool := db.NewConnection()
	defer dbPool.Close() // Ensure connections close when server shuts down

	// 2. Initialize Handlers with the DB pool
	appHandler := &handlers.Handler{DB: dbPool}

	// 3. Setup Routes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "NexaAudit API is running securely!")
	})

	// 4. New Upload Route
	http.HandleFunc("app/v1/upload", appHandler.UploadDocument)

	// 5. Start Server
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
