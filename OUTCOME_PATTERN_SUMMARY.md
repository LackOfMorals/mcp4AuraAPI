# MCP Outcome Pattern - Implementation Summary

## What Was Created

### New Files

1. **outcome_types.go** - Core data structures for outcomes
2. **outcome_registry.go** - Central registry managing all outcomes
3. **outcome_specs.go** - MCP tool specifications for the three tools
4. **outcome_handlers.go** - Request handlers for the MCP tools
5. **README.md** - Complete documentation and usage guide

### Modified Files

1. **tools_register.go** - Updated to register the three new tools

### New MCP Tools

The server now exposes these tools:

1. **list-outcomes** - Lists all available operations (read-only)
2. **get-outcome-details** - Gets details for a specific outcome (read-only)
3. **execute-outcome** - Executes an outcome (may be read-only depending on the outcome)
4. **list-instances** - Legacy tool (kept for backwards compatibility)

## Quick Start

### Build and Run

```bash
# From the project root
go build -o bin/mcp-aura-api ./cmd/mcp-aura-api
./bin/mcp-aura-api
```

### Test the Tools

Once the server is running, you can test the new pattern:

```bash
# 1. List available outcomes
echo '{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"list-outcomes","arguments":{}}}' | ./bin/mcp-aura-api

# 2. Get details for a specific outcome
echo '{"jsonrpc":"2.0","id":2,"method":"tools/call","params":{"name":"get-outcome-details","arguments":{"outcome_id":"list-instances"}}}' | ./bin/mcp-aura-api

# 3. Execute the outcome
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"execute-outcome","arguments":{"outcome_id":"list-instances","parameters":{}}}}' | ./bin/mcp-aura-api
```

## Architecture Benefits

### Before (Single Tool Pattern)
- One monolithic tool per operation
- User must know exact tool name
- Token-heavy for discovery

### After (Outcome Pattern)
- Three generic tools serving multiple outcomes
- Self-documenting through list-outcomes
- Token-efficient (only fetch what you need)
- Easy to extend (just add to registry)

## Adding New Outcomes

All new outcomes are added to `outcome_registry.go`:

1. Create a `register<Outcome>()` method
2. Call it in `NewOutcomeRegistry()`
3. Implement the execution logic
4. Add a case in `ExecuteOutcome()`

No changes needed to tool specs or handlers!

## Migration Path

### Phase 1: Both Patterns (Current)
- New three-tool pattern available
- Legacy `list-instances` tool still works
- Clients can migrate at their own pace

### Phase 2: New Pattern Only (Future)
- Remove legacy tool from `tools_register.go`
- All clients use outcome pattern
- Simpler maintenance

## Example: Adding "Get Instance Details" Outcome

```go
// 1. In NewOutcomeRegistry()
func NewOutcomeRegistry() *OutcomeRegistry {
    registry := &OutcomeRegistry{
        outcomes: make(map[string]*Outcome),
    }
    
    registry.registerListInstancesOutcome()
    registry.registerGetInstanceDetailsOutcome() // NEW
    
    return registry
}

// 2. Add the registration method
func (r *OutcomeRegistry) registerGetInstanceDetailsOutcome() {
    r.outcomes["get-instance-details"] = &Outcome{
        ID:          "get-instance-details",
        Name:        "Get Instance Details",
        Description: "Get detailed information about a specific Neo4j Aura instance",
        Type:        OutcomeTypeRead,
        ReadOnly:    true,
        Parameters: []OutcomeParameter{
            {
                Name:        "instance_id",
                Type:        "string",
                Description: "The ID of the instance to retrieve",
                Required:    true,
            },
        },
    }
}

// 3. Add execution logic
func executeGetInstanceDetails(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    instanceID, ok := parameters["instance_id"].(string)
    if !ok {
        return mcp.NewToolResultError("instance_id parameter is required"), nil
    }
    
    instanceInfo, err := deps.AClient.Instances.Get(instanceID)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("Failed to get instance: %v", err)), nil
    }
    
    jsonData, _ := json.MarshalIndent(instanceInfo.Data, "", "  ")
    return mcp.NewToolResultText(string(jsonData)), nil
}

// 4. Add case in ExecuteOutcome()
switch id {
case "list-instances":
    return executeListInstances(ctx, deps)
case "get-instance-details":
    return executeGetInstanceDetails(ctx, parameters, deps)
default:
    return mcp.NewToolResultError(fmt.Sprintf("execution not implemented for outcome: %s", id)), nil
}
```

That's it! No tool specs or handlers to update.

## Token Efficiency Example

### Old Pattern (list-instances tool)
```
User: "What can I do with Aura instances?"
Claude: Calls list-instances tool
Result: Full JSON array of all instances with all details
Tokens: ~500-1000+ depending on instance count
```

### New Pattern
```
User: "What can I do with Aura instances?"
Claude: Calls list-outcomes
Result: Just outcome summaries
Tokens: ~100-200

User: "Tell me about list-instances"
Claude: Calls get-outcome-details with id="list-instances"
Result: Just that outcome's details
Tokens: ~50-100

User: "List my instances"
Claude: Calls execute-outcome with id="list-instances"
Result: Full instance data
Tokens: ~500-1000+ as needed
```

**Total: Same data, but only when needed!**

## Next Steps

1. **Verify the Create API Call**: Check `internal/tools/outcomes/CREATE_INSTANCE_NOTES.md` for details on the API call that may need adjustment
2. Test the three new tools and the create-instance outcome
3. Consider adding more outcomes:
   - get-instance-details
   - delete-instance
   - pause-instance
   - resume-instance
   - snapshot-instance
4. Once satisfied, remove legacy `list-instances` tool
5. Document for end users

## Recently Added

### create-instance Outcome ✨

Added a new outcome for creating Neo4j Aura instances with the following parameters:
- **name**: Instance name (required)
- **cloud_provider**: 'gcp', 'aws', or 'azure' (required)
- **region**: Cloud region (required)
- **memory**: Memory size like '2GB', '8GB' (required)
- **type**: 'free', 'professional', or 'enterprise' (required)
- **version**: Neo4j version (optional, defaults to '5')

**Important**: The API call structure in `executeCreateInstance()` may need adjustment based on the actual aura-client library API. See `CREATE_INSTANCE_NOTES.md` for details.

## Questions?

See `README.md` in this directory for detailed documentation.
