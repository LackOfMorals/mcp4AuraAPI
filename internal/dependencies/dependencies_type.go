// This contains a single type to avoid circular import elsewhere
// Done for now whilst I consider another approach

package dependencies

import (
	"github.com/LackOfMorals/aura-client"
	"github.com/LackOfMorals/mcp4AuraAPI/internal/config"
	"github.com/LackOfMorals/mcp4AuraAPI/internal/outcomes"
)

// Dependencies contains all dependencies needed to achieve an outcome
type Dependencies struct {
	AClient  *aura.AuraAPIClient
	Config   *config.Config
	OutComes *outcomes.OutcomeRegistry
}
