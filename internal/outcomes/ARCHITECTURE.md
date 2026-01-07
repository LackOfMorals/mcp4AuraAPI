```
┌─────────────────────────────────────────────────────────────────────────┐
│                     MCP Tool Pattern Flow                             │
└─────────────────────────────────────────────────────────────────────────┘

┌──────────┐
│  Client  │
│ (Claude) │
└────┬─────┘
     │
     │ 1. What operations are available?
     ├──────────────────────────────────────────────────────┐
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  list-Tools                │
     │                                      │  (Tool_handlers.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  ToolRegistry              │
     │                                      │  GetAllSummaries()            │
     │                                      │  (Tool_registry.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │ ◄─────────────────────────────────────────────────────┤
     │ Returns:                                              │
     │ [{"id": "list-instances",                            │
     │   "name": "List Instances", ...}]                    │
     │                                                       │
     │ 2. Tell me more about "list-instances"              │
     ├──────────────────────────────────────────────────────┐
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  get-Tool-details          │
     │                                      │  (Tool_handlers.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  ToolRegistry              │
     │                                      │  GetTool("list-instances") │
     │                                      │  (Tool_registry.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │ ◄─────────────────────────────────────────────────────┤
     │ Returns:                                              │
     │ {"id": "list-instances",                             │
     │  "parameters": [], ...}                              │
     │                                                       │
     │ 3. Execute "list-instances"                          │
     ├──────────────────────────────────────────────────────┐
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  execute-Tool              │
     │                                      │  (Tool_handlers.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  ToolRegistry              │
     │                                      │  ExecuteTool(...)          │
     │                                      │  (Tool_registry.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  executeListInstances()       │
     │                                      │  (Tool_registry.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  AuraAPIClient                │
     │                                      │  - Instances.List()           │
     │                                      │  - Instances.Get(id)          │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │ ◄─────────────────────────────────────────────────────┤
     │ Returns:                                              │
     │ [{"id": "abc", "name": "my-db",                      │
     │   "status": "running", ...}]                         │
     │                                                       │
     └───────────────────────────────────────────────────────┘


┌─────────────────────────────────────────────────────────────────────────┐
│                         Key Components                                   │
└─────────────────────────────────────────────────────────────────────────┘

┌─────────────────────┐
│  Tool_types.go   │  Defines data structures
│  - Tool          │  - Tool: Full operation definition
│  - ToolSummary   │  - ToolSummary: Brief overview
│  - ToolParameter │  - ToolParameter: Parameter spec
└─────────────────────┘

┌──────────────────────┐
│ Tool_registry.go  │  Central management
│  - NewToolRegistry│  - Registers all Tools
│  - GetAllSummaries   │  - Returns Tool list
│  - GetTool        │  - Returns Tool details
│  - ExecuteTool    │  - Routes to execution logic
│  - execute*()        │  - Individual Tool implementations
└──────────────────────┘

┌─────────────────────┐
│ Tool_specs.go    │  MCP Tool Definitions
│  - ListToolsSpec │  - Tool: list-Tools
│  - GetTool...    │  - Tool: get-Tool-details
│  - ExecuteTool...│  - Tool: execute-Tool
└─────────────────────┘

┌──────────────────────┐
│ Tool_handlers.go  │  MCP Request Handlers
│  - ListTools...   │  - Handles list-Tools requests
│  - GetToolDetails │  - Handles get-Tool-details requests
│  - ExecuteTool... │  - Handles execute-Tool requests
└──────────────────────┘


┌─────────────────────────────────────────────────────────────────────────┐
│                    Adding New Tools                                   │
└─────────────────────────────────────────────────────────────────────────┘

Step 1: Define the Tool in Tool_registry.go
┌─────────────────────────────────────────────────────────────────────────┐
│ func (r *ToolRegistry) registerNewTool() {                        │
│     r.Tools["new-Tool"] = &Tool{                               │
│         ID:          "new-Tool",                                      │
│         Name:        "New Tool",                                      │
│         Description: "...",                                              │
│         Parameters:  []ToolParameter{...},                           │
│     }                                                                    │
│ }                                                                        │
└─────────────────────────────────────────────────────────────────────────┘

Step 2: Register it in NewToolRegistry()
┌─────────────────────────────────────────────────────────────────────────┐
│ func NewToolRegistry() *ToolRegistry {                            │
│     registry := &ToolRegistry{...}                                   │
│     registry.registerListInstancesTool()                             │
│     registry.registerNewTool()  // ADD THIS                          │
│     return registry                                                     │
│ }                                                                        │
└─────────────────────────────────────────────────────────────────────────┘

Step 3: Implement the execution
┌─────────────────────────────────────────────────────────────────────────┐
│ func executeNewTool(ctx, params, deps) (*mcp.CallToolResult, error) {│
│     // Your implementation here                                         │
│     return mcp.NewToolResultText(result), nil                           │
│ }                                                                        │
└─────────────────────────────────────────────────────────────────────────┘

Step 4: Add case in ExecuteTool()
┌─────────────────────────────────────────────────────────────────────────┐
│ func (r *ToolRegistry) ExecuteTool(...) {                         │
│     switch id {                                                          │
│     case "list-instances":                                               │
│         return executeListInstances(...)                                 │
│     case "new-Tool":  // ADD THIS                                     │
│         return executeNewTool(...)                                    │
│     }                                                                    │
│ }                                                                        │
└─────────────────────────────────────────────────────────────────────────┘

Done! The three MCP tools automatically expose your new Tool.
```
