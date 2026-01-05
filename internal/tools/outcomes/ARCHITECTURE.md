```
┌─────────────────────────────────────────────────────────────────────────┐
│                     MCP Outcome Pattern Flow                             │
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
     │                                      │  list-outcomes                │
     │                                      │  (outcome_handlers.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  OutcomeRegistry              │
     │                                      │  GetAllSummaries()            │
     │                                      │  (outcome_registry.go)        │
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
     │                                      │  get-outcome-details          │
     │                                      │  (outcome_handlers.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  OutcomeRegistry              │
     │                                      │  GetOutcome("list-instances") │
     │                                      │  (outcome_registry.go)        │
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
     │                                      │  execute-outcome              │
     │                                      │  (outcome_handlers.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  OutcomeRegistry              │
     │                                      │  ExecuteOutcome(...)          │
     │                                      │  (outcome_registry.go)        │
     │                                      └────────────────┬──────────────┘
     │                                                       │
     │                                      ┌────────────────▼──────────────┐
     │                                      │  executeListInstances()       │
     │                                      │  (outcome_registry.go)        │
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
│  outcome_types.go   │  Defines data structures
│  - Outcome          │  - Outcome: Full operation definition
│  - OutcomeSummary   │  - OutcomeSummary: Brief overview
│  - OutcomeParameter │  - OutcomeParameter: Parameter spec
└─────────────────────┘

┌──────────────────────┐
│ outcome_registry.go  │  Central management
│  - NewOutcomeRegistry│  - Registers all outcomes
│  - GetAllSummaries   │  - Returns outcome list
│  - GetOutcome        │  - Returns outcome details
│  - ExecuteOutcome    │  - Routes to execution logic
│  - execute*()        │  - Individual outcome implementations
└──────────────────────┘

┌─────────────────────┐
│ outcome_specs.go    │  MCP Tool Definitions
│  - ListOutcomesSpec │  - Tool: list-outcomes
│  - GetOutcome...    │  - Tool: get-outcome-details
│  - ExecuteOutcome...│  - Tool: execute-outcome
└─────────────────────┘

┌──────────────────────┐
│ outcome_handlers.go  │  MCP Request Handlers
│  - ListOutcomes...   │  - Handles list-outcomes requests
│  - GetOutcomeDetails │  - Handles get-outcome-details requests
│  - ExecuteOutcome... │  - Handles execute-outcome requests
└──────────────────────┘


┌─────────────────────────────────────────────────────────────────────────┐
│                    Adding New Outcomes                                   │
└─────────────────────────────────────────────────────────────────────────┘

Step 1: Define the outcome in outcome_registry.go
┌─────────────────────────────────────────────────────────────────────────┐
│ func (r *OutcomeRegistry) registerNewOutcome() {                        │
│     r.outcomes["new-outcome"] = &Outcome{                               │
│         ID:          "new-outcome",                                      │
│         Name:        "New Outcome",                                      │
│         Description: "...",                                              │
│         Parameters:  []OutcomeParameter{...},                           │
│     }                                                                    │
│ }                                                                        │
└─────────────────────────────────────────────────────────────────────────┘

Step 2: Register it in NewOutcomeRegistry()
┌─────────────────────────────────────────────────────────────────────────┐
│ func NewOutcomeRegistry() *OutcomeRegistry {                            │
│     registry := &OutcomeRegistry{...}                                   │
│     registry.registerListInstancesOutcome()                             │
│     registry.registerNewOutcome()  // ADD THIS                          │
│     return registry                                                     │
│ }                                                                        │
└─────────────────────────────────────────────────────────────────────────┘

Step 3: Implement the execution
┌─────────────────────────────────────────────────────────────────────────┐
│ func executeNewOutcome(ctx, params, deps) (*mcp.CallToolResult, error) {│
│     // Your implementation here                                         │
│     return mcp.NewToolResultText(result), nil                           │
│ }                                                                        │
└─────────────────────────────────────────────────────────────────────────┘

Step 4: Add case in ExecuteOutcome()
┌─────────────────────────────────────────────────────────────────────────┐
│ func (r *OutcomeRegistry) ExecuteOutcome(...) {                         │
│     switch id {                                                          │
│     case "list-instances":                                               │
│         return executeListInstances(...)                                 │
│     case "new-outcome":  // ADD THIS                                     │
│         return executeNewOutcome(...)                                    │
│     }                                                                    │
│ }                                                                        │
└─────────────────────────────────────────────────────────────────────────┘

Done! The three MCP tools automatically expose your new outcome.
```
