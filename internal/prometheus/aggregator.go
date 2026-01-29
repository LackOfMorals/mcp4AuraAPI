package prometheus

import (
	"context"
	"fmt"
	"math"
	"time"
)

// MetricsAggregator provides high-level metrics aggregation for Neo4j instances
type MetricsAggregator struct {
	client *Client
}

// NewMetricsAggregator creates a new metrics aggregator
func NewMetricsAggregator(prometheusURL string) *MetricsAggregator {
	return &MetricsAggregator{
		client: NewClient(prometheusURL),
	}
}

// HealthSummary represents the overall health of an instance
type HealthSummary struct {
	InstanceID      string            `json:"instance_id"`
	Timestamp       string            `json:"timestamp"`
	Resources       ResourceMetrics   `json:"resources"`
	Query           QueryMetrics      `json:"query"`
	Connections     ConnectionMetrics `json:"connections"`
	Storage         StorageMetrics    `json:"storage"`
	OverallStatus   string            `json:"overall_status"`
	Issues          []string          `json:"issues,omitempty"`
	Recommendations []string          `json:"recommendations,omitempty"`
}

// ResourceMetrics represents resource utilization
type ResourceMetrics struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	DiskIOOps          float64 `json:"disk_io_ops,omitempty"`
}

// QueryMetrics represents query performance
type QueryMetrics struct {
	QueriesPerSecond float64 `json:"queries_per_second"`
	AvgLatencyMs     float64 `json:"avg_latency_ms"`
	P95LatencyMs     float64 `json:"p95_latency_ms,omitempty"`
	P99LatencyMs     float64 `json:"p99_latency_ms,omitempty"`
}

// ConnectionMetrics represents connection pool status
type ConnectionMetrics struct {
	ActiveConnections int     `json:"active_connections"`
	MaxConnections    int     `json:"max_connections"`
	UsagePercent      float64 `json:"usage_percent"`
}

// StorageMetrics represents storage utilization
type StorageMetrics struct {
	UsedGB        float64 `json:"used_gb,omitempty"`
	AvailableGB   float64 `json:"available_gb,omitempty"`
	UsagePercent  float64 `json:"usage_percent,omitempty"`
	PageCacheHits float64 `json:"page_cache_hits_percent,omitempty"`
}

// PerformanceDiagnosis represents detailed performance analysis
type PerformanceDiagnosis struct {
	InstanceID      string                 `json:"instance_id"`
	AnalysisWindow  string                 `json:"analysis_window"`
	CurrentState    HealthSummary          `json:"current_state"`
	Trends          map[string]string      `json:"trends"`
	Bottlenecks     []string               `json:"bottlenecks"`
	SlowQueries     []SlowQuery            `json:"slow_queries,omitempty"`
	Recommendations []string               `json:"recommendations"`
	Metadata        map[string]interface{} `json:"metadata,omitempty"`
}

// SlowQuery represents a slow query analysis
type SlowQuery struct {
	QueryPattern string  `json:"query_pattern,omitempty"`
	AvgTimeMs    float64 `json:"avg_time_ms"`
	Count        int     `json:"count,omitempty"`
	Impact       string  `json:"impact"`
}

// GetInstanceHealth retrieves comprehensive health metrics for an instance
func (m *MetricsAggregator) GetInstanceHealth(ctx context.Context, instanceID string) (*HealthSummary, error) {
	summary := &HealthSummary{
		InstanceID: instanceID,
		Timestamp:  time.Now().Format(time.RFC3339),
		Issues:     make([]string, 0),
	}

	// Get CPU usage
	if cpu, err := m.getCPUUsage(ctx, instanceID); err == nil {
		summary.Resources.CPUUsagePercent = cpu
	}

	// Get memory usage
	if mem, err := m.getMemoryUsage(ctx, instanceID); err == nil {
		summary.Resources.MemoryUsagePercent = mem
	}

	// Get query rate
	if qps, err := m.getQueryRate(ctx, instanceID); err == nil {
		summary.Query.QueriesPerSecond = qps
	}

	// Get connection metrics
	if conn, err := m.getConnectionMetrics(ctx, instanceID); err == nil {
		summary.Connections = conn
	}

	// Get page cache hit rate
	if hitRate, err := m.getPageCacheHitRate(ctx, instanceID); err == nil {
		summary.Storage.PageCacheHits = hitRate
	}

	// Analyze health and generate issues/recommendations
	m.analyzeHealth(summary)

	return summary, nil
}

// DiagnosePerformance performs detailed performance analysis
func (m *MetricsAggregator) DiagnosePerformance(ctx context.Context, instanceID string, windowMinutes int) (*PerformanceDiagnosis, error) {
	diagnosis := &PerformanceDiagnosis{
		InstanceID:      instanceID,
		AnalysisWindow:  fmt.Sprintf("Last %d minutes", windowMinutes),
		Trends:          make(map[string]string),
		Bottlenecks:     make([]string, 0),
		Recommendations: make([]string, 0),
		Metadata:        make(map[string]interface{}),
	}

	// Get current state
	currentHealth, err := m.GetInstanceHealth(ctx, instanceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get current health: %w", err)
	}
	diagnosis.CurrentState = *currentHealth

	// Analyze trends over the window
	end := time.Now()
	start := end.Add(-time.Duration(windowMinutes) * time.Minute)

	// CPU trend
	if trend, err := m.analyzeTrend(ctx, instanceID, "cpu", start, end); err == nil {
		diagnosis.Trends["cpu"] = trend
	}

	// Memory trend
	if trend, err := m.analyzeTrend(ctx, instanceID, "memory", start, end); err == nil {
		diagnosis.Trends["memory"] = trend
	}

	// Query rate trend
	if trend, err := m.analyzeTrend(ctx, instanceID, "query_rate", start, end); err == nil {
		diagnosis.Trends["query_rate"] = trend
	}

	// Identify bottlenecks
	m.identifyBottlenecks(diagnosis)

	// Generate recommendations
	m.generateRecommendations(diagnosis)

	return diagnosis, nil
}

// Helper methods for specific metrics

func (m *MetricsAggregator) getCPUUsage(ctx context.Context, instanceID string) (float64, error) {
	query := fmt.Sprintf(`rate(process_cpu_seconds_total{instance="%s"}[5m]) * 100`, instanceID)
	result, err := m.client.Query(ctx, query)
	if err != nil {
		return 0, err
	}

	if len(result.Data.Result) == 0 {
		return 0, nil
	}

	return getFloatValue(result.Data.Result[0].Value)
}

func (m *MetricsAggregator) getMemoryUsage(ctx context.Context, instanceID string) (float64, error) {
	query := fmt.Sprintf(`(neo4j_memory_heap_used_bytes{instance="%s"} / neo4j_memory_heap_max_bytes{instance="%s"}) * 100`, instanceID, instanceID)
	result, err := m.client.Query(ctx, query)
	if err != nil {
		return 0, err
	}

	if len(result.Data.Result) == 0 {
		return 0, nil
	}

	return getFloatValue(result.Data.Result[0].Value)
}

func (m *MetricsAggregator) getQueryRate(ctx context.Context, instanceID string) (float64, error) {
	query := fmt.Sprintf(`rate(neo4j_transaction_started_total{instance="%s"}[5m])`, instanceID)
	result, err := m.client.Query(ctx, query)
	if err != nil {
		return 0, err
	}

	if len(result.Data.Result) == 0 {
		return 0, nil
	}

	return getFloatValue(result.Data.Result[0].Value)
}

func (m *MetricsAggregator) getConnectionMetrics(ctx context.Context, instanceID string) (ConnectionMetrics, error) {
	metrics := ConnectionMetrics{}

	// Active connections
	activeQuery := fmt.Sprintf(`neo4j_bolt_connections_opened_total{instance="%s"} - neo4j_bolt_connections_closed_total{instance="%s"}`, instanceID, instanceID)
	if result, err := m.client.Query(ctx, activeQuery); err == nil && len(result.Data.Result) > 0 {
		if val, err := getFloatValue(result.Data.Result[0].Value); err == nil {
			metrics.ActiveConnections = int(val)
		}
	}

	// Max connections (example - would need actual metric)
	metrics.MaxConnections = 100 // Would query from config or metric

	if metrics.MaxConnections > 0 {
		metrics.UsagePercent = (float64(metrics.ActiveConnections) / float64(metrics.MaxConnections)) * 100
	}

	return metrics, nil
}

func (m *MetricsAggregator) getPageCacheHitRate(ctx context.Context, instanceID string) (float64, error) {
	query := fmt.Sprintf(`(rate(neo4j_page_cache_hits_total{instance="%s"}[5m]) / (rate(neo4j_page_cache_hits_total{instance="%s"}[5m]) + rate(neo4j_page_cache_faults_total{instance="%s"}[5m]))) * 100`, instanceID, instanceID, instanceID)
	result, err := m.client.Query(ctx, query)
	if err != nil {
		return 0, err
	}

	if len(result.Data.Result) == 0 {
		return 0, nil
	}

	return getFloatValue(result.Data.Result[0].Value)
}

// analyzeTrend analyzes whether a metric is increasing, decreasing, or stable
func (m *MetricsAggregator) analyzeTrend(ctx context.Context, instanceID, metricType string, start, end time.Time) (string, error) {
	var query string

	switch metricType {
	case "cpu":
		query = fmt.Sprintf(`rate(process_cpu_seconds_total{instance="%s"}[5m]) * 100`, instanceID)
	case "memory":
		query = fmt.Sprintf(`(neo4j_memory_heap_used_bytes{instance="%s"} / neo4j_memory_heap_max_bytes{instance="%s"}) * 100`, instanceID, instanceID)
	case "query_rate":
		query = fmt.Sprintf(`rate(neo4j_transaction_started_total{instance="%s"}[5m])`, instanceID)
	default:
		return "unknown", fmt.Errorf("unknown metric type: %s", metricType)
	}

	result, err := m.client.QueryRange(ctx, query, start, end, 1*time.Minute)
	if err != nil || len(result.Data.Result) == 0 {
		return "insufficient_data", err
	}

	values := result.Data.Result[0].Values
	if len(values) < 2 {
		return "insufficient_data", nil
	}

	// Get first and last values
	firstVal, err := getFloatValue(values[0])
	if err != nil {
		return "error", err
	}

	lastVal, err := getFloatValue(values[len(values)-1])
	if err != nil {
		return "error", err
	}

	// Calculate percentage change
	change := ((lastVal - firstVal) / firstVal) * 100

	if math.Abs(change) < 5 {
		return "stable", nil
	} else if change > 0 {
		return fmt.Sprintf("increasing (+%.1f%%)", change), nil
	} else {
		return fmt.Sprintf("decreasing (%.1f%%)", change), nil
	}
}

// analyzeHealth determines overall status and identifies issues
func (m *MetricsAggregator) analyzeHealth(summary *HealthSummary) {
	issues := 0

	// Check CPU
	if summary.Resources.CPUUsagePercent > 80 {
		summary.Issues = append(summary.Issues, fmt.Sprintf("High CPU usage: %.1f%%", summary.Resources.CPUUsagePercent))
		issues++
	}

	// Check memory
	if summary.Resources.MemoryUsagePercent > 85 {
		summary.Issues = append(summary.Issues, fmt.Sprintf("High memory usage: %.1f%%", summary.Resources.MemoryUsagePercent))
		issues++
	}

	// Check connections
	if summary.Connections.UsagePercent > 80 {
		summary.Issues = append(summary.Issues, fmt.Sprintf("Connection pool near capacity: %.1f%%", summary.Connections.UsagePercent))
		issues++
	}

	// Check page cache
	if summary.Storage.PageCacheHits < 90 {
		summary.Issues = append(summary.Issues, fmt.Sprintf("Low page cache hit rate: %.1f%%", summary.Storage.PageCacheHits))
		issues++
	}

	// Determine overall status
	if issues == 0 {
		summary.OverallStatus = "healthy"
	} else if issues <= 2 {
		summary.OverallStatus = "warning"
	} else {
		summary.OverallStatus = "critical"
	}

	// Generate basic recommendations
	if summary.Resources.CPUUsagePercent > 70 {
		summary.Recommendations = append(summary.Recommendations, "Consider reviewing query performance and optimizing expensive operations")
	}
	if summary.Storage.PageCacheHits < 95 {
		summary.Recommendations = append(summary.Recommendations, "Consider increasing page cache size for better performance")
	}
}

// identifyBottlenecks identifies performance bottlenecks from diagnosis
func (m *MetricsAggregator) identifyBottlenecks(diagnosis *PerformanceDiagnosis) {
	current := diagnosis.CurrentState

	if current.Resources.CPUUsagePercent > 80 {
		diagnosis.Bottlenecks = append(diagnosis.Bottlenecks, "CPU saturation - limiting query throughput")
	}

	if current.Resources.MemoryUsagePercent > 85 {
		diagnosis.Bottlenecks = append(diagnosis.Bottlenecks, "Memory pressure - may cause GC pauses")
	}

	if current.Connections.UsagePercent > 75 {
		diagnosis.Bottlenecks = append(diagnosis.Bottlenecks, "Connection pool saturation - clients may experience delays")
	}

	if current.Storage.PageCacheHits < 90 {
		diagnosis.Bottlenecks = append(diagnosis.Bottlenecks, "Excessive disk I/O - page cache misses degrading performance")
	}
}

// generateRecommendations generates actionable recommendations
func (m *MetricsAggregator) generateRecommendations(diagnosis *PerformanceDiagnosis) {
	current := diagnosis.CurrentState

	// CPU recommendations
	if current.Resources.CPUUsagePercent > 80 {
		diagnosis.Recommendations = append(diagnosis.Recommendations,
			"Immediate: Review and optimize slow queries",
			"Consider: Scale up instance CPU capacity")
	}

	// Memory recommendations
	if current.Resources.MemoryUsagePercent > 85 {
		diagnosis.Recommendations = append(diagnosis.Recommendations,
			"Immediate: Review heap and page cache sizing",
			"Consider: Upgrade to instance with more memory")
	}

	// Connection recommendations
	if current.Connections.UsagePercent > 75 {
		diagnosis.Recommendations = append(diagnosis.Recommendations,
			"Review: Application connection pooling configuration",
			"Consider: Increase connection pool limits if appropriate")
	}

	// Cache recommendations
	if current.Storage.PageCacheHits < 95 {
		diagnosis.Recommendations = append(diagnosis.Recommendations,
			"Increase page cache allocation for better I/O performance",
			"Review working set size vs. available cache")
	}

	// Check for increasing trends
	if cpuTrend, ok := diagnosis.Trends["cpu"]; ok && cpuTrend != "stable" && cpuTrend != "insufficient_data" {
		diagnosis.Recommendations = append(diagnosis.Recommendations,
			fmt.Sprintf("Monitor CPU trend closely: %s", cpuTrend))
	}
}
