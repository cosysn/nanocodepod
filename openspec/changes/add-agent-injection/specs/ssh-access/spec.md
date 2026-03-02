## ADDED Requirements

### Requirement: SSH connection information displayed
After a workspace starts with agent enabled, the system SHALL display SSH connection information to the user.

#### Scenario: Connection info displayed
- **WHEN** workspace starts successfully with agent enabled
- **THEN** the system SHALL display connection details including host, port, and username

### Requirement: SSH access via password authentication
The system SHALL allow SSH access to the container using password authentication with the configured password.

#### Scenario: SSH connection established
- **WHEN** user connects via SSH using the displayed credentials
- **THEN** the user SHALL have shell access inside the container

### Requirement: SSH port allocation
The system SHALL allocate a unique port for each workspace's agent (shared by SSH and gRPC) to avoid port conflicts.

#### Scenario: Unique port per workspace
- **WHEN** multiple workspaces are started with agents
- **THEN** each workspace SHALL have a unique port for both SSH and gRPC

#### Scenario: SSH and gRPC on same port
- **WHEN** agent is configured with port 22001
- **THEN** SSH and gRPC SHALL both be accessible on port 22001

### Requirement: Agent status in workspace metadata
The system SHALL store and expose agent status in workspace metadata.

#### Scenario: Agent status tracked
- **WHEN** agent is injected and started
- **THEN** workspace metadata SHALL reflect agent status as "running" and include the port

### Requirement: Agent status in list command
The system SHALL display agent status in the workspace list output.

#### Scenario: List shows agent status
- **WHEN** user runs `codepod list`
- **THEN** each workspace SHALL show its agent status (running/stopped)
