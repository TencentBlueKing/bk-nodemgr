# bk-nodemgr Artifact Map

Use this reference to classify bk-nodemgr project knowledge before choosing a target artifact.

## Root AGENTS.md

Root `AGENTS.md` owns repository-wide agent rules:

- required tools and retrieval-led habits
- repo-wide language, compression, editing, and verification policies
- global architecture boundaries and anti-patterns
- top-level where-to-look indexes for the monorepo

Avoid putting these in root `AGENTS.md`:

- long conceptual explanations
- human onboarding prose
- reusable workflow tutorials that should be skills
- subtree-only details that should be scoped `AGENTS.md`
- temporary plans, progress, or unfinished status

## Scoped AGENTS.md

Scoped `AGENTS.md` owns agent behavior for one subtree:

- local responsibilities and boundaries
- where-to-look indexes for that subtree
- local commands when they differ from repo defaults
- local naming, type-flow, dependency, or error-handling constraints
- local anti-patterns that prevent recurring mistakes

When recommending scoped `AGENTS.md`, read `.agents/skills/scoped-agentsmd/SKILL.md` first. It owns pipe-index format, scope confirmation, and validation rules.

Avoid scoped `AGENTS.md` for:

- repo-wide policy
- full human-facing tutorials
- durable concepts better explained in docs
- reusable agent workflows better expressed as skills

## README.md

README files are stable human entrypoints.

Put content here when a human reader needs:

- purpose and scope
- quick start or usage path
- module overview
- stable examples
- links to deeper docs

Avoid README content that is:

- agent-only instruction such as “load this skill first”
- temporary status such as “unfinished”, “later”, or “in progress”
- low-level implementation call chains when the reader needs behavior
- detailed API Gateway reference material

When actually writing README content, hand off to `.agents/skills/readme-logic-first/SKILL.md`.

## docs/**

Docs own durable knowledge that is too detailed for README and not an agent execution rule.

Put content here when it explains:

- domain concepts and terminology
- design decisions or architecture
- developer workflows and extension guides
- operation and troubleshooting behavior
- integration and third-party mapping
- FAQ material
- API design and development guidance

Always choose a concrete docs subarea. Do not recommend only `docs/` when subdirectories exist.

## apigw/apidocs/{zh,en}

`apigw/apidocs/{zh,en}` owns API Gateway reference documentation:

- one API operation per zh/en document pair
- endpoint URL, method, permissions, version, request and response examples
- field-level request and response reference
- operation-specific behavior needed by API consumers

Use `.agents/skills/api-doc/SKILL.md` for actual writing.

Do not put broad API design rationale or developer workflow here; use `docs/api/` or `docs/developer/` instead.

## support-files/changelog/{zh,en}

`support-files/changelog/{zh,en}` owns versioned changelog and release-note artifacts:

- version change records
- release notes
- Full Changelog links
- zh/en changelog page content

Use `.agents/skills/changelog-doc/SKILL.md` for changelog writing and `.agents/skills/release-version-alignment/SKILL.md` for release version alignment.

Do not put evergreen design or operation docs here.

## .agents/skills/<name>

Project skills own reusable agent workflows.

Promote content to a skill only when it has:

- triggering situations
- a repeatable workflow or decision flow
- expected output shape
- validation gates or common mistakes
- enough reuse to justify discovery overhead

Avoid skills for:

- one-off task plans
- static project facts
- plain documentation that humans should read directly
- constraints that an agent must obey in every scoped operation

Read `.agents/skills/bk-nodemgr-how-to/SKILL.md` before changing project skill routing or deciding that a new task class needs router coverage.

## issue/task/plan

Use issue, task, plan, or conversation artifacts for temporary execution state:

- current progress
- unfinished ideas
- short-lived investigation notes
- open options not yet accepted as durable project truth

Do not turn temporary status into README/docs/AGENTS content.
