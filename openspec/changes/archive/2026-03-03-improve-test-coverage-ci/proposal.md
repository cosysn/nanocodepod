## Why

Currently, the project lacks sufficient test coverage which makes it difficult to ensure code quality and reliability. Additionally, e2e tests are not organized in a standard location, making maintenance harder. Setting a clear CI gate threshold will help maintain consistent quality standards.

## What Changes

1. **Increase unit test coverage to over 80%** - Add or enhance unit tests to achieve >=80% code coverage
2. **Extract e2e tests to test directory** - Move e2e tests from current location to a dedicated `test/` directory
3. **Modify CI gate threshold to 80%** - Update CI configuration to enforce >=80% test coverage as a required check

## Capabilities

### New Capabilities
- `test-coverage-enforcement`: Defines the coverage threshold requirements and validation process
- `test-organization`: Defines the structure and organization of tests (unit, e2e)

### Modified Capabilities
- (none - this is a new capability setup)

## Impact

- **CI Configuration**: `.github/workflows/` - will be modified to add coverage gates
- **Test Files**: Unit tests will be added/enhanced, e2e tests will be reorganized
- **Dependencies**: May require adding test coverage tools (e.g., coverage reporters)
