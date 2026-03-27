---
name: router-permission-supplement
description: Use when users want help choosing permission actions and resources for each handler in one router group, especially when they mention 补权限, 鉴权, action/resource mapping, permission helper rules, or want a final handler-level permission matrix.
---

# Router Permission Supplement

## Overview

Use this skill to narrow a router permission task down to its core decision:
- for each handler, choose the correct `action`
- for each handler, choose the correct `resource`
- output the final permission matrix

This skill is **mapping-first**.
Its job is to help the user produce a router-scoped permission table, not to own implementation planning, completion review, or API documentation workflows.

## When to Use

Use this skill when the user wants to:
- decide which handlers in a router group need permission checks
- choose the right `auth.ActionXxx` for each handler
- choose the right resource helper / resource type for each handler
- confirm router-local permission helper rules such as `buildBizResources(...)`
- review whether a proposed handler-to-action/resource mapping makes sense
- generate a final permission matrix before coding

Do **not** use this skill when:
- the task is unrelated to router/API authorization
- the user wants implementation planning, task breakdown, or rollout planning
- the user wants completion auditing, code verification, or API doc follow-up
- the task is only about IAM provider internals with no router handler mapping question

## Core Principle

Treat router permission supplementation as a **handler mapping problem**:

1. choose one router group
2. enumerate its handlers
3. decide which handlers need auth
4. choose the action for each protected handler
5. choose the resource helper for each protected handler
6. output the final permission matrix

If the router uses a router-specific resource pattern, prefer router-local helper rules over broad abstractions.

## Required Inputs

Always gather these before producing the final answer:

1. **Target service**
   - `internal/backend/router/**`
   - `internal/application/router/**`
   - `internal/file/router/**`
   - `internal/relay/router/**`

2. **Target router group**
   - example: `api-v3/node/agent`
   - example: `api-v3/node/proxy`

If the user does not specify the router group, stop and ask them to choose it first.

## Repo Anchors You Must Consult

Always ground the mapping in repo facts, not assumptions.

### Router facts
- the target router group's route registration file
- the handler files for that router group
- `internal/backend/router/AGENTS.md` or the equivalent service-local router guidance

### Auth facts
- `internal/backend/auth/action.go`
- `internal/backend/auth/resource.go`
- `internal/backend/auth/auth.go`

### Reference pattern
- one existing router that already implements permission checks cleanly
- in this repo, `internal/backend/router/api-v3/node/agent/` is the primary reference

## Search Boundary and Stop Conditions

Keep the investigation intentionally small.

### Minimum read set

After the router group is confirmed, read in this order:
1. the router group's route registration file (`Load()` / route registration)
2. the handler files for the endpoints in that router group
3. `internal/backend/auth/action.go`
4. `internal/backend/auth/resource.go`
5. one reference router that already implements the pattern cleanly

Only expand beyond this minimum set when one of these is true:
- authorizer wiring is unclear
- the router's endpoint inventory is still unclear
- the user explicitly asks for broader comparison
- the router uses a special callback/proxy/download path that changes auth entry behavior

Do **not** scan unrelated router groups just because they exist.
Do **not** broaden to other services once the router group is chosen unless the user explicitly asks.

### Stop condition

Stop searching and produce the matrix as soon as all of these are true:
- the endpoint inventory is known
- the candidate protected handler set is known
- the action mapping is good enough to present
- the resource helper mapping is good enough to present
- any remaining uncertainty can be surfaced as `NEEDS USER CONFIRMATION`

Do **not** keep searching for perfect certainty once the matrix can be written.

## Workflow

### Phase 1: Select scope

Confirm:
- service
- router group

If the router group is ambiguous, stop and ask.

### Phase 2: Inventory handlers

Enumerate:
- all registered endpoints
- handler files implementing them
- which handlers already have permission checks
- which handlers still lack them

Do not infer from filenames alone; confirm by route registration and handler code.

### Phase 3: Confirm protected handlers

Present the inventory and confirm with the user:
- which handlers need auth
- which handlers should remain unguarded, if any

If the user says “all handlers in this router,” record that explicitly.

### Phase 4: Define resource helper rules

Before mapping per-handler actions, define the raw-resource-to-auth-resource rules for this router group.

For each helper, identify:
- helper name
- raw input type (biz ID, host, package ID, network unit ID, etc.)
- produced auth system
- produced auth resource type

If a router already has an established helper file pattern, mirror it.

### Phase 5: Map action and resource per handler

For each protected handler, confirm:
- required action(s)
- required resource helper(s)
- whether multiple checks are `single`, `AND`, or `OR`
- where the raw resource comes from in that handler

If anything is uncertain, mark it as `NEEDS USER CONFIRMATION` in the matrix.

## Output Rules

Always output a **permission matrix**.

Use this exact structure:

| Endpoint | Handler | Needs Auth | Action(s) | Resource Helper(s) | Raw Resource Source | Multi-Check Semantics | Notes |
|---|---|---|---|---|---|---|---|
| `/example` | `example.go:Handler` | yes | `auth.ActionXxx` | `buildXxxResources(...)` | `bizIDs from fetched hosts` | `single` / `AND` / `OR` | ... |

Rules:
- use canonical action names from `internal/backend/auth/action.go`
- use canonical resource types/systems from `internal/backend/auth/resource.go`
- write helper names exactly when known
- if no helper exists yet, describe the intended helper shape clearly
- call out unclear mappings as `NEEDS USER CONFIRMATION`
- never hide uncertainty in prose outside the table

Optionally add a very short summary before the matrix, but do not replace the matrix with prose.

## Common Mistakes

- Starting implementation before the router group is confirmed
- Assuming all handlers need auth without showing the endpoint inventory
- Using non-canonical action or resource names
- Skipping the resource helper design step
- Mixing up raw resource source with auth resource output
- Forgetting `AND` / `OR` semantics for multi-check handlers
- Expanding into implementation planning, completion review, or API docs even though those are outside this skill's scope

## Quick Reference

**Core deliverable:**
- one router-scoped permission matrix

**Primary repo references:**
- `internal/backend/router/AGENTS.md`
- `internal/backend/auth/action.go`
- `internal/backend/auth/resource.go`
- `internal/backend/auth/auth.go`
- `internal/backend/router/api-v3/node/agent/`

## Final Guardrails

- Stay router-group scoped.
- Ask the user to choose the router group before going deep.
- Use canonical repo definitions for action/resource names.
- Prefer router-local auth resource helpers when the pattern is router-specific.
- Output the permission matrix every time.
- Do not expand into implementation planning, completion-check, testing strategy, or API documentation unless the user asks separately.
