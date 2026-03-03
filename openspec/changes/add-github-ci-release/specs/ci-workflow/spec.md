## ADDED Requirements

### Requirement: CI workflow runs on PR and push
The CI workflow SHALL trigger on every pull request and push to main branch.

#### Scenario: PR triggers CI
- **WHEN** a pull request is created or updated
- **THEN** CI workflow SHALL run all tests and checks

#### Scenario: Push to main triggers CI
- **WHEN** code is pushed to main branch
- **THEN** CI workflow SHALL run all tests and checks

### Requirement: Unit tests execution
The CI workflow SHALL run all unit tests with code coverage.

#### Scenario: Tests pass
- **WHEN** all unit tests pass
- **THEN** CI SHALL report success

#### Scenario: Tests fail
- **WHEN** any unit test fails
- **THEN** CI SHALL report failure

### Requirement: Code linting
The CI workflow SHALL run golangci-lint to check code quality.

#### Scenario: Lint passes
- **WHEN** golangci-lint reports no errors
- **THEN** CI SHALL continue

#### Scenario: Lint fails
- **WHEN** golangci-lint reports errors
- **THEN** CI SHALL fail

### Requirement: Security scanning
The CI workflow SHALL run govulncheck to detect vulnerabilities.

#### Scenario: No vulnerabilities
- **WHEN** govulncheck finds no vulnerabilities
- **THEN** CI SHALL continue

#### Scenario: Vulnerabilities found
- **WHEN** govulncheck finds vulnerabilities
- **THEN** CI SHALL report warnings (not fail)

### Requirement: Code coverage threshold
The CI workflow SHALL enforce code coverage greater than 80%.

#### Scenario: Coverage above threshold
- **WHEN** code coverage is greater than 80%
- **THEN** CI SHALL pass

#### Scenario: Coverage below threshold
- **WHEN** code coverage is 80% or less
- **THEN** CI SHALL fail
