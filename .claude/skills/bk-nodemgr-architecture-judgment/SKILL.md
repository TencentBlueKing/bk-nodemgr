---
name: bk-nodemgr-architecture-judgment
description: Use when making or reviewing bk-nodemgr architecture decisions, especially cross-layer router/service/storage/proto/pkg/types changes, new pkg helpers/interfaces/converters/adapters, proto/front contract changes, architecture drift self-review, boundary leak checks, or explicit deepening opportunity scans.
---

# bk-nodemgr Architecture Judgment

## Overview

Use this skill to make architecture judgment explicit before or after risky bk-nodemgr work. It is an advisory guardrail and review lens, not a refactor executor. It helps decide whether a change should stop, proceed with warnings, or continue through the existing project path.

This skill never authorizes code changes by itself. Implementation still requires a separate user request or approved plan.

## When to Use

Use this skill for:

- `pre-flight`: before adding or designing cross-layer work across `router`, `handler`, `service`, `storage`, `DAO`, `proto`, `pkg/types`, or `pkg/proto`.
- `post-flight`: after implementation, before PR handoff, or when reviewing a diff for architecture drift.
- `deepening scan`: when the user explicitly asks to scan for architecture friction, boundary leaks, shallow modules, or deepening opportunities.
- New `pkg` helpers, interfaces, converters, adapters, proto/front fields, or shared abstractions.
- Tenant, auth, permission, storage format, public API, or trust-boundary decisions.

## When Not to Use

Do not use this skill as the primary guide for:

- Local Go style, naming, lint, or error-handling issues with no architecture signal.
- Ordinary implementation that stays inside one layer and follows a dominant local pattern.
- Frontend visual/UI work.
- Git, release, changelog, or PR creation workflows unless architecture judgment is part of the review.
- Deepening scans that the user did not ask for.

When a narrow project skill owns the surface, load that skill too. This skill supplements scoped `AGENTS.md` and narrow project skills; it does not replace them.

## Evidence First

Before judging architecture, gather local evidence:

1. Read root `AGENTS.md` and nearest scoped `AGENTS.md` files for touched paths.
2. Find 2-3 representative implementations in the same service, layer, or package.
3. Identify the current data flow and ownership: `router -> service -> storage/DAO`, `pkg/types`, `pkg/proto`, and proto/front contract points.
4. Report when evidence is missing instead of inventing a rule.

## Mode Selection

| User intent or task signal                                                                   | Mode                   | Scope                                                       |
| -------------------------------------------------------------------------------------------- | ---------------------- | ----------------------------------------------------------- |
| Before a risky change, new endpoint, new shared helper, or uncertain layer placement         | `pre-flight`           | Lightweight reminder or design risk check                   |
| Review current diff, PR, completed implementation, boundary leak, or architecture drift      | `post-flight`          | Findings-first self-review                                  |
| Explicit request to scan a module or area for deepening opportunity or architecture friction | `deepening scan`       | Heavy scan with temp HTML candidate report                  |
| User already selected one deepening candidate and wants to explore interface alternatives    | Candidate design brief | Load `references/design-it-twice.md` if alternatives matter |

## Stop / Warn / Continue

Emit exactly one final verdict. Precedence is `Stop` over `Warn` over `Continue`.

| Verdict    | Use when                                                                                                                                                                                                                                                                                                                       | Required behavior                                                                                                         |
| ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------------------------------------- |
| `Stop`     | A hard `AGENTS.md` rule is violated; proto/front contract, public handler/interface, storage data format, tenant/auth/permission/trust boundary, irreversible action, or long-lived cross-component decision is materially unresolved; evidence is insufficient to choose safely; deepening-scan user choice is still pending. | Do not recommend or begin implementation. State the blocker, evidence, and one precise decision or fact needed to resume. |
| `Warn`     | The path is reversible and locally actionable, but has documented architecture risk: under-proven `pkg` helper/interface/converter/adapter, one-adapter hypothetical seam, terminology conflict, duplication risk, or justified drift.                                                                                         | Continue only with the warning, simplest safe path, and verification requirement recorded.                                |
| `Continue` | Evidence backs the decision, scoped instructions and dominant local patterns agree, and no unresolved architecture-level risk remains.                                                                                                                                                                                         | State why proceeding is safe and preserve implementation guardrails.                                                      |

Common hard stops in bk-nodemgr:

- Adding proto/front fields before endpoint semantics are clear.
- Putting service-specific logic into `pkg`.
- Passing proto structs through business logic instead of converting at `pkg/proto` boundaries.
- Creating a public interface, adapter, or shared helper for one current implementation.
- Mutating stable handler, interface, or proto fields instead of additive extension.
- Adding speculative workers, Helm knobs, feature flags, or shared abstractions for future needs.

## Output Contracts

Every response from this skill includes:

- `Mode`
- `Verdict`
- `Decision`
- `Evidence`
- `Required next action`
- `Side effects`

### Architecture Pre-flight

Use for work-before reminders and design risk checks. Keep companion-trigger output light.

Required shape:

```text
Mode: pre-flight
Verdict: Stop | Warn | Continue
Risk level: low | medium | high
Decision: [architecture decision being made]
Evidence: [AGENTS/scoped rules and local patterns checked]
Watch points:
- [3-5 concrete risks or guardrails]
Blocking question: none | [one precise question]
Recommended path: [simplest safe path]
Side effects: none
```

Ask at most one blocking question. Put additional concerns in `Watch points`.

### Architecture Self-review

Use for completed work, current diffs, PR prep, and architecture drift checks. Lead with findings.

Drift taxonomy:

- `Contract drift`: API, storage, proto, return, error, or behavior semantics changed.
- `Boundary/layer drift`: responsibility moved to the wrong `handler`, `service`, `storage`, converter, `internal`, or `pkg` layer.
- `Dependency drift`: dependency direction, interface seam, adapter, or cross-service coupling changed.
- `Scope/YAGNI drift`: speculative fields, abstractions, options, workers, or unrelated cleanup appeared.
- `Terminology/model drift`: established domain terms or model ownership changed inconsistently.
- `Side-effect drift`: writes, external I/O, background work, or irreversible actions appeared outside the approved design.
- `Verification drift`: promised checks were skipped or no longer prove the intended contract.

Required shape:

```text
Mode: post-flight
Verdict: Stop | Warn | Continue
Decision: [whether implementation matches intended architecture]
Findings: [severity-ordered; say none if none]
Drift risks: locality / seam / proto boundary / pkg ownership / test surface
Evidence: [diff, AGENTS/scoped rules, local references]
Required follow-up: none | ask user | adjust implementation
Side effects: none
```

Map `Stop` to blocking/serious review findings and `Warn` to important findings. `Continue` means no architecture-judgment finding; it does not suppress ordinary code-review findings.

### Architecture Deepening Candidates

Use only when the user explicitly asks for a scan. Load `references/deepening-scan.md`.

Required shape:

```text
Mode: deepening scan
Verdict: Stop | Warn | Continue
Decision: candidate scan only; no approved implementation
Evidence: [scanned scope and representative files]
Report: [absolute temp HTML path]
Top recommendation: [candidate id and why]
Choices: [numbered candidates matching the report]
Required next action: Which candidate should we explore?
Side effects: temporary HTML report only
```

Do not present implementation steps as approved work. Stop at user choice.

### Candidate Design Brief

Use only after the user selects a candidate. If materially different interface designs are plausible, load `references/design-it-twice.md`.

Required shape:

```text
Mode: candidate design brief
Verdict: Stop | Warn | Continue
Decision: [chosen candidate and design question]
Evidence: [candidate report and local code references]
Constraints: [contracts, dependencies, AGENTS rules]
Recommended path: [if enough evidence exists]
Required next action: [continue grilling, run design-it-twice, or ask one blocker]
Side effects: none unless user approves ADR/domain note
```

## Dependency Categories

Use dependency category to judge seam placement and testing surface.

| Category            | bk-nodemgr examples                                                   | Judgment                                                                                                    |
| ------------------- | --------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| In-process          | Pure helpers, converters, domain orchestration, in-memory calculation | Usually deepenable without a public port. Test through the owning module interface when tests are in scope. |
| Local-substitutable | Mongo/Redis paths covered by `testsuite/support`                      | Can use real local stand-ins. Keep seams internal unless callers need them.                                 |
| Remote but owned    | bk-nodemgr or BlueKing-owned HTTP/gRPC/queue boundaries               | Define a port only when production and test adapters are both justified.                                    |
| True external       | Third-party systems or SDKs not owned here                            | Inject the external dependency behind an adapter/mock; do not leak SDK types into business code.            |

One adapter means a hypothetical seam. Two adapters means a real seam.

## Terminology

Use architecture vocabulary for judgment:

- `module`, `interface`, `implementation`, `depth`, `seam`, `adapter`, `leverage`, `locality`, `shallow module`, `deletion test`.

Use bk-nodemgr vocabulary for code ownership:

- `router`, `handler`, `service`, `storage`, `DAO`, `pkg/types`, `pkg/proto`, `proto boundary`, `converter`, `contextx`, `tenant/user`, `permission action/resource`.

Do not rename project concepts into upstream architecture vocabulary. If terms conflict across `AGENTS.md`, proto, `pkg/types`, API JSON, frontend, and docs, report the conflict as `Warn` and use the dominant local term.

## Side-effect Policy

- `pre-flight`: read, search, and analyze only.
- `post-flight`: read, search, analyze, and run read-only inspection commands only.
- `deepening scan`: the only allowed write is one standalone HTML report under the OS temp directory.
- Do not edit repository files, mutate Git history, install dependencies, publish, start persistent services, auto-repair code, or transition into implementation.
- ADR or domain-model notes are suggestions until the user explicitly approves writing them.

Ask before writing persistent decision documents. A good prompt is: `Want me to record this as an ADR/domain note?`

## Companion Integration

- `bk-nodemgr-how-to`: load this skill for cross-layer placement, competing designs, dependency direction, shared-contract impact, or architecture drift signals, then load the narrower project skill for the affected surface.
- `code-review`: use `post-flight` only when a diff affects architecture boundaries, dependency direction, shared contracts, or departs from the intended design.
- `api-scaffold`: may use `pre-flight` before scaffolding high-risk endpoints, but should not duplicate this skill's matrix.
- Ordinary companion usage must stay lightweight and must not generate a deepening HTML report.

## References

- `references/deepening-scan.md`: explicit deepening scan workflow and temp HTML report contract.
- `references/design-it-twice.md`: alternative interface design protocol after a candidate is selected.

## Common Mistakes

| Mistake                                             | Better move                                                                |
| --------------------------------------------------- | -------------------------------------------------------------------------- |
| Treating this skill as permission to refactor       | Return a verdict and wait for explicit implementation approval.            |
| Turning every risky task into a deepening scan      | Use lightweight `pre-flight` unless the user explicitly asks for a scan.   |
| Adding a `pkg` helper because one endpoint needs it | Keep logic in the owning package until real cross-package reuse is proven. |
| Creating an interface for one implementation        | Wait for a second real adapter or a test adapter that proves the seam.     |
| Reporting many blockers                             | Ask one precise blocking question and put the rest in watch points.        |
| Letting proto structs pass through business layers  | Convert at `pkg/proto` boundaries and keep business logic on domain types. |

## Validation Prompts

Use the approved RED prompts as the first GREEN validation set:

1. Cross-layer backend API with possible `pkg` helper: expect `pre-flight`, one verdict, no edits.
2. Current diff architecture drift review: expect `post-flight`, drift taxonomy, required follow-up.
3. `dpmgr` deepening opportunity scan: expect temp HTML candidate report, top recommendation, user choice, no repository edits.
