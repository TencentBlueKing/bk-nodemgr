# BKSAAS KNOWLEDGE BASE

## OVERVIEW

`bksaas` manages BK SaaS related concepts and shared integration capabilities.

## WHERE TO LOOK

- Shared concepts root: `pkg/thirdparty/bksaas/`
- Header/auth related handling: `pkg/thirdparty/bksaas/header/`
- Login/auth flows: `pkg/thirdparty/bksaas/bklogin/`

## CONVENTIONS

- Keep BK SaaS common abstractions centralized under this subtree.
- Separate cross-cutting header/auth helpers from concrete login flow logic.

## ANTI-PATTERNS

- Do not duplicate BK SaaS header/auth logic across unrelated packages.
