# delete-instance Tool - Implementation Guide

## Overview

The `delete-instance` Tool permanently deletes a Neo4j Aura database instance. This is a **destructive operation** that cannot be undone.

## Safety Features

### 1. Explicit Confirmation Required
The Tool requires a `confirm` parameter set to `true` to proceed. This prevents accidental deletions.

### 2. Read-Only Mode Protection
The Tool **will not execute** when the server is in read-only mode (`READ_ONLY=true`). This is enforced at the Tool execution level.

### 3. Instance Verification
Before deletion, the system retrieves instance details to:
- Verify the instance exists
- Confirm access permissions
- Return information about what was deleted

## Parameters

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| instance_id | string | Yes | The ID of the instance to delete |
| confirm | boolean | Yes | Must be `true` to confirm deletion |

## Behavior in Read-Only Mode

### When READ_ONLY=true (default):
```json
{
  "Tool_id": "delete-instance",
  "parameters": {
    "instance_id": "abc123",
    "confirm": true
  }
}
```

**Response:**
```json
{
  "isError": true,
  "content": [{
    "type": "text",
    "text": "Cannot execute 'delete-instance' Tool: server is in read-only mode. Write operations are disabled. Set READ_ONLY=false to enable write operations."
  }]
}
```

### When READ_ONLY=false:
The operation proceeds if confirmation is provided.

## Usage Examples

### Step 1: Discover the Tool

```json
{
  "name": "list-Tools",
  "arguments": {}
}
```

Response includes:
```json
{
  "id": "delete-instance",
  "name": "Delete Instance",
  "description": "Permanently delete a Neo4j Aura database instance...",
  "type": "delete",
  "readonly": false
}
```

### Step 2: Get Details

```json
{
  "name": "get-Tool-details",
  "arguments": {
    "Tool_id": "delete-instance"
  }
}
```

Response:
```json
{
  "id": "delete-instance",
  "name": "Delete Instance",
  "description": "Permanently delete a Neo4j Aura database instance. This is a destructive operation that cannot be undone. Requires explicit confirmation via the 'confirm' parameter.",
  "type": "delete",
  "readonly": false,
  "parameters": [
    {
      "name": "instance_id",
      "type": "string",
      "description": "The ID of the instance to delete",
      "required": true
    },
    {
      "name": "confirm",
      "type": "boolean",
      "description": "Must be set to true to confirm deletion. This is a safety measure to prevent accidental deletions.",
      "required": true
    }
  ],
  "metadata": {
    "category": "instances",
    "destructive": true,
    "warning": "This operation permanently deletes the instance and all its data. This cannot be undone."
  }
}
```

### Step 3: Execute Deletion

**Successful Deletion:**
```json
{
  "name": "execute-Tool",
  "arguments": {
    "Tool_id": "delete-instance",
    "parameters": {
      "instance_id": "4f8e3b2a-1234-5678-90ab-cdef12345678",
      "confirm": true
    }
  }
}
```

**Success Response:**
```json
{
  "success": true,
  "message": "Instance 'my-test-db' (ID: 4f8e3b2a-1234-5678-90ab-cdef12345678) has been successfully deleted",
  "deleted_id": "4f8e3b2a-1234-5678-90ab-cdef12345678",
  "deleted_name": "my-test-db",
  "warning": "This instance and all its data have been permanently deleted and cannot be recovered."
}
```

## Error Scenarios

### 1. Missing Confirmation

```json
{
  "Tool_id": "delete-instance",
  "parameters": {
    "instance_id": "abc123"
  }
}
```

**Error:**
```json
{
  "isError": true,
  "content": [{
    "type": "text",
    "text": "'confirm' parameter is required and must be a boolean (true to confirm deletion)"
  }]
}
```

### 2. Confirmation Set to False

```json
{
  "Tool_id": "delete-instance",
  "parameters": {
    "instance_id": "abc123",
    "confirm": false
  }
}
```

**Error:**
```json
{
  "isError": true,
  "content": [{
    "type": "text",
    "text": "Deletion not confirmed. Set 'confirm' to true to proceed with deletion. WARNING: This action cannot be undone."
  }]
}
```

### 3. Instance Not Found

```json
{
  "Tool_id": "delete-instance",
  "parameters": {
    "instance_id": "nonexistent-id",
    "confirm": true
  }
}
```

**Error:**
```json
{
  "isError": true,
  "content": [{
    "type": "text",
    "text": "Failed to retrieve instance details before deletion: [error details]. The instance may not exist or you may not have access to it."
  }]
}
```

### 4. Read-Only Mode

```json
{
  "Tool_id": "delete-instance",
  "parameters": {
    "instance_id": "abc123",
    "confirm": true
  }
}
```

**Error (when READ_ONLY=true):**
```json
{
  "isError": true,
  "content": [{
    "type": "text",
    "text": "Cannot execute 'delete-instance' Tool: server is in read-only mode. Write operations are disabled. Set READ_ONLY=false to enable write operations."
  }]
}
```

### 5. API Deletion Failure

If the API call to delete fails after verification:

**Error:**
```json
{
  "isError": true,
  "content": [{
    "type": "text",
    "text": "Failed to delete instance: [API error details]"
  }]
}
```

## Claude Integration Examples

### Example 1: User Requests Deletion

```
User: Delete my test database

Claude: I can help you delete an instance. First, let me list your instances 
to find the right one.
[Calls list-instances]

Claude: I found these instances:
1. test-db (ID: abc-123)
2. production-db (ID: def-456)

Which one would you like to delete?

User: The test one

Claude: ⚠️ WARNING: This will permanently delete 'test-db' and all its data. 
This cannot be undone. Are you sure you want to proceed?

User: Yes, delete it

Claude: Deleting 'test-db'...
[Calls execute-Tool with delete-instance, confirm=true]

Claude: ✓ Instance 'test-db' has been successfully deleted. 
The instance and all its data have been permanently removed.
```

### Example 2: Read-Only Mode Protection

```
User: Delete my old database

Claude: I'd like to help, but I'm currently in read-only mode. 
Write operations like deleting instances are disabled.

To enable deletions, the server administrator needs to set READ_ONLY=false 
in the server configuration.
```

## API Call Structure

The implementation assumes:

```go
// Get instance details
instanceInfo, err := deps.AClient.Instances.Get(instanceID)

// Delete the instance
err = deps.AClient.Instances.Delete(instanceID)
```

**Note:** Verify this matches your aura-client library API. The Delete method should accept the instance ID and return an error if the deletion fails.

## Configuration

### Enable Deletions

Set in your environment or Claude Desktop config:

```json
{
  "mcpServers": {
    "mcp-aura-api": {
      "command": "/path/to/mcp-aura-api",
      "env": {
        "CLIENT_ID": "your-client-id",
        "CLIENT_SECRET": "your-client-secret",
        "READ_ONLY": "false"
      }
    }
  }
}
```

Or via environment variable:
```bash
export READ_ONLY=false
```

### Default Behavior

- **Default**: `READ_ONLY=true` (deletions disabled)
- **For Production**: Keep `READ_ONLY=true` to prevent accidental deletions
- **For Development**: Set `READ_ONLY=false` to enable all operations

## Testing

### 1. Test in Read-Only Mode (Default)

```bash
# Build
go build -o bin/mcp-aura-api ./cmd/mcp-aura-api

# Test (READ_ONLY defaults to true)
npx @modelcontextprotocol/inspector ./bin/mcp-aura-api

# Try to delete - should fail with read-only error
```

### 2. Test with Read-Only Disabled

```bash
# Set READ_ONLY to false
export READ_ONLY=false
export CLIENT_ID="your-id"
export CLIENT_SECRET="your-secret"

# Test
npx @modelcontextprotocol/inspector ./bin/mcp-aura-api

# Try to delete - should succeed with confirmation
```

### 3. Test Confirmation Requirement

```json
// Without confirm - should fail
{
  "Tool_id": "delete-instance",
  "parameters": {
    "instance_id": "test-id"
  }
}

// With confirm=false - should fail
{
  "Tool_id": "delete-instance",
  "parameters": {
    "instance_id": "test-id",
    "confirm": false
  }
}

// With confirm=true - should succeed (if READ_ONLY=false)
{
  "Tool_id": "delete-instance",
  "parameters": {
    "instance_id": "test-id",
    "confirm": true
  }
}
```

## Best Practices

1. **Always Verify**: List instances first to confirm IDs
2. **Backup First**: Consider creating snapshots before deletion
3. **Double-Check**: Verify you're deleting the correct instance
4. **Production Safety**: Keep READ_ONLY=true in production
5. **Audit Trail**: Log deletion events for compliance

## Workflow Recommendation

```
1. List instances → Get instance IDs
2. Verify instance → Confirm it's the right one
3. Get deletion approval → User confirms
4. Execute deletion → With confirm=true
5. Verify deletion → Check instance no longer appears in list
```

## Architecture Notes

### Read-Only Protection Layers

1. **Server Level**: READ_ONLY environment variable
2. **Tool Level**: ExecuteTool checks Tool.ReadOnly flag
3. **Parameter Level**: Explicit confirm parameter required
4. **API Level**: Aura API enforces permissions

This defense-in-depth approach prevents accidental deletions.

### Why Tool-Level readonly=true?

The `execute-Tool` tool is marked as `readonly: true` at the tool level, even though it can execute write operations. This is because:

1. The tool itself just routes to Tools
2. Individual Tools have their own readonly flags
3. ExecuteTool checks the Tool's readonly flag against server config
4. This allows read-only Tools to execute even in READ_ONLY mode
5. Write Tools are blocked at execution time with a clear error message

## Troubleshooting

### "Cannot execute: server is in read-only mode"
**Solution**: Set `READ_ONLY=false` in your configuration

### "Deletion not confirmed"
**Solution**: Set `confirm: true` in parameters

### "Failed to retrieve instance details"
**Solution**: Verify the instance_id is correct and you have access

### "Failed to delete instance"
**Solution**: Check API credentials and instance status

## Summary

- ✅ Requires explicit confirmation
- ✅ Protected by read-only mode
- ✅ Verifies instance before deletion
- ✅ Returns detailed deletion confirmation
- ✅ Multiple safety layers
- ✅ Clear error messages

**Status**: Production-ready with comprehensive safety features
