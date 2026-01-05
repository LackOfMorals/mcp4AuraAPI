package tools

import (
	"github.com/LackOfMorals/aura-client"
	"github.com/LackOfMorals/mcp4AuraAPI/internal/config"
)

// ToolDependencies contains all dependencies needed by tools
type ToolDependencies struct {
	AClient *aura.AuraAPIClient
	Config  *config.Config
}
