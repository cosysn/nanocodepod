## ADDED Requirements

### Requirement: Server runs in WSL

The system SHALL run a server component inside WSL that manages Docker containers.

#### Scenario: Start server in WSL
- **WHEN** user starts codepod server
- **THEN** the server starts inside WSL and listens for client connections

### Requirement: Server manages containers

The server SHALL handle all Docker container operations.

#### Scenario: Server creates workspace container
- **WHEN** client requests workspace creation
- **THEN** server creates Docker container with agent binary copied inside

#### Scenario: Server starts workspace
- **WHEN** client requests workspace start
- **THEN** server starts the container and returns connection info

#### Scenario: Server stops workspace
- **WHEN** client requests workspace stop
- **THEN** server stops the container

### Requirement: Agent copied into container

The server SHALL copy the agent binary into each container during creation.

#### Scenario: Copy agent to container
- **WHEN** server creates a new container
- **THEN** the agent binary is copied into the container at configured path

### Requirement: gRPC communication

The client and server SHALL communicate via gRPC.

#### Scenario: Client connects to server
- **WHEN** client starts and connects to server
- **THEN** gRPC connection is established for command execution
