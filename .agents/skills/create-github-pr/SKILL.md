---
name: create-github-pr
description: Use when creating a GitHub pull request for the current branch. Handles the complete workflow - analyzing changes, creating an issue with proper templates, reading PR conventions (title format, labeler rules), and creating a PR that follows project standards. Trigger when user says "create PR", "submit pull request", "准备合并", "提交 PR", or mentions merging their branch. Always use this skill for PR creation to ensure compliance with project conventions.
---

# Create GitHub PR

A skill for creating GitHub pull requests that follow project conventions, with human confirmation checkpoints at key decision points.

## When to Use

Use this skill whenever the user wants to create a pull request. Common phrases:
- "Create a PR for this branch"
- "Submit a pull request"
- "准备合并到主分支"
- "提交 PR"
- "Open a pull request"

## Workflow Overview

The skill follows this sequence with human confirmation at critical steps:

1. **Analyze Changes** → Gather branch info, commits, diffs
2. **Create Issue** → Draft issue content → ✅ **CONFIRM** → Create on GitHub
3. **Read PR Conventions** → Parse title format rules, labeler config
4. **Create PR** → Draft PR content → ✅ **CONFIRM** → Create on GitHub

## Step 1: Analyze Current Branch

Gather context about what's being merged. The comparison base must be the latest canonical upstream `master`, not a stale tracking ref and not the fork's `origin/master` unless `origin` is the canonical upstream.

```bash
# Resolve the real upstream and base before diffing
git remote -v
git branch -vv
git fetch --prune [upstream-remote] master

# Run these in parallel after [upstream-remote]/master is updated
git status
git log --oneline -10
git merge-base HEAD [upstream-remote]/master
git diff $(git merge-base HEAD [upstream-remote]/master)..HEAD --stat
git diff $(git merge-base HEAD [upstream-remote]/master)..HEAD
```

Base resolution rules:
1. Prefer the remote URL pointing to the canonical project repository (for bk-nodemgr: `TencentBlueKing/bk-nodemgr`) and fetch its `master` before diffing.
2. If `origin` points to the user's fork, do not use `origin/master` as the comparison base unless the user explicitly confirms it is synced to canonical `master`.
3. If `git branch -vv` shows the tracking branch as `[gone]`, ignore that tracking ref for diff scope; it may be stale or absent.
4. If no canonical upstream remote exists, ask the user which remote/branch is the PR base before drafting the issue.
5. If the user provides explicit commit hashes or a commit range, analyze exactly that range and do not replace it with whole-branch diff output.

Extract:
- Current branch name
- Tracking remote (if any), but do not trust it as diff base when stale/gone
- Canonical upstream remote and freshly fetched base branch (usually `upstream/master` or the remote that points to `TencentBlueKing/bk-nodemgr`)
- Commit messages
- Files changed (with line counts)
- Actual code changes

Summarize the changes in 2-3 sentences focusing on **what** changed and **why** (infer from commit messages and code).

### Identify Remote Configuration

Detect fork vs direct-push workflow:

```bash
git remote -v
```

Parse output to identify:
- **Canonical upstream remote**: The remote URL pointing to the project repository (for bk-nodemgr: `TencentBlueKing/bk-nodemgr`). It may be named `upstream`, `origin`, or something else; identify it from `git remote -v`, not from the remote name alone.
- **Fork remote**: User's fork (e.g., `zouxingyuks/bk-nodemgr`). Do not use its `master` as diff base unless explicitly confirmed synced to canonical `master`.

If fork remote exists:
1. Push branch to fork remote, not origin
2. Use `gh api` method for PR creation with `head="[fork-owner]:[branch]"`
3. Handle branch divergence with force push to fork if needed

Example detection:

```bash
# Find fork remote (non-origin push URLs)
git remote -v | grep push | grep -v "TencentBlueKing"
```

## Step 2: Create GitHub Issue

### Find Issue Templates

Check for issue templates:

```bash
ls .github/ISSUE_TEMPLATE/
```

Common templates:
- `bug_report.yml` - For bug fixes
- `feature_request.yml` - For new features
- `discussion.yml` - For proposals/refactors

### Draft Issue Content

Based on the change analysis, draft:
- **Template**: Choose the most appropriate template
- **Template title prefix**: Read the template's `title:` value and preserve it exactly (for example, `feature_request.yml` uses `[FEATURE]: `). The final issue title must start with that prefix unless the user explicitly asks otherwise.
- **Title**: Concise issue summary after the template prefix. Use Chinese for bk-nodemgr issue titles by default because the public issue templates are bilingual with Chinese-first labels.
- **Body language**: Write the issue body in Chinese by default for bk-nodemgr. Keep proper nouns, code identifiers, package names, labels, and technical keywords in English where natural.
- **Body structure**: Fill the selected template fields in the same order as the template, using the template's Chinese-first field labels where practical.
- **Labels**: Suggest labels from the selected issue template and any clearly applicable project labels (e.g., `kind/features`, `module/pkg`).

Why this matters: GitHub issue forms provide default title prefixes and Chinese-first field labels, but `gh issue create` does not automatically apply the template contract when passing a custom `--title` and `--body`. The agent must carry the template prefix and language convention into the draft explicitly.

### ✅ Checkpoint 1: Confirm Issue

Present to user:

```
I'll create an issue with:
- Template: [template name]
- Template title prefix: [exact template title prefix]
- Title: [proposed title including prefix]
- Language: Chinese by default for bk-nodemgr issues
- Labels: [suggested labels]

Body:
[formatted body content using selected template fields]

Should I proceed? (yes/no, or suggest changes)
```

Wait for confirmation. If user suggests changes, revise and confirm again.

### Create Issue

Once confirmed, create via GitHub CLI:

```bash
gh issue create \
  --repo [owner/repo] \
  --title "[title]" \
  --body "[body]" \
  --label "[label1,label2]"
```

Capture the issue URL and number from output.

Immediately verify the created issue:

```bash
gh issue view [issue-number] --repo [owner/repo] --json title,body,labels,url
```

Confirm the returned title still includes the template prefix and the body language matches the confirmed draft. If it does not, update the issue before continuing to PR drafting.

## Step 3: Read PR Conventions

Projects often have PR title format requirements and auto-labeling rules. Read these files:

```bash
# Check for PR title lint rules
cat .github/workflows/pr-title-lint.yml

# Check for auto-labeler config
cat .github/labeler.yml
```

### Parse Title Format

Look for `title-regex` in the workflow file. Common format:

```
^(feat|fix|docs|style|refactor|perf|test|chore):\s[a-zA-Z0-9\s,.\\-_()]+\s--issue=#\d+(,\s?#\d+)*\s*$
```

This means:
- **Type prefix**: `feat|fix|docs|style|refactor|perf|test|chore`
- **Colon + space**: `: `
- **English description**: Letters, numbers, spaces, punctuation
- **Issue reference**: ` --issue=#[number]`

### Parse Auto-Labels

From `labeler.yml`, identify which labels will be auto-applied based on changed files. For example:

```yaml
module/pkg:
  - changed-files:
      - any-glob-to-any-file: 'pkg/**/*'
```

If your changes touch `pkg/`, the `module/pkg` label will be auto-applied.

## Step 4: Create Pull Request

### Determine Base Branch

Use the same freshly resolved base branch from Step 1. Check the workflow file for target branches (usually `master` or `main`) and ensure the PR base matches the diff base used for analysis. If they differ, re-run the analysis against the intended PR base before drafting the PR.

```bash
git remote show [upstream-remote] | grep "HEAD branch"
```

### Draft PR Content

Based on the changes and issue, draft:

**Title** (must match regex):
```
[type]: [english description] --issue=#[issue-number]
```

Examples:
- `feat: add user authentication module --issue=#6510`
- `refactor: unify storage WrapFn closure pattern --issue=#9012`
- `fix: resolve login timeout issue --issue=#1234, #1235`

### Language Selection

Determine PR body language based on:
1. **Issue language**: If linked issue uses Chinese, PR body should use Chinese
2. **User request**: If user explicitly requests a language, follow it
3. **Project convention**: For bk-nodemgr, prefer Chinese PR bodies when the issue is Chinese
4. **Fallback default**: English only when the issue/user/project convention does not indicate a localized language

**Note**: PR title MUST remain English (regex requirement), but issue titles do not use the PR title regex. Do not apply the PR title format to GitHub issue titles; issue titles must follow the selected issue template's `title:` prefix.

**Body (English template)**:

```markdown
## Summary
[2-3 sentence overview of what changed and why]

## Changes
- **File 1**: [what changed]
- **File 2**: [what changed]

## Motivation
[Why this change was needed - expand on the "why"]

closes #[issue-number]
```

**Body (Chinese template)**:

```markdown
## 概述
[2-3 句概述变更内容和原因]

## 变更内容
- **文件1**: [变更说明]
- **文件2**: [变更说明]

## 动机
[为什么需要这个变更]

closes #[issue-number]
```

At Checkpoint 2, if issue was written in Chinese, present Chinese template by default.

### ✅ Checkpoint 2: Confirm PR

Present to user:

```
I'll create a PR with:

Title: [proposed title]
Base: [base-branch]
Head: [owner:branch-name]

Body:
[formatted body]

Auto-labels (from labeler.yml): [predicted labels]

Should I proceed? (yes/no, or suggest changes)
```

Wait for confirmation. If user suggests changes, revise and confirm again.

### Create PR

Determine the correct method based on remote setup:

**If pushing to fork** (common in open source):
```bash
gh api repos/[upstream-owner]/[repo]/pulls --method POST \
  -f title="[title]" \
  -f head="[fork-owner]:[branch]" \
  -f base="[base-branch]" \
  -f body="[body]" \
  --jq '.html_url'
```

**If pushing to same repo**:
```bash
gh pr create \
  --repo [owner/repo] \
  --base [base-branch] \
  --head [branch] \
  --title "[title]" \
  --body "[body]"
```

Capture the PR URL from output.

## Step 5: Report Results

Present to user:

```
✅ PR created successfully!

Issue: [issue-url]
PR: [pr-url]

Auto-applied labels: [labels from labeler.yml]

The PR title follows the project's format: [type]: [description] --issue=#[number]
```

## Error Handling

### Title Format Validation Failed

If the PR creation fails due to title format:
1. Show the regex pattern from `pr-title-lint.yml`
2. Show the attempted title
3. Explain what's wrong (missing colon space, non-English chars, wrong issue format, etc.)
4. Suggest a corrected title
5. Ask user to confirm before retrying

### Issue Template Not Found

If no issue templates exist:
1. Create a simple issue with title and body
2. Use generic labels like `enhancement` or `bug`
3. Inform user that the project doesn't use issue templates

### Remote/Branch Issues

If `gh pr create` fails with "Head sha can't be blank":
1. Check if branch is pushed: `git branch -vv`
2. If not pushed, push first: `git push -u [remote] [branch]`
3. If pushing to fork, use the `gh api` method instead

## Tips for Success

1. **Always read project conventions first** - Don't assume standard formats
2. **Use exact regex patterns** - PR title lint is strict, one wrong character fails CI
3. **Predict auto-labels** - Tell user which labels will be applied so they're not surprised
4. **Link issue and PR** - Use `closes #[number]` in PR body for auto-linking
5. **English descriptions only** - Most projects require English in titles even if comments are in other languages
6. **Multiple issues** - Format as `--issue=#100, #101` (comma + space)

## Common Patterns

### Type Selection

- `feat` - New functionality added
- `fix` - Bug correction
- `refactor` - Code restructuring without behavior change
- `perf` - Performance improvement
- `docs` - Documentation only
- `style` - Formatting, whitespace, etc.
- `test` - Test additions or corrections
- `chore` - Build, tools, dependencies

### Description Writing

Good descriptions are:
- **Specific**: "add JWT authentication" not "add auth"
- **Action-oriented**: Start with verb (add, fix, refactor, update)
- **Concise**: 5-10 words, no unnecessary details
- **English**: Even if codebase comments are in other languages

Bad examples:
- "Add new feature" (too vague)
- "添加新功能" (not English)
- "feat add feature" (missing colon)
- "FEAT: add feature" (type must be lowercase)

## Confirmation Checkpoints

The skill has two mandatory confirmation points:

1. **After drafting issue** - User reviews title, body, labels before creation
2. **After drafting PR** - User reviews title, body, predicted labels before creation

At each checkpoint:
- Present the full content clearly formatted
- Explain what will happen next
- Wait for explicit confirmation (yes/no)
- If user says no or suggests changes, revise and re-confirm
- Never proceed without confirmation

This ensures the user has full control over what gets created on GitHub.
