## ADDED Requirements

### Requirement: CLI connects to server via HTTP

The system SHALL use HTTP for communication between CLI and WSL server.

#### Scenario: CLI discovers server port
- **WHEN** CLI needs to connect to WSL server
- **THEN** the CLI SHALL read the port from server stdout or use a default port

#### Scenario: CLI sends request to server
- **WHEN** CLI needs to perform a WSL operation
- **THEN** the CLI SHALL send an HTTP request to the server with operation details

### Requirement: Protocol supports Docker operations

The system SHALL define API endpoints for Docker operations.

#### Scenario: Docker ps request
- **WHEN** CLI sends GET /docker/ps to the server
- **THEN** the server SHALL return JSON array of containers

#### Scenario: Docker run request
- **WHEN** CLI sends POST /docker/run with container config
- **THEN** the server SHALL execute docker run and stream output

#### Scenario: Docker pull request
- **WHEN** CLI sends POST /docker/pull with image name
- **THEN** the server SHALL execute docker pull and stream progress

### Requirement: Protocol supports file operations

The system SHALL define API endpoints for file system operations.

#### Scenario: File read request
- **WHEN** CLI sends GET /fs/read?path=/some/path
- **THEN** the server SHALL return file contents

#### Scenario: File write request
- **WHEN** CLI sends POST /fs/write with path and content
- **THEN** the server SHALL write the file and return success/failure

### Requirement: Protocol includes health endpoint

The system SHALL provide a simple health check endpoint.

#### Scenario: Health check
- **WHEN** CLI sends GET /health to the server
- **THEN** the server SHALL return HTTP 200 with status OK
