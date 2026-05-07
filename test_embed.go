package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

func main() {
	// Securely load the key from .env
	_ = godotenv.Load()
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("❌ GEMINI_API_KEY not found in .env file")
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, option.WithAPIKey(apiKey))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	fmt.Println("🔍 Asking Google what EMBEDDING models this key can access...")

	iter := client.ListModels(ctx)
	found := false
	for {
		m, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Fatalf("Error fetching models: %v", err)
		}

		// We only want to print models that support generating embeddings
		for _, method := range m.SupportedGenerationMethods {
			if method == "embedContent" {
				fmt.Println("✅ Available Embedding Model:", m.Name)
				found = true
				break
			}
		}
	}
	if !found {
		fmt.Println("❌ No embedding models found for this API key.")
	}
	fmt.Println("Done!")
}
