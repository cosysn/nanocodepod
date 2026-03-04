## Context

The project currently has low test coverage (~15%) and lacks proper CI gate enforcement. The CI workflow has coverage checking commented out. There is an existing `tests/` directory with `e2e` and `integration` subdirectories, but tests are not organized consistently. Unit tests exist in `codepod/internal/*` packages.

## Goals / Non-Goals

**Goals:**
- Increase unit test coverage to >=80% across all packages in `codepod/internal/`
- Organize e2e tests in a dedicated `test/` directory
- Configure CI to enforce 80% coverage threshold as a required gate

**Non-Goals:**
- Add integration tests (already exists in `tests/integration/`)
- Refactor existing working code just for coverage
- Add tests for generated code or protobuf files

## Decisions

1. **Coverage tool**: Use `go test -coverprofile` with `go tool cover` (already in use)
   - Already configured in CI
   - No additional dependencies needed

2. **Test organization**: Keep unit tests alongside source code in `*_test.go` files
   - Go standard practice
   - E2e tests move to `codepod/tests/e2e/` (already exists)

3. **Coverage threshold**: Set 80% as the required threshold
   - Industry standard for good coverage
   - Comment in CI shows this was the original intent

4. **CI gate behavior**: Fail the build if coverage drops below 80%
   - Uncomment the existing threshold check in CI workflow
   - Use simple bc comparison for reliability

## Risks / Trade-offs

- **Risk**: Some packages may be difficult to test (e.g., WSL, Docker integration)
  - **Mitigation**: Use interfaces and dependency injection for testable code; may need to exclude certain packages from coverage

- **Risk**: Increasing coverage may lead to shallow tests just for numbers
  - **Mitigation**: Focus on testing core business logic and critical paths

- **Risk**: Breaking existing CI if coverage drops during transition
  - **Mitigation**: Add coverage incrementally; only enable gate after reaching threshold

## Migration Plan

1. First, enable the coverage check in CI (uncomment threshold)
2. Add unit tests incrementally to raise coverage
3. Verify e2e tests are in `codepod/tests/e2e/`
4. Run CI to confirm 80% threshold passes
5. The change is complete when CI passes with 80% threshold enforced
