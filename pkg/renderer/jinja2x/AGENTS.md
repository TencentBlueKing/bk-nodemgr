# JINJA2X KNOWLEDGE BASE

## OVERVIEW

`jinja2x` wraps a staticx-packaged Python jinja2 binary as an embedded subprocess renderer. It provides `Handler.Render(tmpl, content)` to render Jinja2 templates via temp-file IPC, and is one of two `renderer.IHandler` implementations (the other being `gotemplate`).

## STRUCTURE

```
jinja2x/
|- jinja2x.go        # public API: Handler, Render, New
|- jinja2_bin.go     # binary lifecycle + subprocess execution with process group management
|- jinja2x_test.go   # table-driven render tests
|- jinja2/
|  |- jinja2.go                # package doc only
|  |- jinja2_linux_amd64.go   # embed binaries/jinja2_exec_amd64
|  |- jinja2_linux_arm64.go   # embed binaries/jinja2_exec_arm64
|  `- binaries/                # pre-built staticx-packaged Python binaries
```

## WHERE TO LOOK

| Task | File | Notes |
|------|------|-------|
| Add render options/flags | `jinja2x.go` | `Render()` builds CLI args from template + content |
| Modify subprocess lifecycle | `jinja2_bin.go` | Binary extraction, process group, zombie reaping |
| Update embedded binary | `jinja2/binaries/` | Replace binary, rebuild with staticx |
| Add test cases | `jinja2x_test.go` | Table-driven, covers variable/substitution/loop/error |

## CONVENTIONS

### Process Group Management (staticx-specific)

The jinja2_exec binary is a staticx-packaged Python process. At runtime staticx forks a child to run the actual interpreter, creating a parent-child hierarchy. If the parent dies unexpectedly, the Python child can become an orphaned zombie adopted by init (or the backend process). All subprocess execution must follow this pattern:

1. `Setpgid: true` — create a new process group so all descendants can be managed as a unit
2. `Pdeathsig: syscall.SIGKILL` — ensure children receive SIGKILL when the Go parent dies
3. `cmd.Cancel` kills the entire process group (`-pgid`) on context cancellation
4. `cmd.WaitDelay = 0` — immediate kill after cancellation, no grace period
5. Zombie reaping loop via `syscall.Wait4(-pgid, WNOHANG)` after `cmd.Run()` returns, with `maxWaitAttempts = 100` safety limit

Reference implementation: `pkg/wmix/wmi_bin.go` (same staticx process group pattern).

### Binary Lifecycle

- Embedded binary is extracted once to a temp file via `sync.Once` singleton (`jinja2ExecBin`)
- Temp file is chmod 0700 for executability
- Cleanup removes the temp file but is not called automatically on process exit (acceptable for long-running backend)

### IPC via Temp Files

- Template string → temp file (`-t` flag)
- Context JSON → temp file (`-c` flag)
- Render result ← temp file (`-o` flag)
- All temp files use `pkg/runtime/tmp` with explicit `defer` cleanup

### Context Usage

`jinja2x.go:Render()` passes `context.Background()` to `jinja2ExecRunCmd` — there is no caller-controlled cancellation. The process group management ensures clean termination even without explicit context cancellation.

## ANTI-PATTERNS

- Do not remove `Setpgid` / `Pdeathsig` / `Cancel` / `Wait4` — staticx-forked children will become zombies
- Do not add service-specific rendering logic here — this is a shared renderer, orchestration belongs in `internal/`
- Do not edit `jinja2/binaries/` binaries directly — rebuild via staticx pipeline
- Do not use `context.Background()` alternatives without understanding the zombie-reaping implications
