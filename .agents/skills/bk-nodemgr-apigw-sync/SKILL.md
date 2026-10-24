---
name: bk-nodemgr-apigw-sync
description: Use when a bk-nodemgr development task needs to adjust apigw-sync inputs or integration surfaces, including apigw/definition.yaml, apigw/resources.yaml, apigw/apidocs/, the apigw-sync image, Dockerfile, scripts, Helm values/templates, or build references; also when the agent needs the upstream bkpaas-python-sdk sync documentation. Not for executing gateway synchronization or release publication.
---

# bk-nodemgr apigw-sync

## Role

Help a bk-nodemgr developer adjust the repository's API Gateway synchronization inputs and integration surfaces. This is a lightweight project reference and routing skill: follow the concrete task, inspect only the relevant local sources, and query upstream documentation when the behavior belongs to `apigw-manager`.

Do not turn an adjustment request into a full consistency audit. Do not execute gateway synchronization, deletion, release, or publication operations.

## Local source map

Use the current checkout as the primary source of truth for this project. Start from the smallest relevant set:

| Surface               | Primary paths                                                                                            | Responsibility                                                                                                                 |
| --------------------- | -------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------ |
| Gateway definition    | `apigw/definition.yaml`                                                                                  | Gateway visibility, stages/backends, release metadata, docs, grants, related applications, and other definition-level settings |
| Gateway resources     | `apigw/resources.yaml`                                                                                   | OpenAPI paths/methods/schemas plus `x-bk-apigateway-resource` gateway policy and authentication metadata                       |
| Resource docs         | `apigw/apidocs/zh/`, `apigw/apidocs/en/`                                                                 | Operation-specific Chinese and English Markdown docs, normally keyed by `operationId`                                          |
| Sync image and script | `install/images/bk-nodemgr-apigw-sync/Dockerfile`, `install/images/bk-nodemgr-apigw-sync/support-files/` | Base image, bundled sync inputs, command arguments, and publication controls                                                   |
| Helm integration      | `install/helm/bk-nodemgr/values.yaml`, `install/helm/bk-nodemgr/templates/apigw-sync-*`                  | Deployment-time image and definition configuration, mounts, and sync Job wiring                                                |
| Build entrypoints     | `Makefile`                                                                                               | Image build context and API-doc packaging entrypoints                                                                          |

Read secondary entrypoints only when the task reaches them, such as the Docker Compose template, FAQ, CI, or release files. Do not assume a checked-in Swagger file automatically generates `apigw/resources.yaml`.

## Upstream reference

Keep these repository paths and default documentation entrypoints available for on-demand lookup:

- Repository: `TencentBlueKing/bkpaas-python-sdk`
- Configuration and resource-doc semantics: `sdks/apigw-manager/docs/sync_apigateway.md`
  - <https://github.com/TencentBlueKing/bkpaas-python-sdk/blob/master/sdks/apigw-manager/docs/sync_apigateway.md>
- Docker image, mounts, scripts, and command execution: `sdks/apigw-manager/docs/sync-apigateway-with-docker.md`
  - <https://github.com/TencentBlueKing/bkpaas-python-sdk/blob/master/sdks/apigw-manager/docs/sync-apigateway-with-docker.md>

Query only the relevant section or linked source. Do not copy either document into this skill or treat a `master` example, image tag, default argument, or newly documented feature as proof of compatibility with the image used by this checkout. When version behavior matters, compare the local Dockerfile and deployment image references before making a claim.

## Task routing

Keep the work at the apigw-sync integration boundary and hand off narrower responsibilities:

| Request                                                                       | Use                                                                    |
| ----------------------------------------------------------------------------- | ---------------------------------------------------------------------- |
| Add or revise an endpoint's standardized bilingual API reference              | `api-doc`                                                              |
| Prepare or verify release version fields and release-alignment scope          | `release-version-alignment`                                            |
| Call or debug an already deployed Node Manager API Gateway                    | `bk-nodemgr-api-debug`                                                 |
| Review generic service-startup config drift across Helm, helmfile, or Compose | `bk-nodemgr-config-sync` when that broader scope is actually requested |

Provide the relevant local paths and the concrete adjustment context when handing off. Do not duplicate another skill's full workflow inside this skill.

## Working pattern

1. Identify the developer's requested adjustment and select the smallest local surface from the source map.
2. Read the current local file and its direct build/deployment reference before relying on upstream examples.
3. Query the applicable upstream section only when it explains `apigw-manager` behavior, syntax, arguments, mounts, or compatibility.
4. Make the narrow change requested by the developer, preserving existing project naming and ownership boundaries.
5. Verify the changed local surface with the task-appropriate static or documentation check. Stop and report when the request depends on an unverified public contract or version assumption; do not invent a replacement.

The final response follows the task: describe the relevant paths, changes, evidence, handoffs, and unresolved decisions only when useful. There is no mandatory audit matrix or fixed maintenance report.

## Out of scope

- Running `sync-apigateway.sh`, a sync Job, Docker/Helm deployment, resource deletion, version creation, or gateway publication.
- Treating local checkout files as proof that a gateway stage is deployed or published.
- Replacing `api-doc`, `release-version-alignment`, or `bk-nodemgr-api-debug` with a second copy of their workflow.
- Automatically making `definition.yaml`, `resources.yaml`, Helm values, and deployment templates identical when their difference may be intentional.
