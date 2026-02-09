# TEST KNOWLEDGE BASE

## OVERVIEW

`test/` contains integration/API test harnesses and a dedicated mock server for multi-service validation.

## STRUCTURE

```
test/
|- cases/               # API test cases by service and route hierarchy
|- mock-server/         # mock service used by test flows
|- helper/              # shared test helpers
|- Makefile             # build/test/clean targets
`- test.go              # test flags and endpoint configuration
```

## WHERE TO LOOK

| Task | Location | Notes |
|------|----------|-------|
| Service API test behavior | `test/cases/**` | Organized by backend/application/file and route levels |
| Mock integration behavior | `test/mock-server/**` | CMDB/BKRepo simulation endpoints |
| Test startup flags | `test/test.go` | Backend/application/file endpoint flags |
| Test execution commands | `test/Makefile`, `test/README.md` | Build and run flow |

## CONVENTIONS

- Keep integration tests independent; avoid order dependence.
- Use generated/randomized test data suffixes to reduce collisions.
- Prefer clear assertion errors for faster triage.

## ANTI-PATTERNS

- Do not move integration tests into service runtime directories.
- Do not hardcode target endpoints when flags/config are available.
- Do not couple tests to local mutable state in `test/build` artifacts.
