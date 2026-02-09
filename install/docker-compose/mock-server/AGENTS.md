# MOCK-SERVER DOCKER-COMPOSE KNOWLEDGE BASE

## OVERVIEW

Independent docker-compose deployment for mock-server, simulating third-party APIs (CMDB, BK-Repo) for development and testing.

## STRUCTURE

```
mock-server/
|- deploy.sh                              # generate/install/uninstall/clean
|- mock-server.env                        # environment variables
|- templates/config/mock-server.yml.tpl   # application config template
|- templates/service/mock-server.yml.tpl  # docker-compose template
`- .gitignore
```

## CONVENTIONS

- Fully independent from bk-nodemgr, no shared networks or dependencies.
- `generate.sh` is shared at `../generate.sh`, do not duplicate.
- Environment variable prefix: `MOCK_SERVER_` (not `BK_MOCK_SERVER_`).
- Generate artifacts (`etc/mock-server.yml`, `mock-server.yml`) must not be committed.
- `mockData` in env is single-line JSON, rendered as-is into YAML config.

## COMMANDS

```bash
bash deploy.sh generate    # render templates only
bash deploy.sh install     # generate + start container
bash deploy.sh uninstall   # stop container
bash deploy.sh clean       # stop + remove generated files
```
