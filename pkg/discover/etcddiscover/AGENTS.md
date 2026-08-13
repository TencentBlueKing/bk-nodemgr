|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:pkg/discover/etcddiscover
|Overview:etcd-backed `discover.IProvider` for registration, discovery, lease keepalive, watch/list cache sync, TLS, and etcd-log bridging
|Boundary:owns all etcd-specific client/TLS/lease/watch/cache/log adaptation|callers own service-specific `discover.Instance` payload production and consumption
|Structure:pkg/discover/etcddiscover:{README.md,etcd.go,etcd_test.go,etcd_logger_test.go,AGENTS.md}
|Where to look:package intent/boundary:README.md:etcd discovery purpose and “all etcd logic stays here” rule
|Where to look:provider implementation:etcd.go:`ProviderEtcd`,`NewProviderEtcd`,`WithWatch`,`WithDiscoverPathPrefix`,`Start`,`GracefulShutdown`
|Where to look:registration lifecycle:etcd.go:`Register`→`register`→lease `Grant`→`Put` with lease→`KeepAlive`→local cache|`Update`/`Deregister` require `metaKeyLeaseID`
|Where to look:read sync path:etcd.go:`startWatching`/`watch` event sync + `keepListing`/`list` periodic full sync→`cacheInstances`→query APIs
|Where to look:query APIs:etcd.go:`GetAllService`,`GetAllEndpoint`,`GetEndpoint`,`SelectEndpoints`:read provider-managed cache and delegate endpoint selection to `discover.SelectEndpoints`
|Where to look:etcd client config/logging:etcd.go:`initTLS`,`newEtcdClientConfig`,`etcdLoggerCore`,`mapEtcdLogLevel`,`etcdEntryKeyValues`
|Where to look:behavior tests:etcd_test.go:integration-style provider CRUD/query/select/shutdown tests require `.env` with `ETCD_ENDPOINT`
|Where to look:logger tests:etcd_logger_test.go:etcd zap core mapping, deterministic key-value extraction, config logger injection
|Conventions:implement parent `pkg/discover` contracts exactly; public surface remains `discover.Instance`/`discover.Endpoint`/`discover.ServiceName`/`discover.EndpointName`
|Conventions:keep etcd raw types private to this package except internal `clientv3` use; never leak `clientv3` responses through `IProvider`
|Conventions:preserve `metaKeyLeaseID` whenever replacing registered instances; lease id is required for update put-with-lease and revoke flows
|Conventions:use `config.Etcd` + `initTLS()` for endpoints/auth/TLS; keep defaults local (`defaultEtcdPrefix`,`defaultEtcdDialTimeout`,`defaultEtcdLeaseTTLSec`,`defaultListTickTime`)
|Conventions:watch/list cache is the read model; query methods should not perform ad-hoc etcd reads outside the provider sync path
|Conventions:use `pkg/logger` for project logs; etcd internal zap logs must route through `newEtcdClientConfig()`/`etcdLoggerCore`
|Conventions:shutdown order matters: deregister local instances before cancel because `Deregister` uses `provider.ctx` to revoke leases
|Anti-patterns:do not add service-specific routing, weighting, health, or metadata policy here; callers own payload semantics beyond `discover.Instance`
|Anti-patterns:do not hardcode endpoints/credentials/TLS paths outside `config.Etcd`; do not bypass `initTLS()` for TLS setup
|Anti-patterns:do not write directly to etcd without updating in-memory `localInstances` and preserving lease metadata
|Anti-patterns:do not duplicate root `pkg/discover` selector/error/model logic; extend root contracts first when behavior is generic
|Dependencies:internal project:{pkg/config,pkg/discover,pkg/logger,pkg/runtime/conv,pkg/runtime/ssl}|external:{go.etcd.io/etcd/client/v3,go.uber.org/zap}
|Commands:targeted tests need reachable etcd + `.env`:`go test ./pkg/discover/etcddiscover`|logger-only tests can be filtered with `-run 'Test.*Logger|TestEtcdEntry|TestMapEtcd'`
