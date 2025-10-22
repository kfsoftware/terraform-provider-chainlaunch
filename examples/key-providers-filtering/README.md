# Key Providers Filtering Example

This example demonstrates advanced usage of the key providers data source with various filtering options.

## What This Example Does

- Fetches all available key providers
- Filters key providers by type (database)
- Filters key providers by name (partial match)
- Demonstrates accessing provider details through data sources
- Creates an organization using the default key provider

## Features Demonstrated

### Data Source Filtering

1. **All Providers**: Fetch all key providers without filtering
   ```hcl
   data "chainlaunch_key_providers" "all" {}
   ```

2. **Filter by Type**: Get only database providers
   ```hcl
   data "chainlaunch_key_providers" "database" {
     type_filter = "database"
   }
   ```

3. **Filter by Name**: Search providers by name (case-insensitive)
   ```hcl
   data "chainlaunch_key_providers" "by_name" {
     name_filter = "Default"
   }
   ```

### Default Provider Access

The data source automatically identifies the default provider:

```hcl
data.chainlaunch_key_providers.all.default_id
data.chainlaunch_key_providers.all.default_name
data.chainlaunch_key_providers.all.default_type
```

## Configuration

The example uses the following default values:

- **Chainlaunch URL**: `http://localhost:8100`
- **Username**: `admin`
- **Password**: `admin123`

## Usage

1. Navigate to this directory:
   ```bash
   cd examples/key-providers-filtering
   ```

2. Review the configuration:
   ```bash
   terraform plan
   ```

3. Create the organization:
   ```bash
   terraform apply
   ```

## Expected Outputs

- `all_providers`: List of all available key providers
- `database_providers`: Filtered list of database providers
- `providers_by_name`: Providers matching the name filter
- `default_provider`: Details of the default key provider
- `using_default_provider`: Example of using the default provider ID

## Resources Created

- 1 Hyperledger Fabric organization using the default key provider

## Clean Up

```bash
terraform destroy
```
