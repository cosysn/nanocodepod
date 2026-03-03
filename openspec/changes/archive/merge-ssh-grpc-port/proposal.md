## Why

Currently, the agent uses TWO separate ports: SSH on port 22 (mapped to host port N) and gRPC on port 23 (mapped to host port N+1). However, the original design explicitly specified that SSH and gRPC should share the same port using protocol detection/mux. This bug results in unnecessary port consumption and complicates firewall rules.

## What Changes

- Implement port multiplexing so SSH and gRPC share a single port
- Remove the second port allocation (container port 23)
- Keep only container port 22 for both SSH and gRPC connections
- Protocol detection based on first bytes (gRPC uses HTTP/2 magic bytes)
- Update port pool to allocate only one port per workspace
- Update Docker port bindings to remove the gRPC-specific mapping

## Capabilities

### New Capabilities

- `single-port-agent`: Agent accepts both SSH and gRPC connections on a single port using protocol multiplexing

### Modified Capabilities

- `ssh-access`: Modify to support gRPC on same port (already designed, not fully implemented)
- `grpc-api`: Modify to support sharing port with SSH (already designed, not fully implemented)

## Impact

- **Workspace Package**: Remove port+1 allocation, update port bindings
- **Agent Package**: Implement port mux to handle both SSH and gRPC on same port
- **CLI**: No changes needed (still shows single port)
- **Docker**: Remove second port mapping in container configuration
