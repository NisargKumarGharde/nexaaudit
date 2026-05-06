package ai

import (
	"context"
	"encoding/json"
	"fmt"
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
	apiKey := "----"
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
