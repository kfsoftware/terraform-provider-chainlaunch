# Complete Example

This comprehensive example demonstrates all major features of the Chainlaunch Terraform provider including data sources, resource creation, and various configuration patterns.

## What This Example Does

- Fetches all available key providers with filtering
- Creates multiple organizations using different key provider selection methods
- Demonstrates comprehensive output configuration
- Shows how to read existing organizations (example commented out)

## Features Demonstrated

### Data Sources

1. **Key Providers Data Source**
   - Fetch all key providers
   - Filter by type (database, AWS KMS, etc.)
   - Access default provider details

2. **Organization Data Source** (optional, currently commented out)
   - Read existing organization by ID or MSP ID
   - Access organization details (MSP ID, description)
   - Update the ID in the example to match an existing organization in your instance

### Resources

1. **Multiple Organizations**
   - Create Organization 1 using default key provider
   - Create Organization 2 using filtered database provider
   - Both with full configuration (name, MSP ID, description, provider_id)

### Outputs

- All key providers list
- Default provider details
- Existing organization details
- Created organizations details with IDs and configuration

## Configuration

The example uses the following default values:

- **Chainlaunch URL**: `http://localhost:8100`
- **Username**: `admin`
- **Password**: `admin123`

Override these by creating a `terraform.tfvars` file:

```hcl
chainlaunch_url      = "https://your-chainlaunch-instance.com"
chainlaunch_username = "your-username"
chainlaunch_password = "your-password"
```

## Usage

1. Navigate to this directory:
   ```bash
   cd examples/complete
   ```

2. Review the configuration:
   ```bash
   terraform plan
   ```

   You should see:
   - 2 organizations to be created
   - 2 data sources to be read (key providers)
   - Multiple outputs configured

3. Apply the configuration:
   ```bash
   terraform apply
   ```

4. View all outputs:
   ```bash
   terraform output
   ```

5. View specific output:
   ```bash
   terraform output org1_details
   ```

## Expected Outputs

- `all_key_providers`: Complete list of key providers
- `default_provider`: Default key provider (ID, name, type)
- `org1_details`: Created organization 1 details
- `org2_details`: Created organization 2 details

Note: The `existing_organization` output is commented out by default. To use it, uncomment the data source and output in `main.tf` and update the ID to match an existing organization.

## Resources Created

- 2 Hyperledger Fabric organizations
  - CompleteOrg1 (CompleteOrg1MSP) - using default provider
  - CompleteOrg2 (CompleteOrg2MSP) - using database provider

## Key Concepts

### Provider Selection

**Method 1: Using Default Provider**
```hcl
provider_id = tonumber(data.chainlaunch_key_providers.all.default_id)
```

**Method 2: Using Filtered Provider**
```hcl
provider_id = tonumber(data.chainlaunch_key_providers.database.providers[0].id)
```

### Data Source vs Resource

- **Data Source**: Read existing resources (non-destructive)
  ```hcl
  # Query by ID
  data "chainlaunch_fabric_organization" "existing" {
    id = "1"
  }

  # Or query by MSP ID (recommended - more stable)
  data "chainlaunch_fabric_organization" "existing" {
    msp_id = "Org1MSP"
  }
  ```

- **Resource**: Create/manage resources (creates infrastructure)
  ```hcl
  resource "chainlaunch_fabric_organization" "org1" {
    msp_id      = "Org1MSP"
    description = "First organization"
    provider_id = data.chainlaunch_key_providers.all.default_provider_id
  }
  ```

## Clean Up

To destroy all created resources:

```bash
terraform destroy
```

Note: This will only destroy resources created by Terraform (the 2 new organizations). Data sources (reading existing resources) are not affected.
