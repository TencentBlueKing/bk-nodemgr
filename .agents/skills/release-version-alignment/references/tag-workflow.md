# Release Tag Workflow Notes

## Purpose

Use this reference when the user is deciding **where** to place a release tag and **when** the tag should be created relative to Helm version bump work.

## Core rule

Do **not** assume the `feat: release version ...` commit is the final tag target.

In this repo, the safer rule is:

> tag the final release landing commit, not the first release-version commit you notice.

## What this means in practice

Before creating a tag such as `v3.0.1-alpha.17`, confirm:

1. which commit is the actual release landing point
2. whether the required Helm version bumps are already included there
3. whether any final release-only follow-up commit still needs to land first

If Helm chart or values version bumps are still missing, the tag is probably premature.

## Release commit vs final tag target

Typical pattern:

- a release-related commit appears first
- follow-up release changes land after it
- the final tag is attached to the later commit

So this question matters:

```text
Does this commit represent the final releasable state, or only an earlier release step?
```

If the answer is unclear, do not tag yet.

## Minimal pre-tag checklist

Before tagging, check:

- target branch or target commit is explicitly known
- `install/helm/bk-nodemgr/Chart.yaml` is correct
- `install/helm/bk-nodemgr/values.yaml` is correct
- no final release follow-up commit is still pending

Only after these are true should the tag be created.

## Safe ordering

Preferred order:

1. confirm final release commit
2. confirm Helm version files are aligned
3. merge or prepare the minimal release follow-up PR if needed
4. create the release tag on the final landing commit

This avoids tagging too early and then having to move a published tag.

## If a tag was created too early

If the tag already exists on the wrong commit:

1. stop and identify the correct target commit first
2. confirm the user really wants the remote tag moved
3. warn that moving a pushed tag changes remote history
4. only then retarget the tag

Do not rewrite a pushed tag based on guesswork.

## Quick decision table

| Situation                                                       | Action                                               |
| --------------------------------------------------------------- | ---------------------------------------------------- |
| Release commit exists, but Helm version bump is still missing   | Do not tag yet                                       |
| Tag candidate commit already contains all final release changes | Tag this final commit                                |
| User only points to a release-related commit message            | Verify whether later release follow-up commits exist |
| Tag already pushed, but final release commit changed            | Ask before moving the remote tag                     |
| Unsure whether release state is final                           | Delay tagging and verify first                       |

## Lightweight tag note

If repo history shows release tags are lightweight tags, keep following that convention unless the user explicitly asks for an annotated tag.
