|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow references/AGENTS-compression-guide.md (pipe-index format, concise, no prose/code blocks)
|Scope:testsuite
|Overview:Package-level Go integration test support for bk-nodemgr; owns reusable test fixtures and local integration environment examples
|Overview:This is NOT legacy `test/`; do not use `test/Makefile`, `test/helper`, or `test/mock-server` as patterns for this scope
|Structure:testsuite:{integration.env.example,support/AGENTS.md,support/*.go}
|Where to look:local env example:testsuite/integration.env.example|copy to ignored testsuite/integration.env for user-maintained Mongo/Redis/CMDB
|Where to look:helper APIs:testsuite/support:{RequireMongoDatabase,RequireMongoClient,RequireRedisClientWithKeyPrefix,RequireRedisClient,RequireCMDBTarget}
|Where to look:first Mongo pilot:pkg/dao/mongo/host/handler_test.go|integration-tag DAO test using support.RequireMongoDatabase
|Where to look:first Redis pilots:pkg/redsync/handler_test.go|pkg/rediscache/redis_test.go|integration-tag Redis tests using support.RequireRedisClientWithKeyPrefix
|Dependencies:Go tests only|testcontainers-go for Docker source|mongo-driver/go-redis for clients|Docker SDK only for host-network containers without PortBindings
|Conventions:All real service/database tests using this scope must be `//go:build integration`; default `go test ./...` must not start Docker or connect external services
|Conventions:Use `testsuite/support`; do not place test-only helpers in `pkg/` or business `internal/`; production code must not import this subtree
|Conventions:Environment variables must use `NODEMGR_TEST_*`; never add bare `MONGO_*`, `REDIS_*`, or project-agnostic test env keys here
|Conventions:`NODEMGR_TEST_ENV_SOURCE=auto|env|docker`; auto uses testsuite/integration.env if present, otherwise Docker; existing but invalid env must fail, not fallback
|Conventions:Docker network mode is `NODEMGR_TEST_DOCKER_NETWORK=bridge|host`; host mode is for Docker plugins that deny port publishing and must be run serially per fixed service port
|Conventions:Mongo tests get isolated generated databases and clean them in `t.Cleanup`; Redis tests get generated key prefixes and clean prefix keys in `t.Cleanup`; CMDB tests get a ready apiserver target config and construct package-specific clients plus standalone legacy headers outside support
|Conventions:Default Docker images track bundled chart app versions or verified upstream standalone images:{Mongo docker.io/bitnamilegacy/mongodb:6.0.10-debian-11-r8,Redis docker.io/bitnamilegacy/redis:7.0.12-debian-11-r34,CMDB ccr.ccs.tencentyun.com/bk.io/cmdb-standalone@sha256:1987ff3dcf56debc6961d16e5669404c02816e7e31e7f7711fb3e054a1f06302}
|Conventions:CMDB env source must use disposable/non-production apiserver origins only; CMDB Docker bridge binds standalone to 0.0.0.0 by default when NODEMGR_TEST_CMDB_HOST_IP is empty
|Anti-patterns:no legacy `test/` coupling|no shared global DB/key state|no package-level `sync.Once` fixtures against shared services|no hardcoded `addr` Redis clients
|Anti-patterns:no committing testsuite/integration.env|no weakening failing integration assertions to fit helper behavior|no speculative services beyond current Mongo/Redis/CMDB use
|Anti-patterns:no parallel Docker host-network Mongo/Redis/CMDB package runs on the same fixed port; prefer env source or bridge for parallelism
|Child AGENTS:testsuite/support/AGENTS.md
|Commands:default support compile=GOTOOLCHAIN=local go test ./testsuite/support
|Commands:integration support compile=GOTOOLCHAIN=local go test -tags=integration ./testsuite/support
|Commands:host Mongo pilot compile=GOTOOLCHAIN=local go test -tags=integration ./pkg/dao/mongo/host -run '^$' -count=1
|Commands:Redis pilots compile=GOTOOLCHAIN=local go test -tags=integration ./pkg/redsync ./pkg/rediscache -run '^$' -count=1
