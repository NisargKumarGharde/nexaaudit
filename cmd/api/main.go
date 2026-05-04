package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/NisargKumarGharde/nexaaudit/internal/db"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "NexaAudit API is running securely!")
}

func main() {
	// 1. Initialize Database Connection
	dbPool := db.NewConnection()
	defer dbPool.Close() // Ensure connections close when server shuts down

	// 2. Setup Routes
	http.HandleFunc("/health", healthCheck)

	// 3. Start Server
	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
