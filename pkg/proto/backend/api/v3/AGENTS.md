# BACKEND API V3 KNOWLEDGE BASE

## OVERVIEW

`pkg/proto/backend/api/v3` contains the protocol layer for **NodeMgr Backend API v3**.

- Source of truth (proto definitions): `proto/backend/api/v3/*.proto`
- Generated artifacts (do not touch): `pkg/proto/backend/api/v3/*.pb.go`
- Handwritten supplements (the real work surface): `pkg/proto/backend/api/v3/*.go` (non-`*.pb.go`)

This package is responsible for:
- API wire models (request/response structures) generated from proto
- Validation and error organization on top of generated types
- Conversions between proto structs and `pkg/types`
- Default-value normalization for fields where proto zero-values are ambiguous

This package is NOT responsible for:
- Business flow logic
- Storage/DAO orchestration

## WHERE TO LOOK

- Proto contracts: `proto/backend/api/v3/*.proto`
- Generated Go protobuf files (do not read/modify): `pkg/proto/backend/api/v3/*.pb.go`
- Handwritten extensions (preferred reading/editing targets):
  - `common.go`, `constant.go` (shared helpers/constants)
  - `iam.go`, `globalsettings.go` (backend-specific protocol helpers)
  - `deploy_policy.go`, `schedule_workflow.go`, `sync.go` (domain converters/validators)
  - `node_agent.go`, `node_proxy.go`, `node_workflow.go`, `plugin.go`, `plugin_workflow.go`, `process.go`, `cipher.go`, `configpolicy.go` (per-domain helpers)
  - Other non-`*.pb.go` files in this directory are also handwritten supplements.

## CONVENTIONS

- **Never read or edit `*.pb.go` by hand.** They are generated from proto and will be overwritten.
- Prefer editing handwritten `.go` files only.
- Keep proto structs lifecycle confined to this package:
  - Convert proto -> `pkg/types` before business logic.
  - Convert `pkg/types` -> proto structs for responses.
- Avoid declaring brand-new parallel types in this package; extend generated structs via methods/helpers.
- Handle “semantic defaults” in conversion paths (commonly in `AutoConvert()`):
  - If `nil` vs `0` has different meaning, normalize explicitly instead of relying on proto getters.

## ANTI-PATTERNS

- Reading/editing generated `*.pb.go` as if it were business logic.
- Using proto structs directly in upper-layer business modules.
- Duplicating conversion/validation logic outside this package.
- Introducing new types that mirror generated messages without a strong reason.

## HOW TO CHANGE

- If you need to change request/response fields or RPC contracts:
  1) Edit proto: `proto/backend/api/v3/*.proto`
  2) Regenerate: `cd proto && make clean && make all`
  3) Update handwritten converters/validators in this directory to match the new contract.

## TESTING

- Basic compile/test: `go test ./pkg/proto/backend/api/v3`
- Focused regression checks (if related): `go test ./pkg/proto/backend/api/v3 -run 'TestDeployPolicy'`

