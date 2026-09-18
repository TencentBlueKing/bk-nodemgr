# Design It Twice

Use this reference after the user selects a deepening candidate, or during `pre-flight` when there are materially different valid decompositions and choosing wrong would create long-lived cost.

Design it twice means comparing designs. It does not mean implementing twice, adding speculative abstractions, or running a second implementation phase.

## When to Load

Load this reference when one or more are true:

- Ownership or layer placement is genuinely uncertain.
- Two or more designs can satisfy the current requirement through different seams.
- A choice affects proto/front contracts, `pkg/types`, public interfaces, storage format, tenant/auth/permission behavior, or cross-service dependency direction.
- A selected deepening candidate needs a new module interface.
- The first design idea is attractive but expensive to reverse.

Do not load it for local implementation details, naming-only alternatives, or one obvious existing project pattern.

## Frame the Problem Space

Before comparing designs, state:

- The selected candidate or decision at stake.
- Applicable `AGENTS.md` and scoped rules.
- Dominant local patterns found in the same service/layer.
- Current contracts that must not silently change.
- Dependency category: in-process, local-substitutable, remote but owned, or true external.
- What sits behind the proposed seam.
- Verification surface and rollback cost.

This frame is evidence, not a proposal.

## Produce Two Real Designs

Create at least two independently viable designs at comparable detail. They must differ in seam placement, ownership, or interface shape, not just naming or minor implementation mechanics.

Useful design constraints:

- Minimal interface: 1-3 entry points with maximum leverage.
- Caller-centric: make the common caller trivial while keeping internals hidden.
- Ownership-first: keep business logic in the owning service/package and delay `pkg` exposure.
- Adapter-first: use a port only when production and test adapters are both justified.

Each design must include:

- Interface shape: types, methods, params, invariants, ordering, and error modes when relevant.
- Usage example from the likely caller.
- Hidden implementation behind the seam.
- Dependency and adapter strategy.
- Contract and data-model impact.
- Verification path.
- Trade-offs in depth, locality, leverage, and reversibility.

## Compare Designs

Compare every design using the same dimensions:

| Dimension                          | Question                                                                                                 |
| ---------------------------------- | -------------------------------------------------------------------------------------------------------- |
| Local pattern reuse                | Does it follow 2-3 representative bk-nodemgr implementations?                                            |
| Responsibility and layer ownership | Does logic sit in the right `handler`, `service`, `storage`, DAO, converter, `internal`, or `pkg` layer? |
| Dependency direction               | Does high-level code depend on stable abstractions instead of concrete infrastructure?                   |
| Contract and data-model impact     | Does it avoid unnecessary proto/front/API/storage contract change?                                       |
| Terminology consistency            | Does it reuse dominant domain terms?                                                                     |
| Scope and side effects             | Does it avoid speculative fields, options, workers, and unrelated cleanup?                               |
| Reversibility and rollback cost    | Can it be changed later without breaking public consumers or persisted data?                             |
| Verification cost                  | Can behavior be proven through a stable interface?                                                       |

Prefer the smallest design that satisfies current requirements through existing paths. A smaller design is not better if it hides a public contract decision, weakens a trust boundary, or forces duplicated future changes.

## Output Contract

```text
Mode: candidate design brief
Verdict: Stop | Warn | Continue
Decision: [recommended design]
Evidence: [candidate, local patterns, constraints]
Design A: [interface, usage, hidden implementation, dependency strategy, trade-offs]
Design B: [interface, usage, hidden implementation, dependency strategy, trade-offs]
Comparison: [same dimensions for all designs]
Recommendation: [strong opinion or hybrid]
Required next action: [none | one blocking question | implementation plan request]
Side effects: none
```

Use `Stop` when no safe recommendation can be made without a missing product/API/security/data decision. Use `Warn` when one design is acceptable but carries explicit architecture debt. Use `Continue` when evidence supports a design and guardrails are clear.

## Common Mistakes

| Mistake                                            | Better move                                                                |
| -------------------------------------------------- | -------------------------------------------------------------------------- |
| Comparing two names for the same design            | Change seam placement, ownership, or interface shape.                      |
| Designing an interface for a single implementation | Keep it private unless a second adapter or test adapter proves the seam.   |
| Picking the flexible design by default             | Prefer current requirement leverage and reversibility over future options. |
| Hiding proto/front contract impact                 | Treat public contract questions as `Stop`.                                 |
| Turning recommendation into implementation         | Ask for an implementation plan or explicit go-ahead first.                 |
