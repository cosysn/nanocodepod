## ADDED Requirements

### Requirement: Server starts on-demand

The system SHALL start the WSL server when the first CLI request arrives and keep it running for subsequent requests.

#### Scenario: Server starts on first request
- **WHEN** CLI makes a request to WSL server and no server is running
- **THEN** the system SHALL start the server in the background and respond to the request

#### Scenario: Server handles multiple requests
- **WHEN** CLI makes multiple requests to the server
- **THEN** the system SHALL reuse the same server instance for all requests

### Requirement: Server exposes Docker operations

The system SHALL provide Docker operations through the WSL server.

#### Scenario: List containers via server
- **WHEN** CLI requests container list from the server
- **THEN** the server SHALL execute `docker ps` inside WSL and return the output

#### Scenario: Run container via server
- **WHEN** CLI requests to run a container through the server
- **THEN** the server SHALL execute `docker run` with the provided arguments and stream output

#### Scenario: Pull image via server
- **WHEN** CLI requests to pull an image through the server
- **THEN** the server SHALL execute `docker pull` and stream progress to CLI

### Requirement: Server handles file operations

The system SHALL provide file system operations through the WSL server.

#### Scenario: Read file via server
- **WHEN** CLI requests to read a file path
- **THEN** the server SHALL read the file inside WSL and return contents

#### Scenario: Write file via server
- **WHEN** CLI requests to write content to a path
- **THEN** the server SHALL write the content to the file inside WSL

### Requirement: Server provides health check

The system SHALL provide a health endpoint to verify server is running.

#### Scenario: Health check request
- **WHEN** CLI sends a health check request to the server
- **THEN** the server SHALL return a successful response indicating it's alive
