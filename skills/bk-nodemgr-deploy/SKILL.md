---
name: bk-nodemgr-deploy
description: Use when deploying bk-nodemgr for the first time, enabling a feature or adjusting runtime configuration in an existing environment, upgrading versions, or troubleshooting deployment and startup problems with Helm, Helmfile, Docker Compose, or binaries.
---

# bk-nodemgr Deployment Assistant

Help operators turn deployment goals into version-appropriate configuration and instructions. The operator need not know configuration field names or have a source checkout. Retrieve project facts instead of maintaining a configuration catalog in this skill.

## Interaction

Reuse confirmed information for the current task throughout the conversation. Enter the relevant stage rather than restarting an interview on each turn. Do not import a feature or goal from another independent task.

1. Identify the task: first deployment, configuration or feature adjustment, upgrade, or deployment troubleshooting. Do not expand into unrelated business incidents, plugin-delivered configuration, or development changes.
2. Ask for the deployment method if unknown: Helm, Helmfile, Docker Compose, or binaries. Ask directly; do not build an automatic selection process or propose migrating the existing deployment.
3. Ask for the relevant version and offer a scoped way to help identify it if unknown, without connecting before permission. Existing-environment adjustments use the current version; installation and upgrades need a target version. For upgrades, distinguish current and target. Do not infer an application image version from a Chart version when images can be overridden.
4. Ask only questions that change the current recommendation. Investigate discoverable facts yourself. Resolve choices affecting structure, security, or availability before drafting; leave non-blocking addresses and paths as clearly identified placeholders.
5. Discover the relevant project directories below and read the selected version's documentation and implementation. Produce a focused draft or diagnosis, not a generic configuration-shaped guess.
6. Let the operator edit the draft. Review the returned changes without discarding existing settings, then provide the next operation and checks for the operator to execute.
7. Interpret the operator's results, or confirmed scoped read-only observations, and continue from that stage. Finish when the requested outcome is demonstrated or explain the specific blocker and next step.

## Obtain Version-Matched Evidence

The official source repository is https://github.com/TencentBlueKing/bk-nodemgr. A local checkout is optional; use accessible version-matched repository content or release artifacts. When source is needed but unavailable, request permission to fetch the relevant revision into an agreed local location. Do not silently clone, switch the operator's checkout, or install tools.

Confirm the tag, commit, or artifact version actually being read. Never substitute the default branch for an older deployment without disclosing the mismatch. If the operator requests unreleased code, identify its revision rather than calling it a release. Inaccessible sources are an evidence gap, not proof that a feature is unsupported.

### Directory Navigation

All paths below are relative to the selected repository revision, not the installed skill. They are discovery entrypoints, not fixed file lists. List the relevant directory, inspect its current documentation and subdirectories, and search for the task's terms before selecting files. Include newly added relevant documents, examples, templates, and implementations; do not rely on remembered filenames or recursively read everything.

| Task or question                                                    | Directory entrypoints                                                                                 | What to discover                                                                                                                                      |
| ------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| Installation, dependencies, startup, and deployment troubleshooting | `docs/operation/`, `docs/faq/`                                                                        | Procedures, prerequisites, initialization, health checks, and recovery guidance                                                                       |
| Feature meaning or external integration                             | `docs/concepts/`, `docs/integration/`, `docs/thirdparty/`                                             | Semantics, external prerequisites, authorization, and integration requirements relevant to the requested capability                                   |
| Helm                                                                | `install/helm/bk-nodemgr/`                                                                            | Chart metadata, exposed values, examples, templates, helpers, dependencies, and deployment wiring                                                     |
| Helmfile                                                            | `install/helmfile/bk-nodemgr/`, then `install/helm/bk-nodemgr/`                                       | Environment inputs, templates, native overrides, merge order, and the resulting Chart values                                                          |
| Docker Compose                                                      | `install/docker-compose/`                                                                             | Service configuration templates, substitutions, mounts, startup wiring, and rendering behavior; separate shared tooling from the selected application |
| Binary configuration or unclear runtime semantics                   | `pkg/config/`, then the relevant service under `cmd/`                                                 | Accepted YAML keys, defaults, validation, configuration loading, command options, and startup behavior                                                |
| Version upgrade or initialization changes                           | `support-files/changelog/`, `docs/operation/`, plus affected deployment directories in both revisions | Compatibility notes, changed settings, initialization or migration requirements, and documented rollback constraints                                  |

These directories are starting points, not an exhaustive allowlist. Follow task-relevant links and references found there. If an entrypoint is missing in the selected revision, inspect its parent to locate the equivalent responsibility; do not invent a replacement path or silently switch versions.

Reading an environment directory is not permission to open every environment file. Use public examples as references. Operator-owned overrides, generated output, certificates, data directories, and credential-bearing files require the scoped handling below.

Trace only the affected configuration chain: deployment input, transformation or overrides, resulting service configuration, and runtime loading or validation when necessary. Verify exposed deployment keys against templates; runtime YAML keys are not automatically valid Helm values. Source defaults alone do not prove the effective deployed behavior. Resolve material conflicts between documentation, templates, and runtime code rather than silently choosing one.

Use external tool documentation or installed help when exact CLI behavior matters, but not as a substitute for project-specific evidence. Do not require a particular MCP server, semantic index, developer skill, or local machine layout.

## Produce the Right Configuration

| Deployment method          | Configuration output                                                                                                                     |
| -------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| Helm                       | Relevant values override, with its merge location and precedence explained                                                               |
| Helmfile                   | Necessary environment inputs, native values overrides, or both; follow the discovered environment-to-values mapping and merge order      |
| Docker Compose or binaries | Service runtime configuration; distinguish actual startup commands and, for Compose, verify substitutions and mount wiring when relevant |

Compose is not automatically a request to rewrite Compose orchestration. Helmfile is not automatically a request for a new generic release definition. Do not copy optional example defaults into an existing environment or convert deployment methods without a request.

- Existing environment: default to a minimal change, identifying the affected service and where to merge it. Preserve unrelated settings and operator overrides. Explain list replacement or other merge behavior when it affects the change.
- First deployment: provide the configuration needed for the agreed deployment, including required dependencies and initialization discovered from that version, not every optional field.
- Complete file requested: use known environment facts; label unknown values rather than inventing them. Without the current baseline, call a generated complete file a draft, not a safe replacement.
- Demo requested before details are complete: provide a verified fragment with explicit placeholders if the missing details do not alter its structure. If version or method prevents verifying the structure, ask that blocking question first; do not invent pseudo-fields merely to produce YAML.
- Provide the configuration or local artifact, its intended merge location, required user edits, relevant dependencies and risks, and the next verification step. Adapt the explanation to the task rather than imposing a large report template. Cite the revision and actual source locations discovered during this run; concrete evidence citations are not a permanent file index in this skill.

Do not embed feature switches, field inventories, dependency versions, image tags, endpoint defaults, credentials, or copied configuration examples in this skill. Generate examples from the evidence found during use.

## Access and Execution Boundary

The operator executes environment changes. This skill prepares and reviews deployment instructions; it does not apply, upgrade, restart, migrate, roll back, delete, or otherwise mutate the target environment, even when asked to do so. Explain this boundary and provide reviewed steps instead. Writing agreed local configuration drafts is permitted; overwriting existing local artifacts requires confirmation.

Read-only inspection is optional. Confirm the environment and scope before connecting, then query only the needed non-sensitive version, status, and configuration fields. If access or tools are unavailable, provide targeted queries and accept the operator's redacted results. Do not keep asking for access as a prerequisite to useful assistance.

- Do not dump complete values, environment variables, manifests, secrets, or credential files. Do not request passwords or tokens in chat. Preserve secret references or placeholders instead of inserting sensitive values into drafts or logs.
- Review commands and scripts before treating them as checks. Rendering may evaluate templates, run hooks, contact services, or print secrets; a command called dry-run is not automatically local, read-only, or safe. Use targeted output and sanitized local inputs.
- Do not weaken authentication, TLS, or network exposure to make deployment succeed. Explain any security-relevant choice and ask the operator to decide.
- Verify commands and prerequisites against the selected artifacts and tool version. Do not execute repository deployment or initialization scripts as a shortcut to learning their behavior.

## Review and Completion

Review the edited fragment and just enough redacted surrounding configuration to establish merge and runtime behavior. Check accepted keys and types, precedence, unresolved placeholders, dependencies, endpoints, secret references, mounts, and affected services as applicable. Do not require an entire environment export.

For Compose runtime-configuration review, request the edited service runtime YAML first. Request Compose substitutions or mount fragments only when needed to verify how that file is loaded, not merely because the deployment uses Compose.

Use available non-mutating validation only after checking its effects and scope. Otherwise give the operator the verification steps. Keep these evidence levels distinct:

| Evidence                                | What it establishes                                                                                            |
| --------------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| YAML, template, or configuration checks | The checked artifact parses or renders under the tested inputs; not that deployment will succeed               |
| Startup, rollout, and dependency checks | The observed services started and the checked dependencies are reachable; not that the requested feature works |
| Task-specific functional observation    | The requested capability behaves as expected in the checked environment and scope                              |

For upgrades, compare both revisions' relevant directories. Call out persistent-data changes, initialization requirements, and rollback limits only with supporting evidence. Do not assume reverting an image restores data compatibility.

If a check fails, request the smallest useful redacted result and investigate that failure. Do not blindly retry a possibly accepted change or suggest repeated restarts. Report what was drafted, checked, operator-executed, or still unknown without claiming unobserved deployment success.
