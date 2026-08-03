---
name: bk-nodemgr-naming-boundary
description: Use when adding, changing, or reviewing bk-nodemgr API/RPC/proto fields, pkg/types fields, DAO options/BSON terms, router/service/storage helpers, module names, cross-layer terms, naming drift, or local variable/helper names that reveal boundary smells like data, info, helper, manager, or process.
---

# bk-nodemgr naming boundary

## Overview

Naming is a boundary declaration. A name claims ownership, lifecycle, and cross-layer meaning, so naming review in bk-nodemgr must prove the term belongs to the layer where it appears and still maps cleanly across neighboring contracts.

This skill standardizes naming-boundary reports. It complements existing bk-nodemgr skills; it doesn't replace API scaffolding, architecture review, Go style, generic Go naming, glossary persistence, or permission matrix work.

## When to Use

Use this skill for naming decisions that can affect a boundary:

- Public API, RPC, proto, API JSON, `pkg/types`, DAO option, BSON, frontend, or docs terms.
- Router, service, storage, converter, module, or helper names that imply ownership or lifecycle.
- Cross-layer drift where the same concept appears under different names.
- Local variable or helper names only when names such as `data`, `info`, `helper`, `manager`, or `process` hide mixed responsibility or wrong layer ownership.

Do not use this skill for every ordinary local variable. Use `golang-naming` for generic Go identifier rules when no bk-nodemgr boundary smell exists.

## Evidence Floor

Before returning `Continue`, gather and state this evidence:

1. Root `AGENTS.md` plus the nearest scoped `AGENTS.md` for touched paths.
2. Existing term search in `pkg/types`, `proto/**`, and the same-service router, service, and storage code.
3. `front/src` and docs search when the term reaches frontend or documented API behavior.
4. At least one analogous implementation in the same layer.
5. Explicit mapping across proto, `pkg/types`, API JSON, DAO BSON, frontend, or docs whenever those layers apply.

If any required evidence is missing, don't return `Continue`.

## Verdicts

- `Stop`: public contract, cross-layer, or data-model naming lacks evidence, conflicts with existing terms, or has unclear boundary ownership.
- `Warn`: local or internal helper naming exposes mixed responsibility or mild terminology drift, but the issue is reversible.
- `Continue`: evidence, boundary owner, and cross-layer mapping are clear.

Emit exactly one final verdict. Prefer `Stop` over `Warn`, and `Warn` over `Continue`.

## Naming Boundary Review

Use this exact report shape:

```text
Naming Boundary Review

Existing Terms
- [terms found, where they live, and whether they match]

Boundary
- [owner layer, lifecycle claim, and whether the name belongs there]

Cross-layer Alignment
- [proto / pkg/types / API JSON / DAO BSON / frontend / docs mapping, or not applicable]

Questions
- [blocking or non-blocking questions, or none]

Recommendation
- [keep, rename, split, search more, or hand off]

Verdict: Stop | Warn | Continue
```

Keep the report short. If the naming question expands into broader design, hand off instead of embedding another skill's framework.

## Companion and Hand-off Rules

- `api-scaffold` owns endpoint scaffolding, proto generation, converter scaffolding, routes, and handler stubs.
- `bk-nodemgr-architecture-judgment` owns full architecture tradeoffs, boundary drift, shared abstractions, and cross-component design risk.
- `bk-nodemgr-code-style` owns bk-nodemgr Go readability, control flow, comments, initialization, and local style consistency.
- `golang-naming` owns generic Go identifier rules when names don't reveal bk-nodemgr boundary concerns.
- `domain-modeling` owns glossary and ADR persistence after the term or decision is resolved.
- `router-permission-supplement` owns permission action and resource matrix naming.

## RED Baseline Lessons

The baseline behavior was often conservative and useful, but inconsistent. This skill exists to add:

- A stable `Naming Boundary Review` report with one formal verdict.
- A clear evidence floor before `Continue`.
- Explicit cross-layer mapping for public and persisted terms.
- A hand-off contract so naming review doesn't duplicate other project skills.

## Eval Notes

Eval prompts live in `evals/evals.json`. Keep run artifacts, timing, grading, benchmarks, and review files under `/tmp` unless the user explicitly asks otherwise.
