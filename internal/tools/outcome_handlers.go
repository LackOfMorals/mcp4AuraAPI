package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// Global registry instance - initialized once
var registry *OutcomeRegistry

func init() {
	registry = NewOutcomeRegistry()
}

// ListOutcomesHandler returns a handler function for listing all available Outcomes
func ListOutcomesHandler(deps *Outcomes.OutcomeDependencies) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		summaries := registry.GetAllSummaries()

		jsonData, err := json.MarshalIndent(summaries, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize Outcomes: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// GetOutcomeDetailsHandler returns a handler function for getting Outcome details
func GetOutcomeDetailsHandler(deps *Outcomes.OutcomeDependencies) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Type assert Arguments to map[string]interface{}
		arguments, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments format"), nil
		}

		// Extract Outcome_id from request
		OutcomeID, ok := arguments["Outcome_id"].(string)
		if !ok || OutcomeID == "" {
			return mcp.NewToolResultError("Outcome_id parameter is required and must be a string"), nil
		}

		// Get the Outcome details
		Outcome, err := registry.GetOutcome(OutcomeID)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		jsonData, err := json.MarshalIndent(Outcome, "", "  ")
		if err != nil {
			return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize Outcome details: %v", err)), nil
		}

		return mcp.NewToolResultText(string(jsonData)), nil
	}
}

// ExecuteOutcomeHandler returns a handler function for executing an Outcome
func ExecuteOutcomeHandler(deps *Outcomes.OutcomeDependencies) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// Type assert Arguments to map[string]interface{}
		arguments, ok := request.Params.Arguments.(map[string]interface{})
		if !ok {
			return mcp.NewToolResultError("invalid arguments format"), nil
		}

		// Extract Outcome_id from request
		OutcomeID, ok := arguments["Outcome_id"].(string)
		if !ok || OutcomeID == "" {
			return mcp.NewToolResultError("Outcome_id parameter is required and must be a string"), nil
		}

		// Extract parameters (optional, defaults to empty map)
		var parameters map[string]interface{}
		if paramsVal, exists := arguments["parameters"]; exists {
			if params, ok := paramsVal.(map[string]interface{}); ok {
				parameters = params
			} else {
				return mcp.NewToolResultError("parameters must be an object/map"), nil
			}
		} else {
			parameters = make(map[string]interface{})
		}

		// Execute the Outcome
		return registry.ExecuteOutcome(ctx, OutcomeID, parameters, deps)
	}
}
