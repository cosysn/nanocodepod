## Why

Currently, the project lacks automated CI/CD workflows. Every code change requires manual testing and release processes, which is error-prone and time-consuming. Adding GitHub Actions workflows will automate testing, linting, security scanning, and releases.

## What Changes

- Add GitHub Actions CI workflow for automated testing on PR/push
- Add GitHub Actions release workflow for semantic versioning releases
- CI workflow includes: unit tests, linting, security checks, code coverage validation
- CI requires code coverage > 80% to pass
- Release workflow uses semantic versioning and creates GitHub releases

## Capabilities

### New Capabilities

- `ci-workflow`: GitHub Actions CI workflow with tests, lint, security, coverage
- `release-workflow`: GitHub Actions release workflow with semantic versioning

### Modified Capabilities

- None - this is a new capability

## Impact

- **GitHub Actions**: New workflows in `.github/workflows/`
- **Development**: All PRs and commits will be automatically tested
- **Releases**: Semantic versioning releases with automated changelog
