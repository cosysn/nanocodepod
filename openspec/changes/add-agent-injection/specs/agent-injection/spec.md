## ADDED Requirements

### Requirement: Agent injection enabled by default
When a workspace is started with the `codepod start` or `codepod up` command, the system SHALL automatically inject the codepod-agent binary into the development container unless explicitly disabled.

#### Scenario: Default agent injection
- **WHEN** user runs `codepod start myworkspace` without any agent flag
- **THEN** the system SHALL inject the agent into the workspace container

#### Scenario: Agent injection disabled
- **WHEN** user runs `codepod start myworkspace --no-agent`
- **THEN** the system SHALL NOT inject the agent into the workspace container

### Requirement: Agent binary deployment
The system SHALL copy the codepod-agent binary from the host to the container at `/usr/local/bin/codepod-agent`.

#### Scenario: Agent binary copied
- **WHEN** agent injection is triggered
- **THEN** the system SHALL copy the codepod-agent binary to the container

#### Scenario: Agent binary already exists
- **WHEN** agent injection is triggered and agent already exists in container
- **THEN** the system SHALL skip copying and reuse existing binary

### Requirement: SSH server dependencies
The system SHALL install required SSH server dependencies (openssh-server) in the container before starting the agent.

#### Scenario: Dependencies installed
- **WHEN** agent injection is triggered
- **THEN** the system SHALL install openssh-server in the container

### Requirement: Agent process management
The system SHALL start the agent process inside the container as a background process.

#### Scenario: Agent started successfully
- **WHEN** agent injection completes
- **THEN** the agent SSH server SHALL be running on the configured port inside the container

### Requirement: Environment variable configuration
The system SHALL pass environment variables from the host to the agent process in the container.

#### Scenario: Agent port configured via environment variable
- **WHEN** user sets CODEPOD_AGENT_PORT environment variable before running `codepod start`
- **THEN** the agent SHALL use the specified port inside the container

#### Scenario: Agent password configured via environment variable
- **WHEN** user sets CODEPOD_AGENT_PASSWORD environment variable before running `codepod start`
- **THEN** the agent SHALL use the specified password for SSH authentication

#### Scenario: Default environment values
- **WHEN** user runs `codepod start` without setting any agent environment variables
- **THEN** the agent SHALL use default port (22001) and default password ("codepod")

### Requirement: Agent as container init process (PID 0)
The system SHALL run agent as the container's entrypoint (PID 0) instead of `sleep infinity`.

#### Scenario: Agent runs as PID 0
- **WHEN** container starts
- **THEN** agent SHALL run as PID 0 (init process) in the container

#### Scenario: Agent runs SSH server and waits
- **WHEN** agent runs as PID 0
- **THEN** agent SHALL start SSH server and wait for connections

### Requirement: Agent forks child processes for user tasks
The system SHALL allow agent to fork child processes to handle user commands.

#### Scenario: Agent forks child process
- **WHEN** user connects via SSH and runs a command
- **THEN** agent SHALL fork a child process to execute the command

#### Scenario: Agent reaps zombie processes
- **WHEN** child process terminates
- **THEN** agent SHALL reap (wait for) the child process to prevent zombies

#### Scenario: Agent inherits environment from container
- **WHEN** agent runs as PID 0
- **THEN** the agent SHALL have access to container environment variables (CODEPOD_AGENT_PORT, CODEPOD_AGENT_PASSWORD)
