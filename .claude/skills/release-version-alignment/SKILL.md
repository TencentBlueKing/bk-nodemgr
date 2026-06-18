---
name: release-version-alignment
description: Use when the user wants to start or prepare a bk-nodemgr version release, or verify/fix release version alignment for an existing version. Trigger on new-version language such as "prepare release", "next release", "准备新版本", "准备 alpha.N", and on alignment language such as "补 helm", "version not aligned", "missing chart update", "check if ready to tag". Do not use for unrelated Helm config changes, CI/CD failures, tag rewrite/movement, rollback, or broad release-process explanation.
---

# Release Version Alignment

Use this skill when preparing a new bk-nodemgr version release or verifying that release-alignment files match the target version.

This skill handles two scenarios:

1. **Primary (complete release PR)**: prepare a new version release by aligning Helm versions, API Gateway release metadata/resources, versioned changelog files, and PR metadata.
2. **Secondary (Helm-only follow-up)**: verify or fix missing Helm release-alignment fields for an already identified version.

## When to Use

**Primary scenario:**

- `我先要出一个新版本`
- `准备发布新版本`
- `准备 alpha.18`
- `开始准备下一个版本`
- `发布新版本`
- `准备一个新版本 release PR`

**Secondary scenario:**

- `给这个版本补一下 helm values`
- `release commit 还少 helm 改动`
- `把 alpha.17 的 chart 版本补上`
- `这个版本可以打 tag 了吗？`
- `helm 版本对齐了没？`
- `发版前检查一下 helm`
- `chart version、appVersion、image.tag 没对齐`
- `注意不要自动关闭这个 issue`
- `release 后补 helm PR`
- `chart 跟镜像版本没对齐`
- `补一下 release 遗漏的 helm 版本`
- `只补 install/helm 版本，不做别的`
- `发版尾差一个 helm follow-up`
- `检查下 helm 是不是漏了`
- `helm 版本号还没跟上`

Do not use this skill for tag surgery, CI/CD debugging, unrelated Helm configuration changes, rollback operations, or conceptual release-process explanations.

## Fastpath Rules

1. **Start from evidence, not guesswork.**
   Read `install/AGENTS.md`, then inspect the user-mentioned release commit, target version, or nearest recent release commit that shows the intended bump pattern.

2. **Separate complete release from Helm-only follow-up.**
   Do not apply the secondary 4-Helm-file constraint to a primary complete release PR. A complete release may legitimately include API Gateway files and versioned changelog files.

3. **Only touch release-alignment fields.**
   - `Chart.yaml`: `version`, `appVersion`
   - `install/helm/bk-nodemgr/values.yaml`: `image.tag`, `apiManagerImage.tag`, `apigwSync.config.release.version`, `apigwSync.config.release.comment`
   - `install/helm/mock-server/values.yaml`: `image.tag`
   - `apigw/definition.yaml`: `release.version`, `release.comment`
   - `support-files/changelog/{en,zh}/`: add the target version file only

4. **Treat API Gateway release version separately from app version.**
   `apigw/definition.yaml release.version` and `apigwSync.config.release.version` use the API Gateway release number, such as `2.14.1`. The app release string, such as `v3.0.1-alpha.40`, belongs in `release.comment`.

5. **Do not carry unrelated history into the PR.**
   If the current branch contains unrelated commits against the target `master` branch, create a clean branch and carry only the release-alignment changes.

6. **Keep commits and PR text intentionally short.**
   This PR is about release alignment, not feature explanation.

7. **Default to `refs`, not `closes`.**
   Only use `closes #...` when the user explicitly wants the issue auto-closed after merge.

8. **Stop when the target version or scope is unclear.**
   If you cannot quickly confirm the release commit, target version, API Gateway release version when needed, or expected file set, stop and ask instead of inventing a bump.

9. **Do not turn this into tag surgery.**
   If the request shifts to tag timing, remote tag movement, or release landing uncertainty, read `references/tag-workflow.md` and treat it as a different task.

## Workflow

### Step 1 - Determine target version and mode

**Scenario A: User requests a new version release**

1. Find the latest version tag:

   ```bash
   git tag --sort=-version:refname | head -1
   ```

2. Automatically infer the next version:
   - If latest is `v3.0.1-alpha.17` -> next is `v3.0.1-alpha.18`
   - If latest is `v3.0.1-beta.5` -> next is `v3.0.1-beta.6`
   - Pattern: increment the last numeric component

3. Confirm briefly with the user, then proceed without waiting for explicit confirmation:

   ```text
   准备 v3.0.1-alpha.18（当前最新：v3.0.1-alpha.17）
   ```

4. Use the **complete release PR** scope in Step 2.

**Scenario B: User provides a version or release/alignment commit**

Read `install/AGENTS.md`, then inspect the user-mentioned commit or version.

Extract only the facts you need:

- target app version string, such as `v3.0.1-alpha.40`
- whether the task is complete release PR or Helm-only follow-up
- which files were updated
- which fields changed: `version`, `appVersion`, `image.tag`, `apiManagerImage.tag`, `apigwSync.config.release.*`, `apigw/definition.yaml release.*`
- for complete release PRs, the API Gateway release version, such as `2.14.1`

If you cannot extract the target version, mode, or expected field set from evidence, stop and ask the user.

### Step 2 - Confirm the minimal target diff

**Primary complete release PR expected files:**

- `apigw/definition.yaml`
- `apigw/resources.yaml`
- `install/helm/bk-nodemgr/Chart.yaml`
- `install/helm/bk-nodemgr/values.yaml`
- `install/helm/mock-server/Chart.yaml`
- `install/helm/mock-server/values.yaml`
- `support-files/changelog/en/<version>_<YYYY-MM-DD>.md`
- `support-files/changelog/zh/<version>_<YYYY-MM-DD>.md`

Primary release field expectations:

- `apigw/definition.yaml`: `release.version`, `release.comment`
- `bk-nodemgr/Chart.yaml`: `version`, `appVersion`
- `bk-nodemgr/values.yaml`: `image.tag`, `apiManagerImage.tag`, `apigwSync.config.release.version`, `apigwSync.config.release.comment`
- `mock-server/Chart.yaml`: `version`, `appVersion`
- `mock-server/values.yaml`: `image.tag`
- changelog files: heading `## [Version: <target version>] - YYYY-MM-DD`

**Secondary Helm-only follow-up allowed files:**

- `install/helm/bk-nodemgr/Chart.yaml`
- `install/helm/bk-nodemgr/values.yaml`
- `install/helm/mock-server/Chart.yaml`
- `install/helm/mock-server/values.yaml`

Most Helm-only follow-ups update:

- `bk-nodemgr/Chart.yaml`: `version`, `appVersion`
- `bk-nodemgr/values.yaml`: `image.tag`, and `apiManagerImage.tag` when release evidence shows the apigw-sync image lagging
- `mock-server/Chart.yaml`: `version`, `appVersion`
- `mock-server/values.yaml`: `image.tag`

If evidence says the follow-up should touch fewer files, that is acceptable as long as every touched field is release-alignment-only. For example, a follow-up may only bump `install/helm/bk-nodemgr/values.yaml apiManagerImage.tag` when that is the sole lagging field.

If evidence says the task should touch files outside the mode-specific scope, stop and confirm with the user before expanding scope.

### Step 3 - Check branch cleanliness before editing

Check:

- current branch name
- working tree contents
- diff against the target `master` branch
- whether unrelated commits are already present

Use this rule:

- **Only target release-alignment changes in working tree, no unrelated commits** -> continue on current branch
- **Unrelated commits exist, or working tree has non-alignment changes** -> create clean branch

Unrelated commits means: commits that touch files outside the mode-specific scope or touch allowed files for reasons other than release alignment.

If a clean branch is needed, prefer a stable branch name such as:

- `pr/helm-release-alpha17`
- `pr/helm-release-alpha18`

### Step 4 - Edit alignment fields only

Apply the smallest possible alignment change. Do not modify chart dependency structure, unrelated values, templates, runtime settings, API Gateway routing policy, or changelog content unrelated to the target version.

Respect `install/AGENTS.md`:

- keep chart value naming backward-compatible when possible
- do not change chart dependency structure casually
- do not reintroduce deprecated Helm fields

### Step 5 - Commit in atomic units

For Helm-only follow-ups touching both charts, prefer two commits:

1. `bk-nodemgr` chart + values
2. `mock-server` chart + values

Recommended Helm-only commit messages:

- `feat: bump bk-nodemgr helm chart to vX.Y.Z-alpha.N --issue=#1234`
- `feat: bump mock-server helm chart to vX.Y.Z-alpha.N --issue=#1234`

For a single lagging field, use a narrow message:

- `chore: bump bk-nodemgr-apigw-sync image to vX.Y.Z-alpha.N`

For a complete release PR, keep release-alignment changes together unless the user asks for a different split.

Read `references/example-pr.md` when you need concrete branch, commit, or PR wording examples.

### Step 6 - Verify only what matters

Run the bundled verifier when the working tree represents the intended release-alignment diff:

```bash
bash .claude/skills/release-version-alignment/scripts/verify-alignment.sh <target-version> master
```

The verifier checks:

1. **File scope check**

   Expected:
   - complete release PR: only the primary file set listed in Step 2
   - Helm-only follow-up: only allowed Helm files, possibly a subset

2. **Version alignment check**

   Confirm every touched release field matches the appropriate version:
   - app version fields equal `X.Y.Z-alpha.N` or `vX.Y.Z-alpha.N`
   - `apiManagerImage.tag` is checked when present in the diff or when validating complete release alignment
   - API Gateway `release.version` fields match each other, while `release.comment` fields equal the target app version

3. **Diff sanity check**

   Expected: Helm diffs only change version/appVersion/tag/API Gateway release metadata lines, with no template, dependency, or runtime setting changes.

4. **Changelog check**

   Expected for complete release PRs: both localized changelog files exist and contain the target version heading:

   ```bash
   support-files/changelog/en/<target-version>_*.md
   support-files/changelog/zh/<target-version>_*.md
   ```

   `release.md` is historical and is not the default changelog target for this skill.

5. **Gateway resource sync check**

   Compare the previous release version and target release version as release evidence, not as permission to silently expand a Helm-only PR.

   Expected rule:
   - list every changed file matching `docs/api/swagger/backend/api/v3/*.swagger.json` in that version window
   - if any changed backend swagger file changes a gateway-visible contract (path, method, operationId, request/response schema, auth/resource extension, permission metadata, or description published through `apigw/resources.yaml`), then `apigw/resources.yaml` must also be updated in the same version window
   - `topo.swagger.json` is only one example; do not treat it as the only swagger file that can require API Gateway sync

   Important:
   - first report changed backend swagger files and the exact API Gateway impact decision
   - if a gateway-visible swagger change lacks an `apigw/resources.yaml` update, stop and report the release follow-up as incomplete
   - only add `apigw/resources.yaml` after the user confirms the scope or the user already asked for a complete release PR

If any check fails:

- For file scope, version alignment, diff sanity, or gateway resource sync failures: stop and report the mismatch.
- For changelog failures: delegate changelog generation using `changelog-doc`, wait for user confirmation, then resume verification.

Do not over-expand verification into unrelated Helm template analysis unless edited files themselves show a real problem.

### Step 7 - Handle missing changelog

If Step 6 Check 4 fails for a complete release PR:

1. Stop the current workflow; do not proceed to PR creation.
2. Delegate changelog generation:

   ```text
   task(
     category="writing",
     load_skills=["changelog-doc"],
     run_in_background=false,
     description="Generate changelog for version",
     prompt="Generate zh/en versioned changelog files for {TARGET_VERSION} under support-files/changelog/{zh,en}/. Use the changelog-doc skill."
   )
   ```

3. Wait for user confirmation of the generated changelog.
4. Resume from Step 6 and re-run verification.

### Step 8 - API Gateway resource sync

For complete release PRs, include API Gateway release metadata and resource sync when evidence requires it:

- `apigw/definition.yaml`: update `release.version` and `release.comment`
- `install/helm/bk-nodemgr/values.yaml`: update `apigwSync.config.release.version` and `apigwSync.config.release.comment`
- `apigw/resources.yaml`: update gateway-visible resource contracts when swagger/proto evidence requires it

If Step 6 Check 5 detects swagger changes that require `apigw/resources.yaml` updates:

1. Run the extraction script:

   ```bash
   bash .claude/skills/release-version-alignment/scripts/extract-backend-swagger-changes.sh {FROM_VERSION} {TO_VERSION}
   ```

2. Report the changed swagger files and exact API Gateway impact decision.

3. If the current task is Helm-only follow-up, stop and ask before adding API Gateway files. If the current task is a complete release PR, proceed only after confirming the required `apigw/resources.yaml` update.

Important:

- Not all swagger changes need to be synced to API Gateway.
- Users may need to adjust descriptions, permissions, or routing policies.
- If no gateway-visible swagger changes were detected, skip `apigw/resources.yaml` edits.

### Step 9 - Create the PR

Before creating the PR, read:

- `.github/workflows/pr-title-lint.yml`
- `.github/labeler.yml`

Current repo constraints that matter here:

- base branch is `master`
- PR title must match `type: subject --issue=#...`
- subject must be English
- `install/**/*` changes auto-apply `kind/core`

If a fork remote exists, prefer pushing the branch to the fork and opening the PR from fork -> upstream.

Default Helm-only PR title:

- `feat: bump helm release version to vX.Y.Z-alpha.N --issue=#1234`

Default complete-release PR title:

- `feat: bump release version to vX.Y.Z-alpha.N --issue=#1234`

Default PR body:

```md
## Summary

- align release version to vX.Y.Z-alpha.N

refs #1234
```

Keep the body short unless the user asks for more detail.

## Output Expectations

Report at least:

1. the reference release commit or release evidence used
2. the target version
3. the mode: complete release PR or Helm-only follow-up
4. which Helm files and fields were changed
5. whether `apiManagerImage.tag` was checked or changed
6. which API Gateway files and fields were changed, if any
7. which changelog files were created or verified
8. which backend swagger files changed, whether they require an `apigw/resources.yaml` sync check, the result, and whether the sync was completed
9. whether a clean branch was created
10. the commit list
11. the final branch name
12. the PR title / base / head
13. the PR URL
14. whether the issue reference uses `refs` or `closes`

## When Not to Use

Do not use this skill when:

- the user wants broad release-process explanation without preparing or checking a concrete version
- the request is really about tag rewrite, moving a pushed remote tag, or release landing recovery
- the request is really about CI/CD or pipeline failure diagnosis
- the target version or release evidence cannot be confirmed quickly
- the user's request mixes version alignment with unrelated Helm changes, such as ingress config changes
- the user wants to understand why version alignment is needed rather than execute or verify it

In those cases, switch to broader research, tag workflow, or debugging instead of forcing the fast path.

## References

Read these as needed:

- `install/AGENTS.md` - Helm scope constraints and anti-patterns
- `references/example-pr.md` - branch, commit, and PR wording examples
- `.github/workflows/pr-title-lint.yml` - PR title regex and English-subject requirement
- `.github/labeler.yml` - auto-label expectations for `install/**/*`
- `references/tag-workflow.md` - tag timing, landing order, and tag-move caution
- `create-github-pr/SKILL.md` - broader PR workflow when the fast path no longer fits
