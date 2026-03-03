## ADDED Requirements

### Requirement: Single port for SSH and gRPC
The agent SHALL accept both SSH and gRPC connections on a single TCP port using protocol detection.

#### Scenario: SSH connection on shared port
- **WHEN** a client connects to the agent port and sends SSH protocol bytes first
- **THEN** the agent SHALL handle the connection as SSH and provide shell access

#### Scenario: gRPC connection on shared port
- **WHEN** a client connects to the agent port and sends HTTP/2 PRI message (gRPC magic bytes)
- **THEN** the agent SHALL handle the connection as gRPC and provide gRPC service

### Requirement: Protocol detection
The agent SHALL detect the protocol by peeking at the first bytes of the connection:
- If first bytes match HTTP/2 PRI magic bytes (`PRI * HTTP/2.0\r\n\r\nSM\r\n\r\n`), treat as gRPC
- Otherwise, treat as SSH

#### Scenario: gRPC magic bytes detection
- **WHEN** connection starts with HTTP/2 PRI bytes
- **THEN** agent routes to gRPC handler

#### Scenario: SSH default detection
- **WHEN** connection starts with non-gRPC bytes
- **THEN** agent routes to SSH handler

### Requirement: Port allocation
The workspace manager SHALL allocate only ONE port per workspace for the agent (both SSH and gRPC).

#### Scenario: Single port allocation
- **WHEN** a new workspace is created
- **THEN** the port pool SHALL allocate a single port for the agent

### Requirement: Docker port mapping
The Docker container SHALL map only container port 22 to the host port.

#### Scenario: Single port mapping
- **WHEN** container port bindings are configured
- **THEN** only port 22/tcp is mapped to host port (not 23/tcp)
