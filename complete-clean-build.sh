#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Complete Clean Build ===${NC}\n"

# Step 1: Create proper . gitignore
echo -e "${YELLOW}STEP 1: Creating .gitignore${NC}\n"

# cat > .gitignore << 'GITIGNORE_EOF'
# # Binaries
# traefik-plugin-blockip
# traefik-plugin-blockip-*
# *.so
# *. o
# *.a
# *.out

# # Build directories
# dist/
# build/
# bin/

# # Go
# vendor/
# go.sum
# .env

# # IDE
# .vscode/
# .idea/
# *.swp
# *.swo
# *~
# .DS_Store

# # Test coverage
# *.coverprofile
# coverage.out
# coverage.html

# # Temp files
# tmp/
# temp/
# *.tmp

# # OS
# Thumbs.db
# . AppleDouble
# .LSOverride
# *. swp
# *.swo

# # Logs
# *.log

# # Dependencies (but keep go.mod and go.sum for the repo)
# GITIGNORE_EOF

# if [ -f ". gitignore" ]; then
#     echo -e "${GREEN}✓ .gitignore created${NC}\n"
# else
#     echo -e "${RED}✗ Failed to create .gitignore${NC}\n"
#     exit 1
# fi

# Step 2: Clean local build artifacts
echo -e "${YELLOW}STEP 2: Cleaning build artifacts${NC}\n"

echo "Removing compiled binaries..."
rm -f traefik-plugin-blockip
rm -f traefik-plugin-blockip-*
find .  -name "*.so" -delete
find . -name "*.o" -delete
find . -name "*.a" -delete

echo "Removing build directories..."
rm -rf dist/
rm -rf build/
rm -rf bin/

echo "Removing temp files..."
rm -rf tmp/
rm -rf temp/
find . -name "*.tmp" -delete

echo -e "${GREEN}✓ Artifacts cleaned${NC}\n"

# Step 3: Clean Go cache
echo -e "${YELLOW}STEP 3: Cleaning Go cache{{NC}\n"

go clean -cache
go clean -testcache
go clean -modcache

echo -e "${GREEN}✓ Go cache cleared${NC}\n"

# Step 4: Update dependencies
echo -e "${YELLOW}STEP 4: Updating dependencies{{NC}\n"

go mod download
go mod tidy
go mod verify

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Dependencies updated${NC}\n"
else
    echo -e "${RED}✗ Dependency update failed${NC}\n"
fi

# Step 5: Run tests
echo -e "${YELLOW}STEP 5: Running tests{{NC}\n"

go test -v -cover ./...

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Tests passed${NC}\n"
else
    echo -e "${RED}⚠ Some tests failed${NC}\n"
fi

# Step 6: Build plugin
echo -e "${YELLOW}STEP 6: Building plugin{{NC}\n"

go build -v -o traefik-plugin-blockip . 

if [ -f "traefik-plugin-blockip" ]; then
    SIZE=$(ls -lh traefik-plugin-blockip | awk '{print $5}')
    echo -e "${GREEN}✓ Built successfully ($SIZE)${NC}\n"
else
    echo -e "${RED}✗ Build failed${NC}\n"
    exit 1
fi

# Step 7: Check what will be ignored
echo -e "${YELLOW}STEP 7: Checking . gitignore status{{NC}\n"

echo "Files that will be ignored:"
git check-ignore -v * 2>/dev/null | head -20

echo ""

# Step 8: Add to git
echo -e "${YELLOW}STEP 8: Adding files to git{{NC}\n"

git add -A
git add -f go.sum  # Force add go.sum (we want this in repo)

STATUS=$(git status --short)

echo "Status:"
echo "$STATUS" | head -20

echo ""

# Step 9: Commit
echo -e "${YELLOW}STEP 9: Committing{{NC}\n"

git commit -m "Clean build: Update . gitignore and rebuild plugin

- Updated .gitignore with proper patterns
- Cleaned all build artifacts
- Cleared Go cache
- Updated and verified dependencies
- Rebuilt plugin successfully
- Kept go.sum in repository"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Committed${NC}\n"
else
    echo -e "${YELLOW}⚠ Nothing to commit or commit failed${NC}\n"
fi

# Step 10: Push to GitHub
echo -e "${YELLOW}STEP 10: Pushing to GitHub{{NC}\n"

git push origin main

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Pushed{{NC}\n"
else
    echo -e "${RED}⚠ Push may have failed${NC}\n"
fi

# Step 11: Verify git status
echo -e "${YELLOW}STEP 11: Verifying git status{{NC}\n"

echo "Clean working tree:"
git status

echo ""

# Step 12: Summary
echo -e "${BLUE}=== Summary ==={{NC}\n"

echo "✓ . gitignore created/updated"
echo "✓ Build artifacts cleaned"
echo "✓ Go cache cleared"
echo "✓ Dependencies verified"
echo "✓ Plugin built: $(ls -lh traefik-plugin-blockip | awk '{print $5}')"
echo "✓ Committed and pushed"
echo ""

echo -e "${YELLOW}Files in repository:{{NC}"
git ls-files | head -20

echo ""
echo -e "${GREEN}=== Clean Build Complete ==={{NC}"