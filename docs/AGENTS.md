# DOCS KNOWLEDGE BASE

## OVERVIEW

`docs/` is the documentation index for API, developer, operation, and concept materials.

## STRUCTURE

```
docs/
|- api/           # API process specs and swagger artifacts
|- developer/     # setup, compile, plugin/API dev docs
|- operation/     # deployment/ops playbooks
|- concepts/      # domain concepts and architecture notes
`- README.md      # docs navigation entry
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| API development procedure | `docs/api/API接口开发流程.md` | Proto-to-router-to-storage process |
| Local setup/toolchain | `docs/developer/setup.md` | Go/proto/gopls setup references |
| Build and compile docs | `docs/developer/compile.md` | Build usage and packaging notes |
| Docs index | `docs/README.md` | Top-level doc category routing |

## CONVENTIONS

- Keep docs task-oriented and anchored to concrete paths/commands.
- When API contracts change, update docs alongside proto/router changes.
- Prefer concise checklists/tables for multi-step workflows.
- Write docs from the target reader's task/decision perspective: preserve reader-needed facts, links, limits, risks, and actions; keep author process, draft status, evidence gaps, and unnecessary internal details out of the final docs body; do not remove key evidence just to make the page shorter or template-shaped.

## ANTI-PATTERNS

- Do not leave docs stale after changing API contracts or build flow.
- Do not duplicate large command blocks already maintained in root docs unless scope differs.
- Avoid mixing operational procedures into developer-only sections.
