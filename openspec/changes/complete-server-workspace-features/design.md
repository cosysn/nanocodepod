## Context

The WSL server mode currently supports basic workspace CRUD operations but lacks features that exist in the standalone CLI:
- Git repository cloning
- Local directory mounting
- Devcontainer support for custom Docker images
- Agent binary injection
- IDE auto-launch

These features are essential for developer workflows. The server runs in WSL and manages Docker containers on behalf of the CLI client.

## Goals / Non-Goals

**Goals:**
- Add Git clone support during workspace creation
- Add local directory as workspace source
- Add devcontainer.json parsing and image building
- Add agent injection capability
- Add IDE auto-launch via protocol handlers

**Non-Goals:**
- Refactoring existing workspace CRUD logic (reuse where possible)
- Supporting multiple agents per workspace
- IDE session management/persistence

## Decisions

1. **Reuse existing codepod workspace code** - The standalone CLI at `codepod/internal/workspace/` already has implementations for git clone, local dir, devcontainer, and agent injection. These will be adapted for server use rather than rewritten.

2. **Extend existing API endpoints** - Rather than creating new endpoints, extend `/workspace/create` with optional fields:
   - `git_url`, `git_branch` for Git clone
   - `local_dir` for local directory
   - `devcontainer_path` for devcontainer
   - `inject_agent` boolean for agent injection

3. **IDE launch happens on client side** - The server returns workspace connection info (host, port, SSH details), and CLI client launches IDE via local protocol handler. This keeps server responsibility focused.

4. **Agent injection uses container exec** - Agent binary is copied into running container and started via `docker exec`, similar to standalone CLI approach.

## Risks / Trade-offs

- **Risk**: Git operations may be slow for large repos → Mitigation: Add progress indication, consider shallow clone option
- **Risk**: Devcontainer builds may fail → Mitigation: Return build logs to client, allow retry
- **Risk**: Agent injection requires openssh-client in container → Mitigation: Install via container entrypoint or agent init

## Migration Plan

1. Extend server API with new optional fields
2. Update CLI to pass new options to server
3. Add workspace methods that delegate to existing codepod internal packages
4. Test end-to-end with each feature

## Open Questions

- Should git credentials be stored/retrieved for private repos?
- Should devcontainer images be cached to avoid rebuilds?
