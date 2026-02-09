# BKLOGIN KNOWLEDGE BASE

## OVERVIEW

`bklogin` handles BK SaaS login/auth identity related integration flows.

## WHERE TO LOOK

- Login client/wrapper: `pkg/thirdparty/bksaas/bklogin/bklogin.go`
- Business adapter: `pkg/thirdparty/bksaas/bklogin/handler.go`
- Auth identity handling: `pkg/thirdparty/bksaas/bklogin/auth_identity.go`
- Types: `pkg/thirdparty/bksaas/bklogin/types.go`

## CONVENTIONS

- Keep login provider details inside this package.
- Expose stable handler interface to upstream callers.

## ANTI-PATTERNS

- Do not spread login token parsing and identity mapping to callers.
