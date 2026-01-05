package outcomes

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func ListInstancesSpec() mcp.Tool {
	return mcp.NewTool("list-instances",
		mcp.WithDescription(`
		Retrieve a list of instances in Neo4j Aura cloud.  An empty list means there are no instances or the user does not have access.`),
		mcp.WithTitleAnnotation("List instances"),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithIdempotentHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
		mcp.WithOpenWorldHintAnnotation(true),
	)
}
