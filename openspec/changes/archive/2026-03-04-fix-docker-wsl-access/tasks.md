## 1. WSL Docker Detection

- [x] 1.1 Add function to check if Docker is available in WSL distribution (in wsl/platform.go)
- [x] 1.2 Add function to detect Docker availability on Windows (try native, fallback to WSL)
- [x] 1.3 Add WSL detection flag to platform type detection

## 2. WSL-Aware Docker Client Wrapper

- [x] 2.1 Create WSLDockerClient wrapper struct in docker/client.go
- [x] 2.2 Implement command execution via WSL for all DockerClient methods
- [x] 2.3 Add path translation for volume bindings (Windows path to WSL path)
- [x] 2.4 Implement fallback logic: try native Windows Docker first, then WSL

## 3. Docker Client Factory Integration

- [x] 3.1 Modify docker.New() to detect platform and Docker availability
- [x] 3.2 Return appropriate DockerClient implementation based on detection
- [x] 3.3 Add clear error messages when Docker is not found anywhere

## 4. Testing

- [x] 4.1 Add unit tests for WSL Docker detection
- [x] 4.2 Add unit tests for path translation
- [x] 4.3 Test fallback chain behavior
- [x] 4.4 Verify existing tests still pass
