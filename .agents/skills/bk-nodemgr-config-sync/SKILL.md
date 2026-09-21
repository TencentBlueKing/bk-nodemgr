---
name: bk-nodemgr-config-sync
description: Use when editing or reviewing bk-nodemgr service startup config in pkg/config, Helm values/templates, helmfile environment templates, or docker-compose config templates; especially when pkg/config fields, YAML tags, defaults, validation, or deployment config surfaces may drift.
---

# bk-nodemgr Config Sync

## Overview

`pkg/config` is the runtime contract for bk-nodemgr service startup configuration. Helm, helmfile, and docker-compose are deployment adapters that must generate YAML compatible with that contract.

Use this skill in two modes:

- **Implementation mode**: when changing startup config, update every required config surface.
- **Review mode**: when reviewing startup config changes, verify every required surface matches and report drift.

This skill uses an evidence checklist. Do not invent an automated checker, CI gate, or render pipeline unless the user explicitly asks for one.

## Scope

This skill covers service startup config only:

- `pkg/config/application.go`
- `pkg/config/backend.go`
- `pkg/config/file.go`
- `pkg/config/relay.go`
- service startup wiring in `cmd/*`
- Helm chart config surfaces under `install/helm/bk-nodemgr/`
- helmfile environment mapping under `install/helmfile/bk-nodemgr/`
- docker-compose local config surfaces under `install/docker-compose/bk-nodemgr/`

Do not use this skill for:

- config policy business configuration
- plugin package configuration
- node/plugin runtime-delivered configuration
- unrelated files that merely contain `config` in the name

## Mode Selection

Use **Implementation mode** when:

- the user asks to add, change, remove, or rename a `pkg/config` startup config field
- a task changes YAML tags, defaults, or `Validate()` behavior in `pkg/config`
- the user asks to sync Helm, helmfile, or docker-compose after a config change
- the task is a bugfix caused by a missing deployment config value

Use **Review mode** when:

- the user asks for review and the diff touches startup config surfaces
- a PR changes `pkg/config/**`
- a PR changes Helm configmaps, values, deployment config wiring, helmfile templates, or docker-compose config templates
- the user asks whether config changes line up across deployment surfaces

Review mode is read-only. Do not modify files in Review mode.

## Config Surface Map

| Service     | Runtime config              | Helm values/configmap                                                                                 | Helm deployment                                                 | Helmfile                                                                                                                     | Docker Compose config                                                               | Compose service                                                                                                           |
| ----------- | --------------------------- | ----------------------------------------------------------------------------------------------------- | --------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------- |
| application | `pkg/config/application.go` | `install/helm/bk-nodemgr/values.yaml`, `install/helm/bk-nodemgr/templates/application-configmap.yaml` | `install/helm/bk-nodemgr/templates/application-deployment.yaml` | `install/helmfile/bk-nodemgr/templates/application.yaml.gotmpl`, `install/helmfile/bk-nodemgr/environments/example/env.yaml` | `install/docker-compose/bk-nodemgr/templates/config/bk-nodemgr-application.yml.tpl` | `install/docker-compose/bk-nodemgr/templates/service/bk-nodemgr.yml.tpl`, `install/docker-compose/bk-nodemgr/serviced.sh` |
| backend     | `pkg/config/backend.go`     | `install/helm/bk-nodemgr/values.yaml`, `install/helm/bk-nodemgr/templates/backend-configmap.yaml`     | `install/helm/bk-nodemgr/templates/backend-deployment.yaml`     | `install/helmfile/bk-nodemgr/templates/backend.yaml.gotmpl`, `install/helmfile/bk-nodemgr/environments/example/env.yaml`     | `install/docker-compose/bk-nodemgr/templates/config/bk-nodemgr-backend.yml.tpl`     | `install/docker-compose/bk-nodemgr/templates/service/bk-nodemgr.yml.tpl`, `install/docker-compose/bk-nodemgr/serviced.sh` |
| file        | `pkg/config/file.go`        | `install/helm/bk-nodemgr/values.yaml`, `install/helm/bk-nodemgr/templates/file-configmap.yaml`        | `install/helm/bk-nodemgr/templates/file-deployment.yaml`        | `install/helmfile/bk-nodemgr/templates/file.yaml.gotmpl`, `install/helmfile/bk-nodemgr/environments/example/env.yaml`        | `install/docker-compose/bk-nodemgr/templates/config/bk-nodemgr-file.yml.tpl`        | `install/docker-compose/bk-nodemgr/templates/service/bk-nodemgr.yml.tpl`, `install/docker-compose/bk-nodemgr/serviced.sh` |
| relay       | `pkg/config/relay.go`       | no standard chart surface currently                                                                   | no standard deployment surface currently                        | no standard helmfile surface currently                                                                                       | no standard compose surface currently                                               | no standard compose surface currently                                                                                     |

Do not invent Relay Helm, helmfile, or docker-compose surfaces unless the user explicitly asks to add relay deployment support.

## Field Classification

Classify each config change before syncing it.

| Class                    | Meaning                                                                                                                                           | Sync rule                                                               |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| Deployment-configurable  | Operators or local developers should set the value: endpoint, port, auth, token/JWT, TLS, path, repo, feature switch, logging, tracing, profiling | Sync all relevant surfaces by default                                   |
| Runtime-only             | Internal default or validation behavior that should not become deployment input                                                                   | Keep in `pkg/config`; record omission reason                            |
| Kubernetes-runtime value | Pod IP, pod name, release-generated service DNS, mounted cert file path, or value filled by initContainer                                         | Keep in Helm deployment/init logic; do not copy mechanically to Compose |
| Compose-local value      | LAN IP, local port, local path, or placeholder used for quick local experience                                                                    | Sync Compose only when it affects local startup or experience           |

Docker Compose is not mechanically equivalent to Helm. Sync Compose when the field affects local startup or local experience, especially ports, endpoints, auth, token/JWT, TLS, paths, repo, feature switches, tracing, logging, or profiling.

## Known Intentional Omissions

These are allowed only when current file evidence confirms the pattern:

- `tracing.instanceID`: Helm ConfigMap omits it because Deployment injects `__TRACING_INSTANCEID__` from pod metadata or explicit values.
- `advertiseIPV4` and `advertiseIPV6`: Helm ConfigMap may keep placeholders and Deployment initContainer fills pod IPs or explicit overrides.
- `gracefulShutdownTimeoutSec`: Helm helper may derive it from `terminationGracePeriodSeconds / 2` when absent.
- `workflow.gracefulShutdownTimeoutSeconds`: Helm helper may derive it from `terminationGracePeriodSeconds / 2` when absent.
- `RelayService`: no standard Helm, helmfile, or docker-compose surface currently.

Do not use this list blindly. Cite the current path and behavior when relying on an omission.

## Implementation Mode

1. Identify the changed service config struct and YAML path.
   - Read the relevant `pkg/config/<service>.go` fields, YAML tags, constructor defaults, and `Validate()` behavior.
   - Check `cmd/<service>` startup flags when the config file path or startup command changes.
2. Classify the field using the field classification table.
3. Update `pkg/config` first.
   - Keep defaults in `New*Service()`.
   - Keep startup validation in `Validate()`.
   - Preserve existing YAML tag naming style.
4. Sync Helm.
   - Update `install/helm/bk-nodemgr/values.yaml` when the chart should expose the value.
   - Update `install/helm/bk-nodemgr/templates/<service>-configmap.yaml` when the service YAML needs the field.
   - Update `install/helm/bk-nodemgr/templates/<service>-deployment.yaml` only for runtime placeholders, ports, mounts, env, volumes, or startup command changes.
   - Update `install/helm/bk-nodemgr/README.md` only when chart user-facing behavior changes.
5. Sync helmfile.
   - Update `install/helmfile/bk-nodemgr/templates/<service>.yaml.gotmpl` when helmfile should pass the value into chart values.
   - Update `install/helmfile/bk-nodemgr/environments/example/env.yaml` for every helmfile-exposed value.
   - Do not force every environment override to grow the field unless an existing override already needs alignment.
6. Sync docker-compose when the field affects local startup or local experience.
   - Update `install/docker-compose/bk-nodemgr/templates/config/bk-nodemgr-<service>.yml.tpl`.
   - Update `install/docker-compose/bk-nodemgr/bk-nodemgr.env` only when a new placeholder is needed and it is safe to edit.
   - Update `install/docker-compose/bk-nodemgr/templates/service/*.tpl` or `serviced.sh` only for ports, mounts, env, volumes, or startup process changes.
7. Produce a Config Surface Matrix before final response.

If the user explicitly limits scope, such as "only change Helm", respect that scope. Still report the unsynced Helmfile and docker-compose risks in the final response.

## Review Mode

Review mode is a static evidence check by default. Do not render Helm, helmfile, or docker-compose unless the user asks or the diff specifically risks template syntax or generation failure.

For every changed config field, answer:

- Which service owns it?
- What YAML path does it map to?
- Is the field/default/validation present in `pkg/config`?
- Is Helm values exposure correct when needed?
- Is Helm ConfigMap rendering correct?
- Is Helm Deployment wiring correct when the field affects runtime placeholders, ports, mounts, env, volumes, or startup command?
- Is helmfile gotmpl mapping correct?
- Is helmfile example env present?
- Is docker-compose template coverage correct if the field affects local experience?
- Is any omission intentional, and what file evidence proves it?

### Severity

Use surface-sensitive severity:

- Helm drift is usually `High` because it affects the primary chart deployment surface.
- Helmfile drift is usually `Medium` because helmfile is a production environment mapping entrypoint.
- Docker Compose drift is usually `Low`, but use `Medium` when it breaks local startup or a core local experience path.
- If a runtime default fully preserves behavior but the value is no longer configurable on a deployment surface, report at least `Medium` for Helm and at least `Low` for Compose.

## Required Matrix

Always produce this matrix in Implementation mode. Use it in Review mode when any drift is found or when the user asks for a complete check.

| Service | YAML path      | pkg/config                   | Helm values | Helm ConfigMap | Helm Deployment | Helmfile gotmpl | Helmfile example env | Docker Compose | Status             |
| ------- | -------------- | ---------------------------- | ----------- | -------------- | --------------- | --------------- | -------------------- | -------------- | ------------------ |
| backend | `example.path` | yes: `pkg/config/backend.go` | yes/no/n/a  | yes/no/n/a     | yes/no/n/a      | yes/no/n/a      | yes/no/n/a           | yes/no/n/a     | ok/missing/omitted |

Status rules:

- `ok`: required surface is present and consistent.
- `missing`: required surface is absent or inconsistent.
- `omitted`: absence is intentional and backed by file evidence.

## Implementation Final Response

Report in Chinese main text with exact paths:

- changed runtime config fields
- synced Helm files
- synced helmfile files
- synced docker-compose files
- intentional omissions with reasons
- Config Surface Matrix
- verification performed or not performed

## Review Final Response

Use code review style:

- findings first, ordered by severity
- each finding includes `file:line`, source field, YAML path, missing surface, and runtime impact
- then include checked surfaces
- if no drift is found, say no config surface drift was found and list checked surfaces

## Common Mistakes

- Treating Helm, helmfile, and docker-compose as documentation instead of deployment adapters.
- Updating `pkg/config` and Helm ConfigMap but forgetting helmfile example env.
- Copying Kubernetes-runtime placeholders into docker-compose templates.
- Marking `RelayService` as missing Helm/Compose support when no relay deployment surface exists.
- Reporting known omissions without checking the current files.
- Silently honoring a user-limited scope without reporting unsynced config surface risk.

## Validation Prompts

Project-specific prompts live in `evals/evals.json`. They cover implementation sync, review drift detection, and user-limited scope handling.
