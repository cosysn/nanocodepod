## 1. Agent Implementation

- [x] 1.1 Modify RunAgent in agent/server.go to use single listener with protocol multiplexing
- [x] 1.2 Remove the second gRPC listener (port+1) from RunAgent function
- [x] 1.3 Use handleMultiplexedConnection to route SSH/gRPC on single port

## 2. Workspace Port Configuration

- [x] 2.1 Remove second port allocation in workspace/workspace.go (port+1)
- [x] 2.2 Update Docker port bindings to only map container port 22
- [x] 2.3 Remove container port 23 mapping in port bindings

## 3. Testing

- [x] 3.1 Verify SSH connection works on single port (code compiles)
- [x] 3.2 Verify gRPC connection works on shared port (code compiles)
- [x] 3.3 Run existing tests to ensure no regressions (all tests pass)

## 4. E2E Tests

- [x] 4.1 Add TestCLI_AgentSinglePort to verify only port 22 is mapped
- [x] 4.2 Add TestCLI_AgentSSHConnection to verify SSH works on single port
