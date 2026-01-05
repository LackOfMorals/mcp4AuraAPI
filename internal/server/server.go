package server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client"
	"github.com/LackOfMorals/mcp4AuraAPI/internal/config"
	"github.com/mark3labs/mcp-go/server"
)

// Neo4jMCPServer represents the MCP server instance
type Neo4jMCPServer struct {
	MCPServer *server.MCPServer
	config    *config.Config
	aClient   *aura.AuraAPIClient
	version   string
}

// NewNeo4jMCPServer creates a new MCP server instance
// The config parameter is expected to be already validated
func NewNeo4jMCPServer(version string, cfg *config.Config) *Neo4jMCPServer {
	mcpServer := server.NewMCPServer(
		"mcp-aura-api",
		version,
		server.WithToolCapabilities(true),
		server.WithInstructions("This MCP server provides tools for interacting with Neo4j Aura API "),
	)

	// Create the client to Aura API
	auraClient, _ := aura.NewClient(
		aura.WithCredentials(cfg.ClientId, cfg.ClientSecret),
		aura.WithTimeout(120*time.Second),
	)

	return &Neo4jMCPServer{
		MCPServer: mcpServer,
		config:    cfg,
		version:   version,
		aClient:   auraClient,
	}
}

// Start initializes and starts the MCP server using stdio transport
func (s *Neo4jMCPServer) Start() error {
	slog.Info("Starting MCP Aura API Server...")
	err := s.verifyRequirements()
	if err != nil {
		return err
	}

	// Register tools
	if err := s.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}
	slog.Info("Started MCP Aura API Server. Now listening for input...")
	// Note: ServeStdio handles its own signal management for graceful shutdown
	return server.ServeStdio(s.MCPServer)
}

// verifyRequirements check the Neo4j requirements:
func (s *Neo4jMCPServer) verifyRequirements() error {

	return nil
}

// Stop gracefully stops the server
func (s *Neo4jMCPServer) Stop() error {
	slog.Info("Stopping MCP Aura API Server...")
	// Currently no cleanup needed - the MCP server handles its own lifecycle
	return nil
}
