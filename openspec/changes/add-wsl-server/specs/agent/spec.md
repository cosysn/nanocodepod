## ADDED Requirements

### Requirement: Agent runs inside development container

The system SHALL run the agent inside development containers created by the server.

#### Scenario: Server creates development container
- **WHEN** server creates a development container
- **THEN** the server SHALL inject codepod-agent into the container

#### Scenario: Agent executes tasks
- **WHEN** server sends a task to the agent
- **THEN** the agent SHALL execute the task inside the container and return the result

#### Scenario: Agent communicates with server via gRPC
- **WHEN** agent needs to report status or receive commands
- **THEN** the agent SHALL communicate with the server via gRPC

### Requirement: Agent functionality

The agent SHALL provide the following capabilities inside the container.

#### Scenario: Run commands
- **WHEN** server sends a command to agent
- **THEN** the agent SHALL execute the command inside the container and return output

#### Scenario: Stream output
- **WHEN** command produces long output
- **THEN** the agent SHALL stream output back to the server in real-time

#### Scenario: Health reporting
- **WHEN** server queries agent status
- **THEN** the agent SHALL report its status (running, idle, processing)

#### Scenario: SSH access (existing)
- **WHEN** user needs terminal access to container
- **THEN** the agent SHALL provide SSH access to the container
