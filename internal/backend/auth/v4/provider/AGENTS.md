|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow .agents/skills/scoped-agentsmd/references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:internal/backend/auth/v4/provider
|Overview:V4 resource queries and runtime attribute enrichment|typed query validation and resource dispatch|storage queries and callback response shaping
|Ownership:V3 and V4 providers are independently owned|no cross-version auth/provider imports|keep V4 behavior and changes local to this version
|Protocol:V4-owned DTOs and two callback methods={list_instance,fetch_instance_info}|Handler.DispatchMethod owns method dispatch and typed filter conversion; resource providers receive typed requests
|Where to look:query implementation:handler.go:{NewHandler,RegisterProvider,RequestMethod,DispatchMethod,ListInstance,FetchInstanceInfo,FetchResourceAttributes}
|Where to look:contracts:{iface.go,provider.go}:IHandler/IDispatcher/IQueryHandler/IInstanceLister/IAttributeEnricher|router consumes IDispatcher; runtime keeps typed query/enrichment interfaces|IProvider implements ListInstance and FetchInstanceInfo|Request carries typed Filter, Page and Requires
|Where to look:resource providers:{networkarea.go,networkunit.go,package.go,packagetype.go}:networkarea|networkunit|package|package_type
|Where to look:paths and pagination:provider.go:{BuildIAMPath,ParseParentFromIAMPath,paginateInstances}|resource constants and display names:constants.go|callback errors:ERROR_CODES.md
|Where to look:registration and authorized-scope expansion:internal/backend/auth/v4/auth_iamv4.go:{NewProviderHandler,listAllResourceIDs}|enumerate all pages; reject changing totals, short pages, empty/duplicate IDs
|Where to look:transport boundary:internal/backend/router/api-v3/iam/v4/v4.go:{basicAuthMiddleware,handleResourceCallback,callbackError}|request validation and page conversion:pkg/proto/backend/api/v3/iam.go
|Where to look:router permission resources:internal/backend/auth/resource.go:root auth.BuildPackageResources constructs canonical types.AuthResource slices; routers must use this helper, not version-specific providers
|Conventions:resource IDs returned by callbacks must match authorization resource IDs|generic permission resource construction belongs to root auth, not this provider package
|Current Flow:resolve/validate tenant -> tenant-scoped BasicAuth token check -> V4 JSON validation and page conversion -> DispatchMethod -> typed query handler -> provider/storage -> HTTP 200 data envelope|echo X-Request-Id on success and failure
|Conventions:list page={page>=1,page_size=1..1000} -> types.Page offset/limit with overflow guard via V4 proto conversion; no V3 pagination defaults|transport uses standard Gin proto JSON binding|fetch filter.ids=non-null string array, <=1000 nonempty IDs; empty array returns [] without storage queries
|Conventions:fetch requires is top-level; omitted/empty selects all supported attributes, unknown attributes ignored, id always returned|InstanceInfo.MarshalJSON flattens attributes and emits _bk_iam_path_ as one string; runtime BuildIAMPath/FetchResourceAttributes retain arrays|enrichment batches IDs by MaxFetchInstanceIDs
|Conventions:list without parent enumerates current-tenant candidates, not authorization grants|parent and keyword filters intersect|deterministic pages and filtered Count|explicit empty lists, including missing instances and out-of-bounds pages
|Conventions:data sources unchanged:topo storage for networkarea/networkunit; fixed package types; DistinctNameReleasePlugin/DistinctNameReleasePluginBinTool and fixed names for packages|filter names -> sort -> distinctNamesWithPagination; no release-version List truncation
|Conventions:package ID=canonical name|fetch order=plugin -> pluginbintool -> fixed types|cross-type same-name packages are outside current business scenarios; preserve defensive validation, do not add deduplication, change IDs or precedence
|Conventions:storage/query failures use logger.G.Biz(ctx) before wrapping|propagate errors with fmt.Errorf and %w|never swallow errors with Warn
|Errors:ERROR_CODES.md|ErrInvalidArgument -> 400 INVALID_ARGUMENT|invalid credentials -> 401 UNAUTHENTICATED|ErrNotFound -> 404 NOT_FOUND|unexpected/token lookup/storage/serialization failures -> 500 INTERNAL|router owns error envelope; preserve error chains
|Extension:define ResourceTypeXxx and XxxProvider -> implement two-method IProvider -> reuse version-local models/filters -> register in internal/backend/auth/v4/auth_iamv4.go:NewProviderHandler|protocol changes require separate approval
|Testing:{provider_test.go,provider_topology_test.go}:hand-written storage mocks|empty/out-of-bounds pagination|error propagation|keyword filtering|attribute serialization and topology paths|follow repo test-delivery rules
|Commands:go test ./internal/backend/auth/v4/provider/... -v
|Anti-patterns:no raw callback body parsing in providers or bypassing handler validation|no legacy callback methods, proto DTOs or policy-expression resolver in V4 providers|no partial nil payload shapes|no heavy business logic: adapter and query orchestration only|no cross-version imports or generic resource builder copies|candidate enumeration never grants permission
