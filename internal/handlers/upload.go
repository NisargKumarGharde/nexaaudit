package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// We use a struct so we can pass the database pool to our handlers
type Handler struct {
	DB *pgxpool.Pool
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	// 1. Parse the incoming form data (limit to 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	// 2. Retrieve the file from the form data
	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// 3. For now, we will mock a user ID (until we add Auth0 later)
	mockUserID := uuid.New()

	// 4. Insert the record into PostgreSQL using pgx
	query := `
	INSERT INTO documents (user_id, file_name, status)
	VALUES ($1, $2, $3)
	RETURNING id, status, uploaded_at`

	var docID uuid.UUID
	var status string
	var uploadedAt string

	err = h.DB.QueryRow(context.Background(), query, mockUserID, header.Filename, "pending").Scan(&docID, &status, &uploadedAt)

	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	// 5. Return a JSON success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Document uploaded successfully",
		"document_id": docID,
		"file_name":   header.Filename,
		"status":      status,
	})
}
