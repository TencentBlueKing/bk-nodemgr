|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:pipe-index format, concise, no prose/code blocks
|Scope:support-files/bkiamv4/migrate/iamv4
|Overview:IAM V4 model API adapter for the migration CLI; follows pkg/thirdparty layering but is separate from the runtime authorization adapter
|References:pkg/thirdparty/AGENTS.md|pkg/thirdparty/iamv3/AGENTS.md|pkg/thirdparty/iamv4/AGENTS.md
|Where to look:scenario interfaces, conversion, complete pagination and aggregation:handler.go
|Where to look:single-request HTTP API methods, endpoint paths, headers and response decoding:iamv4.go
|Where to look:named API request/response types, JSON tags and optional-field semantics:types.go
|Where to look:migration ordering, upsert decisions, dry-run and allowed-update policy:../migrate.go|../resource_type.go
|Conventions:callers depend on handler-layer interfaces; keep the raw client and HTTP wire requests/responses inside this package
|Conventions:raw client methods accept named *XxxReq types and return named XxxResp payloads when the API has data; define these in types.go, never anonymous request structs in HTTP methods
|Conventions:each raw client method performs one API request; list methods expose one page including count/results; handler owns all-page iteration, aggregation and completeness checks
|Conventions:handler converts scenario inputs to API DTOs and normalizes errors; do not expose pagination envelopes or endpoint assembly to the migration runner
|Conventions:API methods model documented protocol capabilities; migration policy such as name-only updates and topology-drift rejection stays in the runner, not the HTTP client
|Conventions:preserve omitted versus explicit empty mutable fields with pointers; preserve documented array request bodies and HTTP success codes
|Conventions:validate migration input at the runner boundary; follow existing System client methods without adding per-request Validate methods or duplicate validation immediately before HTTP calls; keep startup Config.Validate
|Conventions:reuse pkg/rest/client and pkg/thirdparty/apigw; propagate contextx tenant and cancellation; application authentication and header masking remain centralized
|Conventions:check HTTP status and IAM error envelope; wrap causes with %w; distinguish confirmed absence from failed queries; fail on incomplete pagination rather than return partial state
|Anti-patterns:no generic HTTP framework, raw endpoint calls from the runner, client-side pagination loops, migration decisions in handler/client, or duplicated wire DTOs outside types.go
|Anti-patterns:no secrets/auth headers/raw remote messages in logs; no fallback authorization bypass, automatic destructive repair, or runtime callback/token-cache logic in this model adapter
|Commands:go test ./support-files/bkiamv4/migrate/...|make -C support-files/bkiamv4/migrate test
