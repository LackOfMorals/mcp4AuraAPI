# Outcome-Based MCP Tool Pattern

This directory implements a three-tool pattern for MCP operations:

1. **list-outcomes** - Lists all available operations
2. **get-outcome-details** - Gets detailed information about a specific operation
3. **execute-outcome** - Executes the operation

## Architecture

### Key Components

- **outcome_types.go** - Defines the core data structures (`Outcome`, `OutcomeSummary`, `OutcomeParameter`)
- **outcome_registry.go** - Central registry managing all outcomes and their execution
- **outcome_specs.go** - MCP tool specifications for the three tools
- **outcome_handlers.go** - Handlers that connect MCP requests to the registry

### Legacy Files

- **list_instances_spec.go** and **list_instances_handler.go** - Original single-tool implementation (can be removed after migration)

## Usage Flow

```
1. User calls list-outcomes
   → Returns: [{"id": "list-instances", "name": "List Instances", ...}]

2. User calls get-outcome-details with outcome_id="list-instances"
   → Returns: Full outcome details including parameters

3. User calls execute-outcome with outcome_id="list-instances" and parameters={}
   → Executes the operation and returns results
```

## Adding a New Outcome

To add a new outcome (e.g., "create-instance"):

### 1. Register the Outcome in outcome_registry.go

```go
func (r *OutcomeRegistry) registerCreateInstanceOutcome() {
    r.outcomes["create-instance"] = &Outcome{
        ID:          "create-instance",
        Name:        "Create Instance",
        Description: "Create a new Neo4j Aura database instance",
        Type:        OutcomeTypeCreate,
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
                Description: "Cloud provider (gcp, aws, azure)",
                Required:    true,
            },
            {
                Name:        "memory",
                Type:        "string",
                Description: "Memory size (e.g., '2GB', '8GB')",
                Required:    true,
            },
        },
    }
}
```

### 2. Call the Registration in NewOutcomeRegistry()

```go
func NewOutcomeRegistry() *OutcomeRegistry {
    registry := &OutcomeRegistry{
        outcomes: make(map[string]*Outcome),
    }
    
    registry.registerListInstancesOutcome()
    registry.registerCreateInstanceOutcome() // Add this line
    
    return registry
}
```

### 3. Implement the Execution Logic

```go
func executeCreateInstance(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    // Validate required parameters
    name, ok := parameters["name"].(string)
    if !ok {
        return mcp.NewToolResultError("name parameter is required"), nil
    }
    
    cloudProvider, ok := parameters["cloud_provider"].(string)
    if !ok {
        return mcp.NewToolResultError("cloud_provider parameter is required"), nil
    }
    
    memory, ok := parameters["memory"].(string)
    if !ok {
        return mcp.NewToolResultError("memory parameter is required"), nil
    }
    
    // Execute the operation using deps.AClient
    // ... implementation ...
    
    return mcp.NewToolResultText("Instance created successfully"), nil
}
```

### 4. Add the Case in ExecuteOutcome()

```go
func (r *OutcomeRegistry) ExecuteOutcome(ctx context.Context, id string, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    outcome, err := r.GetOutcome(id)
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }

    switch id {
    case "list-instances":
        return executeListInstances(ctx, deps)
    case "create-instance":
        return executeCreateInstance(ctx, parameters, deps)
    // Add more cases here
    default:
        return mcp.NewToolResultError(fmt.Sprintf("execution not implemented for outcome: %s", id)), nil
    }
}
```

## Benefits of This Pattern

1. **Token Efficiency** - Users only fetch the information they need
2. **Discoverability** - list-outcomes provides a catalog of available operations
3. **Flexibility** - Get details before committing to execution
4. **Extensibility** - Easy to add new outcomes without changing the tool interface
5. **Type Safety** - Parameters are clearly defined and validated
6. **Separation of Concerns** - Tool interface is decoupled from implementation

## Migration Note

The original `list-instances` tool is kept for backwards compatibility. Once all clients have migrated to the new pattern, it can be removed from `tools_register.go`.

## Example Session

```
> list-outcomes
[
  {
    "id": "list-instances",
    "name": "List Instances",
    "description": "Retrieve a list of all Neo4j Aura database instances...",
    "type": "list",
    "readonly": true
  }
]

> get-outcome-details { "outcome_id": "list-instances" }
{
  "id": "list-instances",
  "name": "List Instances",
  "description": "Retrieve a list of all Neo4j Aura database instances...",
  "type": "list",
  "readonly": true,
  "parameters": []
}

> execute-outcome { "outcome_id": "list-instances", "parameters": {} }
[
  {
    "id": "abc123",
    "name": "my-database",
    "status": "running",
    ...
  }
]
```
