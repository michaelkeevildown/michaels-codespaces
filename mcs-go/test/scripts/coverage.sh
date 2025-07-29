#!/bin/bash
# MCS Test Coverage Script

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Directories
COVERAGE_DIR="test/coverage"
UNIT_COVERAGE_DIR="$COVERAGE_DIR/unit"
INTEGRATION_COVERAGE_DIR="$COVERAGE_DIR/integration"
E2E_COVERAGE_DIR="$COVERAGE_DIR/e2e"
REPORTS_DIR="test/reports"

# Create directories if they don't exist
mkdir -p "$UNIT_COVERAGE_DIR" "$INTEGRATION_COVERAGE_DIR" "$E2E_COVERAGE_DIR" "$REPORTS_DIR"

echo -e "${BLUE}🧪 MCS Test Coverage Report Generator${NC}"
echo "========================================"

# Unit test coverage
echo -e "${YELLOW}Running unit tests with coverage...${NC}"
go test -race -coverprofile="$UNIT_COVERAGE_DIR/coverage.out" -covermode=atomic ./...

# Generate HTML coverage report
echo -e "${YELLOW}Generating HTML coverage report...${NC}"
go tool cover -html="$UNIT_COVERAGE_DIR/coverage.out" -o "$REPORTS_DIR/coverage.html"

# Generate coverage summary
echo -e "${YELLOW}Generating coverage summary...${NC}"
COVERAGE_PERCENT=$(go tool cover -func="$UNIT_COVERAGE_DIR/coverage.out" | tail -1 | awk '{print $3}')
echo -e "${GREEN}Total Coverage: $COVERAGE_PERCENT${NC}"

# Save summary to file
echo "# Coverage Summary" > "$REPORTS_DIR/coverage_summary.md"
echo "" >> "$REPORTS_DIR/coverage_summary.md"
echo "**Total Coverage:** $COVERAGE_PERCENT" >> "$REPORTS_DIR/coverage_summary.md"
echo "" >> "$REPORTS_DIR/coverage_summary.md"
echo "Generated on: $(date)" >> "$REPORTS_DIR/coverage_summary.md"

echo -e "${GREEN}✅ Coverage report generated successfully!${NC}"
echo -e "${BLUE}📊 HTML Report: $REPORTS_DIR/coverage.html${NC}"
echo -e "${BLUE}📋 Summary: $REPORTS_DIR/coverage_summary.md${NC}"