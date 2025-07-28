#!/bin/bash
# Build script for MCS Docker images

set -e

# Docker Hub organization (change this to your org)
DOCKER_ORG="${DOCKER_ORG:-mcs}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Building MCS Docker images...${NC}"

# Build base image first
echo -e "${YELLOW}Building base image...${NC}"
docker build -f Dockerfile.base -t ${DOCKER_ORG}/code-server-base:latest .

# Build language-specific images
for dockerfile in Dockerfile.*; do
    if [[ "$dockerfile" == "Dockerfile.base" ]]; then
        continue
    fi
    
    # Extract language from filename (e.g., Dockerfile.python -> python)
    lang=${dockerfile#Dockerfile.}
    
    echo -e "${YELLOW}Building ${lang} image...${NC}"
    docker build -f "$dockerfile" -t ${DOCKER_ORG}/code-server-${lang}:latest .
done

echo -e "${GREEN}All images built successfully!${NC}"

# Optional: Push to registry
if [[ "$1" == "--push" ]]; then
    echo -e "${YELLOW}Pushing images to registry...${NC}"
    
    docker push ${DOCKER_ORG}/code-server-base:latest
    
    for dockerfile in Dockerfile.*; do
        if [[ "$dockerfile" == "Dockerfile.base" ]]; then
            continue
        fi
        lang=${dockerfile#Dockerfile.}
        docker push ${DOCKER_ORG}/code-server-${lang}:latest
    done
    
    echo -e "${GREEN}All images pushed successfully!${NC}"
fi

# List built images
echo -e "${GREEN}Built images:${NC}"
docker images | grep "${DOCKER_ORG}/code-server"