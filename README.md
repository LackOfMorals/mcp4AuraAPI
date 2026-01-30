# MCP for Aura Infrastructure Metrics
This MCP Server gathers information from the Prometheus endpoint ( available on  Aura Business Critical And VDC tiers ) for analysis by an LLM / Agent.  

Additionally, as these are supporting capabilities, it can also retrieve the list of Aura instances and detailed configuration information for each. 


## Features

- Simplified health check requiring only instance ID (auto-fetches Prometheus URL)
- Retrieve a summary list of all Neo4j Aura database instances
- Get detailed info for a specific instance (includes Prometheus URL)


## A 3 tool Architecture

There is a risk of overloading LLM / Agents when a MCP Server returns lots of tools with their descriptions.  This consumes the context memory of the LLM / Agent and can , in the worse case, lead to the onset of "context confusion" .  

An alternative approach, and the one taken here with this MCP Server, is to provide three tools 

- list outcomes
  - A summary list of outcomes that can be achieved 
- get outcome details
  - What an outcome can do, how to use it and what to expect to be returned
- execute an outcome
  - Run an outcome

Outcomes are what the LLM / Agent achieves when using it.  This is a different approach to giving a bunch of tools and then reply on them being used in the right order to get a job done. 

Outcomes can be coded directly into the MCP Server, as they are here, or could be stored elsewhere for on demand, dynamic retreival. 



## Prerequisites

- Go 1.25+ (see `go.mod`)
- Client Id and Secret for Neo4j Aura API for an Aura User who has the 'Metrics Gathering Role' assigned to them
- An Aura Business Critical or Virtual Dedicate Cloud instance
- Metrics has been enabled for the instance

See Neo4j Aura Console documentation for how to setup instances for montoring : [Customer Metrics Integration (CMI)](https://neo4j.com/docs/aura/metrics/metrics-integration/introduction/)


> **Note:** At the time of writing, 30th Jan 2026, Admin cannot be used.  It must be a user account who has the Metrics Gathering Role who generates the Client Id and Secret.   If you get a 401 error, this is likely due to the Client Id and Secret are for a user who does not have the required role. 



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

