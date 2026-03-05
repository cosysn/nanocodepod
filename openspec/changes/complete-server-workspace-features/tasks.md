## 1. Server API Extension

- [ ] 1.1 Extend workspace creation request struct to include git_url, git_branch, git_token, local_dir, devcontainer_path, inject_agent fields
- [ ] 1.2 Update workspace create handler to parse new fields
- [ ] 1.3 Add helper methods for git clone, local dir, devcontainer, agent injection in workspace service

## 2. Git Clone Feature

- [ ] 2.1 Create git clone function in workspace service that clones repo to workspace directory
- [ ] 2.2 Add shallow clone option support
- [ ] 2.3 Add branch checkout support
- [ ] 2.4 Handle private repository with token authentication

## 3. Local Directory Feature

- [ ] 3.1 Create local directory copy function that copies files to workspace directory
- [ ] 3.2 Preserve file permissions during copy
- [ ] 3.3 Validate directory exists before copying

## 4. Devcontainer Feature

- [ ] 4.1 Add devcontainer.json parsing function
- [ ] 4.2 Implement Dockerfile-based image build
- [ ] 4.3 Implement direct image pull support
- [ ] 4.4 Return build logs on failure

## 5. Agent Injection Feature

- [ ] 5.1 Add agent binary detection on server
- [ ] 5.2 Implement docker exec to copy agent into container
- [ ] 5.3 Install openssh-client if not present in container
- [ ] 5.4 Start agent as background process in container

## 6. IDE Connect Feature

- [ ] 6.1 Update CLI workspace create command to accept --ide and --connect flags
- [ ] 6.2 Implement VS Code launch via vscode:// protocol
- [ ] 6.3 Implement JetBrains IDE launch via jetbrains:// protocol
- [ ] 6.4 Handle IDE not installed error case

## 7. Linux Provider Support

- [ ] 7.1 Create separate LinuxProvider struct in cli/internal/provider/
- [ ] 7.2 Implement LinuxProvider with Docker detection and local command execution
- [ ] 7.3 Add --provider flag to CLI for explicit provider selection
- [ ] 7.4 Update GetProvider() to support explicit provider selection
- [ ] 7.5 Add provider config file support for ~/.codepod/provider/<name>/config.yaml
