---
name: router-permission-supplement
description: Use when users want to add, review, confirm, or verify router/API permission checks for a chosen router group, especially when they mention 补权限, 鉴权, permission coverage, action/resource mapping, permission helper rules, router permission design, or checking whether permission supplementation is complete.
---

# Router Permission Supplement

## Overview

Use this skill to guide **permission supplementation work for one selected router group**.

This skill is **checklist-first, not execution-first**:
- before implementation, produce a confirmed permission matrix and checklist
- after implementation, verify completion against that matrix and check API docs
- do **not** force `.sisyphus` or OpenSpec usage; those are optional companion workflows

## When to Use

Use this skill when the user wants to:
- add permission checks to existing router APIs
- confirm which endpoints under a router group need authorization
- design or review `action` / `resource` mappings
- introduce or validate router-local auth resource helpers such as `permission_rules.go`
- verify whether a permission supplementation task is fully completed
- update API docs after permission checks are added

Do **not** use this skill when:
- the user only wants to implement code immediately without analysis
- the task is unrelated to router/API authorization
- the task is only about IAM provider callback internals with no router permission supplementation goal

## Core Principle

Treat permission supplementation as a **router-scoped workflow**:

1. choose one router group
2. inspect auth wiring and endpoint inventory
3. confirm which endpoints require auth
4. define router-local resource helper rules
5. confirm action/resource mapping with the user
6. output a matrix and checklist
7. later, verify implementation and API docs against the same matrix

The router package should own its own permission helper rules when the pattern is router-specific.

## Required Inputs

Always gather these before producing the final output:

1. **Target service**
   - `internal/backend/router/**`
   - `internal/application/router/**`
   - `internal/file/router/**`
   - `internal/relay/router/**`

2. **Target router group**
   - example: `api-v3/node/agent`
   - example: `api-v3/plugin/package`

3. **Requested mode**
   - `planning` - create matrix + checklist before implementation
   - `completion-check` - verify implementation after coding is done

4. **Doc target**
   - default: `apigw/apidocs/zh` and `apigw/apidocs/en`

If the user does not specify the router group, stop and ask them to choose it first.

## Repo Anchors You Must Consult

Always ground the workflow in repo facts, not assumptions.

### Router conventions
- `internal/backend/router/AGENTS.md`
- equivalent service-local router guidance if the selected service has one

### Auth facts
- `internal/backend/auth/action.go`
- `internal/backend/auth/resource.go`
- `internal/backend/auth/auth.go`
- `pkg/rest/errf/permission.go`
- `support-files/bkiamv3/templates/0003_bk_nodemgr_actions.json.tpl`

### API doc process
- `.claude/skills/api-doc/SKILL.md`
- `docs/api/API接口开发流程.md`

### Reference implementation pattern
- inspect an existing router that already supplements permissions cleanly
- in this repo, `internal/backend/router/api-v3/node/agent/` is the primary reference

## Workflow

## Search Boundary and Stop Conditions

In `planning` mode, keep the investigation intentionally small.

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
Do **not** broaden the search to other services once the router group has been chosen unless the user explicitly asks for cross-service comparison.

### Planning-mode stop condition

Stop searching and start writing as soon as all four are true:
- the endpoint inventory is known
- auth wiring status is known
- the candidate action/resource/helper mapping is good enough to present
- remaining uncertainty can be surfaced as `NEEDS USER CONFIRMATION`

Once these four are satisfied, produce the matrix and checklist immediately.
Do **not** continue searching for perfect certainty.

### Completion-check stop condition

In `completion-check` mode, focus on:
- code vs expected matrix
- obvious test/document follow-up gaps

Do **not** reopen design exploration unless the implementation clearly diverges from the intended matrix.

### Phase 1: Select scope

Ask the user to confirm:
- service
- router group
- planning vs completion-check mode

If the router group is still ambiguous, stop and ask.

### Phase 2: Inspect auth wiring

Check whether the selected router group already has the prerequisites for permission checks:
- top-level auth middleware path exists
- handler can access an authorizer dependency
- there is already a permission helper file or a natural place to add one

Typical questions to answer:
- Is `Authorizer` injected into the handler/capability layer?
- Does the router already use `BatchCheck` or `Check` anywhere?
- Is there already a router-local helper pattern for auth resources?

If wiring is missing, record it as a **prerequisite blocker** in the checklist.
Only gather enough evidence to answer whether auth prerequisites exist for the selected router group.

### Phase 3: Inventory endpoints

Read the target router package and enumerate:
- all registered endpoints
- handler files implementing them
- which endpoints already have permission checks
- which endpoints still lack them

Do not infer from filenames alone; confirm by route registration and handler code.

Once the endpoint inventory is stable, stop expanding the search surface.

### Phase 4: Confirm protected endpoints

Present the endpoint inventory to the user and ask them to confirm:
- which endpoints require authorization
- which endpoints should remain unguarded, if any

If the user says “all endpoints in this router,” record that explicitly.

### Phase 5: Define router-local resource helper rules

Before discussing per-endpoint actions, define the raw-resource-to-auth-resource rules for this router group.

Prefer router-local helper functions when the pattern is router-specific.

For each helper, identify:
- helper name
- raw input type (biz ID, host, package ID, network unit ID, etc.)
- produced auth system
- produced auth resource type

If a router already has an established helper file pattern, mirror it.

Do **not** jump to global shared abstractions unless the user explicitly asks for a cross-router refactor.

### Phase 6: Confirm action/resource mapping with the user

For each endpoint that needs auth, confirm:
- required action(s)
- required resource helper(s)
- whether multiple checks are `AND` or `OR`
- where in the handler the check belongs

The placement rule should be explicit, for example:
- after request bind
- after fetching hosts/resources
- before business mutation or workflow launch

### Phase 7: Produce the permission matrix

Always output a matrix first.

Use this exact structure:

| Endpoint | Handler | Needs Auth | Action(s) | Resource Helper(s) | Check Timing | Multi-Check Semantics | Notes |
|---|---|---|---|---|---|---|---|
| `/example` | `example.go:Handler` | yes | `auth.ActionXxx` | `buildXxxResources(...)` | after bind, before business logic | single / AND / OR | ... |

Rules:
- use canonical action names from `internal/backend/auth/action.go`
- use canonical resource types/systems from `internal/backend/auth/resource.go`
- call out unclear mappings as `NEEDS USER CONFIRMATION`
- never hide uncertainty in prose
- if the selected router group is clear, do not delay output just to gather more peripheral evidence

### Phase 8: Produce the implementation checklist

After the matrix, output a checklist.

Use this exact structure:

```md
## Permission Supplement Checklist

### Scope
- [ ] Target service and router group confirmed
- [ ] Endpoint inventory confirmed with user
- [ ] Protected endpoint set confirmed

### Auth wiring
- [ ] Router has usable authorizer wiring
- [ ] Permission denial should return `resterrf.PermissionDenied`
- [ ] Existing auth middleware path verified

### Resource helper design
- [ ] Router-local helper rules confirmed
- [ ] Helper file/location confirmed
- [ ] No unnecessary global abstraction introduced

### Endpoint mapping
- [ ] Every protected endpoint has confirmed action(s)
- [ ] Every protected endpoint has confirmed resource helper(s)
- [ ] Multi-check semantics (`AND` / `OR`) confirmed where needed
- [ ] Check timing confirmed for every protected endpoint

### Completion verification
- [ ] Permission checks implemented in handlers
- [ ] Wrong or missing `PermissionDenied` wrapping not present
- [ ] Tests/verification commands selected
- [ ] API docs updated in `apigw/apidocs/**`
```

After the generic checklist, add router-specific checklist items for the selected router group.

### Phase 9: Completion-check mode

When invoked after implementation, do not redesign from scratch.

Instead:
1. reconstruct or read the intended endpoint matrix
2. compare code against the matrix
3. report complete / incomplete / mismatched items
4. check API docs were updated accordingly

Completion-check output should use this template:

```md
## Permission Supplement Completion Check

### Complete
- ...

### Incomplete
- ...

### Mismatch
- Endpoint: ...
  - Expected: ...
  - Found: ...

### API Docs Follow-up
- ...
```

## API Doc Follow-up Rules

After permission supplementation is implemented, verify the interface documentation under `apigw/apidocs/**`.

Follow existing `api-doc` skill conventions for naming and placement.

At minimum, make sure the docs clearly record:
- required action(s)
- required resource(s)
- additional conditions for multi-check endpoints
- permission failure semantics

Recommended documentation block:

```md
### Authorization

- **Required action(s)**: `auth.ActionXxx`
- **Required resource(s)**: `resource helper output / resource rule`
- **Additional conditions**: `AND` / `OR` / prerequisite notes
- **Failure semantics**: returns permission denied when the required checks fail
```

If the project's existing API doc format differs, adapt the wording but keep these four facts present.

## Output Rules

Always produce both:
1. a **permission matrix**
2. a **checklist**

Optionally add a short summary before them, but do not replace either artifact with prose.

## Common Mistakes

- Starting implementation before the router group is confirmed
- Continuing to search after endpoint inventory and wiring status are already clear
- Expanding to unrelated router groups or services without user instruction
- Assuming all endpoints need auth without showing the inventory
- Skipping the router-local helper design step
- Using non-canonical action or resource names
- Forgetting `AND` / `OR` semantics for multi-check endpoints
- Forgetting `resterrf.PermissionDenied` semantics in completion review
- Treating API doc updates as optional after code changes
- Forcing `.sisyphus` / OpenSpec output when the user only asked for a checklist

## Quick Reference

**Default mode:** planning

**Default deliverables:**
- permission matrix
- implementation checklist

**Default doc target:**
- `apigw/apidocs/zh`
- `apigw/apidocs/en`

**Primary repo references:**
- `internal/backend/router/AGENTS.md`
- `internal/backend/auth/action.go`
- `internal/backend/auth/resource.go`
- `pkg/rest/errf/permission.go`
- `.claude/skills/api-doc/SKILL.md`

## Final Guardrails

- Stay router-group scoped.
- Ask the user to choose the router group before going deep.
- Use canonical repo definitions for action/resource names.
- Prefer router-local auth resource helpers when the pattern is router-specific.
- Output matrix + checklist every time.
- In completion-check mode, verify implementation against the matrix rather than inventing a new design.
