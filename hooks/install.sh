#!/bin/bash

# Git hooks installation script for chainlaunch-terraform provider

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "  Chainlaunch Terraform Provider - Git Hooks"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo -e "${YELLOW}ℹ${NC}  Pre-commit hooks are DISABLED by default to save time during development."
echo ""
echo "  Instead, run code quality checks manually when needed:"
echo -e "  ${GREEN}make check-code${NC}    - Run all pre-commit checks"
echo -e "  ${GREEN}make check${NC}         - Run format, lint, and validation checks"
echo ""
echo "  To enable automatic pre-commit hooks, run:"
echo -e "  ${GREEN}ln -sf ../../scripts/check-code.sh .git/hooks/pre-commit${NC}"
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
