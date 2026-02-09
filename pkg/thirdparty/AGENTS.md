# THIRDPARTY KNOWLEDGE BASE

## OVERVIEW

`pkg/thirdparty/` contains integrations with external systems and SDK wrappers.
Encapsulate third-party concepts inside each package and expose scenario-oriented interfaces to internal callers.

## STRUCTURE

```
pkg/thirdparty/
|- <provider>/
|  |- README.md        # provider-specific contract and usage
|  |- <provider>.go    # raw API client/wrapper
|  |- types.go         # request/response mapping types
|  `- handler.go       # domain adapter and conversion boundary
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Add a new external system | `pkg/thirdparty/<provider>/` | Keep provider isolated in its own package |
| Adapt business-facing APIs | `pkg/thirdparty/<provider>/handler.go` | Expose scenario APIs, not raw endpoints |
| Define data mapping | `pkg/thirdparty/<provider>/types.go` | Keep third-party payload shape local |
| Provider-specific caveats | `pkg/thirdparty/<provider>/README.md` | Source of truth for provider behavior |

## CONVENTIONS

- Depend on interfaces defined in handler layer; avoid binding callers to concrete handler implementations.
- Normalize third-party errors in handler layer into project-standard errors.
- Keep conversion logic in provider package to prevent external concepts leaking into business modules.

## ANTI-PATTERNS

- Do not expose raw third-party API structures directly to upper layers.
- Do not bypass handler interfaces by constructing provider clients across business code.
- Do not mix multiple third-party providers in one package.
