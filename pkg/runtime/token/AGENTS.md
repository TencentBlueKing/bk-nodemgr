|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/runtime/token
|OVERVIEW:Creates compact stateless tokens containing arbitrary []byte payloads and embedded expiration times authenticated with truncated HMAC-SHA256
|OVERVIEW:Boundary=business-agnostic token framing, signing, parsing, expiration validation, and empty-key fallback only; callers own payload encoding, authorization, storage lookup, and configured key lifecycle
|Structure:pkg/runtime/token:{token.go,README.md,AGENTS.md}
|WHERE TO LOOK:public API:token.go:{New,WithTTL,Generator.Generate,Generator.Parse,ErrInvalid,ErrExpired}
|WHERE TO LOOK:wire/security contract:README.md:{实现说明,使用限制,功能边界}
|CONVENTIONS:Keep the package stdlib-only, portable, stateless per call, and safe for concurrent use after Generator construction
|CONVENTIONS:Default TTL is 24 hours; WithTTL accepts positive durations; encoded expiration uses unsigned 32-bit Unix seconds with one-second precision
|CONVENTIONS:New uses the package default key when passed an empty key; any non-empty caller key overrides the default
|CONVENTIONS:Wire format=v1(1B)→expiresAt big-endian uint32(4B)→payload(NB)→HMAC-SHA256 tag(12B); encode with base64.RawURLEncoding
|CONVENTIONS:HMAC covers version+expiration+payload under signingDomain; verify with hmac.Equal before reading expiration or returning payload
|CONVENTIONS:Copy constructor keys and parsed payloads; never retain or expose caller-owned mutable byte slices
|CONVENTIONS:ErrInvalid covers malformed/unsupported/unauthentic input; ErrExpired is returned only after successful authentication
|CONVENTIONS:Document token length effects; 16-byte payload produces a 44-character token under v1
|ANTI-PATTERNS:Do not add export IDs, UUID parsing, database access, download policy, authorization, configuration loading, logging, or service imports
|ANTI-PATTERNS:Do not claim confidentiality or use token payloads for secrets; the payload is Base64-readable and only integrity-protected
|ANTI-PATTERNS:Do not shorten the HMAC tag, compare tags with ==/bytes.Equal, trust expiration before verification, or accept empty payloads/non-positive TTLs
|ANTI-PATTERNS:Do not change formatVersion, field order, timestamp width, signingDomain, tagSize, or Base64 encoding without a new version and coordinated producer/consumer rollout
|Commands:gofmt -w pkg/runtime/token/token.go|go test ./pkg/runtime/token|go vet ./pkg/runtime/token
