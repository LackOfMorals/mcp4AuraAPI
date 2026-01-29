# Quick Start - Prometheus Monitoring

## Getting Started

The MCP server now includes Phase 1 Prometheus monitoring capabilities. Here's how to use them:

## 1. Discover Available Outcomes

```
list-outcomes
```

You'll see 7 outcomes including 4 new monitoring ones:
- get-instance-health
- diagnose-performance  
- analyze-resource-usage
- get-query-statistics

## 2. Get Outcome Details

```
get-outcome-details
  Outcome_id: "get-instance-health"
```

Shows required parameters:
- `instance_id` - Your Aura instance ID
- `prometheus_url` - The Prometheus endpoint URL

## 3. Execute Monitoring Outcomes

### Quick Health Check

```
execute-outcome
  Outcome_id: "get-instance-health"
  parameters: {
    "instance_id": "your-instance-id",
    "prometheus_url": "https://your-instance.metrics.neo4j.io/prometheus"
  }
```

**Returns:** Current CPU, memory, query rate, connections, cache hit rate, overall status, and recommendations.

**Use When:** 
- Quick status check
- Dashboard display
- Automated monitoring

---

### Performance Diagnosis

```
execute-outcome
  Outcome_id: "diagnose-performance"
  parameters: {
    "instance_id": "your-instance-id",
    "prometheus_url": "https://your-instance.metrics.neo4j.io/prometheus",
    "window_minutes": 60
  }
```

**Returns:** Current state + trends + bottlenecks + recommendations

**Use When:**
- Database is slow
- Investigating issues
- Understanding patterns

---

### Resource Analysis

```
execute-outcome
  Outcome_id: "analyze-resource-usage"
  parameters: {
    "instance_id": "your-instance-id",
    "prometheus_url": "https://your-instance.metrics.neo4j.io/prometheus"
  }
```

**Returns:** Resource utilization + capacity assessment + scaling recommendations

**Use When:**
- Capacity planning
- Right-sizing decisions
- Cost optimization

---

### Query Statistics

```
execute-outcome
  Outcome_id: "get-query-statistics"
  parameters: {
    "instance_id": "your-instance-id",
    "prometheus_url": "https://your-instance.metrics.neo4j.io/prometheus"
  }
```

**Returns:** Query throughput + latency + performance assessment + optimization advice

**Use When:**
- Query optimization
- Performance benchmarking
- SLA monitoring

## Example Session with Claude

```
You: "Check the health of my database abc123"

Claude: [Discovers get-instance-health outcome]
        "I can check your database health. What's the Prometheus endpoint URL?"

You: "https://abc123.metrics.neo4j.io/prometheus"

Claude: [Executes get-instance-health]
        "Your database is in WARNING status:
        
        Resources:
        - CPU: 78% (high)
        - Memory: 65% (normal)
        
        Queries: 1250/sec, avg latency 125ms
        Connections: 85/100 (85% usage - near capacity)
        
        Issues:
        - High CPU usage limiting throughput
        - Connection pool near capacity
        
        Recommendations:
        - Review slow queries
        - Consider scaling up CPU
        - Increase connection pool if appropriate"
```

## Getting Prometheus URL

For Neo4j Aura instances:
1. Log into Aura Console
2. Select your instance
3. Go to "Monitoring" tab
4. Copy the Prometheus endpoint URL

Format: `https://<instance-id>.metrics.neo4j.io/prometheus`

## Common Patterns

### "Is everything OK?"
→ Use `get-instance-health`
→ Check overall_status field
→ Review issues array

### "Why is it slow?"
→ Use `diagnose-performance` with 60+ minute window
→ Check bottlenecks array
→ Review trends to see if degrading

### "Should I scale?"
→ Use `analyze-resource-usage`
→ Check utilization_assessment
→ Review recommendations

### "Are queries slow?"
→ Use `get-query-statistics`
→ Check assessment
→ Follow optimization advice

## Error Handling

If you see errors:

```json
{
  "error": "Failed to retrieve instance health: ..."
}
```

**Common Issues:**
- Invalid Prometheus URL (check format)
- Network connectivity 
- Instance not exposing metrics
- Authentication failure

**Debug Steps:**
1. Verify Prometheus URL is correct
2. Check instance is running
3. Test URL in browser (should require auth)
4. Check network connectivity

## Tips

1. **Start with health check** - Quick overview before deep dive
2. **Use longer windows for trends** - 120-240 minutes for pattern detection
3. **Check regularly** - Set up periodic health checks
4. **Act on recommendations** - They're specific and actionable
5. **Compare over time** - Track metrics to see improvements

## Integration Example

```bash
#!/bin/bash
# Daily health check script

INSTANCE_ID="your-instance-id"
PROMETHEUS_URL="https://your-instance.metrics.neo4j.io/prometheus"

# Execute health check via MCP
echo '{
  "jsonrpc": "2.0",
  "method": "tools/call",
  "params": {
    "name": "execute-outcome",
    "arguments": {
      "Outcome_id": "get-instance-health",
      "parameters": {
        "instance_id": "'$INSTANCE_ID'",
        "prometheus_url": "'$PROMETHEUS_URL'"
      }
    }
  }
}' | ./bin/mcp-aura-api
```

## Next Steps

- See [PROMETHEUS_INTEGRATION.md](PROMETHEUS_INTEGRATION.md) for detailed documentation
- See [PHASE1_IMPLEMENTATION_SUMMARY.md](PHASE1_IMPLEMENTATION_SUMMARY.md) for architecture details
- Test with MCP Inspector for interactive exploration
- Build custom monitoring dashboards
- Set up automated health checks
