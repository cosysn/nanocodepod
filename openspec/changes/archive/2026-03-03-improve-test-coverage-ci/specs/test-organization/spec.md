## ADDED Requirements

### Requirement: Unit tests shall be co-located with source code
Unit tests MUST be placed in the same package as the code they test, using Go's `*_test.go` naming convention.

#### Scenario: Unit test location
- **WHEN** a developer adds unit tests for a package in `codepod/internal/<pkg>/`
- **THEN** the test file MUST be named `<pkg>_test.go` and placed in the same directory

### Requirement: E2E tests shall be in the test directory
End-to-end tests MUST be placed in `codepod/tests/e2e/` directory.

#### Scenario: E2E test location
- **WHEN** a developer adds e2e tests
- **THEN** the test files MUST be placed in `codepod/tests/e2e/`

### Requirement: Integration tests shall be in the test directory
Integration tests MUST be placed in `codepod/tests/integration/` directory.

#### Scenario: Integration test location
- **WHEN** a developer adds integration tests
- **THEN** the test files MUST be placed in `codepod/tests/integration/`
