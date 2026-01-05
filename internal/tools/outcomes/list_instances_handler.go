package outcomes

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/LackOfMorals/mcp4AuraAPI/internal/tools"
	"github.com/mark3labs/mcp-go/mcp"
)

// ListInstancesHandler returns a handler function for the list-instances tool
func ListInstancesHandler(deps *tools.ToolDependencies) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, _ mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		return handleListInstances(ctx, deps)
	}
}

// handleListInstances retrieves a list of instances using Aura Client
func handleListInstances(ctx context.Context, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {

	type instanceDetail struct {
		Id            string
		Name          string
		CloudProvider string
		Memory        string
		Type          string
		URL           string
		Status        string
	}

	type instanceList []instanceDetail

	if deps.AClient == nil {
		errMessage := "Aura API Client is not initialized"
		slog.Error(errMessage)
		return mcp.NewToolResultError(errMessage), nil
	}

	slog.Info("retrieving instances from aura api")

	// Get the list of instances
	instances, err := deps.AClient.Instances.List()
	if err != nil {
		errMessage := "Failed to list instances"
		slog.Error(errMessage)
	}

	// Create an empty list
	records := instanceList{}

	// Now get the details for each instance
	// And add them to records list
	for _, inst := range instances.Data {
		instanceInfo, err := deps.AClient.Instances.Get(string(inst.Id))
		if err != nil {
			errMessage := "Failed to get instance details for " + string(inst.Name)
			slog.Error(errMessage)
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
		slog.Warn("No instances retrieved.")
		return mcp.NewToolResultText("The tool executed successfully; however, no instance information was returned."), nil
	}

	jsonData, err := json.Marshal(records)
	if err != nil {
		slog.Error("failed to serialize to JSON", "error", err)
		return mcp.NewToolResultError(err.Error()), nil

	}
	return mcp.NewToolResultText(string(jsonData)), nil
}
