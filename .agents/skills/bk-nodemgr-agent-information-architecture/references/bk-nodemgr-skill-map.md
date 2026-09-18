# bk-nodemgr Skill Map

Use this reference when deciding whether project knowledge belongs in `.agents/skills` or whether a skill-router update is needed.

## Project Skill Directory

bk-nodemgr project skills live under `.agents/skills/<name>`.

Before creating or changing a skill placement decision:

1. Search `.agents/skills/*/SKILL.md` for existing ownership.
2. Read `.agents/skills/bk-nodemgr-how-to/SKILL.md` when the task touches skill routing, project workflow discovery, or companion-skill selection.
3. Prefer extending an existing skill when it already owns the workflow.
4. Create a new skill only when the workflow has distinct triggers, steps, output contract, and validation gates.

## Skill Types

| Type                   | Placement                                                              | Owns                                                           |
| ---------------------- | ---------------------------------------------------------------------- | -------------------------------------------------------------- |
| project-local workflow | `.agents/skills/<workflow>/SKILL.md`                                   | repeatable bk-nodemgr task workflow                            |
| project-router         | `.agents/skills/bk-nodemgr-how-to/SKILL.md` or another explicit router | routing across project skills and generic companions           |
| companion reference    | reference file inside a skill                                          | detailed rules loaded only when the parent workflow needs them |

Do not create a skill for static background facts, one-off plans, or always-on scoped constraints. Use docs, issue/task/plan, or `AGENTS.md` instead.

## Router Update Rules

Update `.agents/skills/bk-nodemgr-how-to/SKILL.md` only when the task changes:

- project skill discovery expectations
- the router entrypoint's intended scope
- routing decisions for a task class
- cross-reference expectations for existing skills
- generic vs project-specific companion-skill rules

Do not update the router only because a new skill exists.

## Required Handoffs

When placement selects a target owned by a narrower skill, this skill should stop at placement and synchronization boundaries, then hand off actual writing:

- scoped `AGENTS.md`: `.agents/skills/scoped-agentsmd/SKILL.md`
- README or logic-first docs: `.agents/skills/readme-logic-first/SKILL.md`
- Proto API Gateway reference docs: `.agents/skills/api-doc/SKILL.md`
- changelog or release notes: `.agents/skills/changelog-doc/SKILL.md`
- release version alignment: `.agents/skills/release-version-alignment/SKILL.md`
- router permission mapping: `.agents/skills/router-permission-supplement/SKILL.md`

## Common Placement Checks

- If the content tells agents what to do every time under a path, prefer scoped `AGENTS.md`, not a skill.
- If the content is a reusable procedure triggered by user intent, prefer a skill.
- If the content changes which skill future agents should load, evaluate `bk-nodemgr-how-to` router impact.
- If the content is a long explanation for humans, prefer README or docs instead of a skill.
