## 1. Core Data Structures

- [x] 1.1 Create pkg/resolver package directory
- [x] 1.2 Define Authority struct with Provider, Identity, Path fields
- [x] 1.3 Define DevContainerConfig struct for Hex-JSON parsing

## 2. URI Splitting

- [x] 2.1 Implement SplitAuthorityAndPath function
- [x] 2.2 Add tests for SplitAuthorityAndPath with various inputs
- [x] 2.3 Handle edge cases (empty path, trailing slash)

## 3. Hex-JSON Codec

- [x] 3.1 Implement EncodeToHex function for JSON encoding
- [x] 3.2 Implement DecodeFromHex function for JSON decoding
- [x] 3.3 Add validation for odd-length and invalid Hex characters
- [x] 3.4 Add tests for encode/decode round-trip
- [x] 3.5 Test DevContainer configuration encoding

## 4. Hybrid URI Resolver

- [x] 4.1 Define provider type constants (wsl, ssh-remote, docker-container, dev-container)
- [x] 4.2 Define IdentityParser interface for strategy pattern
- [x] 4.3 Implement PlainIdentityParser for simple providers
- [x] 4.4 Implement HexJSONIdentityParser for dev-container
- [x] 4.5 Implement Resolve function with provider-to-strategy routing
- [x] 4.6 Add error handling for missing provider, missing identity

## 5. Routing Engine

- [x] 5.1 Create pkg/router package
- [x] 5.2 Define AgentType constants (local, workspace, container)
- [x] 5.3 Define RouteNode struct
- [x] 5.4 Define Router struct with parent, children, services
- [x] 5.5 Implement SetParent method
- [x] 5.6 Implement AddChild method
- [x] 5.7 Implement RemoveChild method
- [x] 5.8 Implement FindChild method
- [x] 5.9 Implement RegisterService method
- [x] 5.10 Add thread-safe with RWMutex

## 6. Agent Bootstrapper

- [x] 6.1 Create pkg/bootstrapper package
- [x] 6.2 Define BootstrapConfig struct (TargetType, Target, BinaryPath)
- [x] 6.3 Define Bootstrapper struct with Driver interface
- [x] 6.4 Implement Bootstrap method
- [x] 6.5 Support WSL bootstrap (copy + exec)
- [x] 6.6 Support Docker bootstrap (docker cp + exec)
- [x] 6.7 Support SSH bootstrap (scp + ssh)
- [x] 6.8 Support UDS bootstrap (spawn local process)
- [x] 6.9 Add config validation
- [x] 6.10 Add binary not found error handling

## 7. Provider System

- [x] 7.1 Define Provider interface (Type, Bootstrap, Connect)
- [x] 7.2 Create Local Agent provider map (wsl, ssh-remote, docker-container)
- [x] 7.3 Implement WSLProvider struct
- [x] 7.4 Implement SSHProvider struct
- [x] 7.5 Implement DockerProvider struct
- [x] 7.6 Provider returns RPCClient after bootstrap
- [x] 7.7 Provider handles connection errors
- [x] 7.8 Provider registers child in routing table
- [x] 7.9 Route to existing child if available

## 8. Channel Interface & Factory

- [x] 8.1 Create pkg/channel package
- [x] 8.2 Define Channel interface (Dial, Listen, Close)
- [x] 8.3 Define Conn interface (Read, Write, Close)
- [x] 8.4 Define Listener interface (Accept, Close, Addr)
- [x] 8.5 Implement NewChannel factory function
- [x] 8.6 Add channel type registry for extensibility
- [x] 8.7 Add error types (ErrConnectionRefused, ErrTimeout, etc.)

## 9. UDS Channel

- [x] 9.1 Implement UDSChannel struct
- [x] 9.2 Implement UDS Dial method
- [x] 9.3 Implement UDS Listen method
- [x] 9.4 Implement UDS Conn (Read, Write, Close)
- [x] 9.5 Handle socket file permissions

## 10. SSH Channel

- [x] 10.1 Implement SSHChannel struct with config
- [x] 10.2 Implement SSH Dial with password auth
- [x] 10.3 Implement SSH Dial with key auth
- [x] 10.4 Implement SSH exec command
- [x] 10.5 Implement SSH session/stream

## 11. WSL Channel

- [x] 11.1 Implement WSLChannel struct
- [x] 11.2 Implement WSL Dial via Interop
- [x] 11.3 Implement WSL exec command
- [x] 11.4 Handle WSL distribution lookup

## 12. StdIO Channel

- [x] 12.1 Implement StdIOChannel struct
- [x] 12.2 Wrap os/exec.Cmd as Channel
- [x] 12.3 Handle stdin/stdout/stderr streams

## 13. RPC Transport

- [x] 13.1 Create pkg/rpc package
- [x] 13.2 Define RPCServer interface
- [x] 13.3 Define RPCClient interface
- [x] 13.4 Implement JSON-RPC 2.0 codec
- [x] 13.5 Implement request/response handling
- [x] 13.6 Support notification (no response)
- [x] 13.7 Support batch requests

## 14. RPC + Yamux Integration

- [x] 14.1 Define MuxedConn interface
- [x] 14.2 Wrap Channel with Yamux
- [x] 14.3 Implement Stream() for creating new streams
- [x] 14.4 Support control stream (RPC)
- [x] 14.5 Support data streams (PTY, FS)

## 15. Bidirectional RPC

- [x] 15.1 Server can register callback handlers
- [x] 15.2 Server can call client methods
- [x] 15.3 Handle both directions simultaneously

## 16. Agent Core

- [x] 16.1 Create pkg/agent package
- [x] 16.2 Define Agent struct (Type, Router, Bootstrapper, Resolver, RPC)
- [x] 16.3 Define AgentOption functional options
- [x] 16.4 Implement NewAgent constructor
- [x] 16.5 Implement HandleRequest recursive routing
- [x] 16.6 Implement forward to child logic
- [x] 16.7 Implement bootstrap on missing child
- [x] 16.8 Implement service handler dispatch
- [x] 16.9 Add exit condition detection

## 17. Agent Type Variations

- [x] 17.1 Create NewLocalAgent factory
- [x] 17.2 Create NewWorkspaceAgent factory
- [x] 17.3 Create NewContainerAgent factory
- [x] 17.4 Configure appropriate router behavior per type

## 18. CLI Integration

- [x] 18.1 CLI uses Agent with Local type
- [x] 18.2 CLI parses entry URI and forwards
- [x] 18.3 Handle connection to Local Agent socket

## 19. CLI-Agent RPC Integration

- [x] 19.1 CLI auto-starts Local Agent if not running
- [x] 19.2 Implement UDS socket auto-detection
- [x] 19.3 Implement agent process spawn on missing ] 19.4 CLI connects via socket
- [ RPC to Local Agent
- [x] 19.5 Local Agent registers "Agent" service handlers
- [x] 19.6 Local Agent registers "Resolver" service handlers
- [x] 19.7 Local Agent registers "Router" service handlers
- [x] 19.8 CLI calls "Agent.Route" to forward remaining URI
- [ ] 19.9 Agent returns child RPC connection for direct communication
- [ ] 19.10 Handle RPC errors (connection lost, method not found)

## 20. Provider Routing Integration

- [x] 20.1 Local Agent has multiple providers (WSL, SSH, Docker)
- [x] 20.2 Provider looks up by authority type
- [x] 20.3 Provider bootstrap injects Workspace Agent
- [x] 20.4 Provider establishes connection after bootstrap
- [x] 20.5 Local Agent adds child to routing table
- [x] 20.6 Route to existing child if available
- [x] 20.7 Full flow: CLI → Local → Provider → Workspace Agent

## 21. Recursive Routing (Peeling Onion)

- [ ] 21.1 Local Agent parses dev-container Hex-JSON
- [ ] 21.2 Local Agent forwards to Workspace Agent
- [ ] 21.3 Workspace Agent parses dev-container Hex-JSON
- [ ] 21.4 Workspace Agent determines target container
- [ ] 21.5 Workspace Agent routes to existing Container Agent
- [ ] 21.6 Workspace Agent bootstraps new Container Agent if needed
- [ ] 21.7 Container Agent handles service requests
- [ ] 21.8 Local Agent handles WSL/SSH provider routing
- [ ] 21.9 Workspace Agent handles Docker/DevContainer provider
- [ ] 21.10 Exit condition: path is empty or service prefix

## 22. Integration Tests

- [x] 22.1 Test resolver with all provider types
- [x] 22.2 Test routing engine parent-child operations
- [x] 22.3 Test channel implementations (UDS, SSH, WSL, StdIO)
- [x] 22.4 Test RPC request/response
- [x] 22.5 Test Yamux stream multiplexing
- [x] 22.6 Test agent recursive routing flow
- [ ] 22.7 Test CLI auto-start Local Agent
- [ ] 22.8 Test CLI calls Agent RPC methods
- [ ] 22.9 Test provider bootstrap flow
- [ ] 22.10 Test full CLI → Local → Provider → Workspace routing
- [ ] 22.11 Test full CLI → Local → Workspace → Provider → Container routing

## 23. Documentation & Export

- [x] 23.1 Add package-level documentation
- [x] 23.2 Export public functions/types
- [x] 23.3 Verify packages compile without errors
