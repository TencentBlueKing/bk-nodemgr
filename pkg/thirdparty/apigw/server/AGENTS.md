# APIGW SERVER KNOWLEDGE BASE

## OVERVIEW

`apigw/server` provides server-side handling for API Gateway requests.

## WHERE TO LOOK

- Server integration and middleware: `pkg/thirdparty/apigw/server/server.go`, `midleware.go`
- Auth identity and JWT logic: `auth_identity.go`, `jwt.go`

## CONVENTIONS

- Parse and normalize gateway headers via shared middleware.
- Keep auth/requestid/tenantid behaviors centralized here.

## ANTI-PATTERNS

- Do not parse gateway auth headers in scattered API handlers.
