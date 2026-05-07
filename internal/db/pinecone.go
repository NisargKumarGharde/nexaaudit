package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// QueryResponse matches the exact JSON returned by Pinecone's REST API
type QueryResponse struct {
	Matches []struct {
		ID       string                 `json:"id"`
		Score    float32                `json:"score"`
		Metadata map[string]interface{} `json:"metadata"`
	} `json:"matches"`
}

// CheckDuplicate queries Pinecone via HTTP to see if a similar document exists
func CheckDuplicate(ctx context.Context, vector []float32) (bool, string, error) {
	apiKey := os.Getenv("PINECONE_API_KEY")
	host := os.Getenv("PINECONE_HOST")

	if apiKey == "" || host == "" {
		return false, "", fmt.Errorf("PINECONE_API_KEY or PINECONE_HOST is missing")
	}

	url := fmt.Sprintf("https://%s/query", host)

	payload := map[string]interface{}{
		"vector":          vector,
		"topK":            1,
		"includeMetadata": true,
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Add("Api-Key", apiKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return false, "", err
	}
	defer res.Body.Close()

	var queryRes QueryResponse
	if err := json.NewDecoder(res.Body).Decode(&queryRes); err != nil {
		return false, "", err
	}

	// Evaluate the Score
	if len(queryRes.Matches) > 0 {
		match := queryRes.Matches[0]
		if match.Score >= 0.99 {
			originalDocID := ""
			// Safely extract the original document ID from the metadata map
			if docID, ok := match.Metadata["document_id"].(string); ok {
				originalDocID = docID
			}
			return true, originalDocID, nil
		}
	}

	return false, "", nil
}

// SaveVector stores the new fingerprint in Pinecone via HTTP POST
func SaveVector(ctx context.Context, docID string, vector []float32) error {
	apiKey := os.Getenv("PINECONE_API_KEY")
	host := os.Getenv("PINECONE_HOST")

	url := fmt.Sprintf("https://%s/vectors/upsert", host)

	payload := map[string]interface{}{
		"vectors": []map[string]interface{}{
			{
				"id":     docID,
				"values": vector,
				"metadata": map[string]interface{}{
					"document_id": docID,
				},
			},
		},
	}
	body, _ := json.Marshal(payload)

	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Add("Api-Key", apiKey)
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	// Ensure the request was successful
	if res.StatusCode >= 400 {
		respBody, _ := io.ReadAll(res.Body)
		return fmt.Errorf("pinecone upsert failed: %s", string(respBody))
	}

	return nil
}
