|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow .agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:internal/backend/auth/v3/provider
|Overview:V3-owned IAM callback providers|method validation and resource dispatch|typed filter decoding|storage queries and callback response shaping|policy instance resolution and resource attribute enrichment
|Ownership:V3 and V4 providers are independently owned|no cross-version auth/provider imports|keep V3 behavior and changes local to this version
|Where to look:dispatch implementation:handler.go:NewHandler|RegisterProvider|DispatchMethod|method validation and typed filter dispatch
|Where to look:contracts:{iface.go,provider.go}:IHandler/IDispatcher/IResolver/IAttributeEnricher|IProvider and version-local callback models
|Where to look:resource providers:{networkarea.go,networkunit.go,package.go,packagetype.go}:networkarea|networkunit|package|package_type
|Where to look:policy and pagination:expression_helper.go|resource constants and display names:constants.go|callback error semantics:ERROR_CODES.md
|Where to look:registration:internal/backend/auth/v3/auth_iam.go:version auth.NewProviderHandler constructs handler and registers providers; router consumes the injected dispatcher
|Where to look:transport boundary:internal/backend/router/api-v3/iam/v3/v3.go:handleResourceCallback binds proto JSON and converts filter/page before DispatchMethod
|Where to look:router permission resources:internal/backend/auth/resource.go:root auth.BuildPackageResources constructs canonical types.AuthResource slices; routers must use this helper, not version-specific providers
|Conventions:resource IDs returned by callbacks must match authorization resource IDs|generic permission resource construction belongs to root auth, not this provider package
|Flow:router JSON binding -> filter map and protoBackend.ConvIAMCallbackPageToTypes -> DispatchMethod validates method/resolves resource type -> typed filter conversion -> provider/storage -> callback payload
|Conventions:filter fields decode from map[string]interface{}|pagination limits enforced by proto page conversion|InstanceInfo custom JSON flattens dynamic attributes|AttributeValueID custom UnmarshalJSON supports string/int/bool
|Conventions:deterministic provider methods|explicit empty lists instead of nil slices for unused callback methods|preserve mixed string/number IDs with safe conversions|reuse pkg/runtime/conv|case-insensitive keyword search where supported
|Conventions:package pagination=storage DistinctNameReleasePlugin -> sort -> distinctNamesWithPagination|no List+in-memory dedup: inaccurate Count and unstable pages
|Conventions:storage/query failures use logger.G.Biz(ctx) before wrapping|propagate errors with fmt.Errorf and %w|never swallow errors with Warn
|Errors:follow ERROR_CODES.md|input/validation=parameter errors|unsupported resource type=not found|unexpected provider/storage failures=unknown/internal|preserve error chains
|Extension:define ResourceTypeXxx and XxxProvider -> implement full IProvider in provider.go -> reuse version-local models/filters -> register in internal/backend/auth/v3/auth_iam.go:NewProviderHandler|keep unused callback methods explicitly empty
|Testing:{provider_test.go,provider_topology_test.go}:hand-written storage mocks|empty/out-of-bounds pagination|error propagation|keyword filtering|topology semantics|follow repo test-delivery rules
|Commands:go test ./internal/backend/auth/v3/provider/... -v
|Anti-patterns:no raw callback body parsing in providers or bypassing handler validation|no proto structs as business models|no partial nil payload shapes|no heavy business logic: adapter and query orchestration only|no cross-version imports or generic resource builder copies
