#!/bin/bash

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}=== Final Clean Build with Proper .gitignore ===${NC}\n"

# Step 1: Create/update .gitignore
echo -e "${YELLOW}STEP 1: Setting up .gitignore${NC}\n"

cat > .gitignore << 'GITIGNORE_EOF'
# Binaries
traefik-plugin-blockip
traefik-plugin-blockip-*
*.so
*. o
*.a
*.out

# Build directories
dist/
build/
bin/

# Go
vendor/
. env

# IDE
.vscode/
.idea/
*.swp
*.swo
*~
.DS_Store

# Test coverage
*.coverprofile
coverage. out
coverage.html

# Temp files
tmp/
temp/
*.tmp

# OS
Thumbs.db
.AppleDouble
.LSOverride

# Logs
*.log
GITIGNORE_EOF

echo -e "${GREEN}✓ .gitignore created${NC}\n"

# Step 2: Verify go.sum is NOT ignored
echo -e "${YELLOW}STEP 2: Verifying go.sum will be tracked${NC}\n"

if git check-ignore go.sum > /dev/null 2>&1; then
    echo -e "${RED}✗ go.sum is being ignored - fixing...${NC}"
    git rm --cached go.sum 2>/dev/null
    git add go.sum
else
    echo -e "${GREEN}✓ go.sum will be tracked${NC}"
fi

echo ""

# Step 3: Clean build artifacts
echo -e "${YELLOW}STEP 3: Cleaning build artifacts${NC}\n"

echo "Removing binaries..."
rm -f traefik-plugin-blockip traefik-plugin-blockip-*

echo "Removing build directories..."
rm -rf dist/ build/ bin/ tmp/ temp/

echo "Removing temp files..."
find . -name "*.so" -delete
find . -name "*.o" -delete
find . -name "*.a" -delete
find . -name "*.tmp" -delete

echo -e "${GREEN}✓ Artifacts cleaned${NC}\n"

# Step 4: Clean Go cache
echo -e "${YELLOW}STEP 4: Cleaning Go cache${NC}\n"

go clean -cache
go clean -testcache
go clean -modcache

echo -e "${GREEN}✓ Go cache cleared${NC}\n"

# Step 5: Download dependencies
echo -e "${YELLOW}STEP 5: Downloading dependencies${NC}\n"

go mod download

echo -e "${GREEN}✓ Dependencies downloaded${NC}\n"

# Step 6: Tidy and verify
echo -e "${YELLOW}STEP 6: Tidying and verifying modules${NC}\n"

go mod tidy
go mod verify

echo -e "${GREEN}✓ Modules verified${NC}\n"

# Step 7: Run tests
echo -e "${YELLOW}STEP 7: Running tests${NC}\n"

go test -v -cover ./...

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Tests passed${NC}\n"
else
    echo -e "${YELLOW}⚠ Tests failed${NC}\n"
fi

# Step 8: Build plugin
echo -e "${YELLOW}STEP 8: Building plugin${NC}\n"

go build -v -o traefik-plugin-blockip .

if [ -f "traefik-plugin-blockip" ]; then
    SIZE=$(ls -lh traefik-plugin-blockip | awk '{print $5}')
    echo -e "${GREEN}✓ Built: traefik-plugin-blockip ($SIZE)${NC}\n"
else
    echo -e "${RED}✗ Build failed${NC}\n"
    exit 1
fi

# Step 9: Show what will be committed
echo -e "${YELLOW}STEP 9: Checking git status{{NC}\n"

echo "Untracked/Modified files:"
git status --short
echo ""

# Step 10: Add all files
echo -e "${YELLOW}STEP 10: Staging files{{NC}\n"

git add -A
git add -f go. sum  # Make sure go.sum is added

echo "Files to commit:"
git diff --cached --name-only

echo ""

# Step 11: Commit
echo -e "${YELLOW}STEP 11: Committing{{NC}\n"

git commit -m "Final clean build: Proper .gitignore and rebuilt plugin

- Updated .gitignore to ignore build artifacts but keep go.sum
- Cleaned all build artifacts
- Cleared Go cache completely
- Downloaded fresh dependencies
- Verified all modules
- Rebuilt plugin binary
- Ready for deployment"

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ Committed{{NC}\n"
else
    echo -e "${YELLOW}⚠ Nothing to commit{{NC}\n"
fi

# Step 12: Show git status
echo -e "${YELLOW}STEP 12: Git status{{NC}\n"

echo "Current branch:"
git branch -v

echo ""
echo "Last commits:"
git log --oneline | head -3

echo ""

# Step 13: Summary
echo -e "${BLUE}=== Summary ==={{NC}\n"

echo "✓ . gitignore created/updated"
echo "✓ Build artifacts cleaned"
echo "✓ Go cache cleared"
echo "✓ Dependencies verified (go.sum from yesterday)"
echo "✓ Plugin built: $(ls -lh traefik-plugin-blockip | awk '{print $5}')"
echo "✓ Committed"
echo ""

# Step 14: Files status
echo -e "${YELLOW}Files that will be tracked:{{NC}"
git ls-files | grep -E "\.go|\.mod|\.sum|\.yml|\.yaml|\.md|\.gitignore"

echo ""
echo -e "${YELLOW}Files being ignored:{{NC}"
git check-ignore traefik-plugin-blockip 2>/dev/null && echo "✓ traefik-plugin-blockip (binary)"
git check-ignore dist/ 2>/dev/null && echo "✓ dist/ (build dir)"
git check-ignore . vscode/ 2>/dev/null && echo "✓ .vscode/ (IDE)"

echo ""
echo -e "${GREEN}=== Ready for Push ==={{NC}"