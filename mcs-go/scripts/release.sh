#!/bin/bash
set -e

# Release helper script for MCS

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Get the release type
RELEASE_TYPE="${1:-patch}"

# Get current version from git tags
CURRENT_VERSION=$(git describe --tags --abbrev=0 2>/dev/null || echo "v0.0.0")
CURRENT_VERSION="${CURRENT_VERSION#v}"  # Remove 'v' prefix

# Parse version components
IFS='.' read -r MAJOR MINOR PATCH <<< "${CURRENT_VERSION%-*}"
PRE_RELEASE="${CURRENT_VERSION#*-}"

# Calculate new version based on release type
case "$RELEASE_TYPE" in
    patch)
        NEW_PATCH=$((PATCH + 1))
        NEW_VERSION="v${MAJOR}.${MINOR}.${NEW_PATCH}"
        ;;
    minor)
        NEW_MINOR=$((MINOR + 1))
        NEW_VERSION="v${MAJOR}.${NEW_MINOR}.0"
        ;;
    major)
        NEW_MAJOR=$((MAJOR + 1))
        NEW_VERSION="v${NEW_MAJOR}.0.0"
        ;;
    pre)
        # Handle pre-release versions
        if [[ "$CURRENT_VERSION" == *"-"* ]]; then
            # Already a pre-release, increment the number
            PRE_TYPE="${PRE_RELEASE%.*}"
            PRE_NUM="${PRE_RELEASE##*.}"
            if [[ "$PRE_NUM" =~ ^[0-9]+$ ]]; then
                NEW_PRE_NUM=$((PRE_NUM + 1))
                NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}-${PRE_TYPE}.${NEW_PRE_NUM}"
            else
                NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}-beta.1"
            fi
        else
            # First pre-release
            NEW_VERSION="v${MAJOR}.${MINOR}.${PATCH}-beta.1"
        fi
        ;;
    *)
        echo -e "${RED}Invalid release type: $RELEASE_TYPE${NC}"
        echo "Usage: $0 [patch|minor|major|pre]"
        exit 1
        ;;
esac

echo -e "${GREEN}Current version:${NC} v${CURRENT_VERSION}"
echo -e "${GREEN}New version:${NC}     ${NEW_VERSION}"
echo ""

# Confirm with user
read -p "Create release ${NEW_VERSION}? [y/N] " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Release cancelled."
    exit 0
fi

# Ensure working directory is clean
if [[ -n $(git status --porcelain) ]]; then
    echo -e "${YELLOW}Warning: Working directory has uncommitted changes${NC}"
    read -p "Continue anyway? [y/N] " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo "Release cancelled."
        exit 1
    fi
fi

# Create and push tag
echo -e "${GREEN}Creating tag ${NEW_VERSION}...${NC}"
git tag -a "${NEW_VERSION}" -m "Release ${NEW_VERSION}"

echo -e "${GREEN}Pushing tag to origin...${NC}"
git push origin "${NEW_VERSION}"

echo ""
echo -e "${GREEN}✓ Release ${NEW_VERSION} created successfully!${NC}"
echo ""
echo "GitHub Actions will now:"
echo "  1. Run tests"
echo "  2. Build binaries for all platforms"
echo "  3. Create a GitHub release with changelog"
echo "  4. Upload binaries to the release"
echo ""
echo "Monitor progress at:"
echo "  https://github.com/michaelkeevildown/michaels-codespaces/actions"
echo ""
echo "Once complete, users can install with:"
echo "  curl -fsSL https://raw.githubusercontent.com/michaelkeevildown/michaels-codespaces/main/mcs-go/install.sh | bash"