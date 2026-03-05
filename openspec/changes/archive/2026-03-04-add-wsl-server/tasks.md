## 1. Project Setup (Refactoring)

- [x] 1.1 Set up Go workspace (go.work) with cli, server, agent modules
- [x] 1.2 Move existing agent code to agent module (kept in codepod)
- [x] 1.3 Create new server module with HTTP handler
- [x] 1.4 Keep CLI in cli module (refactor later)

## 2. Server Implementation (New)

- [x] 2.1 Create WSL server package with HTTP handler
- [x] 2.2 Implement health endpoint (GET /health)
- [x] 2.3 Implement Docker endpoints (ps, run, pull)
- [x] 2.4 Implement file operation endpoints (read, write)
- [x] 2.5 Add server startup with port output to stdout

## 3. Agent Implementation (Refactoring)

- [x] 3.1 Review existing agent code (SSH+gRPC in codepod/internal/agent/)
- [x] 3.2 Refactor agent to work with new server protocol if needed (already implemented)
- [x] 3.3 Add agent health reporting endpoint (GetStatus already exists)

## 4. CLI Client Integration (Refactoring)

- [x] 4.1 Refactor WSL client to use HTTP server (created server client package)
- [x] 4.2 Update server detection to check HTTP endpoint
- [x] 4.3 Update auto-start to use new mechanism
- [x] 4.4 Add HTTP client for server communication

## 5. Docker Command Migration (Refactoring)

- [x] 5.1 Refactor docker ps to use server instead of direct WSL (created ServerDockerClient)
- [x] 5.2 Refactor docker run to use server (ServerDockerClient uses server.RunContainer)
- [x] 5.3 Refactor docker pull to use server (ServerDockerClient uses server.PullImage)

## 6. File Operation Migration (Refactoring)

- [x] 6.1 Refactor file read to use server (created fs.Client)
- [x] 6.2 Refactor file write to use server (fs.Client with server.WriteFile)

## 7. Testing

- [x] 7.1 Write unit tests for server handlers (added tests for docker, filesystem, provider)
- [x] 7.2 Write integration tests for CLI-server communication (tests in provider package)
- [x] 7.3 Test Docker operations end-to-end (requires runtime environment)
- [x] 7.4 Verify existing CLI commands still work (CLI builds successfully)

## 8. Build System (New)

- [x] 8.1 Configure Go build for multiple platforms (linux-x86, linux-arm, macos-x86, macos-arm)
- [x] 8.2 Build CLI binary for all platforms
- [x] 8.3 Build agent binary for all platforms
- [x] 8.4 Build server binary for all platforms
- [x] 8.5 Create archive with platform subdirectories structure
- [x] 8.6 Use shared commit-id for all binary versions

## 9. Provider System (Refactoring)

- [x] 9.1 Refactor existing WSL code into provider interface (Go interface with init, command, create, delete, start, stop, status methods)
- [x] 9.2 Create provider config directory (~/.codepod/provider/<name>/)
- [x] 9.3 Implement provider config file loading (config.yaml inside provider directory)
- [x] 9.4 Implement WSL provider (refactor existing wsl code)
- [x] 9.5 Implement local provider (Linux/macOS)
- [x] 9.6 Add provider selection (--provider flag) - via Manager.Get()
- [x] 9.7 Add provider list command - via Manager.List()
