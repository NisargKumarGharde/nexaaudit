package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

type AuditResult struct {
	VendorName  string  `json:"vendor_name"`
	TotalAmount float64 `json:"total_amount"`
	IsFlagged   bool    `json:"is_flagged"`
	RiskReason  string  `json:"risk_reason"`
}

func AnalyzeInvoice(ctx context.Context, fileBytes []byte, mimeType string) (*AuditResult, error) {
	apiKey := "AIzaSyDhiPMndTyK6idkiwYuYMK5EtY93UzwDjY"
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	// 1. Initialize the Gemini Client
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create client: %v", err)
	}
	defer client.Close()

	// 2. Configure the Model (Using 1.5 Pro for multimodal document understanding)
	model := client.GenerativeModel("gemini-2.5-flash")

	// Force the AI to reply in strict JSON format
	model.ResponseMIMEType = "application/json"

	// 3. Define the Prompt
	prompt := `You are an expert financial auditor. Analyze the provided invoice document.
	Extract the following information and return it strictly as JSON:
	- vendor_name: The name of the company issuing the invoice.
	- total_amount: The final total amount as a float (no currency symbols).
	- is_flagged: Set to true ONLY if the invoice looks suspicious, lacks a date, or seems anomalous.
	- risk_reason: If flagged, explain why in one sentence. Otherwise, leave empty.`

	// 4. Send the Request
	resp, err := model.GenerateContent(ctx,
		genai.Text(prompt),
		genai.Blob{
			MIMEType: mimeType,
			Data:     fileBytes,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("gemini generation failed: %v", err)
	}

	// 5. Parse the Response
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil, fmt.Errorf("gemini returned an empty response")
	}

	// Extract the text part from the response
	rawJSON := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])

	// Clean up formatting in case the model adds markdown block ticks
	cleanJSON := strings.TrimPrefix(rawJSON, "```json\n")
	cleanJSON = strings.TrimSuffix(cleanJSON, "\n```")

	// Unmarshal into our Go struct
	var result AuditResult
	if err := json.Unmarshal([]byte(cleanJSON), &result); err != nil {
		return nil, fmt.Errorf("failed to parse AI JSON: %v. Raw output: %s", err, cleanJSON)
	}

	return &result, nil
}

// GenerateEmbedding converts a text string into a 768-dimension vector using direct HTTP
func GenerateEmbedding(ctx context.Context, text string) ([]float32, error) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY environment variable is not set")
	}

	// 1. Point directly to Google's REST API
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-embedding-2:embedContent?key=" + apiKey

	// 2. Build the exact JSON payload Google expects
	payload := map[string]interface{}{
		"model": "models/gemini-embedding-2",
		"content": map[string]interface{}{
			"parts": []map[string]interface{}{
				{"text": text},
			},
		},
		"outputDimensionality": 768,
	}
	body, _ := json.Marshal(payload)

	// 3. Make the request
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("network error: %v", err)
	}
	defer res.Body.Close()

	// 4. Catch the exact response structure
	var result struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode json: %v", err)
	}

	// If Google returns an error, catch it explicitly
	if result.Error != nil {
		return nil, fmt.Errorf("gemini api rejected request: %s", result.Error.Message)
	}

	// If it's still empty, we print the exact text we tried to embed for debugging
	if len(result.Embedding.Values) == 0 {
		return nil, fmt.Errorf("gemini returned empty array for text: '%s'", text)
	}

	return result.Embedding.Values, nil
}
