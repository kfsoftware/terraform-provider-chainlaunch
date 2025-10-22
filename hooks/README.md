# Git Hooks for Chainlaunch Terraform Provider

This directory contains Git hooks to ensure code quality before commits.

## Quick Start

### Install Hooks

```bash
# Using make (recommended)
make hooks

# Or manually
bash hooks/install.sh
```

## What Gets Checked

The pre-commit hook automatically runs these checks:

### ✅ **Go Formatting**
- Checks if Go files are formatted with `gofmt`
- Fails if any files need formatting
- Fix: `make fmt`

### ✅ **Terraform Formatting**
- Checks if Terraform files are formatted with `terraform fmt`
- Fails if any files need formatting
- Fix: `make fmt`

### 🔍 **Go Linting** (optional)
- Runs `golangci-lint` if installed
- Checks for code quality issues
- Fix: `make lint`

### 🔍 **Terraform Linting** (optional)
- Runs `tflint` if installed
- Checks for Terraform best practices
- Fix: `make lint-tf`

### ⚠️ **Common Issues**
- Detects debugging statements (`fmt.Print`, `console.log`, etc.)
- Warns about TODO/FIXME comments (excluding `hooks/` directory)
- Prevents committing large files (>1MB)

## Hook Output Example

```bash
$ git commit -m "Add new feature"

🔍 Running pre-commit checks...

📝 Checking Go formatting...
✓ Go files are formatted

📝 Checking Terraform formatting...
✓ Terraform files are formatted

🔎 Running Go linter...
✓ Go linter passed

🔎 Running Terraform linter...
✓ Terraform linter passed

🔍 Checking for common issues...

✓ All pre-commit checks passed!
```

## Skipping Hooks

Sometimes you need to commit without running hooks:

```bash
# Skip all hooks
git commit --no-verify -m "WIP: work in progress"

# Or use the shorthand
git commit -n -m "WIP: work in progress"
```

**⚠️ Warning**: Only skip hooks when absolutely necessary (e.g., WIP commits, emergency fixes).

## Manual Testing

Test the hook without making a commit:

```bash
# Run the hook directly
.git/hooks/pre-commit

# Or use make commands
make fmt-check
make lint
make check
```

## Installation Options

### Option 1: Simple Git Hook (Current)

✅ Works out of the box
✅ No dependencies
✅ Fast and simple
❌ Limited to pre-commit
❌ No framework features

```bash
make hooks
```

### Option 2: Pre-commit Framework (Advanced)

✅ More powerful
✅ Supports multiple hooks
✅ Language-specific hooks
✅ Better caching
❌ Requires Python
❌ Additional setup

**Install:**
```bash
# Install pre-commit framework
pip install pre-commit

# Or with homebrew
brew install pre-commit

# Install hooks
pre-commit install

# Run manually
pre-commit run --all-files
```

The repository includes both configurations:
- `.git/hooks/pre-commit` - Simple bash script
- `.pre-commit-config.yaml` - Framework configuration

## Troubleshooting

### Hook Not Running

**Problem**: Committed code without hook running

**Check:**
```bash
ls -la .git/hooks/pre-commit
# Should show: -rwxr-xr-x (executable)
```

**Fix:**
```bash
make hooks
```

### Hook Failing on Formatted Code

**Problem**: Hook says files need formatting but they look fine

**Cause**: Different tool versions or cached files

**Fix:**
```bash
# Re-format everything
make fmt

# Try committing again
git commit
```

### Linter Not Found

**Problem**: Hook says linter is not installed

**Optional linters** (recommended but not required):

```bash
# Install golangci-lint
brew install golangci-lint

# Install tflint
brew install tflint
```

The hook will work without these, but they provide better code quality checks.

### Hook Too Slow

**Problem**: Pre-commit hook takes too long

**Solutions:**

1. **Lint only changed files:**
   ```bash
   # Edit .git/hooks/pre-commit
   # Change: golangci-lint run
   # To: golangci-lint run --new-from-rev=HEAD~1
   ```

2. **Skip optional checks:**
   ```bash
   # Comment out tflint section in .git/hooks/pre-commit
   ```

3. **Use --no-verify occasionally:**
   ```bash
   git commit --no-verify
   ```

## Customization

### Adding Custom Checks

Edit `hooks/pre-commit` to add your own checks:

```bash
# Example: Check for sensitive data
if git diff --cached | grep -i 'password\|secret\|api.key' > /dev/null; then
    echo "❌ Found sensitive data in commit"
    FAILED=1
fi
```

### Disabling Specific Checks

Comment out sections you don't need:

```bash
# # 3. Run golangci-lint (if installed)
# echo ""
# echo "🔎 Running Go linter..."
# ... rest of the section
```

### Adjusting Thresholds

Modify warning/error conditions:

```bash
# Change large file threshold
LARGE_FILES=$(... | awk '$5+0 > 5')  # Change from 1MB to 5MB
```

## CI/CD Integration

The same checks should run in CI:

```yaml
# .github/workflows/quality.yml
name: Code Quality

on: [push, pull_request]

jobs:
  checks:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Go
        uses: actions/setup-go@v4

      - name: Setup Terraform
        uses: hashicorp/setup-terraform@v2

      - name: Run all checks
        run: make check
```

## Best Practices

1. **Install hooks immediately** after cloning:
   ```bash
   git clone <repo>
   cd <repo>
   make hooks
   ```

2. **Run checks before committing**:
   ```bash
   make check
   git add .
   git commit
   ```

3. **Fix issues promptly**:
   ```bash
   # Hook failed? Fix and retry
   make fmt
   git add .
   git commit
   ```

4. **Keep tools updated**:
   ```bash
   brew upgrade golangci-lint
   brew upgrade tflint
   ```

5. **Don't skip hooks habitually**:
   - Use `--no-verify` sparingly
   - Fix issues instead of bypassing checks
   - Your team will thank you!

## See Also

- [FORMATTING.md](../FORMATTING.md) - Detailed formatting guide
- [Makefile](../Makefile) - All available commands
- [.pre-commit-config.yaml](../.pre-commit-config.yaml) - Framework configuration
- [golangci-lint](https://golangci-lint.run/) - Go linter documentation
- [tflint](https://github.com/terraform-linters/tflint) - Terraform linter documentation
