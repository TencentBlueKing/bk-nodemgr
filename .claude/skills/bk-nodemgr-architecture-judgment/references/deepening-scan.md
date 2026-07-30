# Deepening Scan

Use this reference only when the user explicitly asks for an architecture friction scan, deepening opportunity scan, shallow-module review, or broad maintainability scan. Do not run this workflow as a companion pre-flight or ordinary code review step.

The scan is candidate discovery, not implementation planning. It may write one temporary HTML report outside the repository, then must stop for user choice.

## Workflow

1. Establish the scan boundary: service, package, router group, workflow, or file cluster.
2. Read root `AGENTS.md` and all nearest scoped `AGENTS.md` files for the scanned paths.
3. Identify representative architecture anchors: router/handler entrypoints, service orchestration, storage/DAO, `pkg/types`, `pkg/proto`, proto files, docs, and tests when relevant.
4. Find 2-3 dominant local patterns before naming friction.
5. Look for concrete deepening opportunities across locality, seam placement, dependency direction, contract stability, package cohesion, converter/helper duplication, terminology, side effects, reversibility, and verification surface.
6. Reject candidates without direct repository evidence.
7. Apply the deletion test to suspected shallow modules: would deleting the module concentrate complexity, or merely move it?
8. Assign each retained candidate `Stop`, `Warn`, or `Continue` using the core skill matrix.
9. Generate one standalone HTML report under the OS temp directory.
10. Return the absolute report path, top recommendation, numbered choices, and ask which candidate the user wants to explore.

Stop there. Do not modify repository files, write ADRs, update domain notes, or start a refactor plan.

## Candidate Signals

Strong candidates usually have evidence like:

- Understanding one business concept requires jumping through many shallow modules.
- A `handler` derives business results that belong in `service` or domain logic.
- `storage` or DAO code owns business policy rather than persistence access.
- Proto structs cross the `pkg/proto` conversion boundary into business logic.
- A `pkg` helper, converter, interface, or adapter is service-specific or has only one real caller.
- Repeated converter/helper logic appears in parallel packages.
- Interface surface is almost as complex as the implementation it hides.
- Tests can only exercise fragments, not observable behavior through a stable interface.
- Local terminology differs across proto, `pkg/types`, API JSON, frontend, and docs.

Weak candidates should be rejected when they are just style preferences, local naming discomfort, isolated complexity with no cross-file cost, or future-proofing without current evidence.

## Dependency Categories

Use these categories to judge seam placement:

| Category            | Use for                                               | Scan implication                                                                                    |
| ------------------- | ----------------------------------------------------- | --------------------------------------------------------------------------------------------------- |
| In-process          | Pure calculation, converters, in-memory orchestration | Prefer deepening inside the owning module; do not add a public port.                                |
| Local-substitutable | Mongo/Redis paths using `testsuite/support`           | Test through local stand-ins when tests are in scope; keep seams internal unless callers need them. |
| Remote but owned    | bk-nodemgr or BlueKing-owned HTTP/gRPC/queue boundary | A port is justified only with real production and test adapters.                                    |
| True external       | Third-party or externally controlled service          | Hide SDK/API details behind injected adapter/mock.                                                  |

One adapter means a hypothetical seam. Two adapters means a real seam.

## HTML Report Contract

Write a single self-contained HTML file to the OS temp directory:

- Resolve temp dir from `$TMPDIR`, then `/tmp` on Linux/macOS, or `%TEMP%` on Windows.
- File name shape: `bk-nodemgr-architecture-review-<timestamp>.html`.
- Do not write the report inside the repository.
- Use CDN assets only when useful; no custom server or eval viewer.

The report must include:

- Repo and scan scope.
- Date/time and method.
- Applicable project constraints from `AGENTS.md` and scoped rules.
- Prioritized candidate cards.
- File and symbol evidence.
- Current architectural cost.
- Proposed direction, without implementation detail.
- Expected benefit in locality, leverage, seam placement, and test surface.
- Risk, blast radius, reversibility, and verification cost.
- Verdict and reason.
- Clear label: `candidate only, not approved work`.
- Numbered choice identifiers matching the chat response.

Candidate card fields:

- `Choice ID`
- `Files`
- `Problem`
- `Evidence`
- `Candidate direction`
- `Benefits`
- `Risks and reversibility`
- `Verification surface`
- `Verdict`
- `Recommendation strength`: `Strong`, `Worth exploring`, or `Speculative`

## Chat Response Contract

After writing the report, respond with:

```text
Mode: deepening scan
Verdict: Stop | Warn | Continue
Decision: candidate scan only; no approved implementation
Evidence: [scanned paths and anchor files]
Report: [absolute temp HTML path]
Top recommendation: [choice id and why]
Choices:
1. [candidate title]
2. [candidate title]
Required next action: Which candidate should we explore?
Side effects: temporary HTML report only
```

Do not include implementation steps, task breakdowns, or code edits. If the scan discovers an urgent `Stop`, still present it as a candidate/blocker and wait for user direction.

## After Selection

Once the user picks a candidate:

- Produce a candidate design brief using the core skill contract.
- Use `references/design-it-twice.md` only when materially different interface or seam designs are plausible.
- Ask before writing ADRs, domain notes, or any repository documentation.
- Treat any refactor as a separate plan or implementation request.
