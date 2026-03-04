## ADDED Requirements

### Requirement: Semantic versioning
The release workflow SHALL use semantic versioning (SemVer) for releases.

#### Scenario: Create release from tag
- **WHEN** a version tag (vX.Y.Z) is pushed
- **THEN** release workflow SHALL create a GitHub release

### Requirement: Binary builds
The release workflow SHALL build binaries for multiple platforms.

#### Scenario: Build CLI binary
- **WHEN** release is triggered
- **THEN** workflow SHALL build codepod CLI binary for linux/amd64 and linux/arm64

#### Scenario: Build agent binary
- **WHEN** release is triggered
- **THEN** workflow SHALL build codepod-agent binary for linux/amd64 and linux/arm64

### Requirement: Release artifacts
The release workflow SHALL upload binaries as release assets.

#### Scenario: Upload assets
- **WHEN** binaries are built
- **THEN** binaries SHALL be uploaded to GitHub release

### Requirement: Changelog generation
The release workflow SHALL generate changelog from commit messages.

#### Scenario: Generate changelog
- **WHEN** release is created
- **THEN** changelog SHALL be included in release notes
