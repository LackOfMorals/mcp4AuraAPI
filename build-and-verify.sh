#!/bin/bash
set -e

echo "======================================"
echo "Building MCP Aura API with Prometheus"
echo "======================================"
echo ""

# Colors
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Build
echo -e "${BLUE}Step 1: Building...${NC}"
go build -o ./bin/mcp-aura-api ./cmd/mcp-aura-api
echo -e "${GREEN}✓ Build successful${NC}"
echo ""

# Verify binary
echo -e "${BLUE}Step 2: Verifying binary...${NC}"
if [ -f "./bin/mcp-aura-api" ]; then
    echo -e "${GREEN}✓ Binary created successfully${NC}"
else
    echo "✗ Binary not found"
    exit 1
fi
echo ""

# Summary
echo -e "${BLUE}Step 3: Implementation Summary${NC}"
echo "Files created:"
echo "  ✓ internal/prometheus/client.go"
echo "  ✓ internal/prometheus/aggregator.go"
echo "  ✓ internal/server/outcome_registry.go (updated)"
echo ""
echo "Outcomes available (7 total):"
echo "  Instance Management:"
echo "    - list-instances"
echo "    - create-instance"
echo "    - delete-instance"
echo "  Prometheus Monitoring (NEW):"
echo "    - get-instance-health"
echo "    - diagnose-performance"
echo "    - analyze-resource-usage"
echo "    - get-query-statistics"
echo ""

# Next steps
echo -e "${BLUE}Next Steps:${NC}"
echo ""
echo "1. Test with MCP Inspector:"
echo "   CLIENT_ID=<id> CLIENT_SECRET=<secret> \\"
echo "   npx @modelcontextprotocol/inspector ./bin/mcp-aura-api"
echo ""
echo "2. Test outcomes:"
echo "   a. Call 'list-outcomes' - should show 7 outcomes"
echo "   b. Call 'get-outcome-details' with 'get-instance-health'"
echo "   c. Call 'execute-outcome' with your Prometheus URL"
echo ""
echo "3. Documentation:"
echo "   - PROMETHEUS_INTEGRATION.md - Full documentation"
echo "   - QUICKSTART_PROMETHEUS.md - Usage guide"
echo "   - IMPLEMENTATION_COMPLETE.md - Summary"
echo ""
echo -e "${GREEN}✓ Phase 1 implementation complete!${NC}"
