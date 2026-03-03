## ADDED Requirements

### Requirement: Agent runs as container init process (PID 0)
The system SHALL run agent as the container's entrypoint (PID 0) instead of `sleep infinity`.

#### Scenario: Agent runs as PID 0
- **WHEN** container starts
- **THEN** agent SHALL run as PID 0 (init process) in the container

#### Scenario: Agent runs SSH server and waits
- **WHEN** agent runs as PID 0
- **THEN** agent SHALL start SSH server and wait for connections

### Requirement: Agent forks child processes for user tasks
The system SHALL allow agent to fork child processes to handle user commands via SSH or gRPC.

#### Scenario: Agent forks child process for SSH command
- **WHEN** user connects via SSH and runs a command
- **THEN** agent SHALL fork a child process to execute the command

#### Scenario: Agent forks child process for gRPC command
- **WHEN** gRPC client sends ExecuteCommand request
- **THEN** agent SHALL fork a child process to execute the command and return result via gRPC

#### Scenario: Agent reaps zombie processes
- **WHEN** child process terminates
- **THEN** agent SHALL reap (wait for) the child process to prevent zombies

### Requirement: Agent provides gRPC service
The system SHALL provide a gRPC service for command dispatch and status reporting.

#### Scenario: gRPC server starts
- **WHEN** agent starts
- **THEN** gRPC server SHALL start on the same port as SSH

#### Scenario: Command dispatch via gRPC
- **WHEN** gRPC client sends ExecuteCommand request
- **THEN** agent SHALL execute the command and return stdout/stderr/status

#### Scenario: Status reporting via gRPC
- **WHEN** gRPC client sends GetStatus request
- **THEN** agent SHALL return system status (process count, memory, etc.)

### Requirement: SSH and gRPC share same port
The system SHALL multiplex SSH and gRPC on the same port using protocol detection.

#### Scenario: SSH connection detected
- **WHEN** client connects and sends SSH protocol banner
- **THEN** agent SHALL handle as SSH connection

#### Scenario: gRPC connection detected
- **WHEN** client connects and sends HTTP/2 magic bytes
- **THEN** agent SHALL handle as gRPC connection

#### Scenario: Single port for both services
- **WHEN** agent is configured with port 22001
- **THEN** both SSH (port 22001) and gRPC (port 22001) SHALL be accessible
