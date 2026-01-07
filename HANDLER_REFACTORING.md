# Handler Function Refactoring

## What Changed

Refactored the Tool execution pattern to use handler functions instead of a switch statement.

## Before (Switch Statement Pattern)

### Problems
1. **Tight Coupling**: ExecuteTool needed to know about every Tool
2. **Maintenance Overhead**: Every new Tool required modifying ExecuteTool
3. **Switch Statement**: Required manual routing logic
4. **Inconsistent Signatures**: Different handlers had different signatures

### Old Code Structure

```go
// ExecuteTool with switch statement
func (r *ToolRegistry) ExecuteTool(...) (*mcp.CallToolResult, error) {
    Tool, err := r.GetTool(id)
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }

    // Check read-only mode
    if !Tool.ReadOnly && deps.Config != nil && deps.Config.ReadOnly {
        return mcp.NewToolResultError(...), nil
    }

    // Manual routing with switch statement ❌
    switch id {
    case "list-instances":
        return executeListInstances(ctx, deps)  // Different signature!
    case "create-instance":
        return executeCreateInstance(ctx, parameters, deps)
    case "delete-instance":
        return executeDeleteInstance(ctx, parameters, deps)
    // Need to add case for every new Tool
    default:
        return mcp.NewToolResultError(...), nil
    }
}

// Handler functions had inconsistent signatures ❌
func executeListInstances(ctx context.Context, deps *tools.ToolDependencies) (...)
func executeCreateInstance(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (...)
```

---

## After (Handler Function Pattern)

### Improvements
1. **Loose Coupling**: ExecuteTool just calls Tool.Handler
2. **Zero Maintenance**: New Tools don't require touching ExecuteTool
3. **Automatic Routing**: Handler is invoked directly
4. **Consistent Signatures**: All handlers use ToolHandler type

### New Code Structure

```go
// Define handler function type
type ToolHandler func(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error)

// Tool includes handler
type Tool struct {
    ID          string
    Name        string
    Description string
    Type        ToolType
    ReadOnly    bool
    Parameters  []ToolParameter
    Metadata    map[string]interface{}
    Handler     ToolHandler  // ✅ Handler function stored here
}

// ExecuteTool is now simple and generic ✅
func (r *ToolRegistry) ExecuteTool(ctx context.Context, id string, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    Tool, err := r.GetTool(id)
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }

    // Check read-only mode
    if !Tool.ReadOnly && deps.Config != nil && deps.Config.ReadOnly {
        return mcp.NewToolResultError(...), nil
    }

    // Execute the handler - no switch needed! ✅
    if Tool.Handler == nil {
        return mcp.NewToolResultError(fmt.Sprintf("no handler registered for Tool: %s", id)), nil
    }

    return Tool.Handler(ctx, parameters, deps)
}

// All handlers now have consistent signature ✅
func executeListInstances(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    // Implementation
}

func executeCreateInstance(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    // Implementation
}

func executeDeleteInstance(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    // Implementation
}
```

---

## Adding a New Tool: Comparison

### Before (4 steps)
```go
// 1. Register Tool
func (r *ToolRegistry) registerNewTool() {
    r.Tools["new-Tool"] = &Tool{...}
}

// 2. Call registration
func NewToolRegistry() *ToolRegistry {
    registry.registerNewTool()
}

// 3. Implement handler
func executeNewTool(...) (*mcp.CallToolResult, error) {
    // Implementation
}

// 4. Add to switch statement ❌
func (r *ToolRegistry) ExecuteTool(...) {
    switch id {
    case "list-instances":
        return executeListInstances(ctx, deps)
    case "new-Tool":  // ADD THIS
        return executeNewTool(ctx, parameters, deps)
    }
}
```

### After (3 steps)
```go
// 1. Register Tool WITH handler
func (r *ToolRegistry) registerNewTool() {
    r.Tools["new-Tool"] = &Tool{
        ID:          "new-Tool",
        Name:        "New Tool",
        Description: "...",
        Handler:     executeNewTool,  // ✅ Just reference the handler
    }
}

// 2. Call registration
func NewToolRegistry() *ToolRegistry {
    registry.registerNewTool()
}

// 3. Implement handler
func executeNewTool(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    // Implementation
}

// 4. ExecuteTool - NO CHANGES NEEDED! ✅
```

---

## Benefits

### 1. Better Encapsulation
Each Tool is self-contained with its own handler. The Tool "knows" how to execute itself.

### 2. Less Boilerplate
No need to maintain a growing switch statement or routing logic.

### 3. Type Safety
The `ToolHandler` type ensures all handlers have the correct signature at compile time.

### 4. Easier Testing
Can test handlers independently without going through ExecuteTool.

### 5. Cleaner Code
ExecuteTool is now just:
- Get the Tool
- Check permissions
- Call the handler

### 6. Consistency
All handlers have the same signature, making the codebase more predictable.

---

## Files Changed

### Modified
1. **Tool_types.go**
   - Added `ToolHandler` type definition
   - Added `Handler` field to `Tool` struct
   - Added necessary imports (context, tools, mcp)

2. **Tool_registry.go**
   - Removed switch statement from `ExecuteTool()`
   - Updated `ExecuteTool()` to call `Tool.Handler`
   - Updated all `register*Tool()` functions to include `Handler` field
   - Updated `executeListInstances()` signature to match `ToolHandler`

3. **README.md**
   - Updated "Adding a New Tool" section
   - Removed step 4 (adding to switch)
   - Clarified handler function pattern

---

## Migration Guide

If you have custom Tools, update them:

```go
// OLD WAY
func (r *ToolRegistry) registerMyTool() {
    r.Tools["my-Tool"] = &Tool{
        ID:          "my-Tool",
        // ... other fields
    }
}

// Then add to switch in ExecuteTool

// NEW WAY
func (r *ToolRegistry) registerMyTool() {
    r.Tools["my-Tool"] = &Tool{
        ID:          "my-Tool",
        // ... other fields
        Handler:     executeMyTool,  // ✅ Add this
    }
}

// Make sure handler matches signature
func executeMyTool(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    // Implementation
}
```

---

## Architecture Pattern

This is a common design pattern called **Strategy Pattern** or **Command Pattern**:

```
Tool = Data + Behavior
├─ Data: ID, Name, Description, Parameters, etc.
└─ Behavior: Handler function

ExecuteTool = Generic Executor
├─ Finds the Tool
├─ Validates permissions
└─ Delegates to Tool's handler
```

This is similar to how HTTP routers work:
```go
router.Handle("/instances", listInstancesHandler)
router.Handle("/instances/create", createInstanceHandler)
router.Handle("/instances/delete", deleteInstanceHandler)

// Router just dispatches to the registered handler
// No switch statement needed!
```

---

## Testing Impact

### Before
```go
// Had to test through ExecuteTool
func TestCreateInstance(t *testing.T) {
    registry := NewToolRegistry()
    result, err := registry.ExecuteTool(ctx, "create-instance", params, deps)
    // Test result
}
```

### After
```go
// Can test handler directly
func TestCreateInstanceHandler(t *testing.T) {
    result, err := executeCreateInstance(ctx, params, deps)
    // Test result
}

// Or test through ExecuteTool
func TestCreateInstanceThroughRegistry(t *testing.T) {
    registry := NewToolRegistry()
    result, err := registry.ExecuteTool(ctx, "create-instance", params, deps)
    // Test result
}
```

Both work, giving you more flexibility in testing!

---

## Summary

| Aspect | Before | After |
|--------|--------|-------|
| Routing | Switch statement | Handler function |
| Adding Tools | 4 steps | 3 steps |
| ExecuteTool changes | Required | Not required |
| Handler signatures | Inconsistent | Consistent |
| Coupling | Tight | Loose |
| Maintainability | Lower | Higher |
| Type safety | Partial | Full |

**Result**: Cleaner, more maintainable, and easier to extend! ✨
