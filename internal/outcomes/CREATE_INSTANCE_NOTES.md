# Create Instance Implementation Notes

## Implementation Status

The `create-instance` Tool has been added with the following configuration:

### Required Parameters
- **name**: Name for the new instance
- **cloud_provider**: 'gcp', 'aws', or 'azure'
- **region**: Cloud region (e.g., 'us-east-1' for AWS, 'us-central1' for GCP, 'eastus' for Azure)
- **memory**: Memory size (e.g., '2GB', '8GB', '16GB', '32GB', '64GB')
- **type**: 'free', 'professional', or 'enterprise'

### Optional Parameters
- **version**: Neo4j version (defaults to "5")

## Important: API Call Structure

The implementation assumes the Aura API client has the following method signature:

```go
deps.AClient.Instances.Create(name, cloudProvider, region, memory, type, version)
```

**You may need to adjust this** based on the actual `aura-client` library API. Common patterns include:

### Option 1: Multiple Parameters (Current Implementation)
```go
instance, err := deps.AClient.Instances.Create(
    req.Name, 
    req.CloudProvider, 
    req.Region, 
    req.Memory, 
    req.Type, 
    req.Version,
)
```

### Option 2: Struct Parameter
```go
instance, err := deps.AClient.Instances.Create(&aura.CreateInstanceRequest{
    Name:          req.Name,
    CloudProvider: req.CloudProvider,
    Region:        req.Region,
    Memory:        req.Memory,
    Type:          req.Type,
    Version:       req.Version,
})
```

### Option 3: Map/Options Pattern
```go
instance, err := deps.AClient.Instances.Create(map[string]string{
    "name":           req.Name,
    "cloud_provider": req.CloudProvider,
    "region":         req.Region,
    "memory":         req.Memory,
    "type":           req.Type,
    "version":        req.Version,
})
```

## How to Find the Correct API

### Check the aura-client Library
```bash
# View the library code
go list -m -json github.com/LackOfMorals/aura-client

# Or check the source
cd $GOPATH/pkg/mod/github.com/\!lack\!of\!morals/aura-client@v1.0.4
```

### Check Documentation
Look for the aura-client documentation at:
- https://github.com/LackOfMorals/aura-client
- https://pkg.go.dev/github.com/LackOfMorals/aura-client

### Test First
Before deploying, test the create functionality:

```go
// In a test file
package Tools

import (
    "testing"
    "github.com/LackOfMorals/aura-client"
)

func TestCreateInstance(t *testing.T) {
    client, _ := aura.NewClient(
        aura.WithCredentials("client_id", "client_secret"),
    )
    
    // Try to call the Create method and see what signature works
    instance, err := client.Instances.Create(...)
    // Adjust based on compilation errors
}
```

## Expected Response Structure

The implementation expects the API to return an object with:

```go
instance.Data.Id
instance.Data.Name
instance.Data.Status
instance.Data.CloudProvider
instance.Data.Memory
instance.Data.Type
instance.Data.ConnectionUrl
```

This matches the structure used in `list-instances`, so it should be consistent.

## Testing the Tool

Once you've verified/adjusted the API call:

```bash
# 1. Build
go build -o bin/mcp-aura-api ./cmd/mcp-aura-api

# 2. Test with MCP Inspector
npx @modelcontextprotocol/inspector ./bin/mcp-aura-api

# 3. List Tools to see create-instance
# 4. Get details: get-Tool-details with Tool_id="create-instance"
# 5. Execute: execute-Tool with Tool_id="create-instance" and parameters={...}
```

## Example Execution

```json
{
  "Tool_id": "create-instance",
  "parameters": {
    "name": "my-test-db",
    "cloud_provider": "gcp",
    "region": "us-central1",
    "memory": "2GB",
    "type": "professional",
    "version": "5"
  }
}
```

Expected result:
```json
{
  "success": true,
  "message": "Instance created successfully",
  "id": "abc123...",
  "name": "my-test-db",
  "status": "creating",
  "cloud_provider": "gcp",
  "memory": "2GB",
  "type": "professional",
  "url": "neo4j+s://..."
}
```

## Validation

The implementation includes validation for:
- ✅ Required parameters (name, cloud_provider, region, memory, type)
- ✅ Valid cloud providers (gcp, aws, azure)
- ✅ Valid instance types (free, professional, enterprise)
- ✅ Non-empty string values
- ⚠️ Memory size format (not validated - assumes API will handle)
- ⚠️ Region validity (not validated - assumes API will handle)

Additional validation could be added based on API requirements.

## Next Steps

1. **Verify API Call**: Check the actual aura-client library to confirm the Create method signature
2. **Adjust if Needed**: Update line 272 in `Tool_registry.go` with the correct API call
3. **Test**: Build and test with real credentials
4. **Add More Validations**: Consider adding region/memory validation if needed
5. **Documentation**: Update user-facing docs once tested
