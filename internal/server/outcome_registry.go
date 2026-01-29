package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/LackOfMorals/aura-client"
	"github.com/LackOfMorals/mcp4AuraAPI/internal/prometheus"
	"github.com/mark3labs/mcp-go/mcp"
)

// Type aliases for Prometheus metrics to avoid import cycles
type MetricsAggregator = prometheus.MetricsAggregator
type ResourceMetrics = prometheus.ResourceMetrics
type QueryMetrics = prometheus.QueryMetrics
type StorageMetrics = prometheus.StorageMetrics

// NewMetricsAggregator is a convenience wrapper
func NewMetricsAggregator(url string) *MetricsAggregator {
	return prometheus.NewMetricsAggregator(url)
}

// OutcomeRegistry manages all available Outcomes
type OutcomeRegistry struct {
	Outcomes map[string]*Outcome
}

// NewOutcomeRegistry creates a new Outcome registry with all available Outcomes
func NewOutcomeRegistry() *OutcomeRegistry {
	registry := &OutcomeRegistry{
		Outcomes: make(map[string]*Outcome),
	}

	// Register all available Outcomes
	// Add more Outcomes here as they are developed
	registry.registerListInstancesOutcome()
	registry.registerGetInstanceDetailsOutcome()
	registry.registerCreateInstanceOutcome()
	registry.registerDeleteInstanceOutcome()

	// Register Prometheus monitoring outcomes
	registry.registerGetInstanceHealthOutcome()
	registry.registerDiagnosePerformanceOutcome()
	registry.registerAnalyzeResourceUsageOutcome()
	registry.registerGetQueryStatisticsOutcome()

	return registry
}

// GetAllSummaries returns summaries of all Outcomes
func (r *OutcomeRegistry) GetAllSummaries() []OutcomeSummary {
	summaries := make([]OutcomeSummary, 0, len(r.Outcomes))
	for _, Outcome := range r.Outcomes {
		summaries = append(summaries, OutcomeSummary{
			ID:          Outcome.ID,
			Name:        Outcome.Name,
			Description: Outcome.Description,
			Type:        Outcome.Type,
			ReadOnly:    Outcome.ReadOnly,
		})
	}
	return summaries
}

// GetOutcome returns the full details of a specific Outcome
func (r *OutcomeRegistry) GetOutcome(id string) (*Outcome, error) {
	Outcome, exists := r.Outcomes[id]
	if !exists {
		return nil, fmt.Errorf("Outcome with ID '%s' not found", id)
	}
	return Outcome, nil
}

// ExecuteOutcome executes a specific Outcome with provided parameters
func (r *OutcomeRegistry) ExecuteOutcome(ctx context.Context, id string, parameters map[string]interface{}, deps *Dependencies) (*mcp.CallToolResult, error) {
	Outcome, err := r.GetOutcome(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	// Check if this is a write operation and we're in read-only mode
	if !Outcome.ReadOnly && deps.Config != nil && deps.Config.ReadOnly {
		return mcp.NewToolResultError(fmt.Sprintf(
			"Cannot execute '%s' Outcome: server is in read-only mode. Write operations are disabled. Set READ_ONLY=false to enable write operations.",
			id,
		)), nil
	}

	// Execute the handler associated with this Outcome
	if Outcome.Handler == nil {
		return mcp.NewToolResultError(fmt.Sprintf("no handler registered for Outcome: %s", id)), nil
	}

	return Outcome.Handler(ctx, parameters, deps)
}

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
	type instanceDetails struct {
		ID              string `json:"id"`
		Name            string `json:"name"`
		Status          string `json:"status"`
		ConnectionURL   string `json:"connection_url"`
		CloudProvider   string `json:"cloud_provider"`
		Region          string `json:"region"`
		Memory          string `json:"memory"`
		Storage         string `json:"storage"`
		Type            string `json:"type"`
		TenantID        string `json:"tenant_id"`
		PrometheusURL   string `json:"prometheus_url,omitempty"`
		CreatedAt       string `json:"created_at,omitempty"`
	}

	details := instanceDetails{
		ID:            instanceInfo.Data.Id,
		Name:          instanceInfo.Data.Name,
		Status:        instanceInfo.Data.Status,
		ConnectionURL: instanceInfo.Data.ConnectionUrl,
		CloudProvider: instanceInfo.Data.CloudProvider,
		Region:        instanceInfo.Data.Region,
		Memory:        instanceInfo.Data.Memory,
		Type:          instanceInfo.Data.Type,
		TenantID:      instanceInfo.Data.TenantId,
	}

	// Handle optional Storage field (it's a pointer)
	if instanceInfo.Data.Storage != nil {
		details.Storage = *instanceInfo.Data.Storage
	}

	// Add Prometheus URL if available
	// The Prometheus endpoint is typically at: https://<instance-id>.metrics.neo4j.io/prometheus
	if instanceInfo.Data.Id != "" {
		details.PrometheusURL = fmt.Sprintf("https://%s.metrics.neo4j.io/prometheus", instanceInfo.Data.Id)
	}

	jsonData, err := json.MarshalIndent(details, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize instance details: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// =============================================================================
// Prometheus Monitoring Outcomes
// =============================================================================

// registerGetInstanceHealthOutcome registers the get-instance-health outcome
func (r *OutcomeRegistry) registerGetInstanceHealthOutcome() {
	r.Outcomes["get-instance-health"] = &Outcome{
		ID:          "get-instance-health",
		Name:        "Get Instance Health",
		Description: "Retrieve comprehensive health metrics for a Neo4j Aura instance from its Prometheus endpoint. Returns current resource utilization (CPU, memory), query performance metrics, connection pool status, and storage statistics. Includes overall health status assessment and actionable recommendations.",
		Type:        OutcomesTypeRead,
		ReadOnly:    true,
		Parameters: []OutcomeParameter{
			{
				Name:        "instance_id",
				Type:        "string",
				Description: "The ID of the instance to check health for",
				Required:    true,
			},
			{
				Name:        "prometheus_url",
				Type:        "string",
				Description: "The Prometheus endpoint URL for the instance",
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

// executeGetInstanceHealth implements the get-instance-health outcome
func executeGetInstanceHealth(ctx context.Context, parameters map[string]interface{}, deps *Dependencies) (*mcp.CallToolResult, error) {
	// Validate parameters
	instanceID, ok := parameters["instance_id"].(string)
	if !ok || instanceID == "" {
		return mcp.NewToolResultError("'instance_id' parameter is required and must be a non-empty string"), nil
	}

	prometheusURL, ok := parameters["prometheus_url"].(string)
	if !ok || prometheusURL == "" {
		return mcp.NewToolResultError("'prometheus_url' parameter is required and must be a non-empty string"), nil
	}

	// Create Prometheus aggregator
	aggregator := NewMetricsAggregator(prometheusURL)

	// Get health summary
	health, err := aggregator.GetInstanceHealth(ctx, instanceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve instance health: %v", err)), nil
	}

	// Serialize result
	jsonData, err := json.MarshalIndent(health, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize health data: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// registerDiagnosePerformanceOutcome registers the diagnose-performance outcome
func (r *OutcomeRegistry) registerDiagnosePerformanceOutcome() {
	r.Outcomes["diagnose-performance"] = &Outcome{
		ID:          "diagnose-performance",
		Name:        "Diagnose Performance",
		Description: "Perform detailed performance analysis of a Neo4j Aura instance over a specified time window. Analyzes current state, identifies trends in resource usage, detects performance bottlenecks, and provides actionable recommendations for optimization. Useful for troubleshooting slow queries, high resource usage, or degraded performance.",
		Type:        OutcomesTypeRead,
		ReadOnly:    true,
		Parameters: []OutcomeParameter{
			{
				Name:        "instance_id",
				Type:        "string",
				Description: "The ID of the instance to diagnose",
				Required:    true,
			},
			{
				Name:        "prometheus_url",
				Type:        "string",
				Description: "The Prometheus endpoint URL for the instance",
				Required:    true,
			},
			{
				Name:        "window_minutes",
				Type:        "number",
				Description: "Time window in minutes to analyze (default: 60)",
				Required:    false,
				Default:     60,
			},
		},
		Metadata: map[string]interface{}{
			"category": "monitoring",
			"phase":    "1",
		},
		Handler: executeDiagnosePerformance,
	}
}

// executeDiagnosePerformance implements the diagnose-performance outcome
func executeDiagnosePerformance(ctx context.Context, parameters map[string]interface{}, deps *Dependencies) (*mcp.CallToolResult, error) {
	// Validate parameters
	instanceID, ok := parameters["instance_id"].(string)
	if !ok || instanceID == "" {
		return mcp.NewToolResultError("'instance_id' parameter is required and must be a non-empty string"), nil
	}

	prometheusURL, ok := parameters["prometheus_url"].(string)
	if !ok || prometheusURL == "" {
		return mcp.NewToolResultError("'prometheus_url' parameter is required and must be a non-empty string"), nil
	}

	// Get window minutes (default to 60)
	windowMinutes := 60
	if wm, ok := parameters["window_minutes"].(float64); ok {
		windowMinutes = int(wm)
	} else if wm, ok := parameters["window_minutes"].(int); ok {
		windowMinutes = wm
	}

	if windowMinutes < 1 || windowMinutes > 1440 {
		return mcp.NewToolResultError("'window_minutes' must be between 1 and 1440 (24 hours)"), nil
	}

	// Create Prometheus aggregator
	aggregator := NewMetricsAggregator(prometheusURL)

	// Perform diagnosis
	diagnosis, err := aggregator.DiagnosePerformance(ctx, instanceID, windowMinutes)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to diagnose performance: %v", err)), nil
	}

	// Serialize result
	jsonData, err := json.MarshalIndent(diagnosis, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize diagnosis data: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// registerAnalyzeResourceUsageOutcome registers the analyze-resource-usage outcome
func (r *OutcomeRegistry) registerAnalyzeResourceUsageOutcome() {
	r.Outcomes["analyze-resource-usage"] = &Outcome{
		ID:          "analyze-resource-usage",
		Name:        "Analyze Resource Usage",
		Description: "Analyze current resource utilization patterns for capacity planning and optimization. Provides detailed breakdown of CPU, memory, storage, and I/O usage with trend analysis. Helps identify right-sizing opportunities and predict when scaling might be needed.",
		Type:        OutcomesTypeRead,
		ReadOnly:    true,
		Parameters: []OutcomeParameter{
			{
				Name:        "instance_id",
				Type:        "string",
				Description: "The ID of the instance to analyze",
				Required:    true,
			},
			{
				Name:        "prometheus_url",
				Type:        "string",
				Description: "The Prometheus endpoint URL for the instance",
				Required:    true,
			},
		},
		Metadata: map[string]interface{}{
			"category": "monitoring",
			"phase":    "1",
		},
		Handler: executeAnalyzeResourceUsage,
	}
}

// executeAnalyzeResourceUsage implements the analyze-resource-usage outcome
func executeAnalyzeResourceUsage(ctx context.Context, parameters map[string]interface{}, deps *Dependencies) (*mcp.CallToolResult, error) {
	// For now, this uses the same underlying health check
	// In a full implementation, this could provide more detailed resource breakdowns
	instanceID, ok := parameters["instance_id"].(string)
	if !ok || instanceID == "" {
		return mcp.NewToolResultError("'instance_id' parameter is required and must be a non-empty string"), nil
	}

	prometheusURL, ok := parameters["prometheus_url"].(string)
	if !ok || prometheusURL == "" {
		return mcp.NewToolResultError("'prometheus_url' parameter is required and must be a non-empty string"), nil
	}

	aggregator := NewMetricsAggregator(prometheusURL)
	health, err := aggregator.GetInstanceHealth(ctx, instanceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to analyze resource usage: %v", err)), nil
	}

	// Format as resource-focused output
	type resourceAnalysis struct {
		InstanceID      string                 `json:"instance_id"`
		Timestamp       string                 `json:"timestamp"`
		Resources       ResourceMetrics        `json:"resources"`
		Storage         StorageMetrics         `json:"storage"`
		Utilization     string                 `json:"utilization_assessment"`
		Recommendations []string               `json:"recommendations"`
	}

	analysis := resourceAnalysis{
		InstanceID:      health.InstanceID,
		Timestamp:       health.Timestamp,
		Resources:       health.Resources,
		Storage:         health.Storage,
		Recommendations: health.Recommendations,
	}

	// Determine utilization assessment
	maxUtil := health.Resources.CPUUsagePercent
	if health.Resources.MemoryUsagePercent > maxUtil {
		maxUtil = health.Resources.MemoryUsagePercent
	}

	if maxUtil < 50 {
		analysis.Utilization = "underutilized - consider right-sizing"
	} else if maxUtil < 70 {
		analysis.Utilization = "optimal"
	} else if maxUtil < 85 {
		analysis.Utilization = "high - monitor closely"
	} else {
		analysis.Utilization = "critical - immediate action needed"
	}

	jsonData, err := json.MarshalIndent(analysis, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize analysis: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// registerGetQueryStatisticsOutcome registers the get-query-statistics outcome
func (r *OutcomeRegistry) registerGetQueryStatisticsOutcome() {
	r.Outcomes["get-query-statistics"] = &Outcome{
		ID:          "get-query-statistics",
		Name:        "Get Query Statistics",
		Description: "Retrieve query performance statistics including throughput (queries per second), latency percentiles, and execution patterns. Useful for understanding query load and identifying optimization opportunities.",
		Type:        OutcomesTypeRead,
		ReadOnly:    true,
		Parameters: []OutcomeParameter{
			{
				Name:        "instance_id",
				Type:        "string",
				Description: "The ID of the instance to get statistics for",
				Required:    true,
			},
			{
				Name:        "prometheus_url",
				Type:        "string",
				Description: "The Prometheus endpoint URL for the instance",
				Required:    true,
			},
		},
		Metadata: map[string]interface{}{
			"category": "monitoring",
			"phase":    "1",
		},
		Handler: executeGetQueryStatistics,
	}
}

// executeGetQueryStatistics implements the get-query-statistics outcome
func executeGetQueryStatistics(ctx context.Context, parameters map[string]interface{}, deps *Dependencies) (*mcp.CallToolResult, error) {
	instanceID, ok := parameters["instance_id"].(string)
	if !ok || instanceID == "" {
		return mcp.NewToolResultError("'instance_id' parameter is required and must be a non-empty string"), nil
	}

	prometheusURL, ok := parameters["prometheus_url"].(string)
	if !ok || prometheusURL == "" {
		return mcp.NewToolResultError("'prometheus_url' parameter is required and must be a non-empty string"), nil
	}

	aggregator := NewMetricsAggregator(prometheusURL)
	health, err := aggregator.GetInstanceHealth(ctx, instanceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to retrieve query statistics: %v", err)), nil
	}

	// Return query-focused metrics
	type queryStats struct {
		InstanceID string        `json:"instance_id"`
		Timestamp  string        `json:"timestamp"`
		Query      QueryMetrics  `json:"query_metrics"`
		Assessment string        `json:"assessment"`
		Advice     []string      `json:"advice,omitempty"`
	}

	stats := queryStats{
		InstanceID: health.InstanceID,
		Timestamp:  health.Timestamp,
		Query:      health.Query,
		Advice:     make([]string, 0),
	}

	// Assess query performance
	if health.Query.AvgLatencyMs < 50 {
		stats.Assessment = "excellent"
	} else if health.Query.AvgLatencyMs < 200 {
		stats.Assessment = "good"
	} else if health.Query.AvgLatencyMs < 500 {
		stats.Assessment = "moderate - review slow queries"
		stats.Advice = append(stats.Advice, "Consider adding indexes for frequently queried properties")
	} else {
		stats.Assessment = "poor - immediate optimization needed"
		stats.Advice = append(stats.Advice, "Review query execution plans", "Add appropriate indexes", "Consider query result caching")
	}

	jsonData, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize statistics: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}

// registerDeleteInstanceOutcome registers the delete-instance Outcome
func (r *OutcomeRegistry) registerDeleteInstanceOutcome() {
	r.Outcomes["delete-instance"] = &Outcome{
		ID:          "delete-instance",
		Name:        "Delete Instance",
		Description: "Permanently delete a Neo4j Aura database instance. This is a destructive operation that cannot be undone. Requires explicit confirmation via the 'confirm' parameter.",
		Type:        OutcomesTypeDelete,
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

// executeDeleteInstance implements the delete-instance Outcome
func executeDeleteInstance(ctx context.Context, parameters map[string]interface{}, deps *Dependencies) (*mcp.CallToolResult, error) {
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
	_, err = deps.AClient.Instances.Delete(instanceID)
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to delete instance: %v", err)), nil
	}

	// Format the response
	type deleteResult struct {
		Success     bool   `json:"success"`
		Message     string `json:"message"`
		DeletedID   string `json:"deleted_id"`
		DeletedName string `json:"deleted_name"`
		Warning     string `json:"warning"`
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

// registerCreateInstanceOutcome registers the create-instance Outcome
func (r *OutcomeRegistry) registerCreateInstanceOutcome() {
	r.Outcomes["create-instance"] = &Outcome{
		ID:          "create-instance",
		Name:        "Create Instance",
		Description: "Create a new Neo4j Aura database instance with specified configuration. Returns the created instance details including ID, name, and connection information.",
		Type:        OutcomesTypeCreate,
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
				Description: "Memory size for the instance ('2GB', '4GB', '8GB', '16GB', '32GB', '48GB', '64GB', '96GB', '128GB', '192GB', '256GB', '384GB', '512GB', '768GB', '1024GB', '1536GB', '2048GB')",
				Required:    true,
			},
			{
				Name:        "type",
				Type:        "string",
				Description: "Instance type: 'free-db', 'professional-db', or 'business-critical','enterprise-db', 'enterprise-ds'",
				Required:    true,
			},
			{
				Name:        "tenantId",
				Type:        "string",
				Description: "The id of the project that the instance will be created in.",
				Required:    true,
			},
		},
		Metadata: map[string]interface{}{
			"category": "instances",
		},
		Handler: executeCreateInstance,
	}
}

// executeCreateInstance implements the create-instance Outcome
func executeCreateInstance(ctx context.Context, parameters map[string]interface{}, deps *Dependencies) (*mcp.CallToolResult, error) {
	// These are supported parameters for creating an instance
	/*
		var supportedMemory = []string{
			"1GB", "2GB", "4GB", "8GB", "16GB", "24GB", "32GB", "48GB", "64GB", "128GB", "192GB", "256GB", "384GB", "512GB",
		}
		var supportedTypes = []string{
			"enterprise-db", "enterprise-ds", "professional-db", "professional-ds", "free-db", "business-critical",
		}
		var supportedCloudProviders = []string{"gcp", "aws", "azure"}
		var supportedVersions = []string{"5"}
		var supportedStorage = []string{
			"2GB", "4GB", "8GB", "16GB", "32GB", "48GB", "64GB", "96GB", "128GB", "192GB", "256GB", "384GB", "512GB",
			"768GB", "1024GB", "1536GB", "2048GB",
		}
	*/

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
	validTypes := map[string]bool{"enterprise-db": true, "enterprise-ds": true, "professional-db": true, "professional-ds": true, "free-db": true, "business-critical": true}
	if !validTypes[instanceType] {
		return mcp.NewToolResultError(fmt.Sprintf("Invalid type '%s'. Must be one of: 'free', 'professional', 'enterprise'", instanceType)), nil
	}

	tenant, ok := parameters["tenantId"].(string)
	if !ok || instanceType == "" {
		return mcp.NewToolResultError("'tenantId' parameter is required"), nil
	}

	version := "5" // default

	// Create the instance using the Aura API client

	instanceDefinition := aura.CreateInstanceConfigData{
		Name:          name,
		CloudProvider: cloudProvider,
		Region:        region,
		Memory:        memory,
		Type:          instanceType,
		Version:       version,
		TenantId:      tenant,
	}

	// Call the Aura API to create the instance
	instance, err := deps.AClient.Instances.Create(&instanceDefinition)
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
		Username      string `json:"User"`
		Password      string `json:"Password"`
	}

	result := createResult{
		Success:       true,
		Message:       "Instance created successfully",
		Id:            instance.Data.Id,
		Name:          instance.Data.Name,
		CloudProvider: instance.Data.CloudProvider,
		Type:          instance.Data.Type,
		URL:           instance.Data.ConnectionUrl,
		Username:      instance.Data.Username,
		Password:      instance.Data.Password,
	}

	jsonData, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return mcp.NewToolResultError(fmt.Sprintf("Failed to serialize results: %v", err)), nil
	}

	return mcp.NewToolResultText(string(jsonData)), nil
}
