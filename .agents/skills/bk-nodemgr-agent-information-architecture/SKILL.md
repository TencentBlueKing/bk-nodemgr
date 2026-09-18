---
name: bk-nodemgr-agent-information-architecture
description: Use when deciding where bk-nodemgr project knowledge should live across AGENTS.md, scoped AGENTS.md, README.md, docs, apigw/apidocs, support-files/changelog, .agents/skills, and task/issue/plan artifacts. Trigger when organizing agent rules, project skills, docs taxonomy, README boundaries, API reference docs, changelog content, reusable agent workflows, or conflicts between these artifacts.
---

# Agent Information Architecture (bk-nodemgr)

## Overview

Use this skill before editing long-lived bk-nodemgr documentation or agent guidance. It decides which project artifact should own a piece of knowledge, which local anchors must be checked, and which narrower skill should write the final content.

This is a bk-nodemgr-specific skill. The project skill directory is `.agents/skills`. Do not treat this skill as a generic repository information-architecture guide.

## When to Use

Use this skill when the user asks where to put or how to organize:

- agent instructions, scoped rules, commands, guardrails, or anti-patterns
- README content, especially stable human entrypoints and module overviews
- durable concept, architecture, integration, operation, FAQ, API, or third-party docs
- API Gateway reference documentation under `apigw/apidocs/{zh,en}`
- changelog or release-note content under `support-files/changelog/{zh,en}`
- a repeated agent workflow that might become a project skill
- project skill routing, skill discovery, or `.agents/skills` organization
- conflicting information across AGENTS, README, docs, skills, apidocs, changelog, and task artifacts

Do not use it for ordinary copyediting when the target file is already explicit and the edit is mechanical.

## Core Workflow

1. Discover bk-nodemgr local information architecture before deciding placement.
2. Classify the content by reason-to-change and artifact responsibility.
3. Check the required project anchors for the candidate artifact.
4. If the candidate is docs, choose the concrete docs area, not only `docs/`.
5. If the candidate is a skill or agent workflow, read `.agents/skills/bk-nodemgr-how-to/SKILL.md`.
6. If the candidate is scoped `AGENTS.md`, read `.agents/skills/scoped-agentsmd/SKILL.md`.
7. Output a `Placement Decision` before editing files unless the user gave an exact path and asked for a mechanical edit.
8. Hand off actual writing to the narrower bk-nodemgr skill when one owns the target content.

## Required Discovery

Search only as much as needed for the placement decision, but do not skip these anchors when relevant:

- Root rules: `AGENTS.md`.
- Scoped rules: nearest and nested `AGENTS.md` under the affected subtree.
- Project skills: `.agents/skills/*/SKILL.md`.
- Skill router: `.agents/skills/bk-nodemgr-how-to/SKILL.md` when skill, workflow, or router placement is involved.
- Scoped AGENTS workflow: `.agents/skills/scoped-agentsmd/SKILL.md` when scoped `AGENTS.md` is a candidate.
- README entrypoints: root `README.md` and module README files.
- Docs taxonomy: `docs/README.md`, `docs/AGENTS.md`, and representative files under the candidate docs area.
- API reference docs: `apigw/apidocs/{zh,en}` when documenting Proto API Gateway references.
- Changelog artifacts: `support-files/changelog/{zh,en}` when writing release notes or versioned changelog.

Finding a nearby example is evidence, not authority. Resolve by artifact responsibility and project rules.

## bk-nodemgr Artifact Map

Use `references/bk-nodemgr-artifact-map.md` as the responsibility matrix. Default responsibilities:

| Artifact                           | Owns                                                                                                 | Does Not Own                                                                |
| ---------------------------------- | ---------------------------------------------------------------------------------------------------- | --------------------------------------------------------------------------- |
| `AGENTS.md`                        | repo-wide agent rules, required tools, retrieval policy, global boundaries                           | detailed concepts, one-off plans, reusable workflow tutorials               |
| scoped `AGENTS.md`                 | subtree-specific agent rules, where-to-look indexes, local anti-patterns, scope commands             | repo-wide policy, human onboarding, long explanations                       |
| `README.md`                        | stable human entrypoint, purpose, quick start, usage overview, links                                 | agent-only instructions, transient status, implementation call chains       |
| `docs/**`                          | durable concepts, architecture, developer, operation, integration, third-party, FAQ, API design docs | agent execution rules, API Gateway reference pages, release changelog       |
| `apigw/apidocs/{zh,en}`            | API Gateway reference pages, one API per zh/en pair, operation-specific request/response docs        | API design rationale, developer guides, broad concepts                      |
| `support-files/changelog/{zh,en}`  | versioned changelog, release notes, Full Changelog content                                           | evergreen docs, temporary plans                                             |
| `.agents/skills/<name>`            | reusable project agent workflow with trigger, steps, output contract, validation                     | static facts, one-off notes, constraints agents must always obey in a scope |
| `.agents/skills/bk-nodemgr-how-to` | project skill routing and companion-skill decisions                                                  | mandatory sync for every new skill unless routing semantics change          |
| issue/task/plan                    | temporary status, unfinished plans, current execution notes                                          | durable project truth                                                       |

## Placement Rules

Classify by reason-to-change:

| Reason-to-change                                                                           | Preferred placement                                                            |
| ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------ |
| Agent must obey it throughout the repo                                                     | root `AGENTS.md`                                                               |
| Agent must obey it only under one subtree                                                  | scoped `AGENTS.md`                                                             |
| Human needs a stable entrypoint or usage overview                                          | `README.md`                                                                    |
| Content explains durable domain, architecture, operation, integration, or support behavior | concrete `docs/<subarea>`                                                      |
| Content is API Gateway endpoint reference                                                  | `apigw/apidocs/{zh,en}` and `api-doc` handoff                                  |
| Content is versioned changelog or release note                                             | `support-files/changelog/{zh,en}` and `changelog-doc` handoff                  |
| Content is a reusable project agent workflow                                               | `.agents/skills/<name>`                                                        |
| Content changes which project skills to load                                               | `.agents/skills/bk-nodemgr-how-to/SKILL.md` only when routing semantics change |
| Content is temporary plan, progress, or unfinished status                                  | issue/task/plan/conversation, not README/docs as durable truth                 |

## Docs Taxonomy

Do not treat `docs/` as one bucket. Use `references/bk-nodemgr-docs-taxonomy.md` when docs placement matters.

Current verified docs areas include:

- `docs/api/`
- `docs/concepts/`
- `docs/developer/`
- `docs/faq/`
- `docs/integration/`
- `docs/operation/`
- `docs/thirdparty/`
- `docs/img/`

Always re-check the local docs index before choosing a docs area.

## Skill Placement

Use `references/bk-nodemgr-skill-map.md` when placing or designing skills.

Rules:

- Project skills live under `.agents/skills/<name>`.
- Read `.agents/skills/bk-nodemgr-how-to/SKILL.md` before deciding that new or changed skill routing is needed.
- Do not update `bk-nodemgr-how-to` only because a new skill exists; update it when routing semantics, task-class ownership, or cross-reference expectations change.
- A skill should have triggers, repeatable workflow or decision flow, output contract, validation gates, and enough reuse to justify discovery overhead.

## Handoff Rules

This skill decides placement and synchronization boundaries. When actual writing is needed, load the narrow skill that owns the target:

- scoped `AGENTS.md`: `.agents/skills/scoped-agentsmd/SKILL.md`
- README or logic-first docs: `.agents/skills/readme-logic-first/SKILL.md`
- Proto API Gateway reference docs: `.agents/skills/api-doc/SKILL.md`
- changelog or release notes: `.agents/skills/changelog-doc/SKILL.md`
- release version alignment: `.agents/skills/release-version-alignment/SKILL.md`

## Conflict Resolution

Resolve by artifact responsibility:

1. Agent behavior constraints belong to root or scoped `AGENTS.md`.
2. Reusable agent procedures belong to `.agents/skills`.
3. Long-lived semantics, decisions, and operating behavior belong to `docs/**`.
4. API Gateway endpoint reference belongs to `apigw/apidocs/{zh,en}`.
5. Release notes and versioned changes belong to `support-files/changelog/{zh,en}`.
6. Human entry language belongs to `README.md`.
7. Temporary plans and progress belong to issue/task/plan artifacts.

Stop and ask when the placement or sync would change a public contract, data model, trust boundary, permission model, long-lived architecture, or project-wide skill routing convention.

## Output Contract

Use Chinese main text with English technical terms and exact paths.

```text
Placement Decision
推荐位置：[AGENTS.md | scoped AGENTS.md | README.md | docs/<subarea> | apigw/apidocs/{zh,en} | support-files/changelog/{zh,en} | .agents/skills/<skill> | .agents/skills/<router-skill> | issue/task/plan]
变更原因：[为什么这类内容未来会随什么变化]
已检查证据：
- [AGENTS.md / scoped AGENTS.md / README / docs index / .agents skill / target subtree]
bk-nodemgr 锚点：
- [与本项目相关的目录、skill、docs、规则文件]
应放入：
- [推荐位置应该承载的内容]
不应放入：
- [相邻但不适合的 artifact 及理由]
关联 project skill：
- [无 | .agents/skills/<name>/SKILL.md + 如何影响判断]
如果是 docs：
- 目标 docs 区域：[具体目录]
- 未选择区域：[相邻目录 + 拒绝理由]
如果是 skill：
- skill 类型：[project-local workflow | project-router | companion reference]
- 目标目录：[.agents/skills/<name>]
是否影响 scoped AGENTS.md：[否 | 是，路径与原因]
需要同步：
- [无 | 具体文件和同步原因]
风险：[low | medium | high]
阻塞问题：[无 | 一个必须问用户的问题]
是否执行修改：[yes/no]
```

## Common Mistakes

| Mistake                                                           | Better move                                                                                          |
| ----------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------- |
| Treating this as a generic documentation placement skill          | Apply bk-nodemgr artifact ownership and anchors first                                                |
| Putting agent-only instructions in README                         | Put them in root or scoped `AGENTS.md`, or a skill if they are a reusable workflow                   |
| Putting durable concepts in `AGENTS.md`                           | Put them in the concrete `docs/<subarea>` and link from rules only if agents must know where to look |
| Putting API Gateway reference pages under `docs/api/`             | Use `apigw/apidocs/{zh,en}` and hand off to `api-doc`                                                |
| Putting release notes under normal docs                           | Use `support-files/changelog/{zh,en}` and hand off to `changelog-doc`                                |
| Making every repeated note a skill                                | Require trigger, workflow, output contract, validation, and repeated reuse                           |
| Updating `bk-nodemgr-how-to` for every new skill                  | Update only when routing semantics change                                                            |
| Recommending scoped `AGENTS.md` without reading `scoped-agentsmd` | Read the scoped AGENTS workflow first                                                                |
| Treating `docs/` as a dumping ground                              | Discover and use the local docs taxonomy                                                             |

## References

- `references/bk-nodemgr-artifact-map.md`: project artifact responsibilities and ownership boundaries.
- `references/bk-nodemgr-docs-taxonomy.md`: bk-nodemgr docs, apidocs, and changelog placement rules.
- `references/bk-nodemgr-skill-map.md`: project skill placement and router update rules.

## Validation Prompts

Project-specific prompts live in `evals/evals.json`. They cover `.agents/skills`, `bk-nodemgr-how-to`, scoped `AGENTS.md`, README handoff, API docs, changelog, docs taxonomy, and blocking-question behavior.
