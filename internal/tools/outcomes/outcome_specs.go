package outcomes

import (
	"github.com/mark3labs/mcp-go/mcp"
)

// ListOutcomesSpec returns the tool specification for listing available outcomes
func ListOutcomesSpec() mcp.Tool {
	return mcp.NewTool("list-outcomes",
		mcp.WithDescription(`List all available outcomes (operations) that can be performed on Neo4j Aura resources.
Returns a summary of each outcome including its ID, name, description, type, and whether it's read-only.
Use this to discover what operations are available before getting details or executing them.`),
		mcp.WithTitleAnnotation("List Available Outcomes"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	)
}

// GetOutcomeDetailsSpec returns the tool specification for getting outcome details
func GetOutcomeDetailsSpec() mcp.Tool {
	return mcp.NewTool("get-outcome-details",
		mcp.WithDescription(`Get detailed information about a specific outcome including its full parameters and requirements.
Use this after 'list-outcomes' to understand what parameters are needed before executing an outcome.`),
		mcp.WithTitleAnnotation("Get Outcome Details"),
		mcp.WithString("outcome_id",
			mcp.Required(),
			mcp.Description("The ID of the outcome to get details for (from list-outcomes)")),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	)
}

// ExecuteOutcomeSpec returns the tool specification for executing an outcome
func ExecuteOutcomeSpec() mcp.Tool {
	return mcp.NewTool("execute-outcome",
		mcp.WithDescription(`Execute a specific outcome with the provided parameters.
Use 'list-outcomes' to see available outcomes and 'get-outcome-details' to understand required parameters.`),
		mcp.WithTitleAnnotation("Execute Outcome"),
		mcp.WithString("outcome_id",
			mcp.Required(),
			mcp.Description("The ID of the outcome to execute (from list-outcomes)")),
		mcp.WithObject("parameters",
			mcp.Description("Parameters required for the outcome (see get-outcome-details for parameter specifications)"),
			mcp.AdditionalProperties(true)),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithIdempotentHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false), // Will vary by outcome
		mcp.WithOpenWorldHintAnnotation(false),
	)
}
