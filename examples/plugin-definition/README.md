# Plugin Definition Example

This example demonstrates how to **register a plugin definition** in Chainlaunch using Terraform. This is **Part 1** of the plugin workflow - defining what the plugin is and what it does.

## What This Example Does

1. **Registers Plugin**: Creates a plugin definition from a YAML file
2. **Validates Structure**: Ensures the plugin YAML is valid
3. **Stores Metadata**: Saves plugin information (name, version, description, etc.)

## Two-Step Plugin Workflow

```
┌─────────────────────┐       ┌─────────────────────┐
│  1. Plugin          │       │  2. Plugin          │
│     Definition      │  -->  │     Deployment      │
│  (This Example)     │       │  (Separate Example) │
└─────────────────────┘       └─────────────────────┘
       Register                     Deploy with
       YAML spec                    parameters
```

**This example**: Register the plugin (define WHAT it is)
**Next example**: Deploy the plugin (define HOW to run it)

## Plugin YAML Structure

The `plugin.yaml` file defines:

```yaml
apiVersion: dev.chainlaunch/v1
kind: Plugin
metadata:
  name: hlf-plugin-api          # Unique identifier
  version: '1.0'                # Plugin version
  description: 'API for Fabric' # What it does

spec:
  dockerCompose:
    contents: |
      # Docker Compose configuration with Go templates
      # Uses {{ .parameters.* }} for dynamic values

  parameters:
    # JSON Schema defining what parameters are needed for deployment

  metrics:
    # Prometheus endpoints configuration
```

## Prerequisites

- **Chainlaunch Instance**: Running and accessible
- **Plugin YAML File**: Valid plugin definition (provided as `plugin.yaml`)

## Usage

### 1. Review the Plugin YAML

```bash
cat plugin.yaml
```

This shows the complete plugin specification including:
- Docker Compose configuration
- Parameter schema
- Metrics endpoints
- Documentation

### 2. Apply

```bash
# Initialize
terraform init

# Register the plugin
terraform apply
```

### 3. Verify

```bash
# View outputs
terraform output summary

# Query via API
curl -s -u admin:admin123 http://localhost:8100/api/v1/plugins/hlf-plugin-api \
  | jq '.'
```

## Configuration Options

### Using a File Path (Default)

```hcl
resource "chainlaunch_plugin" "hlf_api" {
  yaml_file_path = "./plugin.yaml"
}
```

### Using Inline Content

```hcl
resource "chainlaunch_plugin" "hlf_api" {
  yaml_content = file("${path.module}/plugin.yaml")
}
```

### Using Templated Content

```hcl
resource "chainlaunch_plugin" "hlf_api" {
  yaml_content = templatefile("${path.module}/plugin.yaml.tpl", {
    plugin_name    = "my-custom-plugin"
    plugin_version = "2.0"
  })
}
```

## What Gets Created

After applying, Chainlaunch stores:

- ✅ **Plugin Metadata**: Name, version, description, author
- ✅ **Docker Compose Template**: Service definitions with Go templates
- ✅ **Parameter Schema**: JSON Schema for deployment validation
- ✅ **Metrics Configuration**: Prometheus endpoint definitions
- ✅ **Documentation**: Inline README and examples

## Important Notes

### Plugin Definition vs Deployment

**Plugin Definition (This Example)**:
- Registers WHAT the plugin is
- Defines the Docker Compose template
- Specifies parameter requirements
- No containers are started
- No parameters are provided

**Plugin Deployment (Next Step)**:
- Deploys a registered plugin
- Provides actual parameter values
- Starts Docker Compose services
- Creates running containers

### Updating Plugins

When you update the plugin YAML:

```bash
# Edit plugin.yaml
vim plugin.yaml

# Apply changes
terraform apply
```

Terraform will update the plugin definition. **Note**: Existing deployments continue running with the old definition until redeployed.

### Deleting Plugins

```bash
terraform destroy
```

**Warning**: You cannot delete a plugin that has active deployments. Stop all deployments first using the `plugin-deployment` example.

## Next Steps

### Deploy This Plugin

Once registered, deploy the plugin:

```bash
cd ../plugin-deployment
terraform apply -var="plugin_name=hlf-plugin-api"
```

See the [plugin-deployment example](../plugin-deployment/) for details.

### Create Your Own Plugin

1. **Copy the template**:
   ```bash
   cp plugin.yaml my-plugin.yaml
   ```

2. **Modify the YAML**:
   - Change `metadata.name` to your plugin name
   - Update Docker Compose services
   - Define your parameter schema
   - Add metrics endpoints

3. **Register it**:
   ```hcl
   resource "chainlaunch_plugin" "my_plugin" {
     yaml_file_path = "./my-plugin.yaml"
   }
   ```

### Query Existing Plugins

Use the data source to reference existing plugins:

```hcl
data "chainlaunch_plugin" "existing" {
  name = "hlf-plugin-api"
}

output "version" {
  value = data.chainlaunch_plugin.existing.metadata_version
}
```

## API Operations

This example uses:

- **POST /plugins**: Register new plugin
- **GET /plugins/{name}**: Query plugin details
- **PUT /plugins/{name}**: Update plugin definition
- **DELETE /plugins/{name}**: Remove plugin (if no deployments)

## Troubleshooting

### Plugin Creation Fails

**Check YAML syntax:**
```bash
# Validate YAML
python -c "import yaml; yaml.safe_load(open('plugin.yaml'))"
```

**Common issues:**
- Invalid YAML indentation
- Missing required fields (name, apiVersion, kind)
- Invalid JSON Schema in parameters section

### Plugin Already Exists

If you get "plugin already exists":

1. **Import existing plugin**:
   ```bash
   terraform import chainlaunch_plugin.hlf_api hlf-plugin-api
   ```

2. **Or delete and recreate**:
   ```bash
   curl -X DELETE -u admin:admin123 http://localhost:8100/api/v1/plugins/hlf-plugin-api
   terraform apply
   ```

### Cannot Delete Plugin

If deletion fails with "plugin has active deployments":

1. **List deployments**:
   ```bash
   curl -s -u admin:admin123 http://localhost:8100/api/v1/plugins/hlf-plugin-api/deployment-status
   ```

2. **Stop deployment**:
   ```bash
   cd ../plugin-deployment
   terraform destroy
   ```

3. **Then delete definition**:
   ```bash
   cd ../plugin-definition
   terraform destroy
   ```

## Example Output

```
plugin_name        = "hlf-plugin-api"
plugin_version     = "1.0"
plugin_description = "Hyperledger Fabric API plugin..."
plugin_author      = "ChainLaunch Team"
plugin_repository  = "https://github.com/kfsoftware/plugin-hlf-api"

summary = <<EOT

╔══════════════════════════════════════════════════════════════╗
║              Plugin Registered Successfully                  ║
╚══════════════════════════════════════════════════════════════╝

Plugin Details:
  Name:        hlf-plugin-api
  Version:     1.0
  API Version: dev.chainlaunch/v1
  ...

Next Steps:
1. Query the plugin: terraform output plugin_name
2. Deploy the plugin using the plugin-deployment example
3. Or query via API: curl -s -u admin:admin123 ...
EOT
```

## Related Examples

- [Plugin Deployment](../plugin-deployment/) - Deploy a registered plugin with parameters
- [Plugin HLF API](../plugin-hlf-api/) - Complete end-to-end example (definition + deployment)

## Additional Resources

- [Plugin YAML Reference](../../docs/plugin-yaml-reference.md)
- [Chainlaunch Plugin API](../../swagger.yaml) - See `/plugins` endpoints
- [Custom Plugin Development Guide](../../docs/custom-plugins.md)
