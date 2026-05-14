---
name: release-version-alignment
description: Use when the user wants to initiate a new version release workflow or verify version alignment for an existing release. PRIMARY SCENARIO (80%): User expresses intent to start/prepare/publish a new version - automatically infers next version number and executes complete release preparation (Helm updates, changelog generation, API Gateway sync, PR creation). Trigger on action-oriented language about "new version", "next release", "prepare release", or "start version X" - even without explicit mention of Helm/changelog. SECONDARY SCENARIO (20%): User identifies missing or misaligned Helm version fields for a specific existing version - creates targeted alignment PR. Trigger on remedial language about "补 helm", "version not aligned", "missing chart update", or "check if ready to tag". DO NOT trigger for: Helm configuration changes unrelated to version bumps, CI/CD pipeline failures, tag surgery/rewrite, architecture changes to Helm structure, or rollback operations.
---

# Release Version Alignment

Use this skill when preparing a new version release or verifying that Helm charts are aligned with the target release version.

This skill handles two scenarios:

1. **Primary (80%)**: Preparing a new version release - automatically infers next version and creates complete release PR
2. **Secondary (20%)**: Version alignment check - verifies existing version alignment or creates follow-up PR

## When to Use

**Primary scenario (most common):**

- `我先要出一个新版本`
- `准备发布新版本`
- `准备 alpha.18`
- `开始准备下一个版本`
- `发布新版本`

**Secondary scenarios (version alignment checks):**

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
  Do not use this skill for broad release planning, tag repair, pipeline debugging, non-Helm release work, or requests that mix version alignment with other Helm changes.

## Fastpath Rules

1. **Start from evidence, not guesswork.**
   Read `install/AGENTS.md`, then inspect the user-mentioned release commit or the nearest recent release commit that shows the intended version bump pattern.

2. **Keep the scope version-only.**
   Default allowed files are:
   - `install/helm/bk-nodemgr/Chart.yaml`
   - `install/helm/bk-nodemgr/values.yaml`
   - `install/helm/mock-server/Chart.yaml`
   - `install/helm/mock-server/values.yaml`

3. **Only touch release-alignment fields.**
   - `Chart.yaml`: `version`, `appVersion`
   - `values.yaml`: `image.tag`

4. **Do not carry unrelated history into the PR.**
   If the current branch contains unrelated commits against the target `master` branch, create a clean branch and carry only the Helm bump.

5. **Keep commits and PR text intentionally short.**
   This PR is about version alignment, not feature explanation.

6. **Default to `refs`, not `closes`.**
   Only use `closes #...` when the user explicitly wants the issue auto-closed after merge.

7. **Stop when the target version or scope is unclear.**
   If you cannot quickly confirm the release commit, target version, or expected file set, stop and ask instead of inventing a bump.

8. **Do not turn this into tag surgery.**
   If the request shifts to tag timing, remote tag movement, or release landing uncertainty, read `references/tag-workflow.md` and treat it as a different task.

## Workflow

### Step 1 — Determine target version

**Scenario A: User requests "我先要出一个新版本" (primary scenario)**

1. Find the latest version tag:

   ```bash
   git tag --sort=-version:refname | head -1
   ```

2. Automatically infer the next version:
   - If latest is `v3.0.1-alpha.17` → next is `v3.0.1-alpha.18`
   - If latest is `v3.0.1-beta.5` → next is `v3.0.1-beta.6`
   - Pattern: increment the last numeric component

3. Confirm with user (brief, one-line):

   ```
   准备 v3.0.1-alpha.18（当前最新：v3.0.1-alpha.17）
   ```

4. Proceed directly to Step 2 without waiting for explicit confirmation.

**Scenario B: User provides specific version or commit (secondary scenario)**

Read `install/AGENTS.md`, then inspect the user-mentioned release commit or version.

Extract only the facts you need:

- target version string
- which Helm files were updated
- whether both charts are part of the follow-up
- which fields changed: `version`, `appVersion`, `image.tag`

Stop searching when you have extracted:

- target version string (e.g., v3.0.1-alpha.17)
- which Helm files were updated (Chart.yaml, values.yaml)
- which fields changed (version, appVersion, image.tag)

If you cannot extract all three from the reference commit, stop and ask the user.

### Step 2 — Confirm the minimal target diff

The default expected diff is a 4-file alignment bump:

- `install/helm/bk-nodemgr/Chart.yaml`
- `install/helm/bk-nodemgr/values.yaml`
- `install/helm/mock-server/Chart.yaml`
- `install/helm/mock-server/values.yaml`

Most follow-up PRs update:

- `bk-nodemgr/Chart.yaml`: `version`, `appVersion`
- `bk-nodemgr/values.yaml`: `image.tag`
- `mock-server/Chart.yaml`: `version`, `appVersion`
- `mock-server/values.yaml`: `image.tag`

If the evidence says the follow-up should touch fewer or different files, stop and confirm with the user before expanding or shrinking scope.

### Step 3 — Check branch cleanliness before editing

Check:

- current branch name
- working tree contents
- diff against the target `master` branch
- whether unrelated commits are already present

Use this rule:

- **Only target Helm changes in working tree, no unrelated commits** → continue on current branch
- **Unrelated commits exist, or working tree has non-Helm changes** → create clean branch

Unrelated commits means: commits that touch files outside `install/helm/**` or touch Helm files for reasons other than version alignment.

If a clean branch is needed, prefer a stable branch name such as:

- `pr/helm-release-alpha17`
- `pr/helm-release-alpha18`

### Step 4 — Edit Helm fields only

Apply the smallest possible alignment change. Do not modify chart dependency structure, unrelated values, or runtime settings that do not belong to this release follow-up.

Respect `install/AGENTS.md`:

- keep chart value naming backward-compatible when possible
- do not change chart dependency structure casually
- do not reintroduce deprecated Helm fields

### Step 5 — Commit in atomic units

If both charts are touched, prefer two commits:

1. `bk-nodemgr` chart + values
2. `mock-server` chart + values

Recommended commit messages:

- `feat: bump bk-nodemgr helm chart to vX.Y.Z-alpha.N --issue=#1234`
- `feat: bump mock-server helm chart to vX.Y.Z-alpha.N --issue=#1234`

If only one chart is part of the follow-up, a single commit is acceptable.

Read `references/example-pr.md` when you need concrete branch, commit, or PR wording examples.

### Step 6 — Verify only what matters

Run these checks:

1. **File scope check:**

   ```bash
   git diff master --name-only
   ```

   Expected: only the 4 Helm files, nothing else.

2. **Version alignment check:**
   Read the 4 files and confirm:
   - `bk-nodemgr/Chart.yaml`: `version` = X.Y.Z-alpha.N, `appVersion` = X.Y.Z-alpha.N
   - `bk-nodemgr/values.yaml`: `image.tag` = X.Y.Z-alpha.N
   - `mock-server/Chart.yaml`: `version` = X.Y.Z-alpha.N, `appVersion` = X.Y.Z-alpha.N
   - `mock-server/values.yaml`: `image.tag` = X.Y.Z-alpha.N

3. **Diff sanity check:**

   ```bash
   git diff master install/helm/
   ```

   Expected: only version/appVersion/image.tag lines changed, no template or dependency changes.

4. **Changelog check:**

   ```bash
   grep "## \[Version: $TARGET_VERSION\]" release.md
   ```

   Expected: `release.md` contains a changelog entry for the target version.

   If the changelog entry is missing:
   - Stop the workflow
   - Delegate to a subagent with `changelog-doc` skill to generate the changelog
   - Wait for user confirmation of the generated changelog
   - Resume verification after changelog is added

5. **Gateway resource sync check:**
   Compare the previous release version and the target release version as release evidence, not as the scope of the current Helm PR.

   Expected rule:
   - list every changed file matching `docs/api/swagger/backend/api/v3/*.swagger.json` in that version window
   - if any changed backend swagger file changes a gateway-visible contract (path, method, operationId, request/response schema, auth/resource extension, permission metadata, or description published through `apigw/resources.yaml`), then `apigw/resources.yaml` must also be updated in the same version window
   - `topo.swagger.json` is only one example; do not treat it as the only swagger file that can require API Gateway sync

   Example:
   - from alpha.N to alpha.N+1, if `node_proxy.swagger.json` changes `NodeProxyUpgradeReq.target_version`, `apigw/resources.yaml` must show the corresponding request schema update

   Important:
   - this is a release-completeness check, not permission to silently expand a Helm-only PR scope
   - first report the changed backend swagger files and the exact API Gateway impact decision
   - if a gateway-visible swagger change lacks an `apigw/resources.yaml` update, stop and report the release follow-up as incomplete; only add `apigw/resources.yaml` after the user confirms the scope or the user already asked for a complete release PR

If any check fails:

- For file scope, version alignment, diff sanity, or gateway resource sync failures: stop and report the mismatch
- For changelog failures: delegate to subagent to generate changelog using `changelog-doc` skill, wait for user confirmation, then resume

Do not over-expand verification into unrelated Helm template analysis unless the edited files themselves show a real problem.

### Step 7 — Handle missing changelog

If Step 6 Check 4 (changelog check) fails:

1. **Stop the current workflow** - do not proceed to PR creation
2. **Delegate changelog generation:**
   ```
   task(
     category="writing",
     load_skills=["changelog-doc"],
     run_in_background=false,
     description="Generate changelog for version",
     prompt="Generate changelog entry for version {TARGET_VERSION} in release.md.
             Use the changelog-doc skill to create a proper release note entry."
   )
   ```
3. **Wait for user confirmation** - the generated changelog must be reviewed and approved by the user
4. **Resume from Step 6** - re-run verification after changelog is confirmed

Only proceed to Step 8 (API Gateway resource sync) after all Step 6 checks pass, including changelog and gateway resource sync validation.

### Step 8 — API Gateway resource sync

If Step 6 Check 5 (gateway resource sync check) detected swagger changes that require `apigw/resources.yaml` updates:

1. **Run the extraction script:**

   ```bash
   bash .claude/skills/release-version-alignment/scripts/extract-backend-swagger-changes.sh {FROM_VERSION} {TO_VERSION}
   ```

   This extracts changed backend swagger files to `.diff/{FROM_VERSION}~{TO_VERSION}/docs/api/swagger/backend/api/v3/`

2. **Prompt user for manual update:**
   Display:

   ```
   Swagger changes detected. Please update apigw/resources.yaml based on:
   .diff/{FROM_VERSION}~{TO_VERSION}/docs/api/swagger/backend/api/v3/

   Changed files:
   - {list of changed swagger files}

   After updating apigw/resources.yaml, confirm to proceed.
   ```

3. **Wait for user confirmation** - do not proceed until user confirms the update is complete

4. **Commit the update:**

   ```bash
   git add apigw/resources.yaml
   git commit -m "feat: sync apigw resources with swagger changes from {FROM_VERSION} to {TO_VERSION}"
   ```

5. **Display commit info:**
   Show the commit hash and stats

Important:

- This step is **semi-automated** - extraction and commit are automatic, but the actual `apigw/resources.yaml` update requires human judgment
- Not all swagger changes need to be synced to API Gateway
- Users may need to adjust descriptions, permissions, or routing policies
- If no swagger changes were detected in Step 6 Check 5, skip this step entirely

### Step 9 — Create the minimal PR

Before creating the PR, read:

- `.github/workflows/pr-title-lint.yml`
- `.github/labeler.yml`

Current repo constraints that matter here:

- base branch is `master`
- PR title must match `type: subject --issue=#...`
- subject must be English
- `install/**/*` changes auto-apply `kind/core`

If a fork remote exists, prefer pushing the branch to the fork and opening the PR from fork -> upstream.

Default PR title:

- `feat: bump helm release version to vX.Y.Z-alpha.N --issue=#1234`

Default PR body:

```md
## Summary

- bump helm chart and values version to vX.Y.Z-alpha.N

refs #1234
```

Keep the body short unless the user asks for more detail.

## Output Expectations

Report at least:

1. the reference release commit or release evidence used
2. the target version
3. which Helm files and fields were changed
4. which backend swagger files changed, whether they require an `apigw/resources.yaml` sync check, the result, and whether the sync was completed
5. whether a clean branch was created
6. the commit list
7. the final branch name
8. the PR title / base / head
9. the PR URL
10. whether the issue reference uses `refs` or `closes`

## When Not to Use

Do not use this skill when:

- the user wants full release planning or broad release investigation
- the required changes go beyond Helm version alignment under `install/helm/**`
- the request is really about tag rewrite, moving a pushed remote tag, or release landing recovery
- the request is really about CI/CD or pipeline failure diagnosis
- the target version or release evidence cannot be confirmed quickly
- the user's request mixes version alignment with other Helm changes (e.g., "bump version and also fix the ingress config")
- the user wants to understand why version alignment is needed (use broader investigation instead)

In those cases, switch to a broader research or release workflow instead of forcing the fast path.

## References

Read these as needed:

- `install/AGENTS.md` - Helm scope constraints and anti-patterns
- `references/example-pr.md` - branch, commit, and PR wording examples
- `.github/workflows/pr-title-lint.yml` - PR title regex and English-subject requirement
- `.github/labeler.yml` - auto-label expectations for `install/**/*`
- `references/tag-workflow.md` - tag timing, landing order, and tag-move caution
- `create-github-pr/SKILL.md` - broader PR workflow when the fast path no longer fits
