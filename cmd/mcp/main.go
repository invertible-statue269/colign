package main

import (
	"log"
	"os"

	"github.com/gobenpark/CoSpec/internal/mcp"
)

func main() {
	log.SetOutput(os.Stderr)
	log.Println("CoSpec MCP Server starting...")

	server := mcp.NewServer(os.Stdin, os.Stdout)
	if err := server.Run(); err != nil {
		log.Fatalf("MCP server error: %v", err)
	}
}
