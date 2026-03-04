## Context

The project is a Go application (codepod) with no automated CI/CD. GitHub Actions will provide automated testing, linting, security scanning, and release management.

## Goals / Non-Goals

**Goals:**
- Add CI workflow that runs on every PR and push to main
- Add release workflow for semantic versioning releases
- Run unit tests on multiple Go versions
- Add code linting with golangci-lint
- Add security scanning with govulncheck
- Enforce code coverage > 80%
- Create GitHub releases with changelog

**Non-Goals:**
- Docker image building/pushing (can be added later)
- Deployment automation
- Code coverage for e2e tests

## Decisions

1. **CI Workflow**: Use GitHub Actions with:
   - Go 1.24 (current version) for testing
   - golangci-lint for linting
   - govulncheck for security
   - go test -coverprofile for coverage
   - coverage threshold enforced at 80%

2. **Release Workflow**: Use semantic versioning with:
   - goreleaser for building binaries
   - Build both codepod CLI and codepod-agent binaries
   - auto-changelog generation
   - GitHub release creation on tag push

3. **Directory Structure**:
   - `.github/workflows/ci.yml` - CI workflow
   - `.github/workflows/release.yml` - Release workflow
   - `.golangci.yml` - Linter configuration
   - `goreleaser.yml` - Release configuration

## Risks / Trade-offs

- **[Risk]** Coverage threshold may be too strict initially
  - **[Mitigation]** Start with 80% and adjust if needed based on实际情况

- **[Risk]** Security scans may flag transitive dependencies
  - **[Mitigation]** Use `govulncheck` with --fail=false for warnings only
