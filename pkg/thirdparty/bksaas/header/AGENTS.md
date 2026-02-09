# BKSAAS HEADER KNOWLEDGE BASE

## OVERVIEW

`bksaas/header` contains BK SaaS header definitions and client adjustment helpers.

## WHERE TO LOOK

- Header definitions: `pkg/thirdparty/bksaas/header/header.go`
- Middleware integration: `pkg/thirdparty/bksaas/header/middlware.go`

## CONVENTIONS

- Keep header key/value policy centralized.
- Apply header normalization through shared helper/middleware.

## ANTI-PATTERNS

- Do not hardcode BK SaaS headers outside this package.
