package outcomes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LackOfMorals/mcp4AuraAPI/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// OutcomeRegistry manages all available outcomes
type OutcomeRegistry struct {
	outcomes map[string]*Outcome
}

// NewOutcomeRegistry creates a new outcome registry with all available outcomes
func NewOutcomeRegistry() *OutcomeRegistry {
	registry := &OutcomeRegistry{
		outcomes: make(map[string]*Outcome),
	}

	// Register all available outcomes
	registry.registerListInstancesOutcome()
	registry.registerCreateInstanceOutcome()
	registry.registerDeleteInstanceOutcome()
	// Add more outcomes here as they are developed

	return registry
}

// GetAllSummaries returns summaries of all outcomes
func (r *OutcomeRegistry) GetAllSummaries() []OutcomeSummary {
	summaries := make([]OutcomeSummary, 0, len(r.outcomes))
	for _, outcome := range r.outcomes {
		summaries = append(summaries, OutcomeSummary{
			ID:          outcome.ID,
			Name:        outcome.Name,
			Description: outcome.Description,
			Type:        outcome.Type,
			ReadOnly:    outcome.ReadOnly,
		})
	}
	return summaries
}

// GetOutcome returns the full details of a specific outcome
func (r *OutcomeRegistry) GetOutcome(id string) (*Outcome, error) {
	outcome, exists := r.outcomes[id]
	if !exists {
		return nil, fmt.Errorf("outcome with ID '%s' not found", id)
	}
	return outcome, nil
}

// ExecuteOutcome executes a specific outcome with provided parameters
func (r *OutcomeRegistry) ExecuteOutcome(ctx context.Context, id string, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
	outcome, err := r.GetOutcome(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Check if this is a write operation and we're in read-only mode
	if !outcome.ReadOnly && deps.Config != nil && deps.Config.ReadOnly {
		return mcp.NewToolResultError(fmt.Sprintf(
			"Cannot execute '%s' outcome: server is in read-only mode. Write operations are disabled. Set READ_ONLY=false to enable write operations.",
			id,
		)), nil
	}

	// Execute the handler associated with this outcome
	if outcome.Handler == nil {
		return mcp.NewToolResultError(fmt.Sprintf("no handler registered for outcome: %s", id)), nil
	}

	return outcome.Handler(ctx, parameters, deps)
}

// registerListInstancesOutcome registers the list-instances outcome
func (r *OutcomeRegistry) registerListInstancesOutcome() {
	r.outcomes["list-instances"] = &Outcome{
		ID:          "list-instances",
		Name:        "List Instances",
		Description: "Retrieve a list of all Neo4j Aura database instances. Returns instance details including name, ID, status, cloud provider, memory size, type, and connection URL.",
		Type:        OutcomeTypeList,
		ReadOnly:    true,
		Parameters:  []OutcomeParameter{}, // No parameters needed for listing
		Metadata: map[string]interface{}{
			"category": "instances",
		},
		Handler: executeListInstances,
	}
}

// executeListInstances implements the list-instances outcome
func executeListInstances(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
	type instanceDetail struct {
		Id            string `json:"id"`
		Name          string `json:"name"`
		CloudProvider string `json:"cloud_provider"`
		Memory        string `json:"memory"`
		Type          string `json:"type"`
		URL           string `json:"url"`
		Status        string `json:"status"`
	}

	type instanceList []instanceDetail

	if deps.AClient == nil {
		return mcp.NewToolResultError("Aura API Client is not initialized"), nil
	}

	// Get the list of instances
	instances, err := deps.AClient.Instances.List()
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to list instances: %v", err)), nil
	}

	// Create an empty list
	records := instanceList{}

	// Get the details for each instance
	for _, inst := range instances.Data {
		instanceInfo, err := deps.AClient.Instances.Get(string(inst.Id))
		if err != nil {
			// Log error but continue with other instances
			continue
		}

		records = append(records, instanceDetail{
			Name:          instanceInfo.Data.Name,
			Id:            instanceInfo.Data.Id,
			Status:        instanceInfo.Data.Status,
			CloudProvider: instanceInfo.Data.CloudProvider,
			Memory:        instanceInfo.Data.Memory,
			Type:          instanceInfo.Data.Type,
			URL:           instanceInfo.Data.ConnectionUrl,
		})
	}

	if len(records) == 0 {
		return mcp.NewToolResultText("No instances found or user does not have access to any instances."), nil
	}

	jsonData, err := json.Marshal(records)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// registerDeleteInstanceOutcome registers the delete-instance outcome
func (r *OutcomeRegistry) registerDeleteInstanceOutcome() {
	r.outcomes["delete-instance"] = &Outcome{
		ID:          "delete-instance",
		Name:        "Delete Instance",
		Description: "Permanently delete a Neo4j Aura database instance. This is a destructive operation that cannot be undone. Requires explicit confirmation via the 'confirm' parameter.",
		Type:        OutcomeTypeDelete,
		ReadOnly:    false,
		Parameters: []OutcomeParameter{
			{
				Name:        "instance_id",
				Type:        "string",
				Description: "The ID of the instance to delete",
				Required:    true,
			},
			{
				Name:        "confirm",
				Type:        "boolean",
				Description: "Must be set to true to confirm deletion. This is a safety measure to prevent accidental deletions.",
				Required:    true,
			},
		},
		Metadata: map[string]interface{}{
			"category":    "instances",
			"destructive": true,
			"warning":     "This operation permanently deletes the instance and all its data. This cannot be undone.",
		},
		Handler: executeDeleteInstance,
	}
}

// executeDeleteInstance implements the delete-instance outcome
func executeDeleteInstance(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
	if deps.AClient == nil {
		return mcp.NewToolResultError("Aura API Client is not initialized"), nil
	}

	// Validate and extract required parameters
	instanceID, ok := parameters["instance_id"].(string)
	if !ok || instanceID == "" {
		return mcp.NewToolResultError("'instance_id' parameter is required and must be a non-empty string"), nil
	}

	// Check for confirmation
	confirm, ok := parameters["confirm"].(bool)
	if !ok {
		return mcp.NewToolResultError("'confirm' parameter is required and must be a boolean (true to confirm deletion)"), nil
	}

	if !confirm {
		return mcp.NewToolResultError("Deletion not confirmed. Set 'confirm' to true to proceed with deletion. WARNING: This action cannot be undone."), nil
	}

	// Get instance details first to return information about what was deleted
	instanceInfo, err := deps.AClient.Instances.Get(instanceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve instance details before deletion: %v. The instance may not exist or you may not have access to it.", err)), nil
	}

	// Delete the instance using the Aura API client
	err = deps.AClient.Instances.Delete(instanceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete instance: %v", err)), nil
	}

	// Format the response
	type deleteResult struct {
		Success       bool   `json:"success"`
		Message       string `json:"message"`
		DeletedID     string `json:"deleted_id"`
		DeletedName   string `json:"deleted_name"`
		Warning       string `json:"warning"`
	}

	result := deleteResult{
		Success:     true,
		Message:     fmt.Sprintf("Instance '%s' (ID: %s) has been successfully deleted", instanceInfo.Data.Name, instanceID),
		DeletedID:   instanceID,
		DeletedName: instanceInfo.Data.Name,
		Warning:     "This instance and all its data have been permanently deleted and cannot be recovered.",
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// registerCreateInstanceOutcome registers the create-instance outcome
func (r *OutcomeRegistry) registerCreateInstanceOutcome() {
	r.outcomes["create-instance"] = &Outcome{
		ID:          "create-instance",
		Name:        "Create Instance",
		Description: "Create a new Neo4j Aura database instance with specified configuration. Returns the created instance details including ID, name, and connection information.",
		Type:        OutcomeTypeCreate,
		ReadOnly:    false,
		Parameters: []OutcomeParameter{
			{
				Name:        "name",
				Type:        "string",
				Description: "Name for the new instance",
				Required:    true,
			},
			{
				Name:        "cloud_provider",
				Type:        "string",
				Description: "Cloud provider: 'gcp', 'aws', or 'azure'",
				Required:    true,
			},
			{
				Name:        "region",
				Type:        "string",
				Description: "Cloud region (e.g., 'us-east-1' for AWS, 'us-central1' for GCP, 'eastus' for Azure)",
				Required:    true,
			},
			{
				Name:        "memory",
				Type:        "string",
				Description: "Memory size for the instance (e.g., '2GB', '8GB', '16GB', '32GB', '64GB')",
				Required:    true,
			},
			{
				Name:        "type",
				Type:        "string",
				Description: "Instance type: 'free', 'professional', or 'enterprise'",
				Required:    true,
			},
			{
				Name:        "version",
				Type:        "string",
				Description: "Neo4j version (e.g., '5', '5.24'). If not specified, uses the latest stable version.",
				Required:    false,
				Default:     "5",
			},
		},
		Metadata: map[string]interface{}{
			"category": "instances",
		},
		Handler: executeCreateInstance,
	}
}

// executeCreateInstance implements the create-instance outcome
func executeCreateInstance(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
	if deps.AClient == nil {
		return mcp.NewToolResultError("Aura API Client is not initialized"), nil
	}

	// Validate and extract required parameters
	name, ok := parameters["name"].(string)
	if !ok || name == "" {
		return mcp.NewToolResultError("'name' parameter is required and must be a non-empty string"), nil
	}

	cloudProvider, ok := parameters["cloud_provider"].(string)
	if !ok || cloudProvider == "" {
		return mcp.NewToolResultError("'cloud_provider' parameter is required and must be one of: 'gcp', 'aws', 'azure'"), nil
	}

	// Validate cloud provider
	validProviders := map[string]bool{"gcp": true, "aws": true, "azure": true}
	if !validProviders[cloudProvider] {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid cloud_provider '%s'. Must be one of: 'gcp', 'aws', 'azure'", cloudProvider)), nil
	}

	region, ok := parameters["region"].(string)
	if !ok || region == "" {
		return mcp.NewToolResultError("'region' parameter is required and must be a non-empty string"), nil
	}

	memory, ok := parameters["memory"].(string)
	if !ok || memory == "" {
		return mcp.NewToolResultError("'memory' parameter is required (e.g., '2GB', '8GB', '16GB', '32GB', '64GB')"), nil
	}

	instanceType, ok := parameters["type"].(string)
	if !ok || instanceType == "" {
		return mcp.NewToolResultError("'type' parameter is required and must be one of: 'free', 'professional', 'enterprise'"), nil
	}

	// Validate instance type
	validTypes := map[string]bool{"free": true, "professional": true, "enterprise": true}
	if !validTypes[instanceType] {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid type '%s'. Must be one of: 'free', 'professional', 'enterprise'", instanceType)), nil
	}

	// Optional version parameter
	version := "5" // default
	if v, exists := parameters["version"]; exists {
		if vStr, ok := v.(string); ok && vStr != "" {
			version = vStr
		}
	}

	// Create the instance using the Aura API client
	// Note: The actual API call structure depends on the aura-client library
	// This is a typical pattern - adjust based on the actual client API
	type CreateInstanceRequest struct {
		Name          string `json:"name"`
		CloudProvider string `json:"cloud_provider"`
		Region        string `json:"region"`
		Memory        string `json:"memory"`
		Type          string `json:"type"`
		Version       string `json:"version"`
	}

	req := CreateInstanceRequest{
		Name:          name,
		CloudProvider: cloudProvider,
		Region:        region,
		Memory:        memory,
		Type:          instanceType,
		Version:       version,
	}

	// Call the Aura API to create the instance
	instance, err := deps.AClient.Instances.Create(req.Name, req.CloudProvider, req.Region, req.Memory, req.Type, req.Version)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to create instance: %v", err)), nil
	}

	// Format the response
	type createResult struct {
		Success       bool   `json:"success"`
		Message       string `json:"message"`
		Id            string `json:"id"`
		Name          string `json:"name"`
		Status        string `json:"status"`
		CloudProvider string `json:"cloud_provider"`
		Memory        string `json:"memory"`
		Type          string `json:"type"`
		URL           string `json:"url,omitempty"`
	}

	result := createResult{
		Success:       true,
		Message:       "Instance created successfully",
		Id:            instance.Data.Id,
		Name:          instance.Data.Name,
		Status:        instance.Data.Status,
		CloudProvider: instance.Data.CloudProvider,
		Memory:        instance.Data.Memory,
		Type:          instance.Data.Type,
		URL:           instance.Data.ConnectionUrl,
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
