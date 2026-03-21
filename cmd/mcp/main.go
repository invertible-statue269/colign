package main

import (
	"log"
	"os"

	"github.com/gobenpark/colign/internal/mcp"
)

func main() {
	log.SetOutput(os.Stderr)
	log.Println("Colign MCP Server starting...")

	apiToken := os.Getenv("COLIGN_API_TOKEN")
	if apiToken == "" {
		log.Fatal("COLIGN_API_TOKEN environment variable is required")
	}

	apiURL := os.Getenv("COLIGN_API_URL")
	if apiURL == "" {
		apiURL = "http://localhost:8080"
	}

	server := mcp.NewServer(os.Stdin, os.Stdout, apiToken, apiURL)
	if err := server.Run(); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
