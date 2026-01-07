package server

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/LackOfMorals/aura-client"
	"github.com/LackOfMorals/mcp4AuraAPI/internal/config"
	"github.com/LackOfMorals/mcp4AuraAPI/internal/dependencies"
	"github.com/LackOfMorals/mcp4AuraAPI/internal/outcomes"
	"github.com/LackOfMorals/mcp4AuraAPI/internal/tools"
	"github.com/mark3labs/mcp-go/server"
)

// Neo4jMCPServer represents the MCP server instance
type Neo4jMCPServer struct {
	MCPServer *server.MCPServer
	config    *config.Config
	aClient   *aura.AuraAPIClient
	aOutcomes *outcomes.OutcomeRegistry
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

	// Register outcomes
	auraOutcomes := outcomes.NewOutcomeRegistry()

	return &Neo4jMCPServer{
		MCPServer: mcpServer,
		config:    cfg,
		version:   version,
		aClient:   auraClient,
		aOutcomes: auraOutcomes,
	}
}

// Start initializes and starts the MCP server using stdio transport
func (s *Neo4jMCPServer) Start() error {
	slog.Info("Starting MCP Aura API Server...")
	err := s.verifyRequirements()
	if err != nil {
		return err
	}

	// Dependencies needed by all outcomes
	outcomeDependencies := dependencies.Dependencies{
		AClient:  s.aClient,
		OutComes: s.aOutcomes,
		Config:   s.config,
	}

	// Register tools
	if err := s.registerTools(outcomeDependencies); err != nil {
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

// registerTools registers all enabled MCP tools and adds them to the  MCP server.
func (s *Neo4jMCPServer) registerTools(deps *dependencies.Dependencies) {
	tools := GetAllTools(deps)
	s.MCPServer.AddTools(tools...)
}

// GetAllTools returns all available tools with their specs and handlers
func GetAllTools(deps *dependencies.Dependencies) []server.ServerTool {
	return []server.ServerTool{
		{
			Tool:    tools.ListOutcomesSpec(),
			Handler: tools.ListOutcomesHandler(deps),
		},
		{
			Tool:    tools.GetOutcomeDetailsSpec(),
			Handler: tools.GetOutcomeDetailsHandler(deps),
		},
		{
			Tool:    tools.ExecuteOutcomeSpec(),
			Handler: tools.ExecuteOutcomeHandler(deps),
		},
	}
}

// Stop gracefully stops the server
func (s *Neo4jMCPServer) Stop() error {
	slog.Info("Stopping MCP Aura API Server...")
	// Currently no cleanup needed - the MCP server handles its own lifecycle
	return nil
}
