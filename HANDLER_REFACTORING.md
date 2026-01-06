# Handler Function Refactoring

## What Changed

Refactored the outcome execution pattern to use handler functions instead of a switch statement.

## Before (Switch Statement Pattern)

### Problems
1. **Tight Coupling**: ExecuteOutcome needed to know about every outcome
2. **Maintenance Overhead**: Every new outcome required modifying ExecuteOutcome
3. **Switch Statement**: Required manual routing logic
4. **Inconsistent Signatures**: Different handlers had different signatures

### Old Code Structure

```go
// ExecuteOutcome with switch statement
func (r *OutcomeRegistry) ExecuteOutcome(...) (*mcp.CallToolResult, error) {
    outcome, err := r.GetOutcome(id)
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }

    // Check read-only mode
    if !outcome.ReadOnly && deps.Config != nil && deps.Config.ReadOnly {
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
    // Need to add case for every new outcome
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
1. **Loose Coupling**: ExecuteOutcome just calls outcome.Handler
2. **Zero Maintenance**: New outcomes don't require touching ExecuteOutcome
3. **Automatic Routing**: Handler is invoked directly
4. **Consistent Signatures**: All handlers use OutcomeHandler type

### New Code Structure

```go
// Define handler function type
type OutcomeHandler func(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error)

// Outcome includes handler
type Outcome struct {
    ID          string
    Name        string
    Description string
    Type        OutcomeType
    ReadOnly    bool
    Parameters  []OutcomeParameter
    Metadata    map[string]interface{}
    Handler     OutcomeHandler  // ✅ Handler function stored here
}

// ExecuteOutcome is now simple and generic ✅
func (r *OutcomeRegistry) ExecuteOutcome(ctx context.Context, id string, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    outcome, err := r.GetOutcome(id)
    if err != nil {
        return mcp.NewToolResultError(err.Error()), nil
    }

    // Check read-only mode
    if !outcome.ReadOnly && deps.Config != nil && deps.Config.ReadOnly {
        return mcp.NewToolResultError(...), nil
    }

    // Execute the handler - no switch needed! ✅
    if outcome.Handler == nil {
        return mcp.NewToolResultError(fmt.Sprintf("no handler registered for outcome: %s", id)), nil
    }

    return outcome.Handler(ctx, parameters, deps)
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

## Adding a New Outcome: Comparison

### Before (4 steps)
```go
// 1. Register outcome
func (r *OutcomeRegistry) registerNewOutcome() {
    r.outcomes["new-outcome"] = &Outcome{...}
}

// 2. Call registration
func NewOutcomeRegistry() *OutcomeRegistry {
    registry.registerNewOutcome()
}

// 3. Implement handler
func executeNewOutcome(...) (*mcp.CallToolResult, error) {
    // Implementation
}

// 4. Add to switch statement ❌
func (r *OutcomeRegistry) ExecuteOutcome(...) {
    switch id {
    case "list-instances":
        return executeListInstances(ctx, deps)
    case "new-outcome":  // ADD THIS
        return executeNewOutcome(ctx, parameters, deps)
    }
}
```

### After (3 steps)
```go
// 1. Register outcome WITH handler
func (r *OutcomeRegistry) registerNewOutcome() {
    r.outcomes["new-outcome"] = &Outcome{
        ID:          "new-outcome",
        Name:        "New Outcome",
        Description: "...",
        Handler:     executeNewOutcome,  // ✅ Just reference the handler
    }
}

// 2. Call registration
func NewOutcomeRegistry() *OutcomeRegistry {
    registry.registerNewOutcome()
}

// 3. Implement handler
func executeNewOutcome(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    // Implementation
}

// 4. ExecuteOutcome - NO CHANGES NEEDED! ✅
```

---

## Benefits

### 1. Better Encapsulation
Each outcome is self-contained with its own handler. The outcome "knows" how to execute itself.

### 2. Less Boilerplate
No need to maintain a growing switch statement or routing logic.

### 3. Type Safety
The `OutcomeHandler` type ensures all handlers have the correct signature at compile time.

### 4. Easier Testing
Can test handlers independently without going through ExecuteOutcome.

### 5. Cleaner Code
ExecuteOutcome is now just:
- Get the outcome
- Check permissions
- Call the handler

### 6. Consistency
All handlers have the same signature, making the codebase more predictable.

---

## Files Changed

### Modified
1. **outcome_types.go**
   - Added `OutcomeHandler` type definition
   - Added `Handler` field to `Outcome` struct
   - Added necessary imports (context, tools, mcp)

2. **outcome_registry.go**
   - Removed switch statement from `ExecuteOutcome()`
   - Updated `ExecuteOutcome()` to call `outcome.Handler`
   - Updated all `register*Outcome()` functions to include `Handler` field
   - Updated `executeListInstances()` signature to match `OutcomeHandler`

3. **README.md**
   - Updated "Adding a New Outcome" section
   - Removed step 4 (adding to switch)
   - Clarified handler function pattern

---

## Migration Guide

If you have custom outcomes, update them:

```go
// OLD WAY
func (r *OutcomeRegistry) registerMyOutcome() {
    r.outcomes["my-outcome"] = &Outcome{
        ID:          "my-outcome",
        // ... other fields
    }
}

// Then add to switch in ExecuteOutcome

// NEW WAY
func (r *OutcomeRegistry) registerMyOutcome() {
    r.outcomes["my-outcome"] = &Outcome{
        ID:          "my-outcome",
        // ... other fields
        Handler:     executeMyOutcome,  // ✅ Add this
    }
}

// Make sure handler matches signature
func executeMyOutcome(ctx context.Context, parameters map[string]interface{}, deps *tools.ToolDependencies) (*mcp.CallToolResult, error) {
    // Implementation
}
```

---

## Architecture Pattern

This is a common design pattern called **Strategy Pattern** or **Command Pattern**:

```
Outcome = Data + Behavior
├─ Data: ID, Name, Description, Parameters, etc.
└─ Behavior: Handler function

ExecuteOutcome = Generic Executor
├─ Finds the outcome
├─ Validates permissions
└─ Delegates to outcome's handler
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
// Had to test through ExecuteOutcome
func TestCreateInstance(t *testing.T) {
    registry := NewOutcomeRegistry()
    result, err := registry.ExecuteOutcome(ctx, "create-instance", params, deps)
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

// Or test through ExecuteOutcome
func TestCreateInstanceThroughRegistry(t *testing.T) {
    registry := NewOutcomeRegistry()
    result, err := registry.ExecuteOutcome(ctx, "create-instance", params, deps)
    // Test result
}
```

Both work, giving you more flexibility in testing!

---

## Summary

| Aspect | Before | After |
|--------|--------|-------|
| Routing | Switch statement | Handler function |
| Adding outcomes | 4 steps | 3 steps |
| ExecuteOutcome changes | Required | Not required |
| Handler signatures | Inconsistent | Consistent |
| Coupling | Tight | Loose |
| Maintainability | Lower | Higher |
| Type safety | Partial | Full |

**Result**: Cleaner, more maintainable, and easier to extend! ✨
