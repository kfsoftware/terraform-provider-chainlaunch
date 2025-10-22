#!/bin/bash
# Pre-commit hook for Terraform provider
# Install: cp hooks/pre-commit .git/hooks/pre-commit && chmod +x .git/hooks/pre-commit

set -e

echo "🔍 Running pre-commit checks..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track if any checks fail
FAILED=0

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
    else
        echo -e "${RED}✗${NC} $2"
        FAILED=1
    fi
}

# 1. Check Go formatting
echo ""
echo "📝 Checking Go formatting..."
GO_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' || true)
if [ -n "$GO_FILES" ]; then
    UNFORMATTED=$(gofmt -l $GO_FILES)
    if [ -n "$UNFORMATTED" ]; then
        echo -e "${RED}✗${NC} Go files need formatting:"
        echo "$UNFORMATTED"
        echo -e "${YELLOW}Run: make fmt${NC}"
        FAILED=1
    else
        print_status 0 "Go files are formatted"
    fi
else
    echo "  No Go files to check"
fi

# 2. Check Terraform formatting
echo ""
echo "📝 Checking Terraform formatting..."
TF_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.tf$' || true)
if [ -n "$TF_FILES" ]; then
    terraform fmt -check -recursive . > /dev/null 2>&1
    print_status $? "Terraform files are formatted"
    if [ $? -ne 0 ]; then
        echo -e "${YELLOW}Run: make fmt${NC}"
    fi
else
    echo "  No Terraform files to check"
fi

# 3. Run golangci-lint (if installed)
echo ""
echo "🔎 Running Go linter..."
if command -v golangci-lint &> /dev/null; then
    if [ -n "$GO_FILES" ]; then
        golangci-lint run --new-from-rev=HEAD~1 > /dev/null 2>&1
        print_status $? "Go linter passed"
        if [ $? -ne 0 ]; then
            echo -e "${YELLOW}Run: make lint${NC}"
        fi
    else
        echo "  No Go files to lint"
    fi
else
    echo -e "${YELLOW}⚠${NC}  golangci-lint not installed (optional)"
    echo -e "  Install: brew install golangci-lint"
fi

# 4. Run tflint (if installed)
echo ""
echo "🔎 Running Terraform linter..."
if command -v tflint &> /dev/null; then
    if [ -n "$TF_FILES" ]; then
        tflint --init > /dev/null 2>&1
        tflint --recursive > /dev/null 2>&1
        print_status $? "Terraform linter passed"
        if [ $? -ne 0 ]; then
            echo -e "${YELLOW}Run: make lint-tf${NC}"
        fi
    else
        echo "  No Terraform files to lint"
    fi
else
    echo -e "${YELLOW}⚠${NC}  tflint not installed (optional)"
    echo -e "  Install: brew install tflint"
fi

# 5. Run go vet on staged Go files
echo ""
echo "🔍 Running go vet..."
GO_CODE_FILES=$(git diff --cached --name-only --diff-filter=ACM | grep '\.go$' | grep -v '_test\.go$' || true)
if [ -n "$GO_CODE_FILES" ]; then
    # Get unique package directories
    GO_PACKAGES=$(echo "$GO_CODE_FILES" | xargs -I {} dirname {} | sort -u | sed 's|^|./|')
    if [ -n "$GO_PACKAGES" ]; then
        go vet $GO_PACKAGES > /dev/null 2>&1
        print_status $? "go vet checks passed"
        if [ $? -ne 0 ]; then
            echo -e "${YELLOW}Run: go vet ./...${NC}"
        fi
    fi
else
    echo "  No Go files to check"
fi

# Check for large files
LARGE_FILES=$(git diff --cached --name-only --diff-filter=ACM | xargs ls -lh 2>/dev/null | awk '$5 ~ /M$/ && $5+0 > 1' || true)
if [ -n "$LARGE_FILES" ]; then
    echo -e "${RED}✗${NC} Found large files (>1MB):"
    echo "$LARGE_FILES"
    FAILED=1
fi

# Final result
echo ""
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All pre-commit checks passed!${NC}"
    echo ""
    exit 0
else
    echo -e "${RED}✗ Some pre-commit checks failed${NC}"
    echo ""
    echo -e "${YELLOW}Fix the issues above or use:${NC}"
    echo -e "  git commit --no-verify  ${YELLOW}(to skip checks)${NC}"
    echo ""
    exit 1
fi
