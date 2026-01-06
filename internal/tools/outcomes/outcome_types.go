package outcomes

import (
	"context"

	"github.com/LackOfMorals/mcp4AuraAPI/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// OutcomeType represents the type/category of an outcome
type OutcomeType string

const (
	OutcomeTypeList   OutcomeType = "list"
	OutcomeTypeRead   OutcomeType = "read"
	OutcomeTypeCreate OutcomeType = "create"
	OutcomeTypeUpdate OutcomeType = "update"
	OutcomeTypeDelete OutcomeType = "delete"
)

// OutcomeHandler is a function that executes an outcome
type OutcomeHandler func(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error)

// Outcome represents a high-level operation that can be performed
type Outcome struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Type        OutcomeType            `json:"type"`
	ReadOnly    bool                   `json:"readonly"`
	Parameters  []OutcomeParameter     `json:"parameters,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	Handler     OutcomeHandler         `json:"-"` // Handler function (not serialized to JSON)
}

// OutcomeParameter represents a parameter required for an outcome
type OutcomeParameter struct {
	Name        string      `json:"name"`
	Type        string      `json:"type"`
	Description string      `json:"description"`
	Required    bool        `json:"required"`
	Default     interface{} `json:"default,omitempty"`
}

// OutcomeSummary is a lightweight version for listing outcomes
type OutcomeSummary struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Type        OutcomeType `json:"type"`
	ReadOnly    bool        `json:"readonly"`
}
