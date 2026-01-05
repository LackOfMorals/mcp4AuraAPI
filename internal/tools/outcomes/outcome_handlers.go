package outcomes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LackOfMorals/mcp4AuraAPI/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// Global registry instance - initialized once
var registry *OutcomeRegistry

func init() {
	registry = NewOutcomeRegistry()
}

// ListOutcomesHandler returns a handler function for listing all available outcomes
func ListOutcomesHandler(deps *tools.ToolDependencies) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		summaries := registry.GetAllSummaries()

		jsonData, err := json.MarshalIndent(summaries, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize outcomes: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// GetOutcomeDetailsHandler returns a handler function for getting outcome details
func GetOutcomeDetailsHandler(deps *tools.ToolDependencies) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract outcome_id from request
		outcomeID, ok := request.Params.Arguments["outcome_id"].(string)
		if !ok || outcomeID == "" {
			return mcp.NewToolResultError("outcome_id parameter is required and must be a string"), nil
		}

		// Get the outcome details
		outcome, err := registry.GetOutcome(outcomeID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonData, err := json.MarshalIndent(outcome, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize outcome details: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// ExecuteOutcomeHandler returns a handler function for executing an outcome
func ExecuteOutcomeHandler(deps *tools.ToolDependencies) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Extract outcome_id from request
		outcomeID, ok := request.Params.Arguments["outcome_id"].(string)
		if !ok || outcomeID == "" {
			return mcp.NewToolResultError("outcome_id parameter is required and must be a string"), nil
		}

		// Extract parameters (optional, defaults to empty map)
		var parameters map[string]interface{}
		if paramsVal, exists := request.Params.Arguments["parameters"]; exists {
			if params, ok := paramsVal.(map[string]interface{}); ok {
				parameters = params
			} else {
				return mcp.NewToolResultError("parameters must be an object/map"), nil
			}
		} else {
			parameters = make(map[string]interface{})
		}

		// Execute the outcome
		return registry.ExecuteOutcome(ctx, outcomeID, parameters, deps)
	}
}
