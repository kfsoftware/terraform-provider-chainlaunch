#!/bin/bash
# Script to initialize Terraform examples with development overrides

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROVIDER_DIR="$(dirname "$SCRIPT_DIR")"

echo "Provider directory: $PROVIDER_DIR"
echo "Example directory: $SCRIPT_DIR/$1"

if [ -z "$1" ]; then
    echo "Usage: $0 <example-directory>"
    echo "Example: $0 fabric-network-complete"
    exit 1
fi

EXAMPLE_DIR="$SCRIPT_DIR/$1"

if [ ! -d "$EXAMPLE_DIR" ]; then
    echo "Error: Example directory not found: $EXAMPLE_DIR"
    exit 1
fi

cd "$EXAMPLE_DIR"

# Remove lock file
echo "Removing .terraform.lock.hcl..."
rm -f .terraform.lock.hcl

# Remove .terraform directory
echo "Removing .terraform directory..."
rm -rf .terraform

# Run terraform init with dev overrides from global ~/.terraformrc
echo "Running terraform init..."
terraform init

echo ""
echo "Done! You can now run:"
echo "  cd $EXAMPLE_DIR"
echo "  terraform plan"
