## ADDED Requirements

### Requirement: Multi-platform build produces single archive

The system SHALL build CLI, agent, and server binaries for all supported platforms (Linux x86, Linux ARM, macOS x86, macOS ARM) in a single build process.

#### Scenario: Build produces archive
- **WHEN** build is triggered with a commit-id
- **THEN** the system SHALL produce an archive containing CLI, agent, and server binaries for all 4 platform combinations

#### Scenario: Archive structure
- **WHEN** archive is extracted
- **THEN** the extracted directory SHALL contain: linux-x86/, linux-arm/, macos-x86/, macos-arm/ subdirectories, each with platform-specific binaries

### Requirement: Build uses shared commit-id

The system SHALL use a single commit-id for CLI, agent, and server builds.

#### Scenario: Shared version
- **WHEN** build is triggered
- **THEN** CLI, agent, and server SHALL all be built from the same commit and share the same version identifier

#### Scenario: Commit-id directory naming
- **WHEN** archive is deployed
- **THEN** the archive SHALL be extracted to ~/.codepod-server/bin/<commit-id>/

### Requirement: Binary components

The archive SHALL contain three binary components.

#### Scenario: CLI binary
- **WHEN** archive is extracted
- **THEN** there SHALL be a `codepod` CLI binary for each platform

#### Scenario: Agent binary
- **WHEN** archive is extracted
- **THEN** there SHALL be a `codepod-agent` binary for each platform

#### Scenario: Server binary
- **WHEN** archive is extracted
- **THEN** there SHALL be a `codepod-server` binary for each platform
