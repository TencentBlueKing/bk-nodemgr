# Git Contribution Guidelines

This document defines the bk-nodemgr contribution requirements from local branches and commits through Pull Requests. Repository CI, the PR template, and the Issue forms remain the authoritative sources for their exact formats; this document explains how those sources fit together and what to check before submission.

## Scope

- Keep one branch focused on one Issue or one behavior change.
- Do not mix unrelated refactoring, formatting, or dependency upgrades into a feature, bug fix, or configuration change.
- Follow the applicable `AGENTS.md` files for the target branch and changed paths.
- Do not disclose unconfirmed security vulnerability details in public Issues; use the project's security reporting channel.

## Branches and Commits

### Branches

Create the working branch from the latest target branch and confirm that the target has not drifted before opening a PR. The branch name should identify the Issue or change theme. Release branches follow the naming rules of the release workflow.

### Commit messages

Use concise, action-oriented English commit messages. Existing project commits and PR titles use a Conventional Commit-style type prefix:

```text
fix: resolve config policy preview error
feat: add plugin config preview
refactor: extract deploy policy calculation flow
chore: update CI workflow
```

| Type       | Use for                                       |
| ---------- | --------------------------------------------- |
| `feat`     | New functionality or user-visible capability  |
| `fix`      | Bug fixes                                     |
| `docs`     | Documentation-only changes                    |
| `refactor` | Behavior-preserving code restructuring        |
| `perf`     | Performance improvements                      |
| `test`     | Adding or correcting tests                    |
| `chore`    | Build, tooling, or auxiliary workflow changes |

When a commit needs an Issue reference, use the repository convention:

```text
fix: resolve config policy preview error --issue=#3642
```

A commit message is change evidence; it does not replace the PR body. Do not infer test results, risk, or complete impact scope from commit messages alone.

## Issue Relationships

Reuse an existing Issue that covers the same behavior before creating a duplicate Issue.

| Situation                                                  | PR body relationship                                                                                                          |
| ---------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| The PR completes the Issue's full acceptance scope         | `closes #N`, `fixes #N`, or `resolves #N`                                                                                     |
| `[DAILY]`, routine, umbrella, or long-lived tracking Issue | `refs #N` or `related to #N`; keep the Issue open                                                                             |
| Staged delivery or partial scope                           | `refs #N` or `related to #N`                                                                                                  |
| Replacement for an old PR                                  | Preserve `Replaces #<old-pr>` and decide the Issue relationship separately                                                    |
| No related Issue                                           | Follow `CONTRIBUTING.md` and write `N/A`; if PR title CI requires an Issue number, confirm the project's Issue handling first |

Use the selected `.github/ISSUE_TEMPLATE/*.yml` form for Issue prefixes, fields, and labels. Do not apply an Issue title prefix to a PR title.

## PR Title

For PRs targeting `master`, `.github/workflows/pr-title-lint.yml` requires:

```text
<type>: <English subject> --issue=#<number>[, #<number>]
```

Requirements:

- `type` is one of `feat`, `fix`, `docs`, `style`, `refactor`, `perf`, `test`, or `chore`.
- `subject` is a concise, action-oriented English description.
- `--issue=#N` is required; separate multiple Issues with a comma and a space.
- The PR body may use Chinese; the English title requirement does not require an English body.
- Do not silently remove an intentional `[WIP]` marker. Check its compatibility with the current title lint rule and report the risk.

Examples:

```text
fix: resolve config policy preview error --issue=#3642
feat: add plugin config preview --issue=#3655
```

## PR Body

Use `.github/PULL_REQUEST_TEMPLATE.md` as the only structure source and fill all five sections:

1. `变更摘要 / Summary`: what changed, why it changed, and which behavior reviewers should inspect.
2. `测试结果 / Test Results`: commands and results from the current HEAD only; use `未测试，原因 / Not tested because` when a check did not run.
3. `风险与回滚 / Risks and Rollback`: compatibility, rollout, permission, data, or workflow risks and an executable rollback or degradation path.
4. `关联 Issue / Related Issues`: the correct `refs` or closing relationship, plus `Replaces #...` for replacement PRs.
5. `涉及范围 / Affected Areas`: applicable proto/generated code, frontend, API contract/route, permission/auth, storage/migration, deployment/config, or docs areas, with impact notes.

For a replacement PR, use the old body only as a source of facts, not as the current structure. Distinguish old PR tests, CI results, and risk conclusions from current HEAD evidence. Do not describe a historical check as current validation unless it was rerun.

## Verification by Changed Area

Choose checks from the actual changed paths and record commands and results in the PR body.

| Changed area                              | Focused checks                                                                                  |
| ----------------------------------------- | ----------------------------------------------------------------------------------------------- |
| Go backend or `pkg/**`                    | Relevant package tests; run `make pre`, `make all`, or `make lint` as scope requires            |
| `proto/**`                                | Run `cd proto && make clean && make all`; confirm `pkg/proto/**` matches the source             |
| `pkg/proto/**`                            | Check the corresponding converter and boundary types; never hand-edit generated `*.pb.go`       |
| `front/**`                                | Run `cd front && pnpm lint`, `pnpm typecheck`, and `pnpm build` as applicable                   |
| API routes, permissions, or auth          | Check route, permission mapping, and authorization impact                                       |
| Storage, migration, deployment, or config | Run scoped checks and describe compatibility and rollback                                       |
| `.github/workflows/**`                    | Check triggers, branches, permissions, secrets, action behavior, and the actual validation path |

For cross-layer changes, trace at least:

```text
proto source -> generated target -> converter -> backend route/service -> frontend API/UI
```

State whether each link changed or was verified unchanged; do not only list file names.

## Pre-submission Checklist

- [ ] The branch is based on the latest target branch and has no unrelated changes.
- [ ] The commit message is concise, action-oriented English with the correct type.
- [ ] A matching Issue was reused and the correct `refs` or closing relationship was chosen.
- [ ] The PR title follows the target branch's CI rule and uses an English subject.
- [ ] The current `.github/PULL_REQUEST_TEMPLATE.md` is used.
- [ ] Test results contain only real validation of the current HEAD; skipped checks include reasons.
- [ ] Risk, rollback, affected areas, and cross-layer impact are documented.
- [ ] No generated `*.pb.go` file was edited manually; proto changes were regenerated.
- [ ] GitHub Actions, labels, related-file checks, and maintainer review have been awaited.

## Rule Ownership

- Developer contribution rules, Issue relationships, title/body explanations, and verification guidance: update this document.
- PR body fields: update `.github/PULL_REQUEST_TEMPLATE.md` and synchronize the relevant section here.
- The actual PR title CI rule: update `.github/workflows/pr-title-lint.yml` and synchronize the format here.
- Issue fields, title prefixes, and default labels: update the relevant `.github/ISSUE_TEMPLATE/*.yml` and synchronize the explanation here.
- Agent-only draft mode, remote-write confirmation, `gh` invocation, and retry safety: keep those rules in `.agents/skills/bk-nodemgr-create-github-pr/SKILL.md`.
