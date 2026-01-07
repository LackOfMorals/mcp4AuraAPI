# Tool-Based MCP Tool Pattern

This directory implements a three-tool pattern for MCP operations:

1. **list-Tools** - Lists all available operations
2. **get-Tool-details** - Gets detailed information about a specific operation
3. **execute-Tool** - Executes the operation

## Architecture

### Key Components

- **Tool_types.go** - Defines the core data structures (`Tool`, `ToolSummary`, `ToolParameter`)
- **Tool_registry.go** - Central registry managing all Tools and their execution
- **Tool_specs.go** - MCP tool specifications for the three tools
- **Tool_handlers.go** - Handlers that connect MCP requests to the registry

### Legacy Files

- **list_instances_spec.go** and **list_instances_handler.go** - Original single-tool implementation (can be removed after migration)

## Usage Flow

```
1. User calls list-Tools
   → Returns: [{"id": "list-instances", "name": "List Instances", ...}]

2. User calls get-Tool-details with Tool_id="list-instances"
   → Returns: Full Tool details including parameters

3. User calls execute-Tool with Tool_id="list-instances" and parameters={}
   → Executes the operation and returns results
```

## Adding a New Tool

To add a new Tool (e.g., "create-instance"):

### 1. Register the Tool in Tool_registry.go

```go
func (r *ToolRegistry) registerCreateInstanceTool() {
    r.Tools["create-instance"] = &Tool{
        ID:          "create-instance",
        Name:        "Create Instance",
        Description: "Create a new Neo4j Aura database instance",
        Type:        ToolTypeCreate,
        ReadOnly:    false,
        Parameters: []ToolParameter{
            {
                Name:        "name",
                Type:        "string",
                Description: "Name for the new instance",
                Required:    true,
            },
            // ... more parameters
        },
        Handler: executeCreateInstance, // Point to the handler function
    }
}
```

### 2. Call the Registration in NewToolRegistry()

```go
func NewToolRegistry() *ToolRegistry {
    registry := &ToolRegistry{
        Tools: make(map[string]*Tool),
    }
    
    registry.registerListInstancesTool()
    registry.registerCreateInstanceTool() // Add this line
    
    return registry
}
```

### 3. Implement the Handler Function

The handler must match the `ToolHandler` signature:

```go
// ToolHandler signature:
// func(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error)

func executeCreateInstance(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    // Validate required parameters
    name, ok := parameters["name"].(string)
    if !ok {
        return mcp.NewToolResultError("name parameter is required"), nil
    }
    
    // Execute the operation using deps.AClient
    // ... implementation ...
    
    return mcp.NewToolResultText("Instance created successfully"), nil
}
```

**That's it!** No need to modify ExecuteTool or add switch cases. The handler is automatically called when the Tool is executed.

## Benefits of This Pattern

1. **Token Efficiency** - Users only fetch the information they need
2. **Discoverability** - list-Tools provides a catalog of available operations
3. **Flexibility** - Get details before committing to execution
4. **Extensibility** - Easy to add new Tools without changing the tool interface
5. **Type Safety** - Parameters are clearly defined and validated
6. **Separation of Concerns** - Tool interface is decoupled from implementation

## Migration Note

The original `list-instances` tool is kept for backwards compatibility. Once all clients have migrated to the new pattern, it can be removed from `tools_register.go`.

## Example Session

```
> list-Tools
[
  {
    "id": "list-instances",
    "name": "List Instances",
    "description": "Retrieve a list of all Neo4j Aura database instances...",
    "type": "list",
    "readonly": true
  }
]

> get-Tool-details { "Tool_id": "list-instances" }
{
  "id": "list-instances",
  "name": "List Instances",
  "description": "Retrieve a list of all Neo4j Aura database instances...",
  "type": "list",
  "readonly": true,
  "parameters": []
}

> execute-Tool { "Tool_id": "list-instances", "parameters": {} }
[
  {
    "id": "abc123",
    "name": "my-database",
    "status": "running",
    ...
  }
]
```
