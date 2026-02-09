# TOOLS KNOWLEDGE BASE

## OVERVIEW

`tools/` is a separate Go module for installer/agent tooling binaries and utility runtime code.

## STRUCTURE

```
tools/
|- go.mod                 # independent module boundary
|- cmd/installer/**       # CLI entry and install command flow
|- internal/**            # installer/plugin/agent handlers
`- pkg/**                 # tooling-specific utilities/logging/retrier
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Tool entrypoint | `tools/cmd/installer/main.go` | Cobra command execution path |
| Install step orchestration | `tools/internal/installer/**` | Node/plugin step pipelines |
| Platform-specific operations | `tools/internal/*/{unix,windows}/**` | OS-specific behavior and safety checks |
| Shared tool utilities | `tools/pkg/**` | Retries, logging, path/process helpers |

## CONVENTIONS

- Treat `tools/go.mod` as an isolated dependency/version context.
- Keep installer steps atomic and explicit (`step.go` pattern is common).
- Preserve OS split (`unix` vs `windows`) for path/process behavior.

## ANTI-PATTERNS

- Do not import application runtime internals into tools unless explicitly required.
- Do not collapse platform-specific code paths into one branch when semantics differ.
- Avoid introducing service API business logic into installer utility packages.
