|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/tracing
|Overview:OpenTelemetry tracing primitives for bk-nodemgr|owns exporter setup→service tracer providers→propagators→resource attributes→lifecycle→process-level global fallback
|Structure:pkg/tracing:{tracing.go,handler.go,service.go,iface.go,config.go,enum.go,types.go,README.md}
|Where to look:package entry/global fallback:tracing.go:{Init,G}:sets project handler and OTel global provider/propagator fallback
|Where to look:provider/exporter lifecycle:handler.go:{New,NewService,initGlobalService,newTracerProvider,newTextMapPropagator,ShutdownAll,RemoveService,newExporter}
|Where to look:public contracts:iface.go:{IHandler,IService}|service.go:{Service.TracerProvider,Service.TracerPropagator,Service.Shutdown}
|Where to look:config validation:config.go:{Config,ExporterConfig,OTLPConfig,ServiceConfig,Validate,DefaultConfig}:sampleRate must stay in 0.0..1.0
|Where to look:service categories/attributes:enum.go:{ServiceCategoryHTTP,ServiceCategoryDB,ServiceCategoryCache,ServiceCategoryAsyncBackend,ServiceCategoryKey}
|Where to look:HTTP server injection:pkg/rest/server:{server.go,middleware.go}:creates service tracer and passes provider/propagator into otelgin
|Where to look:dependency tracing callers:internal/backend/service/service.go:{Redis,MongoDB,file client,workflow}:uses explicit service TraceSampleRate and provider injection
|Where to look:client/workflow propagation:pkg/rest/client/request.go|pkg/workflow:{worker.go,controller.go,trigger_handler.go}|pkg/goasync/handler.go
|Where to look:process profiling boundary:pkg/profiling/profiling.go|cmd/{backend,file,relay}/main.go|cmd/application/webserver.go:separate Pyroscope lifecycle, not tracing-owned
|Dependencies:uses pkg/contextx for init context|OpenTelemetry sdk/trace resource propagation semconv stdouttrace otlptracegrpc|callers inject providers into otelgin redisotel otelmongo restclient workflow
|Conventions:Go runtime config supports only tracing.globalService.traceServiceName and tracing.globalService.traceSampleRate for process-level global fallback service config
|Conventions:default process-level global fallback TraceSampleRate is 0 unless explicitly configured under tracing.globalService
|Conventions:per-service providers are the primary model; callers create or receive tracing.IService through tracing.G().NewService(tracing.ServiceConfig{...})
|Conventions:explicit provider injection wins over OTel globals; pass TracerProvider() and TracerPropagator() through local capabilities/options when a caller supports it
|Conventions:OTel global provider is process-level fallback only; keep tracing.Init and G lazy fallback for libraries/call paths using otel.Tracer or otel.GetTracerProvider
|Conventions:service wiring owns TraceServiceName, ServiceCategory, and TraceSampleRate selection; pkg/tracing only validates/applies ServiceConfig
|Conventions:per-service TraceSampleRate is authoritative for that provider; do not inherit upstream sampled decisions when local config says to reduce/drop sampling
|Conventions:new service names must be stable low-cardinality logical component names; repeated NewService with same ServiceName intentionally returns the existing service
|Conventions:resource attributes must keep service.name, service.namespace, service.instance.id, service.version, deployment.environment, and service.category semantics stable
|Conventions:keep IHandler/IService small and lifecycle-oriented; extend additively only when a real caller needs new provider/exporter behavior
|Conventions:lower-level tracing setup returns wrapped errors; service/cmd startup decides logging and process exit behavior
|Anti-patterns:do not create sdk trace providers/exporters directly in callers when tracing.G().NewService can supply one
|Anti-patterns:do not add or reintroduce globalTraceSampleRate or any flat/global compatibility path; nested tracing.globalService is the only runtime config surface
|Anti-patterns:do not use OTel global state to bypass service-specific provider injection or per-service sample-rate policy
|Anti-patterns:do not reintroduce ParentBased for per-service sampler; keep sdkTrace.TraceIDRatioBased(config.SampleRate) authoritative for explicit per-service providers
|Anti-patterns:do not move business span names, domain attributes, router policy, storage policy, or request-specific tracing strategy into pkg/tracing
|Anti-patterns:do not add trace-id/span-id, request IDs, user IDs, host IDs, or other high-cardinality values as process profiling tags
|Anti-patterns:do not add Pyroscope/trace-profile coupling or new profiling dependencies from this package without an explicit product/architecture decision
|Anti-patterns:do not add speculative exporters, config knobs, hot reload, or public interfaces before an actual service path requires them
|Verification:doc-only change:git diff --check -- pkg/tracing/AGENTS.md|rg '^##' pkg/tracing/AGENTS.md should return no matches|scan manually for code-fence markers
|Verification:tracing code change:go test ./pkg/tracing ./pkg/rest/server|go test ./internal/backend/service ./internal/file/service ./internal/application/service ./internal/relay/service
|Verification:global fallback change:confirm tracing.Init registers otel.SetTextMapPropagator and otel.SetTracerProvider while explicit callers still pass IService providers
|Verification:sampling change:confirm handler.go uses sdkTrace.TraceIDRatioBased(config.SampleRate) for per-service providers and keeps configured local rate authoritative
