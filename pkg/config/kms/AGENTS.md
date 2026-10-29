|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:codegraph (code structure, call paths, impact)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/config/kms
|Overview:single encrypted-envelope→typed-secret boundary via bk-kms-sdk;decrypt global/per-module envelopes with caller-supplied private keys;unmarshal plaintext JSON into envelope structs
|Structure:pkg/config/kms:{kms.go,envelope/}
|Where to look:load/decrypt:kms.go:EnvelopeLoadParam;Load{Application,Backend,File}Envelopes;LoadGlobalEnvelope;internal per-envelope decrypt/unmarshal helpers
|Where to look:typed payloads:envelope/:GlobalEnvelope + {Application,Backend,File}Envelope|Child AGENTS:pkg/config/kms/envelope/AGENTS.md
|Service flow:pkg/config/kms.go KMS.load(file/env I/O)→EnvelopeLoadParam→Load{Application,Backend,File}Envelopes→shared GlobalEnvelope + service module envelope→pkg/config/kms.go Parse*ServiceConfig
|Global-only flow:pkg/config/kms.go KMS.loadGlobal(global file/key only)→LoadGlobalEnvelope→KMS.ParseAdminClientDatabase={etcd,redis,mongodb};KMS.ParseAdminClientBKApp={bkapp_id_secret.default app code/secret};KMS.ParseAdminClientJWT={jwt_server_config keys}
|Conventions:one load entrypoint per service returns shared global + own module type|new service=add entrypoint + module envelope type;do not widen existing signatures (OCP)
|Conventions:global-only consumers use LoadGlobalEnvelope;do not require module private keys|secret input only via caller-assembled EnvelopeLoadParam;no os.Getenv/os.ReadFile here
|Conventions:decrypt/unmarshal errors use fmt.Errorf with %w and failed-envelope context|exported symbols require English comments
|Boundary:file/env I/O and service-config mapping belong in pkg/config/kms.go|load once at startup;fail fast;no retry,cache,rotation or background/runtime refresh
|Anti-patterns:logging envelopes/plaintext/private keys/decrypted fields (including via pkg/logger);service-config mapping here;generic map[string]any passthroughs instead of typed envelope fields
