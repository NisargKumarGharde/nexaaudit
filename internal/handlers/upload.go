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
	// LOUD LOG 1: The exact millisecond the request hits the server
	fmt.Println("\n=================================================")
	fmt.Println("➡️ INCOMING UPLOAD REQUEST RECEIVED!")
	fmt.Println("=================================================")

	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		fmt.Printf("❌ Crash 1 - Form Parse Error: %v\n", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("document")
	if err != nil {
		fmt.Printf("❌ Crash 2 - File Retrieval Error: %v\n", err)
		http.Error(w, "Error retrieving the file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		fmt.Printf("❌ Crash 3 - File Read Error: %v\n", err)
		http.Error(w, "Error reading file", http.StatusInternalServerError)
		return
	}

	mimeType := http.DetectContentType(fileBytes)

	// -- Handle Mock User --
	fmt.Println("🔄 Checking for mock user in database...")
	var mockUserID uuid.UUID
	err = h.DB.QueryRow(context.Background(), "SELECT id FROM users LIMIT 1").Scan(&mockUserID)
	if err != nil {
		fmt.Printf("⚠️ No mock user found. Attempting to create one... (Initial error: %v)\n", err)
		err = h.DB.QueryRow(context.Background(), "INSERT INTO users (email) VALUES ('demo@nexaaudit.com') RETURNING id").Scan(&mockUserID)
		if err != nil {
			fmt.Printf("❌ Crash 4 - DB Mock User Creation Error: %v\n", err)
			http.Error(w, fmt.Sprintf("Failed to create mock user: %v", err), http.StatusInternalServerError)
			return
		}
	}
	fmt.Printf("✅ Mock user ready! ID: %s\n", mockUserID)

	// 1. Initial Insert (Pending State)
	fmt.Println("🔄 Saving initial 'pending' document state to database...")
	insertQuery := `
		INSERT INTO documents (user_id, file_name, status) 
		VALUES ($1, $2, $3) 
		RETURNING id, uploaded_at`

	var docID uuid.UUID
	var uploadedAt time.Time

	err = h.DB.QueryRow(context.Background(), insertQuery, mockUserID, header.Filename, "pending").Scan(&docID, &uploadedAt)
	if err != nil {
		fmt.Printf("❌ Crash 5 - DB Document Insert Error: %v\n", err)
		http.Error(w, fmt.Sprintf("Database error: %v", err), http.StatusInternalServerError)
		return
	}
	fmt.Printf("✅ Document saved as pending! Doc ID: %s\n", docID)

	// 2. Call Gemini AI for extraction
	fmt.Println("🧠 Sending document to Gemini AI for extraction...")
	auditResult, aiErr := ai.AnalyzeInvoice(context.Background(), fileBytes, mimeType)

	// 3. Vector Memory: Check for Duplicates
	var isDuplicate bool
	var originalDocID string
	var vector []float32

	if aiErr == nil {
		fmt.Println("✅ AI Extraction successful! Generating semantic fingerprint...")
		textToEmbed := fmt.Sprintf("Vendor: %s, Amount: %.2f", auditResult.VendorName, auditResult.TotalAmount)

		vector, aiErr = ai.GenerateEmbedding(context.Background(), textToEmbed)
		if aiErr == nil {
			fmt.Println("✅ Fingerprint generated! Checking Pinecone for duplicates...")
			isDuplicate, originalDocID, aiErr = db.CheckDuplicate(context.Background(), vector)
			if aiErr != nil {
				fmt.Printf("⚠️ Pinecone check failed (continuing anyway): %v\n", aiErr)
			}
		} else {
			fmt.Printf("❌ AI Embedding Error: %v\n", aiErr)
		}
	} else {
		fmt.Printf("❌ AI Extraction Error: %v\n", aiErr)
	}

	// 4. Evaluate results and update the database
	finalStatus := "audited"
	if aiErr != nil {
		finalStatus = "failed"
	} else if isDuplicate {
		fmt.Printf("🚨 DUPLICATE DETECTED! Matches Doc ID: %s\n", originalDocID)
		finalStatus = "flagged"
		auditResult.IsFlagged = true
		auditResult.RiskReason = "Duplicate submission detected. Potential fraud attempt."
	} else if auditResult.IsFlagged {
		finalStatus = "flagged"
	}

	if aiErr == nil {
		fmt.Printf("🔄 Updating document status in database to: %s...\n", finalStatus)
		updateQuery := `
			UPDATE documents 
			SET status = $1, total_amount = $2, vendor_name = $3 
			WHERE id = $4`

		_, err = h.DB.Exec(context.Background(), updateQuery, finalStatus, auditResult.TotalAmount, auditResult.VendorName, docID)
		if err != nil {
			fmt.Printf("❌ Crash 6 - DB Document Update Error: %v\n", err)
			http.Error(w, fmt.Sprintf("Failed to update db: %v", err), http.StatusInternalServerError)
			return
		}
		fmt.Println("✅ Database updated successfully!")

		// 5. Save the new fingerprint to Pinecone if it wasn't a duplicate
		if !isDuplicate && vector != nil {
			fmt.Println("💾 Saving new fingerprint to long-term memory (Pinecone)...")
			err = db.SaveVector(context.Background(), docID.String(), vector)
			if err != nil {
				fmt.Printf("⚠️ Failed to save to Pinecone: %v\n", err)
			} else {
				fmt.Println("✅ Fingerprint saved successfully!")
			}
		}
	}

	// 6. Return response
	fmt.Println("🎉 Upload pipeline complete. Sending 201 response to frontend!")
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
