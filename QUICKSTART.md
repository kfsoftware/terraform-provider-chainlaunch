# Quick Start Guide

This guide will help you get started with the Terraform Chainlaunch Provider.

## Prerequisites

1. Go 1.21 or later installed
2. Terraform installed
3. Access to a Chainlaunch instance with API credentials

## Step 1: Build the Provider

```bash
# Download dependencies
make deps

# Build the provider
make build
```

## Step 2: Install the Provider Locally

```bash
# Install to local Terraform plugins directory
make install
```

Alternatively, for development, create `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "registry.terraform.io/kfsoftware/chainlaunch" = "/path/to/terraform-provider-chainlaunch"
  }
  direct {}
}
```

## Step 3: Configure Your Credentials

The provider supports two authentication methods. Choose one:

### Option A: Username/Password Authentication

Create a `terraform.tfvars` file (don't commit this!):

```hcl
chainlaunch_url      = "http://localhost:8100"
chainlaunch_username = "admin"
chainlaunch_password = "admin123"
```

Or set environment variables:

```bash
export CHAINLAUNCH_URL="http://localhost:8100"
export CHAINLAUNCH_USERNAME="admin"
export CHAINLAUNCH_PASSWORD="admin123"
```

### Option B: API Key Authentication

Create a `terraform.tfvars` file (don't commit this!):

```hcl
chainlaunch_url     = "https://your-chainlaunch-instance.com"
chainlaunch_api_key = "your-api-key-here"
```

Or set environment variables:

```bash
export CHAINLAUNCH_URL="https://your-chainlaunch-instance.com"
export CHAINLAUNCH_API_KEY="your-api-key-here"
```

## Step 4: Create Your First Resource

Create a `main.tf` file:

### Using Username/Password Authentication

```hcl
terraform {
  required_providers {
    chainlaunch = {
      source = "registry.terraform.io/kfsoftware/chainlaunch"
    }
  }
}

provider "chainlaunch" {
  url      = var.chainlaunch_url
  username = var.chainlaunch_username
  password = var.chainlaunch_password
}

variable "chainlaunch_url" {
  type = string
}

variable "chainlaunch_username" {
  type = string
}

variable "chainlaunch_password" {
  type      = string
  sensitive = true
}

# Create an organization
resource "chainlaunch_organization" "my_org" {
  name        = "MyOrganization"
  msp_id      = "MyOrgMSP"
  description = "My first Fabric organization"
}

output "organization_id" {
  value = chainlaunch_organization.my_org.id
}
```

Or see [examples/basic-auth.tf](examples/basic-auth.tf) for a complete example.

## Step 5: Initialize and Apply

```bash
# Initialize Terraform
terraform init

# Preview changes
terraform plan

# Apply changes
terraform apply
```

## Step 6: Verify the Resource

```bash
# Show the created organization
terraform show

# Get output values
terraform output
```

## Next Steps

Check out the [README.md](README.md) for more examples, including:
- Creating nodes (peers and orderers)
- Setting up networks
- Configuring key providers
- Using data sources

Or explore the complete example in [examples/main.tf](examples/main.tf).

## Common Commands

```bash
# Generate/regenerate API client from swagger.yaml
make generate

# Format your Terraform files
make fmt

# Run tests
make test

# Clean build artifacts
make clean

# Show all available make targets
make help
```

## Troubleshooting

### Provider not found

If Terraform can't find the provider, make sure:
1. You've run `make install` or configured `dev_overrides` in `~/.terraformrc`
2. The provider binary is in the correct location
3. Your `terraform` block specifies the correct source

### API connection errors

Check:
1. Your Chainlaunch URL is correct and accessible
2. Your API key is valid
3. Network connectivity to the Chainlaunch instance

### Generated client issues

If you need to regenerate the API client:

```bash
make clean
make generate
make build
```

## Getting Help

- Check the [README.md](README.md) for detailed documentation
- Review the [examples/main.tf](examples/main.tf) for usage examples
- Open an issue on GitHub for bugs or questions
