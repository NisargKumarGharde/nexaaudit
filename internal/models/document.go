package models

import (
	"time"

	"github.com/google/uuid"
)

type Document struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	FileName    string    `json:"file_name"`
	Status      string    `json:"status"`
	TotalAmount float64   `json:"total_amount"`
	VendorName  string    `json:"vendor_name"`
	UploadedAt  time.Time `json:"uploaded_at"`
}
