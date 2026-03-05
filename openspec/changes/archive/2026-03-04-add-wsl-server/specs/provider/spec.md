## ADDED Requirements

### Requirement: Provider interface for extensibility

The system SHALL define a provider interface that can be implemented for different environments.

The interface includes the following methods:

#### Method: Init
- **Purpose:** Deploy environment - install server, agent, and setup container
- **WHEN** init is called
- **THEN** the provider SHALL deploy codepod-server and codepod-agent into the target environment and set up the container runtime

#### Method: Command
- **Purpose:** Execute commands in the target environment
- **WHEN** command is called with a command string or script
- **THEN** the provider SHALL execute the command/script in the target environment and return the output
- **NOTE:** For WSL, this pipes the script to WSL for execution (similar to VSCode)

#### Method: Create
- **Purpose:** Create the target environment
- **WHEN** create is called
- **THEN** the provider SHALL create the environment (e.g., create a new WSL distribution)

#### Method: Delete
- **Purpose:** Delete the target environment
- **WHEN** delete is called
- **THEN** the provider SHALL delete the environment and all its resources

#### Method: Start
- **Purpose:** Start the target environment
- **WHEN** start is called
- **THEN** the provider SHALL start the environment (e.g., start WSL distribution)

#### Method: Stop
- **Purpose:** Stop the target environment
- **WHEN** stop is called
- **THEN** the provider SHALL stop the environment to save resources (e.g., stop WSL distribution)

#### Method: Status
- **Purpose:** Check environment status
- **WHEN** status is called
- **THEN** the provider SHALL return the current status of the environment (running, stopped, error, etc.)

#### Scenario: WSL provider implements interface
- **WHEN** using WSL environment
- **THEN** the WSL provider SHALL implement the provider interface

#### Scenario: AWS provider implements interface
- **WHEN** using AWS environment
- **THEN** the AWS provider SHALL implement the provider interface

#### Scenario: Linux provider implements interface
- **WHEN** using Linux environment
- **THEN** the Linux provider SHALL implement the provider interface

#### Scenario: macOS provider implements interface
- **WHEN** using macOS environment
- **THEN** the macOS provider SHALL implement the provider interface

### Requirement: WSL deployment via command method

The system SHALL deploy to WSL by passing deployment script via the command method.

#### Scenario: Init uses command to deploy
- **WHEN** WSL provider init is called
- **THEN** the provider SHALL pass the deployment script to WSL via the command method

#### Scenario: Command pipes script to WSL
- **WHEN** command method receives a script
- **THEN** the WSL provider SHALL pipe the script to WSL for execution (similar to VSCode WSL injection)

#### Scenario: Deployment script downloads archive
- **WHEN** deployment script runs in WSL
- **THEN** the script SHALL download the archive from GitHub release to ~/.codepod-server/bin/<commit-id>/

#### Scenario: Deployment script installs binaries
- **WHEN** archive is downloaded
- **THEN** the script SHALL extract the archive and install CLI, server, and agent binaries

#### Scenario: Deployment script starts server
- **WHEN** binaries are installed
- **THEN** the script SHALL start the codepod-server in the background

### Requirement: Provider configuration file

The system SHALL support provider configuration files that define provider-specific options.

#### Scenario: WSL provider config
- **WHEN** using a WSL provider
- **THEN** the configuration SHALL include: WSL distribution name, data directory path

#### Scenario: Provider config location
- **WHEN** CLI looks for provider configuration
- **THEN** the system SHALL look in ~/.codepod/provider/<provider-name>/config.yaml

#### Scenario: Provider config format
- **WHEN** reading provider configuration
- **THEN** the configuration SHALL be in YAML format with provider-type specific fields

### Requirement: Multiple providers supported

The system SHALL support configuring multiple providers for different environments.

#### Scenario: List available providers
- **WHEN** user lists providers
- **THEN** the system SHALL show all configured providers

#### Scenario: Select provider
- **WHEN** user specifies which provider to use
- **THEN** the CLI SHALL use that provider's configuration for operations

### Requirement: Provider types

The system SHALL support different provider types.

#### Scenario: WSL provider config fields
- **WHEN** provider type is WSL
- **THEN** the configuration SHALL include:
  - `type`: "wsl"
  - `wsl_distro`: WSL distribution name (e.g., "Ubuntu-22.04")
  - `data_dir`: Data directory path inside WSL

#### Scenario: WSL provider config example
- **WHEN** creating a WSL provider config
- **THEN** the config.yaml SHALL look like:
```yaml
type: wsl
wsl_distro: Ubuntu-22.04
data_dir: /home/codepod
```

#### Scenario: Local provider config fields
- **WHEN** provider type is local (Linux/macOS)
- **THEN** the configuration SHALL include:
  - `type`: "local"
  - `data_dir`: Data directory path
  - `socket_path`: (optional) Unix socket path for server communication

#### Scenario: Local provider
- **WHEN** provider type is local (Linux/macOS)
- **THEN** the configuration SHALL include: data_dir, socket_path (optional)

### Requirement: Provider connection

The system SHALL use provider configuration to connect to the target environment.

#### Scenario: Connect to WSL provider
- **WHEN** CLI needs to connect to a WSL provider
- **THEN** the CLI SHALL use the provider's wsl_distro to execute commands inside that WSL

#### Scenario: Connect to local provider
- **WHEN** CLI needs to connect to a local provider
- **THEN** the CLI SHALL connect to the local server using the configured socket path
