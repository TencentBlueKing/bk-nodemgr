# APIGW CLIENT KNOWLEDGE BASE

## OVERVIEW

`apigw/client` centralizes API Gateway client auth and sensitive header handling.

## WHERE TO LOOK

- Client setup: `pkg/thirdparty/apigw/client/client.go`
- Auth/config composition: `pkg/thirdparty/apigw/client/config.go`

## CONVENTIONS

- Keep APIGW authentication behavior and masking logic in this package.
- Reuse shared APIGW header contracts from sibling packages.

## ANTI-PATTERNS

- Do not duplicate gateway auth logic in service callers.
