# Improved Workflow with get-instance-details

## The Better Way

Instead of requiring users to manually provide the Prometheus URL, we now have a two-step workflow:

## Step 1: Get Instance Details
```json
{
  "Outcome_id": "get-instance-details",
  "parameters": {
    "instance_id": "abc123"
  }
}
```

**Returns:**
```json
{
  "id": "abc123",
  "name": "my-database",
  "status": "running",
  "connection_url": "neo4j+s://abc123.databases.neo4j.io",
  "cloud_provider": "gcp",
  "region": "us-central1",
  "memory": "8GB",
  "type": "professional-db",
  "tenant_id": "xyz789",
  "prometheus_url": "https://abc123.metrics.neo4j.io/prometheus"
}
```

## Step 2: Use Prometheus URL for Monitoring

```json
{
  "Outcome_id": "get-instance-health",
  "parameters": {
    "instance_id": "abc123",
    "prometheus_url": "https://abc123.metrics.neo4j.io/prometheus"
  }
}
```

## With Claude - Natural Flow

```
User: "Check the health of my database abc123"

Claude: [Calls get-instance-details with "abc123"]
        [Extracts prometheus_url from response]
        [Calls get-instance-health with both values]
        
        "Your database 'my-database' is healthy:
        - CPU: 45%
        - Memory: 60%
        - All systems normal"
```

## Benefits

✅ **No manual URL lookup** - Automatically fetched from Aura API
✅ **Single source of truth** - Always get the correct URL
✅ **Simpler for users** - Just provide instance ID
✅ **More reliable** - No typos in Prometheus URLs
✅ **Works for all instances** - Even newly created ones

## Updated Outcome Count

We now have **8 outcomes total**:

**Instance Management:**
1. list-instances
2. **get-instance-details** ← NEW!
3. create-instance
4. delete-instance

**Prometheus Monitoring:**
5. get-instance-health
6. diagnose-performance
7. analyze-resource-usage
8. get-query-statistics

## Example: Complete Health Check

```bash
# 1. Get instance details (including Prometheus URL)
execute-outcome --Outcome_id get-instance-details \
  --parameters '{"instance_id": "abc123"}'

# Response includes: "prometheus_url": "https://abc123.metrics.neo4j.io/prometheus"

# 2. Check health using that URL
execute-outcome --Outcome_id get-instance-health \
  --parameters '{
    "instance_id": "abc123",
    "prometheus_url": "https://abc123.metrics.neo4j.io/prometheus"
  }'
```

## For Developers

The `get-instance-details` outcome:
- Takes only `instance_id` parameter
- Calls Aura API to get full instance information
- Automatically constructs Prometheus URL from instance ID
- Returns all relevant details in one call

## Prometheus URL Format

For Neo4j Aura instances, the Prometheus endpoint follows this pattern:
```
https://{instance_id}.metrics.neo4j.io/prometheus
```

The `get-instance-details` outcome automatically constructs this URL, so you don't have to remember the format or risk typos.
