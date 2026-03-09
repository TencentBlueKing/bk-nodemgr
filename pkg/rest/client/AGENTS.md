# CLIENT KNOWLEDGE BASE

## OVERVIEW

`pkg/rest/client` provides the shared HTTP client used by all bk-nodemgr services and third-party adapters.
It encapsulates retry logic, OpenTelemetry tracing, Prometheus metrics, request building, and response parsing.
This package is business-agnostic — no domain logic belongs here.

Key capabilities:
- fluent request builder (`Request`) with verb methods (`Get`, `Post`, `Put`, `Delete`, `Patch`, `Head`)
- retry on `ECONNRESET` for GET requests (max 3 cycles, 20ms delay)
- tolerance latency logging: warns when a request exceeds the configured threshold
- OpenTelemetry span injection per request
- Prometheus metrics collection via `pkg/rest/metrics`
- three response consumers: `Into` (JSON parse), `RawData` (raw bytes), `RawStream` (streaming body)

## WHERE TO LOOK

- Client construction, options, and `IClient` interface: `client.go`
- Request builder and `Result` response consumers (`Into`, `RawData`, `RawStream`): `request.go`
- `Capability`, `Discover`, tracing wiring: `types.go`
- HTTP transport and TLS configuration: `http_client.go`
- OpenTelemetry span attribute constants: `trace.go`

## CONVENTIONS

- Always inject the client via `IClient` interface; never depend on the concrete `*Client` type in callers.
- Use `Capability.ToleranceLatencyTime` to configure per-client latency thresholds; default is `ToleranceLatencyTimeDefault` (500ms).
- Response consumers differ in error thresholds — choose carefully:
  - `Into(obj)` — JSON response; treats 5xx as hard error, passes 4xx through for caller inspection.
  - `RawData()` — raw bytes; same error threshold as `Into`.
  - `RawStream()` — file/binary streaming; treats **all 4xx and 5xx as errors** and closes the body immediately. The caller owns the returned `io.ReadCloser` and must close it.
- `maxErrBodySize` (1024 bytes) caps how much of an error response body is read into the error message — guards against large HTML/XML responses from proxies or object storage.
- All exported identifiers require English doc comments ending with a period.
- Use structured logging via `pkg/logger`; never use `fmt.Println` or `log.*`.

## ANTI-PATTERNS

- Do not place service-specific or business logic in this package.
- Do not add response consumers that parse domain-specific payload shapes — keep consumers generic.
- Do not change the retry policy (verb restriction, cycle count) without verifying idempotency guarantees across all callers.
- Do not bypass `Capability.Discover` by hardcoding endpoint URLs in request builders.
- Do not suppress errors from `Result.Err` silently; every error must be propagated or logged.
- Do not return `r.Body` to callers without documenting that they are responsible for closing it.
