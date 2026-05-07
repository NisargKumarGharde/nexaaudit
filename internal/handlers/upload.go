package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/NisargKumarGharde/nexaaudit/internal/ai"
	"github.com/NisargKumarGharde/nexaaudit/internal/db"
)

type Handler struct {
	DB *pgxpool.Pool
}

func (h *Handler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
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

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(fileBytes)

	// -- Handle Mock User --
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

	// 2. Call Gemini AI for extraction
	fmt.Println("🧠 Sending document to Gemini AI...")
	auditResult, aiErr := ai.AnalyzeInvoice(context.Background(), fileBytes, mimeType)

	// 3. Vector Memory: Check for Duplicates
	var isDuplicate bool
	var originalDocID string
	var vector []float32

	if aiErr == nil {
		fmt.Println("🔍 Generating semantic fingerprint for memory...")
		// We embed a strict string of the extracted data
		textToEmbed := fmt.Sprintf("Vendor: %s, Amount: %.2f", auditResult.VendorName, auditResult.TotalAmount)

		vector, aiErr = ai.GenerateEmbedding(context.Background(), textToEmbed)
		if aiErr == nil {
			isDuplicate, originalDocID, aiErr = db.CheckDuplicate(context.Background(), vector)
			if aiErr != nil {
				fmt.Printf("⚠️ Pinecone check failed: %v\n", aiErr)
			}
		}
	}

	// 4. Evaluate results and update the database
	finalStatus := "audited"
	if aiErr != nil {
		fmt.Printf("❌ AI Error: %v\n", aiErr)
		finalStatus = "failed"
	} else if isDuplicate {
		fmt.Printf("🚨 DUPLICATE DETECTED! Matches Doc ID: %s\n", originalDocID)
		finalStatus = "flagged" // Mark as flagged in DB
		auditResult.IsFlagged = true
		auditResult.RiskReason = "Duplicate submission detected. Potential fraud attempt."
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
			http.Error(w, fmt.Sprintf("Failed to update db: %v", err), http.StatusInternalServerError)
			return
		}

		// 5. Save the new fingerprint to Pinecone if it wasn't a duplicate
		if !isDuplicate && vector != nil {
			fmt.Println("💾 Saving new fingerprint to long-term memory...")
			err = db.SaveVector(context.Background(), docID.String(), vector)
			if err != nil {
				fmt.Printf("⚠️ Failed to save to Pinecone: %v\n", err)
			}
		}
	}

	// 6. Return response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":     "Document processed successfully!",
		"document_id": docID,
		"file_name":   header.Filename,
		"status":      finalStatus,
		"ai_results":  auditResult,
	})
}
