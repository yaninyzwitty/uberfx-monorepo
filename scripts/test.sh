#!/bin/bash

# Test script for Go monorepo

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Running tests for Go monorepo...${NC}"

# Sync workspace
echo -e "${YELLOW}Syncing workspace...${NC}"
go work sync

# Run unit tests
echo -e "${YELLOW}Running unit tests...${NC}"
for service in packages/*/; do
    if [ -f "$service/go.mod" ]; then
        echo -e "${GREEN}Testing $(basename $service)...${NC}"
        cd "$service"
        go test -v -race ./... || exit 1
        cd - > /dev/null
    fi
done

# Run integration tests if requested
if [ "$1" = "integration" ]; then
    echo -e "${YELLOW}Running integration tests...${NC}"
    for service in packages/*/; do
        if [ -f "$service/go.mod" ]; then
            echo -e "${GREEN}Integration testing $(basename $service)...${NC}"
            cd "$service"
            go test -v -race -tags=integration ./... || true
            cd - > /dev/null
        fi
    done
fi

echo -e "${GREEN}All tests completed!${NC}"
