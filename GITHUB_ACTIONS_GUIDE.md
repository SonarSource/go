# GitHub Actions Quick Reference for go Repository

## Overview

This repository now uses GitHub Actions for CI/CD. This guide provides quick answers to common questions.

## Workflows

### 1. Build Workflow (`.github/workflows/build.yml`)

**Triggers:**
- Push to `master`, `branch-*`, or `dogfood-*` branches
- Pull requests
- Merge queue events
- Manual trigger via workflow_dispatch

**What it does:**
- Builds and tests the Go code
- Runs SonarQube analysis on Next platform
- Optionally runs shadow scans to SonarCloud EU and US (when `shadow_scan` label is present on PR)

### 2. Nightly Workflow (`.github/workflows/nightly.yml`)

**Triggers:**
- Scheduled daily at 03:00 UTC
- Manual trigger via workflow_dispatch

**What it does:**
- Same as build workflow
- Always runs shadow scans to SonarCloud EU and US

### 3. PR Cleanup Workflow (`.github/workflows/pr-cleanup.yml`)

**Triggers:**
- When a pull request is closed

**What it does:**
- Automatically cleans up caches and artifacts associated with the PR

## Running Shadow Scans on Pull Requests

To trigger shadow scans on a pull request:

1. Add the `shadow_scan` label to your PR
2. The shadow scan jobs will execute after the main build completes

Shadow scans analyze your code on:
- SonarCloud EU (https://sonarcloud.io)
- SonarCloud US (https://sonarqube.us)

## Manual Workflow Execution

You can manually trigger workflows from the GitHub UI:

1. Go to **Actions** tab
2. Select the workflow (Build or Nightly)
3. Click **Run workflow**
4. Select the branch
5. Click **Run workflow** button

## Viewing Build Results

### GitHub Actions UI
1. Go to **Actions** tab
2. Click on a workflow run to see details
3. Click on a job to see logs

### SonarQube Results
- **Next Platform**: https://next.sonarqube.com/sonarqube
- **SonarCloud EU**: https://sonarcloud.io
- **SonarCloud US**: https://sonarqube.us

## Common Tasks

### Running Tests Locally

```bash
cd src
go test -v ./... -coverprofile=coverage.out
```

### Building Locally

```bash
cd src
go build -v ./...
```

### Using Mise for Go Version

The project uses [mise](https://mise.jdx.dev/) for tool management:

```bash
# Install mise (if not already installed)
curl https://mise.jdx.dev/install.sh | sh

# Install tools specified in mise.toml
mise install

# Go will be available at the version specified in mise.toml (1.25.1)
go version
```

## Troubleshooting

### Build Failures

1. **Check the logs**: Click on the failed job in GitHub Actions
2. **Verify Go version**: Ensure you're using Go 1.25.1 (specified in `mise.toml`)
3. **Check dependencies**: Ensure `src/go.sum` is up to date

### Shadow Scan Not Running

Shadow scans only run when:
- On nightly schedule (daily at 03:00 UTC), OR
- When PR has the `shadow_scan` label

To trigger manually on a PR:
1. Add the `shadow_scan` label to your PR
2. Push a new commit or re-run the workflow

### Cache Issues

If you suspect cache issues:
1. Caches are automatically managed per branch
2. PR cleanup workflow removes caches when PRs close
3. To clear cache for a branch, you can manually delete it from:
   - **Settings** → **Actions** → **Caches**

## Workflow Permissions

The workflows use these permissions:
- **id-token: write** - For Vault OIDC authentication
- **contents: write/read** - For repository access
- **actions: write** - For PR cleanup (to delete caches/artifacts)

## Build Numbers

Build numbers are automatically managed and increment with each build. They're used for:
- SonarQube analysis tracking (`sonar.analysis.buildNumber`)
- Artifact versioning (if applicable)

## Environment Variables

The workflows automatically set these environment variables:
- `BUILD_NUMBER` - Unique build identifier
- `GITHUB_REPO` - Repository full name (SonarSource/go)
- `SONAR_TOKEN` - Retrieved from Vault
- `SONAR_HOST_URL` - Retrieved from Vault or hardcoded for shadow scans

## Runners

This repository uses **self-hosted runners** (`sonar-xs`) because it's a private repository.

## Support

For issues or questions about the CI/CD setup:
1. Check the [Migration Summary](./MIGRATION_SUMMARY.md)
2. Review the [SonarSource CI GitHub Actions documentation](https://github.com/SonarSource/ci-github-actions)
3. Contact the platform team

## Differences from Cirrus CI

If you're familiar with the old Cirrus CI setup:

| Cirrus CI | GitHub Actions |
|-----------|----------------|
| `build_task` | `build` job |
| `shadow_scan_sqc_eu_task` | `shadow_scan_sqc_eu` job |
| `shadow_scan_sqc_us_task` | `shadow_scan_sqc_us` job |
| Runs in custom Docker container | Uses `mise` for Go installation |
| Manual `sonar-scanner` script | Uses `sonarsource/sonarqube-scan-action` |
| Secrets via `VAULT[...]` | Secrets via `vault-action-wrapper` |

