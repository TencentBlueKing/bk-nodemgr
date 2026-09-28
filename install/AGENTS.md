# INSTALL KNOWLEDGE BASE

## OVERVIEW

`install/` contains deployment assets, especially Helm charts and container image support files.

## STRUCTURE

```
install/
|- helm/bk-nodemgr/**            # chart, values, templates, nested dependency charts
|- images/**                     # Dockerfiles and image build assets
`- docker-compose/
   |- generate.sh                # shared template rendering utility
   `- bk-nodemgr/**              # bk-nodemgr docker-compose deployment
```

## WHERE TO LOOK

| Task                         | Location                               | Notes                                           |
| ---------------------------- | -------------------------------------- | ----------------------------------------------- |
| Helm values/template changes | `install/helm/bk-nodemgr/**`           | Main chart plus redis/mongodb/etcd dependencies |
| Image build context assets   | `install/images/**`                    | Dockerfile variants by image target             |
| bk-nodemgr docker-compose    | `install/docker-compose/bk-nodemgr/**` | nodemgr service deployment                      |
| Shared template rendering    | `install/docker-compose/generate.sh`   | Used by docker-compose deployment templates     |

## CONVENTIONS

- Keep chart value naming backward-compatible when possible.
- Track upstream deprecations in chart docs/values and avoid reintroducing removed settings.
- Align build artifacts with root Make targets (`docker-build-server`, `docker-build-apigw-sync`).

## ANTI-PATTERNS

- Do not hardcode runtime values that belong in chart values files.
- Do not change chart dependency structure without validating nested `charts/*` templates.
- Do not ignore documented deprecated Helm fields when editing values.
