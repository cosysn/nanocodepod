## ADDED Requirements

### Requirement: Detect Docker availability in WSL

The system SHALL detect when Docker is installed inside a WSL distribution and not directly accessible from Windows.

#### Scenario: Docker only in WSL
- **WHEN** codepod runs on Windows and the `docker` command fails on Windows
- **THEN** the system SHALL check if Docker is available inside WSL by executing `docker info` via WSL

#### Scenario: Docker available on Windows
- **WHEN** codepod runs on Windows and the `docker` command succeeds on Windows
- **THEN** the system SHALL use the native Windows Docker directly

### Requirement: Execute Docker commands via WSL

The system SHALL execute Docker CLI commands inside the WSL distribution when Docker is only available in WSL.

#### Scenario: Run docker ps via WSL
- **WHEN** Docker is detected as only available in WSL and the user lists containers
- **THEN** the system SHALL execute `wsl -d <distro> docker ps` and return the container list

#### Scenario: Run docker run via WSL
- **WHEN** Docker is detected as only available in WSL and the user creates a container
- **THEN** the system SHALL execute `wsl -d <distro docker run ...` with all appropriate arguments

#### Scenario: Run docker pull via WSL
- **WHEN** Docker is detected as only available in WSL and the user pulls an image
- **THEN** the system SHALL execute `wsl -d <distro> docker pull <image>` and stream output to stdout/stderr

### Requirement: Handle volume paths correctly

The system SHALL handle path translation for volume mounts when running Docker via WSL.

#### Scenario: Volume mount with Windows path
- **WHEN** Docker runs via WSL and a Windows path is provided for volume binding
- **THEN** the system SHALL convert the Windows path to a WSL-compatible path (e.g., `C:\path` to `/mnt/c/path`)

### Requirement: Fallback chain works correctly

The system SHALL try native Docker first, then fall back to WSL-based Docker, with clear error messages.

#### Scenario: Both Windows and WSL Docker available
- **WHEN** codepod runs on Windows and both Windows-native Docker and WSL Docker are available
- **THEN** the system SHALL prefer Windows-native Docker for better performance

#### Scenario: Neither Docker available
- **WHEN** codepod runs and Docker is not available in either Windows or WSL
- **THEN** the system SHALL return a clear error indicating Docker is not found
