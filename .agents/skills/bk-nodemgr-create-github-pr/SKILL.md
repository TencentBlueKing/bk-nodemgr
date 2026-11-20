---
name: bk-nodemgr-create-github-pr
metadata:
  internal: true
description: Use when preparing, creating, updating, replacing, or splitting a GitHub Pull Request in the bk-nodemgr repository, including routine CI, release, proto, frontend, and cross-layer changes. Trigger for “准备 PR”, “create PR”, “submit pull request”, “替代 PR”, or “提交 PR”.
---

# Create a bk-nodemgr Pull Request

This is the bk-nodemgr PR delivery workflow, not a generic GitHub PR tutorial. Read `docs/developer/git-contribution.md` as the project-owned source for branch, commit, Issue, PR title/body, and verification rules. This skill owns agent execution flow and remote-write safety; do not duplicate every developer rule here.

## Operating Modes

Choose the mode from the user's wording before any remote write:

| User intent                                                | Mode          | Allowed by default                                                                  |
| ---------------------------------------------------------- | ------------- | ----------------------------------------------------------------------------------- |
| “准备”“draft”“整理 PR”“帮我看能不能提”                     | `preflight`   | Local reads, existing remote metadata reads, draft output; no push or GitHub writes |
| “创建/提交 PR” with an explicit Issue and branch available | `create`      | Draft first, then separately confirm each remote write                              |
| “更新 PR”“补充描述”“修标题”                                | `update`      | Read the existing PR, draft a complete change, then confirm the specific edit       |
| “替代/恢复旧 PR”“replaces”                                 | `replacement` | Preserve old facts and head context; migrate into the current template              |
| Multiple independent behaviors or partial delivery         | `split`       | Keep each PR reviewable and use non-closing Issue relations until complete          |

`preflight` is the default when intent is ambiguous. Draft confirmation is not authorization to push, create or edit an Issue/PR, change labels, assign reviewers, or modify milestones.

## 1. Read Rules and Determine Scope

Read these before drafting:

- `docs/developer/git-contribution.md` (project submission rules)
- `.github/CONTRIBUTING.md`
- `.github/PULL_REQUEST_TEMPLATE.md` (authoritative PR body structure)
- `.github/workflows/pr-title-lint.yml` (authoritative title regex)
- `.github/labeler.yml` (path-based label prediction)
- `.agents/skills/bk-nodemgr-issue/SKILL.md` for the Issue handoff in step 2
- nearest `AGENTS.md` files for every changed subtree

For code changes, apply the normal bk-nodemgr routing from `bk-nodemgr-how-to`. This workflow does not replace domain skills such as `bk-nodemgr-config-sync`, `bk-nodemgr-contextx`, `router-permission-supplement`, or `code-review`.

Resolve the canonical comparison base from remote URLs, not remote names:

```bash
git remote -v
git branch -vv
git fetch --prune <canonical-remote> master
git merge-base HEAD <canonical-remote>/master
```

Use freshly fetched canonical `master` when remote refresh is authorized. In `preflight`, do not assume a stale tracking ref is current; report its age or use the existing ref without pretending it was refreshed. Do not use a fork's `origin/master` merely because it is named `origin`. If no canonical base can be identified, stop and ask for the target remote and branch.

Capture branch, target base, fork/direct-push topology, pushed status, merge base, commits, changed paths, behavior changes, and existing Issue/PR references. Do not turn unrelated dirty-worktree changes into scope. If the user gives an explicit commit range, analyze that range rather than the whole branch.

## 2. Resolve the Issue and Decide the PR Relationship

Apply the current rules in `docs/developer/git-contribution.md`. If this skill and the document disagree, use the repository source file named by that document for the exact constraint and report the conflict; keep the agent-only safety gate below.

**Required skill: `bk-nodemgr-issue`.** Pass the user's Issue reference, actual behavior and changed scope, and the current mode/remote-write limits. Delegate Issue lookup and reuse, form selection, drafting, creation, updates, and readback to that workflow; do not run a parallel generic Issue flow. Preparation alone does not authorize creating or editing an Issue.

Receive the Issue handoff: repository, number/URL (or no confirmed match), state, covered scope and acceptance, long-lived tracking status, unresolved gaps, and writes performed or still requiring approval. Confirm that it covers this PR's behavior rather than binding a keyword match.

This PR workflow owns the relationship decision: compare the actual delivered scope with the Issue handoff and apply `docs/developer/git-contribution.md` to choose non-closing `refs` or a closing relationship. Keep routine, umbrella, replacement-only, and partial delivery non-closing; replacement context does not itself prove Issue completion. If no matching Issue is confirmed, follow the guide's `N/A`/title-CI handling and ask rather than fabricate a number or force a new Issue.

## 3. Build the Verification Matrix

Use the changed-path matrix and cross-layer checklist in `docs/developer/git-contribution.md`. Only claim checks run against the analyzed HEAD. An old PR, old commit, or old CI result is historical evidence, not current validation. Do not mark a checkbox passed until the command or manual check ran; if a check did not run, state the exact reason and boundary.

## 4. Draft the Current PR Body

Always start from `.github/PULL_REQUEST_TEMPLATE.md` and apply the section guidance in `docs/developer/git-contribution.md`. Do not use an old `Summary/Changes/Motivation` body as structure. Labels do not replace `Affected Areas`.

For a replacement PR, migrate facts from the old PR into this template. Preserve original head/commit context, Issue links, test notes, risk notes, and replacement reference. If tests were not rerun on current HEAD, say so instead of reusing a historical “passed” claim. Preserve an intentional `[WIP]` marker, but report that it may conflict with title lint; never silently remove it.

## 5. Validate Title and Draft Output

For PRs targeting `master`, validate against the title rule in `docs/developer/git-contribution.md` and `.github/workflows/pr-title-lint.yml`. Keep Issue titles separate from PR title rules; Issue forms own their prefixes.

Before any write, show mode and scope/base/head, Issue and relation, exact title, complete current-template body, predicted path labels, affected areas, verification matrix, unverified items, and remote actions still pending.

In `preflight`, finish with `No GitHub writes performed` and list later actions requiring authorization. Do not push merely to make a draft complete.

## 6. Execute Remote Actions Only After Explicit Confirmation

Treat each as a separate gate:

1. push branch to the intended fork or same-repository remote
2. create or edit an Issue through `bk-nodemgr-issue`, using its confirmation and readback gates
3. create or edit the PR
4. change labels, assignees, reviewers, or milestones

Confirmation to review text does not authorize all four. For fork PRs, use `head=<fork-owner>:<branch>` and target the canonical repository. For same-repository branches, use `gh pr create` with explicit `--repo`, `--base`, and `--head`.

Use `gh pr create --body-file` or `gh pr edit --body-file` for multiline bodies. Do not use `gh api -f body=@file`; it can submit the literal path. Do not treat `--dry-run` as a no-write guarantee because command behavior may still push. After each successful PR write, read back with `gh pr view`; correct only after another explicit edit confirmation. Issue readback belongs to `bk-nodemgr-issue`.

If a command fails after possibly creating a resource, query its state before retrying. Never blindly repeat a create request.

## Output Contract

Return a concise delivery report containing Issue URL/number and relation, PR URL/base/head/title, changed scope and affected areas, actual verification commands/results and untested items, risks/rollback, observed versus predicted labels, and remote actions performed or skipped.

For drafting-only work, return complete Issue/PR drafts and the verification matrix but no fabricated URL, “created successfully” claim, or remote-write implication.

## Common Failure Modes

- Creating a duplicate Issue instead of reusing the matching open or `[DAILY]` Issue.
- Using `closes` for routine, umbrella, replacement, or partial work.
- Reusing a stale PR body or old test result without mapping it to current HEAD.
- Treating path labels as a substitute for affected-area and risk analysis.
- Reporting proto regeneration, frontend checks, or manual verification that did not run.
- Hand-editing generated `pkg/proto/**/*.pb.go` files.
- Treating one confirmation as permission to mutate multiple GitHub resources.
- Retrying `gh` create after an ambiguous failure without checking resource state.
