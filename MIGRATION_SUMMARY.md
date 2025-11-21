# Cirrus CI to GitHub Actions Migration Summary

## Migration Overview

This document describes the migration of the SonarSource/go repository from Cirrus CI to GitHub Actions.

## Repository Information

- **Repository**: SonarSource/go
- **Type**: Private Repository
- **Build System**: Go 1.25.1
- **Runners Used**: `sonar-xs` (private repository self-hosted runners)

## Files Created

### 1. `mise.toml`
- **Purpose**: Tool version management
- **Content**: Specifies Go 1.25.1 version

### 2. `.github/workflows/build.yml`
- **Purpose**: Main build workflow for push, pull requests, and merge groups
- **Jobs**:
  - `build`: Builds, tests, and analyzes code on SonarQube Next
  - `shadow_scan_sqc_eu`: Shadow scan to SonarCloud EU (triggered on nightly or with `shadow_scan` label)
  - `shadow_scan_sqc_us`: Shadow scan to SonarCloud US (triggered on nightly or with `shadow_scan` label)
- **Triggers**: push (master, branch-*, dogfood-*), pull_request, merge_group, workflow_dispatch

### 3. `.github/workflows/nightly.yml`
- **Purpose**: Nightly scheduled builds with shadow scans
- **Jobs**: Same as build.yml but triggered via cron schedule
- **Schedule**: Daily at 03:00 UTC
- **Jobs**:
  - `build`: Builds, tests, and analyzes code on SonarQube Next
  - `shadow_scan_sqc_eu`: Shadow scan to SonarCloud EU
  - `shadow_scan_sqc_us`: Shadow scan to SonarCloud US

### 4. `.github/workflows/pr-cleanup.yml`
- **Purpose**: Automatic cleanup of PR resources when PRs are closed
- **Features**: Uses SonarSource/ci-github-actions/pr_cleanup@v1

## Migration Details

### Build Process

The original Cirrus CI configuration used:
- Custom Docker container with Go and sonar-scanner
- Manual sonar-scanner invocation via `.cirrus/analyze.sh`
- Module caching for Go dependencies

The GitHub Actions workflow uses:
- `mise` for Go version management
- `sonarsource/sonarqube-scan-action` for SonarQube analysis
- SonarSource cache action for Go module caching
- SonarSource vault-action-wrapper for secrets management

### Vault Secrets

All secrets are fetched from HashiCorp Vault using OIDC authentication:

**Build Job (Next)**:
- `development/kv/data/next` → `SONAR_HOST_URL` and `SONAR_TOKEN`

**Shadow Scan EU**:
- `development/kv/data/sonarcloud` → `SONAR_TOKEN`
- Hardcoded `SONAR_HOST_URL`: `https://sonarcloud.io`

**Shadow Scan US**:
- `development/kv/data/sonarqube-us` → `SONAR_TOKEN`
- Hardcoded `SONAR_HOST_URL`: `https://sonarqube.us`

### SonarQube Analysis Parameters

All scans use the following parameters (from `.cirrus/analyze.sh`):
- `sonar.projectKey=SonarSource_go`
- `sonar.organization=sonarsource`
- `sonar.analysis.buildNumber` (from get-build-number action)
- `sonar.analysis.repository` (from GitHub context)
- `sonar.cpd.exclusions=**`
- `sonar.go.duration.statistics=true`
- `sonar.go.coverage.reportPaths=coverage.out`
- `sonar.go.tests.reportPaths=test-report.out`
- `sonar.test.inclusions=**/*_test.go`

### Shadow Scans

The original Cirrus CI configuration included shadow scans that run:
- On nightly cron jobs (when branch is master)
- When PR has the `shadow_scan` label

GitHub Actions implementation:
- **build.yml**: Shadow scans run when label `shadow_scan` is present on PRs
- **nightly.yml**: Shadow scans run on scheduled builds (daily at 03:00 UTC)

### Action Versions Used

Following the migration guide exactly, these action versions were used:

| Action | Version (Commit SHA) | Semantic Version |
|--------|---------------------|------------------|
| `actions/checkout` | `08c6903cd8c0fde910a37f88322edcfb5dd907a8` | v5.0.0 |
| `jdx/mise-action` | `5ac50f778e26fac95da98d50503682459e86d566` | v3.2.0 |
| `sonarsource/sonarqube-scan-action` | `fd88b7d7ccbaefd23d8f36f73b59db7a3d246602` | v6.0.0 |
| `SonarSource/vault-action-wrapper` | `@v3` | v3 (semantic versioning) |
| `SonarSource/ci-github-actions/cache` | `@v1` | v1 (semantic versioning) |
| `SonarSource/ci-github-actions/get-build-number` | `@v1` | v1 (semantic versioning) |
| `SonarSource/ci-github-actions/pr_cleanup` | `@v1` | v1 (semantic versioning) |

**Note**: Mise version `2025.7.12` is specified as required by the migration guide.

### Key Differences from Standard Migration

This Go project migration differs from typical Maven/Gradle migrations:
1. **No promote job**: Go projects don't deploy artifacts to Artifactory
2. **Custom sonar-scanner**: Uses manual sonar-scanner invocation instead of build actions
3. **Shadow scans**: Implements multiple platform scanning (Next, SonarCloud EU, SonarCloud US)
4. **Source directory**: Code is in `src/` subdirectory

### Concurrency Control

Workflow-level concurrency control is configured to:
- Group by workflow and PR number (or ref for branches)
- Cancel in-progress runs when new commits are pushed

## Permissions Required

All jobs use these permissions:
- `id-token: write` - Required for Vault OIDC authentication
- `contents: write` - Required for repository access (build job)
- `contents: read` - Required for repository access (shadow scan jobs)
- `actions: write` - Required for PR cleanup job

## Cirrus CI Files

**Important**: During the migration period, the following Cirrus CI files should remain unchanged:
- `.cirrus.star`
- `.cirrus/Dockerfile`
- `.cirrus/analyze.sh`

Both Cirrus CI and GitHub Actions will coexist during the transition period.

## Post-Migration Steps

### 1. Configure Build Number
After migration, configure the GitHub build number using SPEED:
- Login to [SPEED](https://app.getport.io/self-serve?action=update_github_build_number)
- Run the "Update GitHub Build Number" action
- Set the build number to be greater than the latest Cirrus CI build

### 2. Verify Vault Permissions
Ensure the repository has required Vault permissions in `re-terraform-aws-vault/orders`:
```yaml
go:
  auth:
    github: {}
  secrets:
    kv_paths:
      development/kv/data/next: {}
      development/kv/data/sonarcloud: {}
      development/kv/data/sonarqube-us: {}
```

### 3. Test Workflows
1. Create a test branch
2. Open a PR to verify:
   - Build job runs successfully
   - Tests execute correctly
   - SonarQube analysis completes
3. Add `shadow_scan` label to PR to verify shadow scans
4. Close PR to verify cleanup workflow

### 4. Monitor Nightly Builds
- First nightly build will run at 03:00 UTC
- Verify all three scans complete (Next, SonarCloud EU, SonarCloud US)

## Migration Checklist

- [x] Created `mise.toml` with Go 1.25.1
- [x] Created main build workflow (`.github/workflows/build.yml`)
- [x] Created nightly workflow (`.github/workflows/nightly.yml`)
- [x] Created PR cleanup workflow (`.github/workflows/pr-cleanup.yml`)
- [x] Used exact action versions from migration guide
- [x] Pinned third-party actions to commit SHA
- [x] Used `sonar-xs` runner for private repository
- [x] Configured workflow-level concurrency control
- [x] Set up proper permissions (id-token, contents)
- [x] Implemented shadow scans for SonarCloud EU and US
- [x] Used SonarSource vault-action-wrapper for secrets
- [x] Used sonarsource/sonarqube-scan-action for analysis
- [x] Replicated all SonarQube analysis parameters
- [x] No Cirrus CI references found in documentation

## Next Steps

1. **Build Number Configuration**: Set build number in SPEED (> latest Cirrus CI build)
2. **Test First PR**: Create a test PR to validate the workflow
3. **Monitor First Nightly**: Verify nightly workflow executes correctly
4. **Remove Cirrus CI**: After successful validation, remove:
   - `.cirrus.star`
   - `.cirrus/` directory
   - Any Cirrus CI webhook configurations

## References

- [Migration Guide](https://github.com/SonarSource/ci-github-actions/blob/master/.cursor/cirrus-github-migration.md)
- [SonarSource CI GitHub Actions](https://github.com/SonarSource/ci-github-actions)
- [Mise Documentation](https://mise.jdx.dev/)

