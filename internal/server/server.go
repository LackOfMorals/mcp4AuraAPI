package server

import (
	"fmt"
	"log/slog"

	"github.com/mark3labs/mcp-go/server"
	"github.com/neo4j/mcp/internal/config"
)

// Neo4jMCPServer represents the MCP server instance
type Neo4jMCPServer struct {
	MCPServer *server.MCPServer
	config    *config.Config
	version   string
}

// NewNeo4jMCPServer creates a new MCP server instance
// The config parameter is expected to be already validated
func NewNeo4jMCPServer(version string, cfg *config.Config) *Neo4jMCPServer {
	mcpServer := server.NewMCPServer(
		"neo4j-mcp",
		version,
		server.WithToolCapabilities(true),
		server.WithInstructions("This is the Neo4j official MCP server and can provide tool calling to interact with your Neo4j database,"+
			"by inferring the schema with tools like get-schema and executing arbitrary Cypher queries with read-cypher."),
	)

	return &Neo4jMCPServer{
		MCPServer: mcpServer,
		config:    cfg,
		version:   version,
	}
}

// Start initializes and starts the MCP server using stdio transport
func (s *Neo4jMCPServer) Start() error {
	slog.Info("Starting Neo4j MCP Server...")
	err := s.verifyRequirements()
	if err != nil {
		return err
	}

	s.emitStartupEvent()

	// Register tools
	if err := s.registerTools(); err != nil {
		return fmt.Errorf("failed to register tools: %w", err)
	}
	slog.Info("Started Neo4j MCP Server. Now listening for input...")
	// Note: ServeStdio handles its own signal management for graceful shutdown
	return server.ServeStdio(s.MCPServer)
}

// verifyRequirements check the Neo4j requirements:
func (s *Neo4jMCPServer) verifyRequirements() error {

	return nil
}

// Stop gracefully stops the server
func (s *Neo4jMCPServer) Stop() error {
	slog.Info("Stopping Neo4j MCP Server...")
	// Currently no cleanup needed - the MCP server handles its own lifecycle
	// Database service cleanup is handled by the caller (main.go)
	return nil
}
