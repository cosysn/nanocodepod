## 1. Project Structure (Go Workspace)

- [ ] 1.1 Create Go workspace structure:
  - `codepod/` (root workspace)
  - `cli/` module: Windows CLI client
  - `server/` module: Server runs in WSL
  - `agent/` module: Agent runs in containers
- [ ] 1.2 Create release build script
- [ ] 1.3 Build all components (x86, arm64)
- [ ] 1.4 Package into compressed archive

## 2. Server Implementation (runs in WSL)

- [ ] 2.1 Create server main function
- [ ] 2.2 Implement workspace gRPC service (create, start, stop, delete)
- [ ] 2.3 Server copies agent into containers

## 3. gRPC Service Definition

- [ ] 3.1 Define protobuf for workspace operations
- [ ] 3.2 Generate gRPC code

## 4. Provider Interface & Management

- [ ] 4.1 Define Provider interface (Init, Start, Stop, Status, Command)
- [ ] 4.2 Add "provider" concept to config
- [ ] 4.3 Create provider config file: ~/.codepod/providers/<name>/config.yaml
- [ ] 4.4 Provider config options: data_dir, wsl_distribution, server_port
- [ ] 4.5 Implement WSLProvider
- [ ] 4.6 WSL provider: Start/Stop WSL, Status check, Command execution
- [ ] 4.7 WSL provider: shell script injection via WSL pipe
- [ ] 4.8 CLI: codepod provider add wsl --name=xxx
- [ ] 4.9 CLI: codepod provider list
- [ ] 4.10 CLI: codepod use provider <name>

## 5. Deployment

- [ ] 5.1 CLI "codepod deploy" command
- [ ] 5.2 Detect WSL architecture, extract correct binary
- [ ] 5.3 Deploy to ~/.codepod-server/bin/<commitid>/
- [ ] 5.4 Start server in WSL

## 6. Client Integration

- [ ] 6.1 CLI uses gRPC to talk to server (no more WSL interaction!)
- [ ] 6.2 Server address from selected provider

## 7. Cleanup Old WSL Code

- [ ] 7.1 Remove complex wsl$ path handling from CLI
- [ ] 7.2 Remove filepath conversion code
