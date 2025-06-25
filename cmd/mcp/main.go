package main

import (
	"context"
	"log"
	"os"

	"repository-context-protocol/internal/mcp"
)

func main() {
	server := mcp.NewRepoContextMCPServer()

	ctx := context.Background()
	if err := server.Run(ctx); err != nil {
		log.Fatalf("MCP server failed: %v", err)
		os.Exit(1)
	}
}
