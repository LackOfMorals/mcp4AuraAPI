# Prometheus Integration for Neo4j Aura MCP Server

## Overview

Phase 1 of Prometheus integration adds read-only diagnostics outcomes that provide comprehensive monitoring and performance analysis for Neo4j Aura instances using their Prometheus endpoints.

## Architecture

The integration follows the existing outcome-based pattern with three layers:

### 1. Prometheus Client (`internal/prometheus/client.go`)
Low-level HTTP client for Prometheus API:
- Instant queries (`Query`)
- Range queries (`QueryRange`)
- Result parsing and error handling

### 2. Metrics Aggregator (`internal/prometheus/aggregator.go`)
High-level metrics aggregation and analysis:
- Pre-aggregated health summaries
- Trend analysis
- Bottleneck identification
- Actionable recommendations

### 3. Outcomes (`internal/server/outcome_registry.go`)
Four new outcomes integrated with the registry:
- `get-instance-health` - Comprehensive health snapshot
- `diagnose-performance` - Detailed performance analysis with trends
- `analyze-resource-usage` - Capacity planning and optimization
- `get-query-statistics` - Query performance metrics

## Available Outcomes

### get-instance-health

**Purpose:** Quick health check with current metrics

**Parameters:**
- `instance_id` (required): Instance identifier
- `prometheus_url` (required): Prometheus endpoint URL

**Returns:**
```json
{
  "instance_id": "abc123",
  "timestamp": "2025-01-29T10:30:00Z",
  "resources": {
    "cpu_usage_percent": 45.2,
    "memory_usage_percent": 68.5
  },
  "query": {
    "queries_per_second": 1250.5,
    "avg_latency_ms": 45.2
  },
  "connections": {
    "active_connections": 42,
    "max_connections": 100,
    "usage_percent": 42.0
  },
  "storage": {
    "page_cache_hits_percent": 95.8
  },
  "overall_status": "healthy",
  "issues": [],
  "recommendations": []
}
```

**Use Cases:**
- Quick health checks
- Dashboard displays
- Automated monitoring alerts

---

### diagnose-performance

**Purpose:** Deep performance analysis with historical context

**Parameters:**
- `instance_id` (required): Instance identifier
- `prometheus_url` (required): Prometheus endpoint URL
- `window_minutes` (optional, default 60): Analysis time window (1-1440)

**Returns:**
```json
{
  "instance_id": "abc123",
  "analysis_window": "Last 60 minutes",
  "current_state": { /* health summary */ },
  "trends": {
    "cpu": "increasing (+15%)",
    "memory": "stable",
    "query_rate": "increasing (+22%)"
  },
  "bottlenecks": [
    "CPU saturation - limiting query throughput",
    "Connection pool saturation - clients may experience delays"
  ],
  "slow_queries": [],
  "recommendations": [
    "Immediate: Review and optimize slow queries",
    "Consider: Scale up instance CPU capacity"
  ]
}
```

**Use Cases:**
- Troubleshooting performance degradation
- Understanding trend patterns
- Root cause analysis

---

### analyze-resource-usage

**Purpose:** Resource utilization for capacity planning

**Parameters:**
- `instance_id` (required): Instance identifier
- `prometheus_url` (required): Prometheus endpoint URL

**Returns:**
```json
{
  "instance_id": "abc123",
  "timestamp": "2025-01-29T10:30:00Z",
  "resources": {
    "cpu_usage_percent": 45.2,
    "memory_usage_percent": 68.5
  },
  "storage": {
    "page_cache_hits_percent": 95.8
  },
  "utilization_assessment": "optimal",
  "recommendations": []
}
```

**Utilization Levels:**
- `underutilized` - <50% max resource
- `optimal` - 50-70% max resource
- `high` - 70-85% max resource
- `critical` - >85% max resource

**Use Cases:**
- Right-sizing decisions
- Cost optimization
- Capacity planning

---

### get-query-statistics

**Purpose:** Query performance metrics

**Parameters:**
- `instance_id` (required): Instance identifier
- `prometheus_url` (required): Prometheus endpoint URL

**Returns:**
```json
{
  "instance_id": "abc123",
  "timestamp": "2025-01-29T10:30:00Z",
  "query_metrics": {
    "queries_per_second": 1250.5,
    "avg_latency_ms": 45.2
  },
  "assessment": "good",
  "advice": []
}
```

**Assessment Levels:**
- `excellent` - <50ms avg latency
- `good` - 50-200ms
- `moderate` - 200-500ms
- `poor` - >500ms

**Use Cases:**
- Query optimization
- Performance benchmarking
- SLA monitoring

## Usage Examples

### Using with Claude

```
"Can you check the health of my database?"
→ Claude calls list-outcomes to see available operations
→ Claude calls execute-outcome with get-instance-health
→ Claude provides natural language summary

"My database seems slow today. Can you diagnose it?"
→ Claude calls diagnose-performance with 60-minute window
→ Claude analyzes trends and bottlenecks
→ Claude provides actionable recommendations
```

### Direct Tool Invocation

```bash
# List available outcomes
execute-outcome list-outcomes

# Get instance health
execute-outcome execute-outcome \
  --Outcome_id get-instance-health \
  --parameters '{
    "instance_id": "abc123",
    "prometheus_url": "https://metrics.aura.neo4j.io/prometheus"
  }'

# Diagnose performance over 2 hours
execute-outcome execute-outcome \
  --Outcome_id diagnose-performance \
  --parameters '{
    "instance_id": "abc123",
    "prometheus_url": "https://metrics.aura.neo4j.io/prometheus",
    "window_minutes": 120
  }'
```

## Design Principles

### 1. Outcome-Based Pattern
All operations follow the existing three-tool pattern:
- Discovery via `list-outcomes`
- Details via `get-outcome-details`
- Execution via `execute-outcome`

### 2. Token Efficiency
- Server-side aggregation reduces response size
- Pre-calculated summaries avoid raw time series
- Intelligent defaults (5-minute windows, summary statistics)

### 3. Actionable Insights
- Not just metrics, but interpretation
- Clear status indicators (healthy/warning/critical)
- Specific, actionable recommendations

### 4. Progressive Disclosure
- Quick health checks for dashboards
- Detailed diagnosis for troubleshooting
- Specialized views for specific needs

## Neo4j Prometheus Metrics Used

### Core Metrics
- `process_cpu_seconds_total` - CPU usage
- `neo4j_memory_heap_used_bytes` / `neo4j_memory_heap_max_bytes` - Memory
- `neo4j_transaction_started_total` - Query throughput
- `neo4j_bolt_connections_*` - Connection pool
- `neo4j_page_cache_*` - Storage I/O efficiency

### Calculated Metrics
- CPU usage: `rate(process_cpu_seconds_total[5m]) * 100`
- Memory usage: `(heap_used / heap_max) * 100`
- Query rate: `rate(neo4j_transaction_started_total[5m])`
- Cache hit rate: `hits / (hits + faults) * 100`

## Future Enhancements (Phase 2+)

### Phase 2 - Alerting
- `configure-alerts` - Set up threshold-based alerts
- `check-alert-conditions` - Verify alert configurations
- `get-alert-history` - Review triggered alerts

### Phase 3 - Predictive Analytics
- `predict-capacity-needs` - Forecast resource requirements
- `detect-anomalies` - ML-based anomaly detection
- `forecast-resource-usage` - Trend-based predictions

### Phase 4 - Advanced Analysis
- Slow query pattern analysis
- Index recommendation engine
- Workload characterization

## Error Handling

All outcomes handle errors gracefully:

```json
{
  "error": "Failed to retrieve instance health: connection refused"
}
```

Common error scenarios:
- Invalid Prometheus URL
- Network connectivity issues
- Missing metrics (instance not exposing data)
- Invalid time windows

## Performance Considerations

### Query Optimization
- Use 5-minute rate windows for stability
- Instant queries for current state
- Range queries only for trend analysis

### Caching Strategy
Metrics have different cache lifetimes:
- Current state: 30 seconds
- Historical trends: 5 minutes
- Aggregated summaries: 1 minute

### Rate Limiting
- Respect Prometheus query limits
- Batch related metrics when possible
- Use pre-aggregated recording rules where available

## Testing

```bash
# Build
go build -o ./bin/mcp-aura-api ./cmd/mcp-aura-api

# Test with MCP Inspector
npx @modelcontextprotocol/inspector ./bin/mcp-aura-api

# Example test queries
1. list-outcomes → See new monitoring outcomes
2. get-outcome-details with "get-instance-health" → See parameters
3. execute-outcome with actual Prometheus URL → Get real metrics
```

## Integration Checklist

- [x] Prometheus client implementation
- [x] Metrics aggregation layer
- [x] Four Phase 1 outcomes registered
- [x] Type-safe parameter validation
- [x] Comprehensive error handling
- [x] Token-efficient responses
- [ ] Real Prometheus endpoint testing
- [ ] Documentation in README
- [ ] Integration tests
- [ ] Example queries

## Contributing

When adding new Prometheus outcomes:

1. Define metrics queries in `aggregator.go`
2. Create outcome registration in `outcome_registry.go`
3. Implement handler function
4. Add to registry in `NewOutcomeRegistry()`
5. Update documentation
6. Add integration tests

## Questions?

- How to get Prometheus URL for an Aura instance?
- What metrics are available by default?
- How to configure custom metrics?
- Performance tuning recommendations?
