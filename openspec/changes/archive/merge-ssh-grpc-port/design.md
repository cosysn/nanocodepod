## Context

The original design specified that SSH and gRPC should share the same port using protocol detection. The codebase already has multiplexing infrastructure (`MuxListener`, `handleMultiplexedConnection`, etc.) but the current implementation uses two separate listeners:
- SSH on port N
- gRPC on port N+1

This results in:
- Unnecessary port consumption (2 ports per workspace)
- Complicated firewall rules
- Inconsistent with original design

## Goals / Non-Goals

**Goals:**
- Merge SSH and gRPC to use a single port
- Use existing protocol detection infrastructure (HTTP/2 magic bytes detection)
- Update workspace code to allocate only one port
- Update Docker port bindings to map only one port

**Non-Goals:**
- Adding new SSH/gRPC functionality
- Changing authentication mechanisms
- Modifying the agent's core functionality beyond port configuration

## Decisions

1. **Use existing mux infrastructure**: The codebase already has `handleMultiplexedConnection` and connection wrappers for protocol detection. Use this instead of two separate listeners.

2. **Single listener approach**: Modify `RunAgent` to create only one listener on the configured port, routing connections based on protocol detection.

3. **Port pool unchanged**: The port pool already allocates single ports. The workspace code incorrectly allocates `port+1` for gRPC - this is the bug to fix.

4. **Docker port binding**: Remove the container port 23 -> host port+1 mapping. Only map container port 22 to host port.

## Risks / Trade-offs

- **[Risk]** gRPC connection detection may fail for small packets
  - **[Mitigation]** The peek buffer checks for HTTP/2 magic bytes which is the first data sent in any gRPC connection

- **[Risk]** Existing workspaces may have stale port allocations
  - **[Mitigation]** Port pool handles cleanup; new workspaces get fresh ports

## Migration Plan

1. Modify `RunAgent` in `agent/server.go` to use single listener with protocol multiplexing
2. Remove second listener creation (port+1) in workspace code
3. Update Docker port bindings to remove gRPC-specific mapping
4. Update workspace types if needed to reflect single port
