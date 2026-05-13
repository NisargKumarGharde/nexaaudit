package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// The structure of the data we will send to the frontend
type DashboardStats struct {
	TotalDocuments int       `json:"total_documents"`
	TotalValue     float64   `json:"total_value"`
	Anomalies      int       `json:"anomalies"`
	RecentFiles    []DocInfo `json:"recent_files"`
}

type DocInfo struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	Status      string    `json:"status"`
	TotalAmount float64   `json:"total_amount"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

func (h *Handler) GetDashboard(w http.ResponseWriter, r *http.Request) {
	var stats DashboardStats
	stats.RecentFiles = []DocInfo{}
	ctx := context.Background()

	// 1. Calculate Totals directly in PostgreSQL
	h.DB.QueryRow(ctx, "SELECT COUNT(*) FROM documents").Scan(&stats.TotalDocuments)
	h.DB.QueryRow(ctx, "SELECT COALESCE(SUM(total_amount), 0) FROM documents WHERE status != 'failed'").Scan(&stats.TotalValue)
	h.DB.QueryRow(ctx, "SELECT COUNT(*) FROM documents WHERE status = 'flagged'").Scan(&stats.Anomalies)

	// 2. Fetch the 5 most recent documents
	rows, err := h.DB.Query(ctx, "SELECT id, file_name, status, COALESCE(total_amount, 0), uploaded_at FROM documents ORDER BY uploaded_at DESC LIMIT 5")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var doc DocInfo
			rows.Scan(&doc.ID, &doc.FileName, &doc.Status, &doc.TotalAmount, &doc.UploadedAt)
			stats.RecentFiles = append(stats.RecentFiles, doc)
		}
	}

	// 3. Send the JSON to the React frontend
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
