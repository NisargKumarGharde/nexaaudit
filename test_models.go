package main

import (
	"context"
	"fmt"
	"log"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	apiKey := "AIzaSyDJYmKgCn_ibf0lrkKltUTY19kLq1g-TN0"

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	fmt.Println("🔍 Asking Google what models this key can access...")

	iter := client.ListModels(ctx)
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error fetching models: %v", err)
		}

		for _, method := range m.SupportedGenerationMethods {
			if method == "generateContent" {
				fmt.Println("✅ Available Model:", m.Name)
				break
			}
		}
	}
	fmt.Println("Done!")
}
