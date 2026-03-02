## 1. CLI Flag Implementation

- [x] 1.1 Add `--agent` and `--no-agent` flags to `codepod start` command
- [x] 1.2 Add `--agent` and `--no-agent` flags to `codepod up` command

## 2. Environment Variable Support

- [x] 2.1 Read `CODEPOD_AGENT_PORT` environment variable to configure agent port
- [x] 2.2 Read `CODEPOD_AGENT_PASSWORD` environment variable to configure agent password
- [x] 2.3 Pass environment variables to agent process in container via docker exec

## 3. Workspace Manager Updates

- [x] 3.1 Modify `Start` method in workspace package to accept agent injection option
- [x] 3.2 Add `InjectAgent` field to `CreateOptions` struct
- [x] 3.3 Update agent injection logic to respect the flag
- [x] 3.4 Change container entrypoint from `sleep infinity` to agent binary

## 4. Agent as PID 0 (Init Process)

- [x] 4.1 Modify agent to run as container entrypoint (PID 0)
- [x] 4.2 Agent starts SSH server as main function
- [x] 4.3 Implement child process fork/exec for SSH commands
- [x] 4.4 Implement zombie process reaping (wait for terminated children)

## 5. gRPC Service Implementation

- [x] 5.1 Define gRPC protocol buffers (ExecuteCommand, GetStatus)
- [x] 5.2 Implement gRPC server in agent
- [x] 5.3 Implement command dispatch via gRPC
- [x] 5.4 Implement status reporting via gRPC

## 6. SSH + gRPC Port Multiplexing

- [x] 6.1 Implement protocol detection (SSH vs gRPC) on same port
- [x] 6.2 Route connections to appropriate handler based on protocol
- [x] 6.3 Handle HTTP/2 magic bytes for gRPC detection

## 7. SSH Connection Display

- [x] 7.1 Display SSH/gRPC connection info (host, port, username, password) after workspace starts with agent enabled
- [x] 7.2 Show agent status in the output

## 8. List Command Enhancement

- [x] 8.1 Update `codepod list` to show agent status for each workspace

## 9. Testing

- [ ] 9.1 Test agent injection with `--agent` flag
- [ ] 9.2 Test agent injection disabled with `--no-agent` flag
- [ ] 9.3 Test that list command shows agent status
- [ ] 9.4 Test environment variable configuration (CODEPOD_AGENT_PORT, CODEPOD_AGENT_PASSWORD)
- [ ] 9.5 Test agent runs as PID 0 in container
- [ ] 9.6 Test agent forks child processes for SSH commands
- [ ] 9.7 Test gRPC command dispatch
- [ ] 9.8 Test SSH and gRPC on same port
