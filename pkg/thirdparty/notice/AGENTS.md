# NOTICE KNOWLEDGE BASE

## OVERVIEW

`notice` provides standardized integration with BlueKing Notice Center.

## WHERE TO LOOK

- API wrapper and orchestration: `pkg/thirdparty/notice/notice.go`, `pkg/thirdparty/notice/handler.go`
- Data models: `pkg/thirdparty/notice/types.go`
- Fallback/noop path: `pkg/thirdparty/notice/noop_handler.go`

## CONVENTIONS

- Use scenario-oriented handler interfaces; do not expose raw notice APIs.
- Centralize error wrapping and retry behavior in handler layer.
- Keep APIGW auth handling through package config (`Config.VirtualUserConfig`).

## ANTI-PATTERNS

- Do not couple callers to notice center native models.
