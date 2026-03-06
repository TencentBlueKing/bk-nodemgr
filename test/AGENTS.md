# TEST KNOWLEDGE BASE

## OVERVIEW

`test/` contains integration/API test harnesses and a dedicated mock server for multi-service validation.

## STRUCTURE

```
test/
|- cases/               # API test cases by service and route hierarchy
|  ├- precheck/         # environment verification (runs first, organized by domain: cmdb/, gse/, ...)
|  ├- backend/          # backend service API tests
|  ├- application/      # application service API tests
|  └- file/             # file service API tests
|- data/                # preset set data
|- tools/               # CI helper scripts (e.g. gen-mock-config.sh)
|- mock-server/         # mock service used by test flows
|- helper/              # shared test helpers
|- Makefile             # build/test/clean targets with group ordering
`- config.go            # test flags and environment configuration
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Test cases | `test/cases/**` | Includes precheck and service API tests organized by route levels |
| Preset test data | `test/data/cmdb.yaml` | Shared by mock-server and precheck tests |
| Mock server config generation | `test/tools/gen-mock-config.sh` | Merges base config with `data/cmdb.yaml` into mock-server config |
| Mock integration behavior | `test/mock-server/**` | CMDB/BKRepo simulation endpoints |
| Test startup flags | `test/config.go` | Environment configuration through `--env-file` (`env.yaml` by default) |
| Test execution commands | `test/Makefile`, `test/README.md` | `TEST_GROUPS` controls execution order |

> Test execution requires pre-existing `env.yaml` ; repository does not ship this file.

## CONVENTIONS

- Keep integration tests independent; avoid order dependence within the same group.
- Execution order across groups is controlled by `TEST_GROUPS` in `Makefile`; register new groups there.
- Preset data lives in `data/` directory; use helper functions like `helper.LoadXXXMockData()` to load.
- Preset data is append-only: do not modify or delete existing entries; only append new ones (bug fixes allowed with review).
- Use generated/randomized test data suffixes to reduce collisions.
- Use `require` for preconditions, `assert` for value comparisons; keep assertions concise.

## ANTI-PATTERNS

- Do not move integration tests into service runtime directories.
- Do not hardcode target endpoints when flags/config are available.
- Do not hardcode test data that should come from `data/` directory.
- Do not add test groups without registering them in `Makefile` `TEST_GROUPS`.
- Do not couple tests to local mutable state in `test/build` artifacts.
