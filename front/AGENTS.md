# FRONTEND KNOWLEDGE BASE

## OVERVIEW

`front/` is a Vue 3 + TypeScript + Vite application managed by `pnpm`.

## STRUCTURE

```
front/
|- src/               # app source (pages/components/composables/api/modules)
|- public/            # static assets
|- package.json       # scripts and dependency graph
|- .eslintrc.js       # frontend lint rules
`- tsconfig.json      # TS project config
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| App bootstrap/routing | `front/src/main.ts`, `front/src/pages/**` | Entry and page composition |
| API modules | `front/src/api/modules/**` | Keep naming aligned with backend proto APIs |
| Shared UI blocks | `front/src/components/**` | Auto-registered component set |
| Tooling/scripts | `front/package.json` | `dev`, `build`, `lint`, `typecheck`, `test:*` |
| Lint policy | `front/.eslintrc.js` | BlueKing base config + import sorting + type imports |

## CONVENTIONS

- Use `pnpm` (declared package manager is `pnpm@9.8.0`).
- Keep API naming consistent with backend `proto` contracts.
- Follow import sorting and consistent type import/export rules from ESLint config.

## ANTI-PATTERNS

- Do not introduce backend-only assumptions in frontend modules.
- Do not bypass `typecheck`/`lint` for frontend changes.
- Avoid naming drift between frontend API modules and backend proto definitions.
