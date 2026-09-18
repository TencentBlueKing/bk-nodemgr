# bk-nodemgr Docs Taxonomy

Use this reference when the placement decision points to docs, API Gateway reference docs, or changelog artifacts.

## Discovery Steps

1. Read `docs/README.md` if present.
2. Read `docs/AGENTS.md` if present.
3. List the candidate docs subdirectory.
4. Read representative sibling docs before adding a new page.
5. Choose by audience, purpose, and reason-to-change.
6. Report adjacent areas that were rejected.

## docs/** Areas

Current verified `docs/` areas:

| Area                | Owns                                                                                | Common rejection                                                 |
| ------------------- | ----------------------------------------------------------------------------------- | ---------------------------------------------------------------- |
| `docs/api/`         | API design, API development flow, protocol guidance, non-reference API explanations | per-endpoint API Gateway reference pages                         |
| `docs/concepts/`    | durable domain concepts, terminology, model boundaries                              | endpoint examples or operational runbooks                        |
| `docs/developer/`   | contributor workflow, local development, templates, extension guidance              | public API reference pages or release notes                      |
| `docs/faq/`         | focused support questions and concise answers                                       | broad conceptual docs or temporary status                        |
| `docs/integration/` | external caller or adjacent platform integration behavior                           | upstream implementation mapping that belongs in third-party docs |
| `docs/operation/`   | install, deployment, runtime behavior, recovery, runbooks                           | developer-only extension details                                 |
| `docs/thirdparty/`  | upstream systems, external API mapping, verified third-party constraints            | local domain concepts not tied to upstream behavior              |
| `docs/img/`         | images and static assets for docs                                                   | standalone written content                                       |

Do not impose a docs area without checking current local files. Directory names are stable evidence, but nearby content decides the exact placement.

## API Gateway Reference Docs

Use `apigw/apidocs/{zh,en}` for API Gateway reference pages.

Choose this target when the content is:

- one API operation
- generated or aligned with Proto/API Gateway operation IDs
- zh/en endpoint reference documentation
- request/response field tables and examples
- version and permission metadata for API consumers

Hand off actual writing to `.agents/skills/api-doc/SKILL.md`.

Do not choose `apigw/apidocs/{zh,en}` for API design decisions, developer templates, conceptual explanations, or cross-endpoint architecture. Those belong under `docs/api/`, `docs/developer/`, or `docs/concepts/` depending on audience.

## Changelog Artifacts

Use `support-files/changelog/{zh,en}` for versioned changelog and release-note content.

Choose this target when the content is:

- a version change record
- release notes
- a Full Changelog link
- zh/en changelog page content

Hand off actual writing to `.agents/skills/changelog-doc/SKILL.md`. Use `.agents/skills/release-version-alignment/SKILL.md` when the task is preparing or validating a release version.

Do not choose normal docs for release notes unless the user is writing a durable release process guide rather than a specific version artifact.

## Decision Heuristics

- Vocabulary used across code, API, and docs belongs in `docs/concepts/`.
- Third-party caller guidance belongs in `docs/integration/`.
- Upstream system mapping belongs in `docs/thirdparty/`.
- Contributor setup and extension guidance belongs in `docs/developer/`.
- Deploy, operate, recover, or runtime behavior belongs in `docs/operation/`.
- Narrow support questions belong in `docs/faq/`.
- Endpoint reference pages belong in `apigw/apidocs/{zh,en}`.
- Version-specific release content belongs in `support-files/changelog/{zh,en}`.

## Output Requirement

When recommending docs, include:

- selected concrete area
- adjacent areas rejected
- local evidence that supports the selection
- handoff skill if actual writing is required
