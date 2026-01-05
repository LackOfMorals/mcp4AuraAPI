# Using create-instance Outcome - Examples

## Flow Overview

```
1. list-outcomes          → See "create-instance" is available
2. get-outcome-details    → Get parameter specifications
3. execute-outcome        → Create the instance
```

## Step-by-Step Example

### 1. Discover the Outcome

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "tools/call",
  "params": {
    "name": "list-outcomes",
    "arguments": {}
  }
}
```

**Response:**
```json
[
  {
    "id": "list-instances",
    "name": "List Instances",
    "description": "Retrieve a list of all Neo4j Aura database instances...",
    "type": "list",
    "readonly": true
  },
  {
    "id": "create-instance",
    "name": "Create Instance",
    "description": "Create a new Neo4j Aura database instance with specified configuration...",
    "type": "create",
    "readonly": false
  }
]
```

### 2. Get Detailed Parameter Information

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 2,
  "method": "tools/call",
  "params": {
    "name": "get-outcome-details",
    "arguments": {
      "outcome_id": "create-instance"
    }
  }
}
```

**Response:**
```json
{
  "id": "create-instance",
  "name": "Create Instance",
  "description": "Create a new Neo4j Aura database instance with specified configuration. Returns the created instance details including ID, name, and connection information.",
  "type": "create",
  "readonly": false,
  "parameters": [
    {
      "name": "name",
      "type": "string",
      "description": "Name for the new instance",
      "required": true
    },
    {
      "name": "cloud_provider",
      "type": "string",
      "description": "Cloud provider: 'gcp', 'aws', or 'azure'",
      "required": true
    },
    {
      "name": "region",
      "type": "string",
      "description": "Cloud region (e.g., 'us-east-1' for AWS, 'us-central1' for GCP, 'eastus' for Azure)",
      "required": true
    },
    {
      "name": "memory",
      "type": "string",
      "description": "Memory size for the instance (e.g., '2GB', '8GB', '16GB', '32GB', '64GB')",
      "required": true
    },
    {
      "name": "type",
      "type": "string",
      "description": "Instance type: 'free', 'professional', or 'enterprise'",
      "required": true
    },
    {
      "name": "version",
      "type": "string",
      "description": "Neo4j version (e.g., '5', '5.24'). If not specified, uses the latest stable version.",
      "required": false,
      "default": "5"
    }
  ],
  "metadata": {
    "category": "instances"
  }
}
```

### 3. Execute the Outcome

#### Example 1: Create a Professional GCP Instance

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 3,
  "method": "tools/call",
  "params": {
    "name": "execute-outcome",
    "arguments": {
      "outcome_id": "create-instance",
      "parameters": {
        "name": "my-production-db",
        "cloud_provider": "gcp",
        "region": "us-central1",
        "memory": "8GB",
        "type": "professional",
        "version": "5"
      }
    }
  }
}
```

**Success Response:**
```json
{
  "success": true,
  "message": "Instance created successfully",
  "id": "4f8e3b2a-1234-5678-90ab-cdef12345678",
  "name": "my-production-db",
  "status": "creating",
  "cloud_provider": "gcp",
  "memory": "8GB",
  "type": "professional",
  "url": "neo4j+s://4f8e3b2a.databases.neo4j.io"
}
```

#### Example 2: Create an AWS Development Instance

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 4,
  "method": "tools/call",
  "params": {
    "name": "execute-outcome",
    "arguments": {
      "outcome_id": "create-instance",
      "parameters": {
        "name": "dev-testing-db",
        "cloud_provider": "aws",
        "region": "us-east-1",
        "memory": "2GB",
        "type": "professional"
      }
    }
  }
}
```

Note: The `version` parameter is omitted, so it will use the default "5".

#### Example 3: Create an Azure Enterprise Instance

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": 5,
  "method": "tools/call",
  "params": {
    "name": "execute-outcome",
    "arguments": {
      "outcome_id": "create-instance",
      "parameters": {
        "name": "enterprise-analytics",
        "cloud_provider": "azure",
        "region": "eastus",
        "memory": "32GB",
        "type": "enterprise",
        "version": "5.24"
      }
    }
  }
}
```

## Error Handling Examples

### Missing Required Parameter

**Request:**
```json
{
  "outcome_id": "create-instance",
  "parameters": {
    "name": "my-db",
    "cloud_provider": "gcp"
    // Missing region, memory, type
  }
}
```

**Error Response:**
```json
{
  "isError": true,
  "content": [
    {
      "type": "text",
      "text": "'region' parameter is required and must be a non-empty string"
    }
  ]
}
```

### Invalid Cloud Provider

**Request:**
```json
{
  "outcome_id": "create-instance",
  "parameters": {
    "name": "my-db",
    "cloud_provider": "digitalocean",  // Invalid
    "region": "nyc1",
    "memory": "2GB",
    "type": "professional"
  }
}
```

**Error Response:**
```json
{
  "isError": true,
  "content": [
    {
      "type": "text",
      "text": "Invalid cloud_provider 'digitalocean'. Must be one of: 'gcp', 'aws', 'azure'"
    }
  ]
}
```

### Invalid Instance Type

**Request:**
```json
{
  "outcome_id": "create-instance",
  "parameters": {
    "name": "my-db",
    "cloud_provider": "gcp",
    "region": "us-central1",
    "memory": "2GB",
    "type": "basic"  // Invalid
  }
}
```

**Error Response:**
```json
{
  "isError": true,
  "content": [
    {
      "type": "text",
      "text": "Invalid type 'basic'. Must be one of: 'free', 'professional', 'enterprise'"
    }
  ]
}
```

## Common Regions by Cloud Provider

### Google Cloud Platform (GCP)
- us-central1 (Iowa)
- us-east1 (South Carolina)
- us-west1 (Oregon)
- europe-west1 (Belgium)
- asia-southeast1 (Singapore)

### Amazon Web Services (AWS)
- us-east-1 (N. Virginia)
- us-west-2 (Oregon)
- eu-west-1 (Ireland)
- ap-southeast-1 (Singapore)
- ap-northeast-1 (Tokyo)

### Microsoft Azure
- eastus (East US)
- westus2 (West US 2)
- westeurope (West Europe)
- southeastasia (Southeast Asia)
- australiaeast (Australia East)

## Typical Memory Sizes

- **Free Tier**: Usually limited, check Aura documentation
- **Professional**: 2GB, 8GB, 16GB, 32GB
- **Enterprise**: 32GB, 64GB, 128GB, 256GB+

## Instance Types

- **free**: Free tier with limitations (good for learning/testing)
- **professional**: Production-ready with standard features
- **enterprise**: Advanced features, dedicated support, higher SLAs

## Monitoring Instance Creation

After creating an instance, the status will be "creating". You can poll the instance using the `list-instances` outcome to check when it becomes "running":

**Poll Request:**
```json
{
  "outcome_id": "list-instances",
  "parameters": {}
}
```

Look for your instance in the results and check its `status` field. Common statuses:
- `creating` - Instance is being provisioned
- `running` - Instance is ready to use
- `paused` - Instance is paused
- `deleting` - Instance is being deleted

## Integration with Claude

When using with Claude, the conversation would look like:

```
User: Create a new Neo4j database for our analytics project on GCP

Claude: I'll create a Neo4j Aura instance for you. Let me execute that.
[Calls execute-outcome with create-instance]

Claude: I've successfully created your Neo4j instance:
- Name: analytics-project-db
- Status: creating
- Cloud: GCP (us-central1)
- Memory: 8GB
- Type: professional
- Connection URL: neo4j+s://abc123.databases.neo4j.io

The instance is currently being provisioned. It should be ready in a few minutes.
```

## Next Steps After Creation

1. **Wait for Running Status**: Poll until status is "running"
2. **Get Credentials**: Retrieve database credentials (may need another outcome)
3. **Connect**: Use the connection URL with Neo4j drivers
4. **Configure**: Set up users, constraints, indexes as needed
