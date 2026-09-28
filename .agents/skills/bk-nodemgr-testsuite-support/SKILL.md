---
name: bk-nodemgr-testsuite-support
description: Use when adding package-level integration tests in bk-nodemgr that should use testsuite/support helpers, especially RequireMongoDatabase, RequireMongoClient, RequireRedisClient, RequireRedisClientWithKeyPrefix, NODEMGR_TEST_ENV_SOURCE, NODEMGR_TEST_DOCKER_NETWORK, or testsuite/integration.env.
---

# bk-nodemgr testsuite support

## Overview

Use this skill when a new bk-nodemgr package-level integration test needs the project-owned `testsuite/support` helpers. The skill is about using the support layer correctly, not about broad test strategy or legacy test migration.

Current support coverage is MongoDB and Redis. Keep future helpers under the same rules only after a real package-level integration test needs them.

## When to Use

Use this skill for new tests that need one of these project surfaces:

- `testsuite/support`
- `support.RequireMongoDatabase(t)` or `support.RequireMongoClient(t)`
- `support.RequireRedisClientWithKeyPrefix(t)` or `support.RequireRedisClient(t)`
- `NODEMGR_TEST_ENV_SOURCE`, `NODEMGR_TEST_DOCKER_NETWORK`, `NODEMGR_TEST_MONGO_IMAGE`, or `NODEMGR_TEST_REDIS_IMAGE`
- `testsuite/integration.env` or `testsuite/integration.env.example`
- package-level MongoDB or Redis integration tests under `pkg/**` or `internal/**`

Do not use this skill for:

- Pure unit tests.
- Mongo driver `mtest` command-mock tests.
- Non-Mongo/Redis external service tests.
- Bulk migration of old `.env` tests; migration is handled separately.

## Source Anchors

Read these before adding or changing tests:

- `testsuite/AGENTS.md`: testsuite scope rules and commands.
- `testsuite/support/AGENTS.md`: helper API contracts, source selection, Docker network behavior, and anti-patterns.
- `testsuite/integration.env.example`: supported `NODEMGR_TEST_*` keys.
- `testsuite/support/{source,env,docker,mongo,redis}.go`: helper behavior.
- First Mongo pilot: `pkg/dao/mongo/host/handler_test.go`.
- First Redis pilots: `pkg/redsync/handler_test.go`, `pkg/rediscache/redis_test.go`.

## Core Decision Flow

1. Confirm the test genuinely needs a real MongoDB or Redis server.
2. Add `//go:build integration` to any test file that imports `testsuite/support`.
3. Use the service-specific helper instead of creating clients directly.
4. Keep test data isolated with the generated database or key prefix returned by the helper.
5. Verify both default behavior and integration-tag compile behavior.
6. Run a real env or Docker source smoke test when the environment is available.

Default `go test ./...` must never start Docker or connect to user-maintained integration services.

## Mongo Usage

Prefer `RequireMongoDatabase` for DAO/storage tests because it returns a generated isolated database and handles cleanup.

```go
//go:build integration

func testClient(t *testing.T) IHandler {
    t.Helper()

    _, db := support.RequireMongoDatabase(t)
    return New(db)
}
```

Use `RequireMongoClient` only when the test must control database selection itself. If you use it, still create isolated names and clean up any database or collection you create.

Mongo rules:

- Do not use package-level `sync.Once` fixture data against shared Mongo state.
- Do not read bare `MONGO_*` keys.
- Do not call `mongo.Connect` directly in package tests that should use this support layer.
- Keep fixture insertion and assertions on the same generated database.

## Redis Usage

Prefer `RequireRedisClientWithKeyPrefix` because Redis isolation is key-prefix based.

```go
//go:build integration

func testClient(t *testing.T) (Handler, string) {
    t.Helper()

    redisClient, keyPrefix := support.RequireRedisClientWithKeyPrefix(t)
    return New(redisClient), keyPrefix
}
```

Every key created by the test must include the returned prefix:

```go
mutex := h.NewMutex(keyPrefix + ":" + tt.name)
```

Use `RequireRedisClient` only when the test does not create keys or has its own explicit cleanup contract.

Redis rules:

- Do not hardcode `addr` or localhost clients in package tests.
- Do not read bare `REDIS_*` keys.
- Do not create unprefixed keys.
- Do not run host-network Redis packages in parallel; host mode uses fixed `127.0.0.1:6379`.

## Environment Sources

`testsuite/support` selects the integration environment through `NODEMGR_TEST_ENV_SOURCE`:

| Source   | Behavior                                                                            |
| -------- | ----------------------------------------------------------------------------------- |
| `auto`   | Use `testsuite/integration.env` when present; otherwise use Docker.                 |
| `env`    | Require `testsuite/integration.env`; missing or invalid service config fails.       |
| `docker` | Use Docker/testcontainers; may still read image and network settings from env file. |

Config precedence is process environment first, then `testsuite/integration.env`, then empty/default. Do not mutate process env from the file.

All new keys must start with `NODEMGR_TEST_` and must be documented in `testsuite/integration.env.example`.

## Docker Network Rules

`NODEMGR_TEST_DOCKER_NETWORK=bridge|host` controls container startup.

- `bridge`: use testcontainers modules and mapped ports.
- `host`: use Docker SDK with `NetworkMode=host` and no `PortBindings`; this is for Docker plugins that deny port publishing.

Host network mode uses fixed local ports:

- MongoDB: `127.0.0.1:27017`
- Redis: `127.0.0.1:6379`

Run host-network Mongo/Redis packages serially per service port.

Default images track the bundled chart app versions through legacy repository variants:

- MongoDB: `docker.io/bitnamilegacy/mongodb:6.0.10-debian-11-r8`
- Redis: `docker.io/bitnamilegacy/redis:7.0.12-debian-11-r34`

Use `NODEMGR_TEST_MONGO_IMAGE` or `NODEMGR_TEST_REDIS_IMAGE` only when the local/CI Docker environment requires an override.

## Go Test Command Rules

All validation commands in this repo should pin the local toolchain:

```bash
GOTOOLCHAIN=local go test ...
```

Use default tests to prove the package does not start Docker or connect to external services:

```bash
GOTOOLCHAIN=local go test ./<package>
```

Use the integration build tag for any test that imports `testsuite/support`:

```bash
GOTOOLCHAIN=local go test -tags=integration ./<package> ...
```

Host-network Docker source uses fixed service ports, so run Mongo packages serially with other Mongo packages and Redis packages serially with other Redis packages.

## Verification Gate

After adding or changing a test package that uses `testsuite/support`, run the default package test and the integration compile path:

```bash
GOTOOLCHAIN=local go test ./<package>
GOTOOLCHAIN=local go test -tags=integration ./<package> -run '^$' -count=1
```

When an integration environment is available, run at least one real test path.

Env source:

```bash
GOTOOLCHAIN=local NODEMGR_TEST_ENV_SOURCE=env go test -tags=integration ./<package> -run '<TestName>' -count=1 -v
```

Docker host-network source:

```bash
GOTOOLCHAIN=local DOCKER_HOST=<docker-host> NODEMGR_TEST_ENV_SOURCE=docker NODEMGR_TEST_DOCKER_NETWORK=host go test -tags=integration ./<package> -run '<TestName>' -count=1 -v
```

For host-network Docker, do not run Mongo packages or Redis packages in parallel on the same fixed service port.

## Common Mistakes

| Mistake                                                          | Correction                                                                         |
| ---------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| Importing `testsuite/support` from a non-`integration` test file | Add `//go:build integration` or do not use the helper.                             |
| Creating direct Mongo/Redis clients in package tests             | Use the support helper so source selection and cleanup stay consistent.            |
| Reusing package-global DB/client/key state                       | Create state per test via the helper; keep fixture data isolated.                  |
| Falling back from broken env source to Docker                    | Fail fast; an existing but invalid env file means user-maintained config is wrong. |
| Using Redis without the generated key prefix                     | Prefix every Redis key and let cleanup remove only that test's keys.               |
| Adding helper support for a future service                       | Wait until a real package-level integration test needs it.                         |

## Lightweight Eval Scenarios

Use these prompts to sanity-check future edits to this skill:

1. `我要给 pkg/dao/mongo/business 新增一个真实 Mongo DAO integration test，应该怎么接入 testsuite/support？`
2. `我要给 pkg/redsync 新增一个会写 Redis key 的测试，怎么避免 key 污染？`
3. `我要在 hbm 禁止端口发布的 Docker 环境跑 Redis integration tests，应该怎么配置和验证？`

A good answer should load this skill, cite the relevant scoped `AGENTS.md`, choose the right helper, preserve `integration` build tags, explain isolation, and give the default plus integration-tag verification commands.
