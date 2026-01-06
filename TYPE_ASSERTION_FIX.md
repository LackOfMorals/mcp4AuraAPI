# Type Assertion Fix

## Issue

The compiler error occurred because `request.Params.Arguments` is of type `any` (interface{}), which cannot be directly indexed like a map.

```go
// ERROR: Cannot index request.Params.Arguments (type any)
outcomeID, ok := request.Params.Arguments["outcome_id"].(string)
```

## Solution

Type assert `Arguments` to `map[string]interface{}` before accessing its elements:

```go
// CORRECT: First type assert, then access
arguments, ok := request.Params.Arguments.(map[string]interface{})
if !ok {
    return mcp.NewToolResultError("invalid arguments format"), nil
}

outcomeID, ok := arguments["outcome_id"].(string)
if !ok || outcomeID == "" {
    return mcp.NewToolResultError("outcome_id parameter is required and must be a string"), nil
}
```

## Files Fixed

### internal/tools/outcomes/outcome_handlers.go

**GetOutcomeDetailsHandler** - Added type assertion:
```go
arguments, ok := request.Params.Arguments.(map[string]interface{})
if !ok {
    return mcp.NewToolResultError("invalid arguments format"), nil
}
```

**ExecuteOutcomeHandler** - Added type assertion:
```go
arguments, ok := request.Params.Arguments.(map[string]interface{})
if !ok {
    return mcp.NewToolResultError("invalid arguments format"), nil
}
```

**ListOutcomesHandler** - No changes needed (doesn't access Arguments)

## Why This Happened

The MCP library defines `Arguments` as `any` to support different argument structures. We need to type assert to the specific type we expect before we can work with it.

This is standard Go practice when working with `interface{}` or `any` types:

```go
// General pattern
var i interface{} = map[string]string{"key": "value"}

// Can't do this:
// value := i["key"]  // ERROR!

// Must do this:
if m, ok := i.(map[string]string); ok {
    value := m["key"]  // OK!
}
```

## Testing

The code should now compile successfully:

```bash
go build -o bin/mcp-aura-api ./cmd/mcp-aura-api
```

## Additional Safety

The type assertion also adds safety - if `Arguments` is somehow not a map (malformed request), we catch it early and return a clear error message rather than panicking.
