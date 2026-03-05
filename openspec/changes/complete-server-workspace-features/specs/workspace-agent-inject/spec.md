## ADDED Requirements

### Requirement: Inject codepod-agent into workspace container
The server SHALL inject the codepod-agent binary into the running container when inject_agent is true.

#### Scenario: Inject agent successfully
- **WHEN** user creates workspace with inject_agent=true
- **THEN** server copies agent binary into container and starts it as background process

#### Scenario: Agent binary not found
- **WHEN** inject_agent=true but agent binary is not available on server
- **THEN** server returns error indicating agent binary not found

#### Scenario: Container does not support agent
- **WHEN** inject_agent=true but container lacks required dependencies (openssh-client)
- **THEN** server installs required dependencies before starting agent

#### Scenario: Agent starts on correct port
- **WHEN** agent is injected successfully
- **THEN** agent starts and listens on SSH port (22) with gRPC on port+1
