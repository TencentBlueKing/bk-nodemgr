|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/discover
|Overview:generic service discovery contracts, metadata types, error factories, in-memory sample provider, endpoint selectors
|Boundary:root package owns common `discover` model+interfaces|third-party provider implementations live in child packages
|Structure:pkg/discover:{README.md,constant.go,iface.go,errors.go,default.go,selector.go,\*\_test.go,etcddiscover/}
|Where to look:package intent/boundaries:README.md:source of truth for ownership and usage limits
|Where to look:service/endpoint constants:constant.go:declare new shared `ServiceName` or `EndpointName` here
|Where to look:contracts+domain model:iface.go:`Provider`,`Discover`,`Registry`,`Selector`,`Instance`,`Endpoint`,meta helpers
|Where to look:error semantics:errors.go:constructor-style errors for service, endpoint, instance, selector, lifecycle, metadata failures
|Where to look:default provider:default.go:`ProviderDefault` in-memory implementation used as reference/example
|Where to look:selector behavior:selector.go:`RandomSelector`,`RoundRobinSelector`,`SelectEndpoints`
|Conventions:new services/endpoints:add `ServiceName`/`EndpointName` constants in root package before provider-specific usage
|Conventions:provider implementations must satisfy `Provider` and preserve `Discover`/`Registry` method meanings
|Conventions:caller boundary type is `Instance`/`Endpoint`; keep provider-specific metadata hidden behind `Meta` when needed
|Conventions:selectors choose from endpoints only; keep routing/business policy in callers, not in selector implementations
|Conventions:default provider is reference/in-memory behavior; keep it simple and avoid production backend assumptions
|Anti-patterns:do not put etcd/client/TLS/lease/watch/cache logic in root package; use `pkg/discover/etcddiscover`
|Anti-patterns:do not add service-specific routing, health, or weighting policy to root `discover`
|Anti-patterns:do not expose third-party client types through `Provider`, `Discover`, `Registry`, `Instance`, or `Endpoint`
|Anti-patterns:do not duplicate child provider error/metadata rules in root unless they are generic cross-provider contracts
|Child AGENTS:pkg/discover/etcddiscover/AGENTS.md
