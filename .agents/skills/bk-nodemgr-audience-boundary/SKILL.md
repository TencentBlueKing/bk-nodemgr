---
name: bk-nodemgr-audience-boundary
description: Use when writing or reviewing bk-nodemgr reader-facing artifacts, including docs, Issues, PRs, changelogs, API references, release notes, and human-facing READMEs, especially when process notes, draft status, missing evidence, or implementation assumptions might leak into the artifact.
---

# bk-nodemgr Audience Boundary

## Scope

Use this skill for bk-nodemgr artifacts with intended human readers. It defines project-specific readers, source anchors, and handoff boundaries; it does not replace `changelog-doc`, `api-doc`, `readme-logic-first`, Issue/PR templates, or their document formats. Do not use it to turn temporary agent status into permanent documentation.

## Reader-Facing Artifact Gate

Before drafting or revising a reader-facing artifact, identify the final reader, the decision or action they need to take, and the target artifact format. Preserve facts, links, limits, risks, compatibility boundaries, and validation expectations that the reader needs. Remove author process, low-signal framing, draft or unfinished status, evidence gaps, and unrequested implementation plans from the artifact body; report them separately to the author when needed. Do not drop key evidence merely to make the content shorter or more template-shaped.

When the user asks for a problem handoff or Issue, describe the problem facts, impact, risk boundary, and decision points instead of assuming a solution. If the user asks for a solution plan, keep the plan scoped to what the reader needs to decide or execute.

## Route By Reader

| Destination                                                       | Reader's decision                         | Include when supported                                                                                                      | Give only to the author                                                                                             |
| ----------------------------------------------------------------- | ----------------------------------------- | --------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `support-files/changelog/{zh,en}/`                                | Whether and how to upgrade or roll back   | Changes affecting Agent / Plugin / Package, Backend, Installer; compatibility, operational prerequisites, known limitations | Tag existence, compare-range verification, release-alignment field checklist, unresolved migration or rollback plan |
| `apigw/apidocs/{zh,en}/`                                          | How to call an API safely                 | URL, permission, request/response semantics, supported version, asynchronous workflow meaning                               | Proto/handler cross-checks, uncertain introduction version, missing field evidence, review requests                 |
| `docs/operation/`                                                 | How to deploy and operate                 | Confirmed deployment prerequisites, procedures, failure and rollback implications                                           | Local test output, draft status, deployment verification still to perform                                           |
| `docs/concepts/`, `docs/api/`, `docs/developer/`, other `docs/**` | How the system works or is integrated     | Stable domain behavior and constraints for that area's audience                                                             | Source-reading notes, hypotheses, documentation progress                                                            |
| GitHub Issues                                                     | Whether and how to triage or take over    | Problem facts, impact, affected dependency/component, source links, risk boundary, decision points, acceptance expectations | Agent handoff process, generic workflow advice, unsupported implementation plan                                     |
| GitHub PRs                                                        | Whether and how to review or merge        | User-visible change, risk, validation, compatibility, reviewer focus                                                        | Local drafting history, unrelated investigation notes, unresolved checks not ready for review                       |
| Human-facing `README.md`                                          | What this component does and how to start | Stable entrypoint and usage rules                                                                                           | Agent instructions, implementation diary, unfinished task list                                                      |

Use `bk-nodemgr-agent-information-architecture` to choose a destination when ownership is unclear. Determine the audience from the destination and nearby examples; ask only if ambiguity changes which facts belong. Read applicable `AGENTS.md` and the narrow writing skill for the target path.

## Evidence And Delivery

1. Retrieve the source appropriate to the claim: version changes from tag diff and release records; API fields and permissions from Proto, handler, auth and swagger; operator instructions from chart values/templates and operation docs. Proving a feature exists does not prove it was introduced in the target release: check the release window before attributing any change to that version, even in a draft. Do not turn a source excerpt or a passing local check into an unsupported public promise.
2. Sort each claim by recipient. A verified limitation such as an asynchronous `trigger_id` not indicating completion belongs in the API doc and possibly the release note if it changes caller behavior. A request to check whether a target tag exists belongs to the release editor. A proven upgrade prerequisite belongs to the operator; "confirm the target environment's rollback plan" is an editor task until a concrete prerequisite can be stated and supported.
3. Keep a reader-facing draft and an author handoff separate. If an essential version, permission, compatibility or upgrade fact is unverified, ask for the missing decision or evidence and do not write the draft to the final `apigw/apidocs/{zh,en}/` or `support-files/changelog/{zh,en}/` path. Use the conversation or an agreed draft location; do not create sidecar notes by default. A planned compare link may be drafted, but final verification of its target and scope is a publication check for the author, not a changelog bullet.
4. Before delivery, read each locale's body as its reader. Keep supported cautions and required actions; remove author-only `待确认`, evidence status and release checklist prose, not merely their warning markers. Ensure zh/en convey the same supported facts. If an existing project skill instructs placing unverified notes in the published body, report that conflict to the author and keep the artifact a draft rather than silently following either contradictory publication instruction.

Do not silently rewrite historical published changelogs or API references to remove such text; report the precise passage and request a separately scoped correction. This skill does not authorize release or API Gateway resource changes.

## Handoff

When a gap exists, provide a reader-facing draft outside the final publication path and a separate author-facing list of blockers/checks in the conversation. When facts and required reviews are complete, deliver only the relevant document and a brief verification summary to the author; do not manufacture a `待确认` section.
