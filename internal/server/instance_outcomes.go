// =============================================================================
// These are all of the instance related outcomes
// =============================================================================

package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LackOfMorals/aura-client"
	"github.com/mark3labs/mcp-go/mcp"
)

// registerListInstancesOutcome registers the list-instances Outcome
func (r *OutcomeRegistry) registerListInstancesOutcome() {
	r.Outcomes["list-instances"] = &Outcome{
		ID:          "list-instances",
		Name:        "List Instances",
		Description: "Retrieve a list of all Neo4j Aura database instances. Returns instance details including name, ID, status, cloud provider, memory size, type, and connection URL.",
		Type:        OutcomesTypeList,
		ReadOnly:    true,
		Parameters:  []OutcomeParameter{}, // No parameters needed for listing
		Metadata: map[string]interface{}{
			"category": "instances",
		},
		Handler: executeListInstances,
	}
}

// executeListInstances implements the list-instances Outcome
func executeListInstances(ctx context.Context, parameters map[string]interface{}, deps *Dependencies) (*mcp.CallToolResult, error) {

	type instanceSummary aura.ListInstanceData

	type instanceList []instanceSummary

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

	// Fill the list with our instance summary
	for _, inst := range instances.Data {
		records = append(records, instanceSummary{
			Name:          inst.Name,
			Id:            inst.Id,
			Created:       inst.Created,
			CloudProvider: inst.CloudProvider,
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

// registerGetInstanceDetailsOutcome registers the get-instance-details outcome
func (r *OutcomeRegistry) registerGetInstanceDetailsOutcome() {
	r.Outcomes["get-instance-details"] = &Outcome{
		ID:          "get-instance-details",
		Name:        "Get Instance Details",
		Description: "Retrieve detailed information about a specific Neo4j Aura database instance. Returns comprehensive details including name, status, connection URL, cloud provider, region, memory, type, tenant ID, and importantly the Prometheus metrics endpoint URL for monitoring.",
		Type:        OutcomesTypeRead,
		ReadOnly:    true,
		Parameters: []OutcomeParameter{
			{
				Name:        "instance_id",
				Type:        "string",
				Description: "The ID of the instance to retrieve details for",
				Required:    true,
			},
		},
		Metadata: map[string]interface{}{
			"category": "instances",
		},
		Handler: executeGetInstanceDetails,
	}
}

// executeGetInstanceDetails implements the get-instance-details outcome
func executeGetInstanceDetails(ctx context.Context, parameters map[string]interface{}, deps *Dependencies) (*mcp.CallToolResult, error) {
	if deps.AClient == nil {
		return mcp.NewToolResultError("Aura API Client is not initialized"), nil
	}

	// Validate and extract required parameter
	instanceID, ok := parameters["instance_id"].(string)
	if !ok || instanceID == "" {
		return mcp.NewToolResultError("'instance_id' parameter is required and must be a non-empty string"), nil
	}

	// Get the instance details from Aura API
	instanceInfo, err := deps.AClient.Instances.Get(instanceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve instance details: %v. The instance may not exist or you may not have access to it.", err)), nil
	}

	// Format the response with all relevant details
	type instanceDetails aura.GetInstanceData

	details := instanceDetails{
		Id:            instanceInfo.Data.Id,
		Name:          instanceInfo.Data.Name,
		Status:        instanceInfo.Data.Status,
		ConnectionUrl: instanceInfo.Data.ConnectionUrl,
		CloudProvider: instanceInfo.Data.CloudProvider,
		Region:        instanceInfo.Data.Region,
		Memory:        instanceInfo.Data.Memory,
		Storage:       instanceInfo.Data.Storage,
		Type:          instanceInfo.Data.Type,
		TenantId:      instanceInfo.Data.TenantId,
		MetricsURL:    instanceInfo.Data.MetricsURL,
	}

	jsonData, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize instance details: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
