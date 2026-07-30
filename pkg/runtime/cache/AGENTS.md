|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/runtime/cache
|OVERVIEW:Defines portable cache interfaces and process-local runtime cache implementations:{ICache,IKeyGenerator,MemoryCache}
|OVERVIEW:Boundary=business-agnostic cache primitives only; callers own key design, value encoding, TTL policy, security policy, and backend selection
|Structure:pkg/runtime/cache:{iface.go,memory.go,memory_test.go,README.md,AGENTS.md}
|WHERE TO LOOK:public contracts:iface.go:{ICache,IKeyGenerator,NewDefaultKeyGenerator}
|WHERE TO LOOK:in-memory implementation:memory.go:{MemoryCache,NewMemoryCache,Get,Set,SetNX,SetNXWithExpiration,SetWithExpiration,Exists,Delete}
|WHERE TO LOOK:behavior coverage:memory_test.go|copy semantics, TTL expiration, zero-expiration behavior, SetNX semantics, Exists/Delete, nil context
|WHERE TO LOOK:package intent:README.md|high-level cache interface purpose
|CONVENTIONS:Keep this package dependency-light and runtime-only; do not import internal/*, service packages, router packages, storage, IAM, or business-domain types
|CONVENTIONS:Preserve ICache method semantics across implementations:{nil context returns error, Get miss returns error, Set uses backend default TTL, SetWithExpiration controls TTL, SetNX returns false for live existing keys}
|CONVENTIONS:MemoryCache stores []byte values only and must clone bytes on read/write to avoid caller mutation, secret aliasing, and data races
|CONVENTIONS:Explicit zero expiration in SetWithExpiration/SetNXWithExpiration means no expiration; default-expiration methods must keep using backend default TTL
|CONVENTIONS:SetNX behavior should remain atomic for implementations whose backend supports it; keep MemoryCache based on go-cache Add rather than Get→Set when preserving atomicity matters
|CONVENTIONS:Do not start background cleanup goroutines from MemoryCache unless a current caller needs proactive cleanup; lookup-time expiration is sufficient for current process-local use
|CONVENTIONS:When changing cache semantics, update memory_test.go and compare pkg/rediscache behavior before adjusting shared interface expectations
|ANTI-PATTERNS:Do not put IAM token rules, permission policy, tenant decisions, logging, metrics, secret handling policy, or feature-specific cache keys in this package
|ANTI-PATTERNS:Do not change ICache signatures, error meaning, or TTL semantics without coordinating all implementations and callers
|ANTI-PATTERNS:Do not store or expose caller-owned byte slices directly; no zero-copy shortcuts for cached values
|ANTI-PATTERNS:Do not treat go-cache errors as application errors when Add only reports existing live keys; map SetNX existing-key behavior to false,nil
|ANTI-PATTERNS:Do not add Redis/file/distributed cache behavior here; keep backend-specific implementations in their owning packages such as pkg/rediscache
|Commands:go test ./pkg/runtime/cache|golangci-lint run --config=.golangci.yml --timeout=5m --new-from-rev=origin/master
