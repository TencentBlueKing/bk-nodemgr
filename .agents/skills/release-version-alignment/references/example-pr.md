# Release Alignment PR Examples

## Scope

Use this reference when the task is a bk-nodemgr release-alignment PR.

There are two modes:

1. **Complete release PR**: preparing a new version release.
2. **Helm-only follow-up**: fixing missing Helm alignment after the target version is already known.

Do not use Helm-only limits to reject complete release files.

## Complete Release PR

Expected touched files:

- `install/helm/bk-nodemgr/Chart.yaml`
- `install/helm/bk-nodemgr/values.yaml`
- `support-files/changelog/en/vX.Y.Z-alpha.N_YYYY-MM-DD.md`
- `support-files/changelog/zh/vX.Y.Z-alpha.N_YYYY-MM-DD.md`

Conditional API Gateway files:

- `apigw/resources.yaml` only when gateway-visible resources changed.
- `apigw/definition.yaml` in complete release PRs to align release metadata.

Expected fields:

- `install/helm/bk-nodemgr/Chart.yaml`: `version`, `appVersion`
- `install/helm/bk-nodemgr/values.yaml`: `image.tag`, `apiManagerImage.tag`
- changelog heading: `## [Version: vX.Y.Z-alpha.N] - YYYY-MM-DD`
- API Gateway metadata: `apigw/definition.yaml release.version/comment` and `install/helm/bk-nodemgr/values.yaml apigwSync.config.release.version/comment`

Important: complete release PRs always align API Gateway metadata to the target system version without the leading `v`. For target `v3.0.1-alpha.65`, set all four API Gateway metadata fields to `3.0.1-alpha.65`. `apigw/resources.yaml` remains conditional and changes only when gateway-visible resources changed.

Complete-release PR title example:

```text
feat: bump release version to v3.0.1-alpha.40 --issue=#1502
```

Complete-release PR body example:

```md
## Summary

- align release version to v3.0.1-alpha.40

refs #1502
```

## Helm-only Follow-up

Allowed touched files:

- `install/helm/bk-nodemgr/Chart.yaml`
- `install/helm/bk-nodemgr/values.yaml`

A Helm-only follow-up may touch a subset of these files when evidence shows only one lagging field. For example, `install/helm/bk-nodemgr/values.yaml apiManagerImage.tag` may be the only field that needs a bump.

Expected fields:

- `install/helm/bk-nodemgr/Chart.yaml`: `version`, `appVersion`
- `install/helm/bk-nodemgr/values.yaml`: `image.tag`, `apiManagerImage.tag`

If the change needs unrelated Helm settings, templates, dependencies, or API Gateway files in Helm-only mode, stop and re-check scope with the user.

## Branch Naming Examples

- `pr/helm-release-alpha17`
- `pr/helm-release-alpha18`
- `pr/helm-release-alpha40`

Prefer `pr/helm-release-<version-suffix>` and keep the pattern stable.

## Commit Message Examples

For bk-nodemgr chart follow-up:

```text
feat: bump bk-nodemgr helm chart to v3.0.1-alpha.17 --issue=#1502
```

For a single lagging apigw-sync image field:

```text
chore: bump bk-nodemgr-apigw-sync image to v3.0.1-alpha.40
```

Complete release PR changes may stay together unless the user asks for a different split.

## PR Title Examples

Complete release:

```text
feat: bump release version to v3.0.1-alpha.40 --issue=#1502
```

Helm-only follow-up:

```text
feat: bump helm release version to v3.0.1-alpha.17 --issue=#1502
```

Must stay English and must match the repo title rule:

```text
type: subject --issue=#number
```

## Minimal PR Body Example

```md
## Summary

- align release version to v3.0.1-alpha.40

refs #1502
```

## Issue Reference Safety

Default:

```text
refs #1502
```

Only switch to this if the user explicitly wants auto-close:

```text
closes #1502
```

## Quick Decision Table

| Situation                                                             | Action                                                                             |
| --------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| User asks to prepare a concrete new version release                   | Use complete-release mode                                                          |
| User asks only to supplement missing Helm version fields              | Use Helm-only follow-up mode                                                       |
| Current branch only has the target alignment and no unrelated history | Reuse current branch                                                               |
| Current branch has unrelated commits relative to `origin/master`      | Create a clean branch from `origin/master`                                         |
| Helm-only follow-up only needs `apiManagerImage.tag`                  | Touch only that field                                                              |
| Complete release needs localized changelog                            | Use `support-files/changelog/{en,zh}/vX.Y.Z-alpha.N_YYYY-MM-DD.md`                 |
| Complete release has no `apigw/resources.yaml` change                 | Align API Gateway metadata to `vX.Y.Z-alpha.N` without `v`; do not touch resources |
| `apigw/resources.yaml` changed in the release window                  | Update resources and align API Gateway metadata to `vX.Y.Z-alpha.N` without `v`    |
| User says do not close the issue                                      | Use `refs #...`                                                                    |
| User says merge should auto-close the issue                           | Use `closes #...`                                                                  |
