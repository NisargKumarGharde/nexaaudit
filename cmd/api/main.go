package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/rs/cors"

	"github.com/NisargKumarGharde/nexaaudit/internal/db"
	"github.com/NisargKumarGharde/nexaaudit/internal/handlers"
)

func main() {
	// 1. Load the .env file
	err := godotenv.Overload()
	if err != nil {
		log.Println("⚠️ No .env file found, relying on system environment variables")
	}

	// 2. Initialize Database Connection
	dbPool := db.NewConnection()
	defer dbPool.Close()

	// 3. Initialize Handlers
	appHandler := &handlers.Handler{DB: dbPool}

	// 4. Create a NEW isolated router
	mux := http.NewServeMux()

	// Register ALL routes to this specific mux, NOT http.HandleFunc
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "NexaAudit API is running securely!")
	})
	mux.HandleFunc("/api/v1/upload", appHandler.UploadDocument)

	mux.HandleFunc("/api/v1/dashboard", appHandler.GetDashboard)

	// 5. Configure CORS to allow the frontend to talk to the backend
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"}, // Allows localhost:8081, 3000, 5173, etc.
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		AllowCredentials: true,
	})

	// 6. Wrap the router with CORS
	handler := c.Handler(mux)

	// Render dynamically assigns a port via the PORT environment variable
	port := os.Getenv("PORT")
	if port == "" {
		port = "8082" // Fallback for your local machine
	}

	fmt.Printf("🚀 Server starting on port %s...\n", port)
	log.Fatal(http.ListenAndServe(":"+port, handler))
}
