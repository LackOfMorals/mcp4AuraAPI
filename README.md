# MCP for Aura Infrastructure 


## Prerequisites

- Go 1.25+ (see `go.mod`)
- Client Id and Secret for Neo4j Aura API

## Installation
### Clone the repository 

```bash
git clone https://github.com/LackOfMorals/mcp4AuraAPI.git
```

### Install Dependencies

```bash
cd mcp4AuraAPI
go mod download

```


### Build

MCP for Aura API needs to be compiled before use.   Do this with

```Bash
go build -o ./bin/mcp-aura-api ./cmd/mcp-aura-api

```

### Testing

You can test before using with LLMs by using MCP Inspector. Set required environmental variables in MCP Inspector itself or beforehand.  If the latter, then you need to set CLIENT_ID and CLIENT_SECRET before running MCP Inspector. 

```bash
npx @modelcontextprotocol/inspector go run ./cmd/mcp-aura-api

```



## Using with Claude Desktop

```Text
{
  "mcpServers": {
    "mcp-aura-api": {
      "command": "<FULL PATH TO MCP BINARY>",
      "env": {
        "CLIENT_ID":"<YOUR AURA API CLIENT ID>",
        "CLIENT_SECRET":"<YOUR AURA API SECRET ID>"
      }
    }
  }
}
```



