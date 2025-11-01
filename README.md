# Terraform Provider for Chainlaunch

This Terraform provider allows you to manage Hyperledger Fabric resources in Chainlaunch, including organizations, nodes, networks, and key providers.

**Website**: https://chainlaunch.dev

## Requirements

- [Terraform](https://www.terraform.io/downloads.html) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.21 (for development)
- Chainlaunch API access with valid credentials

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/chainlaunch/terraform-provider-chainlaunch.git
cd terraform-provider-chainlaunch

# Build and install
go install
```

### Using Terraform

Add the following to your Terraform configuration:

```hcl
terraform {
  required_providers {
    chainlaunch = {
      source  = "kfsoftware/chainlaunch"
      version = "~> 1.0"
    }
  }
}
```

## Configuration

### Provider Configuration

The provider supports two authentication methods:

#### Option 1: API Key Authentication

```hcl
provider "chainlaunch" {
  url     = "https://your-chainlaunch-instance.com"
  api_key = "your-api-key"
}
```

Or using environment variables:

```bash
export CHAINLAUNCH_URL="https://your-chainlaunch-instance.com"
export CHAINLAUNCH_API_KEY="your-api-key"
```

#### Option 2: Username/Password Authentication

```hcl
provider "chainlaunch" {
  url      = "http://localhost:8100"
  username = "admin"
  password = "admin123"
}
```

Or using environment variables:

```bash
export CHAINLAUNCH_URL="http://localhost:8100"
export CHAINLAUNCH_USERNAME="admin"
export CHAINLAUNCH_PASSWORD="admin123"
```

## Usage Examples

### Creating a Fabric Organization

```hcl
resource "chainlaunch_fabric_organization" "my_org" {
  msp_id      = "MyOrgMSP"
  description = "My Hyperledger Fabric Organization"
}
```

### Creating a Fabric Peer Node

```hcl
resource "chainlaunch_node" "peer" {
  name     = "peer0"
  platform = "fabric"
  type     = "peer"

  config = jsonencode({
    organizationId = chainlaunch_fabric_organization.my_org.id
    port          = 7051
    couchdbPort   = 5984
  })
}
```

### Creating a Fabric Orderer Node

```hcl
resource "chainlaunch_node" "orderer" {
  name     = "orderer0"
  platform = "fabric"
  type     = "orderer"

  config = jsonencode({
    organizationId = chainlaunch_fabric_organization.my_org.id
    port          = 7050
    consensusType = "etcdraft"
  })
}
```

### Creating a Fabric Network

```hcl
resource "chainlaunch_network" "fabric_network" {
  name = "my-fabric-network"
  type = "fabric"

  config = jsonencode({
    channelName = "mychannel"
    organizations = [
      {
        id = chainlaunch_fabric_organization.my_org.id
      }
    ]
    orderers = [
      {
        id = chainlaunch_node.orderer.id
      }
    ]
    peers = [
      {
        id = chainlaunch_node.peer.id
      }
    ]
  })
}
```

### Creating a Key Provider (Vault)

```hcl
resource "chainlaunch_key_provider" "vault" {
  name = "my-vault-provider"
  type = "vault"

  config = jsonencode({
    address = "https://vault.example.com:8200"
    token   = var.vault_token
  })
}
```

### Using Data Sources

#### Fetch an Existing Organization

```hcl
data "chainlaunch_fabric_organization" "existing_org" {
  id = "1"
}

output "org_msp_id" {
  value = data.chainlaunch_fabric_organization.existing_org.msp_id
}
```

#### Fetch an Existing Node

```hcl
data "chainlaunch_node" "existing_peer" {
  id = "5"
}

output "node_status" {
  value = data.chainlaunch_node.existing_peer.status
}
```

#### Fetch an Existing Network

```hcl
data "chainlaunch_network" "existing_network" {
  id   = "3"
  type = "fabric"
}

output "network_name" {
  value = data.chainlaunch_network.existing_network.name
}
```

## Complete Example

Here's a complete example that sets up a Fabric network with organizations, nodes, and a key provider:

```hcl
terraform {
  required_providers {
    chainlaunch = {
      source  = "kfsoftware/chainlaunch"
      version = "~> 1.0"
    }
  }
}

provider "chainlaunch" {
  url     = var.chainlaunch_url
  api_key = var.chainlaunch_api_key
}

# Create Organization
resource "chainlaunch_fabric_organization" "org1" {
  msp_id      = "Org1MSP"
  description = "First organization in the network"
}

# Create Key Provider
resource "chainlaunch_key_provider" "vault" {
  name = "org1-vault"
  type = "vault"

  config = jsonencode({
    address = "https://vault.example.com:8200"
    token   = var.vault_token
  })
}

# Create Orderer Node
resource "chainlaunch_node" "orderer" {
  name     = "orderer.org1.example.com"
  platform = "fabric"
  type     = "orderer"

  config = jsonencode({
    organizationId = chainlaunch_fabric_organization.org1.id
    port          = 7050
    consensusType = "etcdraft"
  })
}

# Create Peer Node
resource "chainlaunch_node" "peer0" {
  name     = "peer0.org1.example.com"
  platform = "fabric"
  type     = "peer"

  config = jsonencode({
    organizationId = chainlaunch_fabric_organization.org1.id
    port          = 7051
    couchdbPort   = 5984
  })
}

# Create Fabric Network
resource "chainlaunch_network" "channel1" {
  name = "mychannel"
  type = "fabric"

  config = jsonencode({
    channelName = "mychannel"
    organizations = [
      {
        id = chainlaunch_fabric_organization.org1.id
      }
    ]
    orderers = [
      {
        id = chainlaunch_node.orderer.id
      }
    ]
    peers = [
      {
        id = chainlaunch_node.peer0.id
      }
    ]
  })

  depends_on = [
    chainlaunch_node.orderer,
    chainlaunch_node.peer0
  ]
}

# Outputs
output "organization_id" {
  value = chainlaunch_fabric_organization.org1.id
}

output "orderer_id" {
  value = chainlaunch_node.orderer.id
}

output "peer_id" {
  value = chainlaunch_node.peer0.id
}

output "network_id" {
  value = chainlaunch_network.channel1.id
}
```

## Resource Reference

### `chainlaunch_fabric_organization`

#### Arguments

- `name` (Required) - The name of the organization
- `msp_id` (Required) - The MSP ID of the organization
- `description` (Optional) - A description of the organization

#### Attributes

- `id` - The unique identifier of the organization
- `created_at` - Timestamp when the organization was created
- `updated_at` - Timestamp when the organization was last updated

### `chainlaunch_node`

#### Arguments

- `name` (Required) - The name of the node
- `platform` (Required) - The blockchain platform (e.g., "fabric", "besu")
- `type` (Required) - The type of node (e.g., "peer", "orderer")
- `config` (Optional) - JSON configuration for the node

#### Attributes

- `id` - The unique identifier of the node
- `status` - The current status of the node
- `created_at` - Timestamp when the node was created
- `updated_at` - Timestamp when the node was last updated

### `chainlaunch_network`

#### Arguments

- `name` (Required) - The name of the network
- `type` (Required) - The type of network (e.g., "fabric", "besu")
- `config` (Optional) - JSON configuration for the network

#### Attributes

- `id` - The unique identifier of the network
- `status` - The current status of the network
- `created_at` - Timestamp when the network was created
- `updated_at` - Timestamp when the network was last updated

### `chainlaunch_key_provider`

#### Arguments

- `name` (Required) - The name of the key provider
- `type` (Required) - The type of key provider (e.g., "vault", "aws-kms")
- `config` (Optional, Sensitive) - JSON configuration for the key provider

#### Attributes

- `id` - The unique identifier of the key provider
- `status` - The current status of the key provider
- `created_at` - Timestamp when the key provider was created
- `updated_at` - Timestamp when the key provider was last updated

## Development

### Building the Provider

```bash
go build -o terraform-provider-chainlaunch
```

### Generating the API Client

The provider uses a generated client from the Chainlaunch OpenAPI specification:

```bash
# Install go-swagger
go install github.com/go-swagger/go-swagger/cmd/swagger@latest

# Generate the client
swagger generate client -f swagger.yaml -t internal/generated
```

### Running Tests

```bash
go test ./...
```

### Using the Provider Locally

1. Build the provider:
```bash
go build -o terraform-provider-chainlaunch
```

2. Create a local provider configuration in `~/.terraformrc`:
```hcl
provider_installation {
  dev_overrides {
    "kfsoftware/chainlaunch" = "/path/to/terraform-provider-chainlaunch"
  }
  direct {}
}
```

3. Run Terraform commands as usual.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Support

- **Website**: https://chainlaunch.dev
- **Issues**: For issues, questions, or contributions, please open an issue in the GitHub repository.
