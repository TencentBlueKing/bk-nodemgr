# Release Helm PR Examples

## Scope

Use this reference when the task is a **version-only** Helm release follow-up for bk-nodemgr.

Expected touched files:

- `install/helm/bk-nodemgr/Chart.yaml`
- `install/helm/bk-nodemgr/values.yaml`
- `install/helm/mock-server/Chart.yaml`
- `install/helm/mock-server/values.yaml`

If the change needs more than these version bumps, stop and re-check scope.

## Branch naming examples

- `pr/helm-release-alpha17`
- `pr/helm-release-alpha18`
- `pr/helm-release-alpha19`

Prefer `pr/helm-release-<version-suffix>` and keep the pattern stable.

## Commit message examples

For bk-nodemgr chart:

```text
feat: bump bk-nodemgr helm chart to v3.0.1-alpha.17 --issue=#1502
```

For mock-server chart:

```text
feat: bump mock-server helm chart to v3.0.1-alpha.17 --issue=#1502
```

Keep one chart per commit.

## PR title example

```text
feat: bump helm release version to v3.0.1-alpha.17 --issue=#1502
```

Must stay English and must match the repo title rule:

```text
type: subject --issue=#number
```

## Minimal PR body example

```md
## Summary
- bump helm chart and values version to v3.0.1-alpha.17

refs #1502
```

## Issue reference safety

Default:

```text
refs #1502
```

Only switch to this if the user explicitly wants auto-close:

```text
closes #1502
```

## Quick decision table

| Situation | Action |
|---|---|
| Current branch only has the target bump and no unrelated history | Reuse current branch |
| Current branch has unrelated commits relative to `origin/master` | Create a clean branch from `origin/master` |
| PR only needs version alignment | Keep PR body minimal |
| User says do not close the issue | Use `refs #...` |
| User says merge should auto-close the issue | Use `closes #...` |
