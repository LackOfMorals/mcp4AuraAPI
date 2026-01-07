package server

import (
	"github.com/LackOfMorals/mcp4AuraAPI/internal/tools"
	"github.com/LackOfMorals/mcp4AuraAPI/internal/tools/Tools"
	"github.com/mark3labs/mcp-go/server"
)

// registerTools registers all enabled MCP tools and adds them to the provided MCP server.
// Tools are filtered according to the server configuration. For example, when the read-only
// mode is enabled (e.g. via the NEO4J_READ_ONLY environment variable or the Config.ReadOnly flag),
// any tool that performs state mutation will be excluded; only tools annotated as read-only will be registered.
// Note: this read-only filtering relies on the tool annotation "readonly" (ReadOnlyHint). If the annotation
// is not defined or is set to false, the tool will be added (i.e., only tools with readonly=true are filtered in read-only mode).
func (s *Neo4jMCPServer) registerTools() error {
	filteredTools := s.getEnabledTools()
	s.MCPServer.AddTools(filteredTools...)
	return nil
}

type toolFilter func(tools []ToolDefinition) []ToolDefinition

type toolCategory int

const (
	instancesCategory toolCategory = 0
)

type ToolDefinition struct {
	category   toolCategory
	definition server.ServerTool
	readonly   bool
}

func (s *Neo4jMCPServer) getEnabledTools() []server.ServerTool {
	filters := make([]toolFilter, 0)

	// If read-only mode is enabled, expose only tools annotated as read-only.
	if s.config != nil && s.config.ReadOnly {
		filters = append(filters, filterWriteTools)
	}

	deps := &tools.ToolDependencies{
		AClient: s.aClient,
		Config:  s.config,
	}

	toolDefs := s.getAllToolsDefs(deps)

	for _, filter := range filters {
		toolDefs = filter(toolDefs)
	}
	enabledTools := make([]server.ServerTool, 0)
	for _, toolDef := range toolDefs {
		enabledTools = append(enabledTools, toolDef.definition)
	}
	return enabledTools
}

func filterWriteTools(tools []ToolDefinition) []ToolDefinition {
	readOnlyTools := make([]ToolDefinition, 0, len(tools))
	for _, t := range tools {
		if t.readonly {
			readOnlyTools = append(readOnlyTools, t)
		}
	}
	return readOnlyTools
}

// getAllToolsDefs returns all available tools with their specs and handlers
func (s *Neo4jMCPServer) getAllToolsDefs(deps *tools.ToolDependencies) []ToolDefinition {

	return []ToolDefinition{
		// Tool-based tools - the new pattern
		{
			category: instancesCategory,
			definition: server.ServerTool{
				Tool:    Tools.ListToolsSpec(),
				Handler: Tools.ListToolsHandler(deps),
			},
			readonly: true,
		},
		{
			category: instancesCategory,
			definition: server.ServerTool{
				Tool:    Tools.GetToolDetailsSpec(),
				Handler: Tools.GetToolDetailsHandler(deps),
			},
			readonly: true,
		},
		{
			category: instancesCategory,
			definition: server.ServerTool{
				Tool:    Tools.ExecuteToolSpec(),
				Handler: Tools.ExecuteToolHandler(deps),
			},
			readonly: true, // Tool itself is read-only; Tool-level checks prevent write operations
		},

		// Add other categories below...
	}
}
