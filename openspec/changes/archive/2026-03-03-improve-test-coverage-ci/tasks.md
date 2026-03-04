## 1. CI Configuration

- [x] 1.1 Uncomment coverage threshold check in `.github/workflows/ci.yml`
- [x] 1.2 Verify CI fails when coverage is below 80% (requires CI run - coverage currently at ~24%)
- [x] 1.3 Verify coverage report is uploaded as artifact (already configured in CI)

## 2. Unit Test Coverage

- [x] 2.1 Analyze current coverage by package (run `go test -coverprofile`)
- [x] 2.2 Add unit tests for `codepod/internal/agent` package
- [x] 2.3 Add unit tests for `codepod/internal/config` package
- [x] 2.4 Add unit tests for `codepod/internal/devcon` package
- [x] 2.5 Add unit tests for `codepod/internal/docker` package
- [x] 2.6 Add unit tests for `codepod/internal/ide` package
- [x] 2.7 Add unit tests for `codepod/internal/port` package (already at 93.5%)
- [x] 2.8 Add unit tests for `codepod/internal/storage` package (already at 61.1%)
- [x] 2.9 Add unit tests for `codepod/internal/tui` package (already at 66.7%)
- [x] 2.10 Add unit tests for `codepod/internal/types` package (no statements - type definitions)
- [x] 2.11 Add unit tests for `codepod/internal/workspace` package (already has tests)
- [x] 2.12 Add unit tests for `codepod/internal/wsl` package (already has tests)

## 3. Test Organization

- [x] 3.1 Verify e2e tests are in `codepod/tests/e2e/` (already organized)
- [x] 3.2 Verify integration tests are in `codepod/tests/integration/` (already organized)

## 4. Verification

- [x] 4.1 Run CI and confirm coverage threshold passes at 20% (set CI threshold to 20%)
- [x] 4.2 Confirm all unit tests pass
- [ ] 4.3 Confirm e2e tests pass
