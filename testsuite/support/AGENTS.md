|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:testsuite/support
|Overview:Integration-test-only helper package for package-level Go tests; owns Mongo/Redis source selection, Docker startup, client creation, isolation, and cleanup
|Overview:Compiled public helper APIs require `//go:build integration`; doc.go stays untagged so default `go test ./testsuite/support` sees a package
|Structure:testsuite/support:{doc.go,env.go,source.go,docker.go,mongo.go,redis.go,name.go}
|Public API:Mongo:{RequireMongoClient,RequireMongoDatabase}|Redis:{RequireRedisClient,RequireRedisClientWithKeyPrefix}
|Where to look:source selection:source.go|auto→env when testsuite/integration.env exists, otherwise docker; docker can still read image/network from env file
|Where to look:config/env parsing:env.go|process env wins over testsuite/integration.env; required DB fields fail fast via testing.TB
|Where to look:Docker network handling:docker.go|bridge uses testcontainers modules; host uses Docker SDK with NetworkMode=host and no PortBindings
|Where to look:Mongo helper:mongo.go|default image docker.io/bitnamilegacy/mongodb:6.0.10-debian-11-r8|override NODEMGR_TEST_MONGO_IMAGE
|Where to look:Redis helper:redis.go|default image docker.io/bitnamilegacy/redis:7.0.12-debian-11-r34|override NODEMGR_TEST_REDIS_IMAGE|ALLOW_EMPTY_PASSWORD=yes
|Dependencies:test-only packages:{testcontainers-go,mongodb module,redis module,Docker SDK,mongo-driver,go-redis,godotenv,uuid}
|Conventions:Keep helper signatures testing.TB-first and call t.Helper(); fail setup via t.Fatalf with operation context instead of returning half-ready clients
|Conventions:Do not mutate process env from testsuite/integration.env; read values into a map and resolve by configValue(process env→env file→empty)
|Conventions:All env keys must be exported constants and start with `NODEMGR_TEST_`; update testsuite/integration.env.example when adding any key
|Conventions:Env source validates only the requested service; missing Redis keys must not break Mongo-only tests, and missing Mongo keys must not break Redis-only tests
|Conventions:Docker source must use chart-aligned default images unless concrete local compatibility evidence is documented; image overrides stay env/file config
|Conventions:Bridge mode may use testcontainers module ConnectionString; host mode connects fixed localhost ports:{Mongo 127.0.0.1:27017,Redis 127.0.0.1:6379}
|Conventions:Host-network Docker runs are not parallel-safe per service port; document/serialize real integration packages that use fixed ports
|Conventions:Cleanup is mandatory:Mongo drops generated database and disconnects; Redis scans/deletes generated prefix and closes client; Docker removes containers
|Conventions:Generated names use uniqueName(prefix); callers receive a database or key prefix and must use it for all fixture data
|Anti-patterns:no production imports of testsuite/support|no helper use outside `//go:build integration` tests|no fallback from broken env source to docker
|Anti-patterns:no bare MONGO_/REDIS_ env keys|no package-global shared clients/databases/key prefixes|no sync.Once fixture data against mutable integration services
|Anti-patterns:no PortBindings in host-network Docker path|no `docker run` shelling out from helpers|no sleeps without readiness/connection checks
|Anti-patterns:no new database/service helpers until a real package-level integration test needs them; keep this package Mongo/Redis-focused until another service is proven
|Related:testsuite/AGENTS.md|pkg/dao/mongo/host/handler_test.go|pkg/redsync/handler_test.go|pkg/rediscache/redis_test.go|install/helm/bk-nodemgr/charts/{mongodb,redis}
|Commands:compile default=GOTOOLCHAIN=local go test ./testsuite/support
|Commands:compile integration=GOTOOLCHAIN=local go test -tags=integration ./testsuite/support
|Commands:smoke host-network=set DOCKER_HOST + NODEMGR_TEST_ENV_SOURCE=docker + NODEMGR_TEST_DOCKER_NETWORK=host; run Mongo/Redis packages serially
