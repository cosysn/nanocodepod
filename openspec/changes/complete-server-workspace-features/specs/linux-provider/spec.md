## ADDED Requirements

### Requirement: Explicit provider selection via --provider flag
The CLI SHALL support explicit provider selection via a --provider flag.

#### Scenario: Select WSL provider
- **WHEN** user runs CLI with --provider=wsl
- **THEN** CLI uses WSLProvider for all operations

#### Scenario: Select Linux provider
- **WHEN** user runs CLI with --provider=linux
- **THEN** CLI uses LinuxProvider for native Linux environments

#### Scenario: Select local provider
- **WHEN** user runs CLI with --provider=local
- **THEN** CLI uses LocalProvider for local macOS/Windows

#### Scenario: Auto-detect provider when not specified
- **WHEN** user does not specify --provider
- **THEN** CLI auto-detects environment and uses appropriate provider

### Requirement: Separate Linux provider implementation
The system SHALL implement a dedicated LinuxProvider for native Linux environments.

#### Scenario: Linux provider detects Docker
- **WHEN** Linux provider checks environment
- **THEN** provider verifies Docker is available and running

#### Scenario: Linux provider discovers server
- **WHEN** Linux provider discovers server
- **THEN** provider reads /tmp/codepod-server-port and checks health endpoint

#### Scenario: Linux provider executes commands locally
- **WHEN** Linux provider executes command
- **THEN** provider runs command via bash directly (not through WSL)

### Requirement: Provider configuration support
The system SHALL support provider selection via configuration file.

#### Scenario: Provider config file
- **WHEN** user creates ~/.codepod/provider/<name>/config.yaml
- **THEN** CLI can use that provider with --provider=<name>
