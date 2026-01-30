// =============================================================================
// Prometheus Monitoring Outcomes
// =============================================================================

package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
)

// registerGetInstanceHealthSimpleOutcome registers a simplified get-instance-health outcome
// that automatically fetches the Prometheus URL from instance details
func (r *OutcomeRegistry) registerGetInstanceHealthOutcome() {
	r.Outcomes["get-instance-health"] = &Outcome{
		ID:          "get-instance-health",
		Name:        "Get Instance Health",
		Description: "Retrieve an overview of the health of an instance using it's instance id.",
		Type:        OutcomesTypeRead,
		ReadOnly:    true,
		Parameters: []OutcomeParameter{
			{
				Name:        "instance_id",
				Type:        "string",
				Description: "The ID of the instance to check health for",
				Required:    true,
			},
		},
		Metadata: map[string]interface{}{
			"category": "monitoring",
			"phase":    "1",
		},
		Handler: executeGetInstanceHealth,
	}
}

// executeGetInstanceHealthSimple implements the simplified get-instance-health outcome
func executeGetInstanceHealth(ctx context.Context, parameters map[string]interface{}, deps *Dependencies) (*mcp.CallToolResult, error) {
	if deps.AClient == nil {
		return mcp.NewToolResultError("Aura API Client is not initialized"), nil
	}

	// Validate parameters
	instanceID, ok := parameters["instance_id"].(string)
	if !ok || instanceID == "" {
		return mcp.NewToolResultError("'instance_id' parameter is required and must be a non-empty string"), nil
	}

	// Step 1: Get instance details to retrieve Prometheus URL
	instanceInfo, err := deps.AClient.Instances.Get(instanceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve instance details: %v. The instance may not exist or you may not have access to it.", err)), nil
	}

	// Step 2: This is our metrics URL to scrape
	prometheusURL := instanceInfo.Data.MetricsURL

	health, err := deps.AClient.Prometheus.GetInstanceHealth(instanceID, prometheusURL)

	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve instance metrics: %v", err)), nil
	}

	rawMetrics, err := deps.AClient.Prometheus.FetchRawMetrics(prometheusURL)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve instance raw metrics: %v", err)), nil
	}

	// Database size
	nodeCount, _ := deps.AClient.Prometheus.GetMetricValue(rawMetrics, "neo4j_database_count_node", nil)
	relCount, _ := deps.AClient.Prometheus.GetMetricValue(rawMetrics, "neo4j_database_count_relationship", nil)

	type enrichedHealth struct {
		TimeStamp         string `json:"Time Stamp"`
		Name              string `json:"name"`
		Id                string `json:"Id"`
		OverallStatus     string `json:"Overall Status"`
		Nodes             string `json:"Nodes"`
		Relationships     string `json:"Relationships"`
		CpuUsage          string `json:"CPU Usage"`
		MemoryUsage       string `json:"Memory Usage"`
		AvgLatencyP50     string `json:"Average Query Latency (P50)"`
		ActiveConnections string `json:"Active Connections"`
		MaxConnections    string `json:"Max Connections"`
		HealthIssueCount  int    `json:"Health Issues"`
	}

	result := enrichedHealth{
		TimeStamp:         fmt.Sprintf("%s", health.Timestamp.Format("2006-01-02 15:00:00")),
		Name:              instanceInfo.Data.Name,
		Id:                instanceInfo.Data.Id,
		OverallStatus:     health.OverallStatus,
		Nodes:             fmt.Sprintf("%.0f", nodeCount),
		Relationships:     fmt.Sprintf("%.0f", relCount),
		MemoryUsage:       fmt.Sprintf("%.2f%%", health.Resources.MemoryUsagePercent),
		CpuUsage:          fmt.Sprintf("%.2f%%", health.Resources.CPUUsagePercent),
		AvgLatencyP50:     fmt.Sprintf("%.2f%%", health.Query.AvgLatencyMS),
		ActiveConnections: fmt.Sprintf("%d", health.Connections.ActiveConnections),
		MaxConnections:    fmt.Sprintf("%d", health.Connections.MaxConnections),
		HealthIssueCount:  len(health.Issues),
	}

	// Serialize result
	jsonData, err := json.MarshalIndent(result, "", "  ")
	//jsonData, err := json.MarshalIndent(health, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize health data: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
