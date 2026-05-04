package main

import (
	"fmt"
	"log"
	"net/http"
)

func healthCheck(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "NexaAudit API is running!")
}

func main() {
	http.HandleFunc("/health", healthCheck)

	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
