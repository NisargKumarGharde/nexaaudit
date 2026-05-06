package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/NisargKumarGharde/nexaaudit/internal/ai"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	DB *pgxpool.Pool
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20) // 10MB limit
	if err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Read the file bytes into memory so we can send them to Gemini
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	// Detect the MIME type (e.g., application/pdf, image/jpeg)
	mimeType := http.DetectContentType(fileBytes)

	// -- Handle the Mock User --
	var mockUserID uuid.UUID
	err = h.DB.QueryRow(context.Background(), "SELECT id FROM users LIMIT 1").Scan(&mockUserID)
	if err != nil {
		err = h.DB.QueryRow(context.Background(), "INSERT INTO users (email) VALUES ('demo@nexaaudit.com') RETURNING id").Scan(&mockUserID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create mock user: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// 1. Initial Insert (Pending State)
	insertQuery := `
		INSERT INTO documents (user_id, file_name, status) 
		VALUES ($1, $2, $3) 
		RETURNING id, uploaded_at`

	var docID uuid.UUID
	var uploadedAt time.Time

	err = h.DB.QueryRow(context.Background(), insertQuery, mockUserID, header.Filename, "pending").Scan(&docID, &uploadedAt)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}

	// 2. Call Gemini AI!
	fmt.Println("🧠 Sending document to Gemini 1.5 Pro...")
	auditResult, aiErr := ai.AnalyzeInvoice(context.Background(), fileBytes, mimeType)

	// 3. Evaluate results and update the database
	finalStatus := "audited"
	if aiErr != nil {
		fmt.Printf("❌ AI Error: %v\n", aiErr)
		finalStatus = "failed"
	} else if auditResult.IsFlagged {
		finalStatus = "flagged"
	}

	if aiErr == nil {
		updateQuery := `
			UPDATE documents 
			SET status = $1, total_amount = $2, vendor_name = $3 
			WHERE id = $4`

		_, err = h.DB.Exec(context.Background(), updateQuery, finalStatus, auditResult.TotalAmount, auditResult.VendorName, docID)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to update database with AI results: %v", err), http.StatusInternalServerError)
			return
		}
	}

	// 4. Return the enriched JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Document processed successfully!",
		"document_id": docID,
		"file_name":   header.Filename,
		"status":      finalStatus,
		"ai_results":  auditResult, // This will inject the structured JSON from Gemini directly into the API response!
	})
}
