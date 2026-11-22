# Helm Chart 升级指南

本文面向升级已有 bk-nodemgr Helm 部署的运维人员，集中说明需要人工处理的 values 和部署行为变化。首次安装及当前参数说明见 [Chart README](README.md)。

## 如何使用本指南

- 以目标 Chart 版本查找对应章节；跨版本升级时，按版本从旧到新检查中间各节，不要只读目标版本一节。
- 版本章节使用 `Chart.yaml` 的 `version`，不是 Helm revision 或镜像标签。单独覆盖 `image.tag`、模块镜像或 `apiManagerImage.tag` 时，还需确认镜像支持目标配置。
- 本指南记录需要用户操作的字段改名、移除、类型变化、默认行为变化，以及依赖、Hook 和存储的升级要求；不是所有新增参数的清单。完整默认值以目标 Chart 包内的 `values.yaml` 和依赖 Chart 为准，产品变化另见 [版本变更记录](https://github.com/TencentBlueKing/bk-nodemgr/tree/master/support-files/changelog/zh)。在线阅读时切换到目标版本 tag，不要混用 `master` 与发行包配置。
- 没有对应章节不表示任意旧版本可以直接升级；仍需对比源版本与目标版本的默认值、模板和依赖，并检查产品变更记录中的数据兼容性要求。

## 通用升级检查

### 维护完整的环境覆盖

环境文件只维护实际需要的自定义配置，不复制整份默认 values。升级前将字段迁移、需要保留的旧默认行为和当前环境的全部覆盖合入 `values-env.yaml`，而不是只写本次修改的片段。

- map 递归合并，列表整体替换。覆盖 `apigwSync.extraEnvVars`、`apigwSync.config.stages` 或 `backend.config.gseDeployConfs` 等列表时，需要保留全部所需条目；列表项不会按 `name` 自动合并，也不会自动获得目标版本新增的默认条目。
- 默认值变化也可能改变未显式配置的环境。需要保留旧行为时，按对应版本说明显式配置，不要直接把全部旧默认值复制进环境文件。
- `--reuse-values` 不是配置迁移工具，可能继续保留旧默认值和废弃字段。README 的升级命令使用 `--reset-values -f values-env.yaml`，以目标 Chart 默认值加完整环境覆盖为准；遗漏的环境覆盖不会从旧 release 自动恢复。
- Helmfile 的 `env.yaml` 不是原生 Chart values。使用 Helmfile 部署时，按 [Helmfile 配置入口与 overrides 边界](https://github.com/TencentBlueKing/bk-nodemgr/blob/master/install/helmfile/bk-nodemgr/README.md) 更新配置，并检查各 release 的渲染结果；不要将原生 Helm 的覆盖文件直接当作 Helmfile environment 使用。

### 检查最终资源与外部副作用

以下命令中的 `<namespace>`、`<target-chart-version>` 和 `<source-chart-version>` 均需替换为实际值；release 名称不是 `bk-nodemgr` 时也需相应替换。`bk` 为 README 中已添加的 Chart 仓库。

```shell
# 保存当前 release 的用户覆盖、资源清单和历史，供核对与回退使用。
helm get values bk-nodemgr -n <namespace> -o yaml > values-before.yaml
helm get manifest bk-nodemgr -n <namespace> > manifest-before.yaml
helm history bk-nodemgr -n <namespace>

# 对比两个版本的默认值，人工将迁移结果合入完整的 values-env.yaml。
helm show values bk/bk-nodemgr --version <source-chart-version> > defaults-before.yaml
helm show values bk/bk-nodemgr --version <target-chart-version> > defaults-target.yaml

# 使用目标 Chart 包和完整环境覆盖检查最终资源。
helm pull bk/bk-nodemgr --version <target-chart-version> --untar --untardir ./chart-target
helm lint ./chart-target/bk-nodemgr -f values-env.yaml
helm template bk-nodemgr ./chart-target/bk-nodemgr -n <namespace> -f values-env.yaml > manifest-target.yaml
```

`values-before.yaml` 仅用于核对当前用户覆盖，不应不经迁移直接当作目标环境文件。检查 manifest 差异时，应区分模板变化与环境配置变化；必要时用源版本 Chart 和原环境覆盖重新渲染。重点检查镜像、服务启动配置、端口、权限、PVC、调度条件及 Hook Job，不能只看 values 的文本差异。

环境文件、Helm release 导出的 values 和渲染结果可能包含凭证、密钥，应按敏感文件保存，不提交到仓库或外传。执行导出和渲染前，在受保护目录设置合适的文件访问权限。`helm pull` 的目标目录应为空，避免混入其他版本文件。本地渲染不会执行 Hook，也不证明数据库、外部平台权限和网络已就绪。

执行升级前，备份 MongoDB 等持久化数据及当前部署配置，确认恢复方式，并检查目标版本的数据库迁移、IAM 和 APIGateway 同步等 Hook 副作用。升级后检查 Pod、Service、Hook Job 的状态及日志，并通过实际访问入口验证登录和业务操作、核对外部平台同步结果；不能仅以 Pod 就绪或 Helm 命令成功判断升级完成。

需要回退时，先评估目标版本写入的数据和外部变更是否兼容旧版本，再按 `helm history` 选定 revision：

```shell
helm rollback bk-nodemgr <revision> -n <namespace>
```

Helm rollback 不会恢复数据库、宿主机文件或撤销 APIGateway/IAM 等外部平台变更；这些内容需按各自备份和恢复流程处理。

## v3.0.1-alpha.114

### Application 的 bk_token 登录接口不再由租户模式选择

**影响范围：** 从 `v3.0.1-alpha.113` 升级、使用 `authType: bk_token` 的单租户环境。新增 `application.config.bkSaaS.bkLogin.useLegacyAPI`，默认值为 `false`；旧版按租户模式选择登录接口，新版改为显式选择。

| 环境 | 升级前 | 升级后默认行为 | 需要的操作 |
| --- | --- | --- | --- |
| 单租户，`authType: bk_token` | `/accounts/get_user/` | `/open/bk-tokens/userinfo/` | 平台仍依赖旧接口时显式设置 `useLegacyAPI: true`；使用新接口时补齐网关配置与权限 |
| 多租户，`authType: bk_token` | `/open/bk-tokens/userinfo/` | 不变 | 保持 `useLegacyAPI: false`，不支持旧接口兼容模式 |
| `authType: bk_ticket` | 原有票据认证接口 | 不变 | 无需因本项调整 |

仍使用旧登录接口的单租户环境，将以下片段合入完整环境覆盖：

```yaml
application:
  config:
    tenantMode: single
    bkSaaS:
      bkLogin:
        useLegacyAPI: true
```

使用新接口时，保持 `useLegacyAPI: false`，并配置 `application.config.bkSaaS.bkLogin` 下可访问 `/open/bk-tokens/userinfo/` 的 `endpoints`、网关应用凭证 `appCode`、`appSecret`，以及所选 `authMode` 所需的认证参数。`loginURL` 仅用于浏览器跳转，不能代替 API 网关地址。

**升级后验证：** 使用实际访问入口完成一次登录，确认用户身份、租户和业务页面访问正常；不能仅以 Application Pod 就绪判断登录兼容性。回退到旧版本前，需恢复与旧接口选择规则匹配的配置，新字段不能改变旧镜像的接口选择逻辑。

## v3.0.1-alpha.115

### 用户标识兼容不再由租户模式选择

**影响范围：** 从 `v3.0.1-alpha.114` 升级的 Backend 和 Application。新增两个服务的 `config.userManager.useLegacyUsername`，默认值均为 `false`。旧版单租户直接将 `login_name` 作为 `bk_username`，新版默认通过用户管理 lookup 查询当前租户的用户标识映射。

以下 7.2/7.3 指蓝鲸平台大版本，不是节点管理版本：

| 蓝鲸平台 | 租户模式 | 需要的操作 |
| --- | --- | --- |
| 7.2 | `single` | 两个服务同时显式设置 `userManager.useLegacyUsername: true`，保留旧用户标识规则 |
| 7.3 | `single` 或 `multiple` | 两个服务保持 `false`，补齐 lookup 调用权限、地址、应用凭证与请求租户配置 |

仍运行蓝鲸平台 7.2 时，将以下片段合入完整环境覆盖；`true` 仅支持单租户模式，不能用于多租户环境：

```yaml
backend:
  config:
    userManager:
      useLegacyUsername: true
application:
  config:
    userManager:
      useLegacyUsername: true
```

蓝鲸平台 7.3 需为节点管理应用开通用户管理 `GET /open/tenant/virtual-users/-/lookup/` 的调用权限，并检查两个服务的 `userManager.endpoints`、应用凭证及请求租户标识；单租户使用 `default` 租户。查询失败、无权限或没有用户映射时返回错误，不回退为 `bk_username = login_name`，不要通过开启旧版兼容来绕过这些问题。

蓝鲸平台从 7.2 升级到 7.3 时，两个服务必须同步将该字段切换为 `false`。该开关与 `application.config.bkSaaS.bkLogin.useLegacyAPI` 相互独立，不能替代上一节的登录接口兼容配置。

Helmfile 在 environment 的 `backend.config.userManager` 和 `application.config.userManager` 中同步配置；Docker Compose 的两个服务配置模板默认同为 `false`，部署到蓝鲸平台 7.2 前须同步改为 `true`。

**升级后验证：** 从实际入口登录并访问业务页面，确认用户身份、租户及用户映射正确，检查 Backend、Application 是否出现 lookup 权限或用户映射错误。回退到旧镜像后，单租户会重新使用旧用户标识规则，新字段不能让旧镜像继续执行 lookup。

## v3.0.1-alpha.116

### Application 内嵌 Wiki 改为默认关闭

**影响范围：** 依赖 Application `/wiki/` 文档入口的部署。旧版始终注册该路由，新版新增 `application.config.wiki.enabled`，默认值为 `false`。使用 Backend 的 `github` 文档链接模式时，无需因本项开启 Wiki。

使用 `backend.config.wiki.mode: application` 或直接依赖内嵌文档入口时，将以下片段合入完整环境覆盖：

```yaml
application:
  config:
    wiki:
      enabled: true
```

此开关不会自动切换 Backend 的链接模式。保留 `backend.config.wiki.application.prefixURL` 指向用户可访问的文档地址，并确保外部入口将 `/wiki/` 转发到 Application；关闭 Wiki 前需另行调整 Backend 链接配置，避免继续生成不可访问的链接。内嵌文档可免登录访问，不应包含敏感信息。

Helmfile 在 environment 的 `application.config.wiki.enabled` 中配置；Docker Compose 或直接配置 Application 时使用 `wiki.enabled: true`。

**升级后验证：** 通过实际业务页面打开文档链接，确认外部地址、`/wiki/` 路由及文档内容可访问。未启用开关时，该入口不提供文档访问。

### IAM V3 同步 Job 开始传递 Action 替换开关

**影响范围：** 启用 `backend.config.iamV3.enable` 的部署。安装和升级的 IAM V3 同步 Job 新增 `iamV3.option.allowActionReplace`，默认值为 `false`；仅在设为 `true` 时向迁移脚本传入 `--allow-action-replace`。

- 无需替换 Action 资源绑定时保持 `false`。若旧环境已自行添加 `iamV3.option.allowActionReplace: true`，升级前必须复核：旧版 Job 未使用该字段，新版会真正执行对应的替换逻辑。
- 需要替换资源绑定时，先确认目标系统、租户及本次同步的全部绑定变更，再临时设置为 `true`。该开关覆盖全目录同步中所有发生绑定变化的 `upsert_action`，不限于单个 Action。
- 替换会清理策略、删除并重建 Action，已有授权不会自动迁移。**策略删除不可逆，关闭开关或回退 Chart 无法恢复已删除策略。** 执行前确认受影响的用户／用户组及权限恢复方式，不要将此开关作为所有环境的升级默认配置。
- 使用包含 `--allow-action-replace` 支持的 server 镜像；策略删除要求 IAM SaaS ≥ 1.4.15、后台 ≥ 1.7.7。Helmfile 将该 Chart values 片段合入对应 environment 的 `overrides/backend.yaml`，不是服务运行时配置。

**升级后验证：** 检查 IAM V3 同步 Job 日志及远端模型，重新配置并验证受影响的权限，完成后将开关恢复为 `false`。失败或结果未知时先核对远端状态，不要盲目重复执行；详细替换边界和失败恢复见 [IAM V3 迁移工具](https://github.com/TencentBlueKing/bk-nodemgr/blob/v3.0.1-alpha.116/support-files/bkiamv3/README.md#显式替换-action-资源绑定)。

### Application 与 Frontend 必须同步升级

本版本移除了旧 Agent／Proxy 安装 Excel 模板接口。单独覆盖组件镜像或独立部署 Frontend 时，必须同步升级 Application 与 Frontend；直接调用旧接口的集成方也需迁移。升级后重新下载模板并验证下载、解析和页面回填，接口路径与响应变化见 [alpha.116 版本变更记录](https://github.com/TencentBlueKing/bk-nodemgr/blob/v3.0.1-alpha.116/support-files/changelog/zh/v3.0.1-alpha.116_2026-11-20.md)。
