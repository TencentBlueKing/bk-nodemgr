# CMD KNOWLEDGE BASE

## OVERVIEW

`cmd/` contains executable entrypoints; keep it thin and delegate runtime behavior to `internal/*/service`.

## STRUCTURE

```
cmd/
|- backend/main.go      # backend HTTP service bootstrap
|- file/main.go         # file HTTP service bootstrap
|- relay/main.go        # relay service bootstrap
`- application/*.go     # multi-command CLI (webserver/scheduler/migrate/...)
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Backend startup flags/run mode | `cmd/backend/main.go` | Cobra command + config load/validate + logger init |
| File service startup | `cmd/file/main.go` | Similar startup path; includes tenant and shutdown watcher |
| Relay startup | `cmd/relay/main.go` | PID-file setup and relay service bootstrap |
| Application command wiring | `cmd/application/root.go` | Subcommands are registered here |

## CONVENTIONS

- Keep startup sequence consistent: parse flags -> load config -> validate -> init logger -> construct service -> start.
- Prefer wiring-only changes in `cmd`; business logic belongs under `internal`.
- Use `pkg/version` for banner/version output and `pkg/config` for typed config loading.

## ANTI-PATTERNS

- Do not add business/domain logic directly in `cmd/*`.
- Do not duplicate service lifecycle logic across entrypoints when it can be shared under `internal/*/service`.
- Do not bypass config validation before service start.
