# GitHub Actions CI/CD

This directory contains GitHub Actions workflows for continuous integration and deployment.

## Workflows

### CI Workflow (`.github/workflows/ci.yml`)

Runs on every push and pull request to the `main` branch.

**Jobs:**

1. **Lint & Format** - Code quality checks
   - Go formatting check (`gofmt`)
   - Go vet
   - Terraform formatting check
   - golangci-lint

2. **Build Provider** - Multi-platform builds
   - Builds for: Linux, macOS, Windows (amd64, arm64)
   - Uploads build artifacts (retained for 7 days)

3. **Unit Tests** - Fast unit test suite
   - Runs all unit tests
   - Generates coverage report
   - Uploads to Codecov (if configured)

4. **Acceptance Tests** - Integration tests
   - Only runs on PRs and main branch pushes
   - Requires test instance credentials (see secrets below)

5. **Generate Documentation** - Validate docs
   - Generates documentation using `tfplugindocs`
   - Validates documentation format
   - Checks for uncommitted changes

6. **Validate Examples** - Terraform validation
   - Validates all example configurations
   - Ensures examples are syntactically correct

### Release Workflow (`.github/workflows/release.yml`)

Runs when you push a version tag (e.g., `v1.0.0`).

**Jobs:**

1. **Build and Release** - Multi-platform builds
   - Builds binaries for: Linux (amd64, arm64), macOS (amd64, arm64), Windows (amd64)
   - Creates tar.gz/zip archives
   - Generates SHA256 checksums
   - Uploads artifacts

2. **Create Release** - GitHub release
   - Combines all checksums
   - Generates release notes from CHANGELOG.md
   - Creates GitHub release with all binaries

3. **Update Documentation** - Auto-commit docs
   - Regenerates documentation
   - Commits to main branch

## Required Secrets

Configure these secrets in your GitHub repository settings:

### For Acceptance Tests (Optional)

- `CHAINLAUNCH_URL` - URL of test Chainlaunch instance
- `CHAINLAUNCH_USERNAME` - Test instance username
- `CHAINLAUNCH_PASSWORD` - Test instance password

The `GITHUB_TOKEN` is automatically provided by GitHub Actions and has write permissions for releases.

## Dependabot

Dependabot is configured (`.github/dependabot.yml`) to:
- Update Go dependencies weekly
- Update GitHub Actions weekly
- Create labeled PRs automatically

## Creating a Release

1. **Update version and commit:**
   ```bash
   # Update CHANGELOG.md
   git add CHANGELOG.md
   git commit -m "chore: prepare release v1.0.0"
   git push
   ```

2. **Create and push tag:**
   ```bash
   git tag -a v1.0.0 -m "Release v1.0.0"
   git push origin v1.0.0
   ```

3. **GitHub Actions will:**
   - Build binaries for all platforms
   - Sign checksums
   - Create GitHub release with changelog
   - Update documentation

## Local Testing

Test the CI workflow locally:

```bash
# Run all checks that CI runs
make check

# Test documentation generation
make docs

# Build binary locally
make build
```

## Workflow Status

Check workflow status:
- Click on the "Actions" tab in GitHub
- View logs for each workflow run
- Download build artifacts from completed runs

## Troubleshooting

**Build fails on "go vet":**
- Run `make lint` locally to see issues
- Fix any vet warnings before pushing

**Documentation check fails:**
- Run `make docs` locally to regenerate
- Commit the updated docs

**Acceptance tests skipped:**
- Tests only run if `CHAINLAUNCH_URL` secret is set
- This is normal for public repos without test instances

**Release build fails:**
- Check that all platforms build successfully locally
- Test build: `GOOS=linux GOARCH=amd64 make build`
