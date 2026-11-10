# APIGateway 同步 FAQ

## 1. 如何通过 Helm 开启 APIGateway 同步镜像？

这个问题对应 Helm 配置里的 `apigwSync` 和 `apiManagerImage`。

`apigwSync.enabled` 开启后，Chart 会创建一个一次性 Job。这个 Job 使用 `apiManagerImage` 指定的 `bk-nodemgr-apigw-sync`
镜像，挂载由 `apigwSync.config` 渲染出的 `/data/definition.yaml`，然后执行镜像内的 `/data/bin/sync-apigateway.sh`。

一个可直接参考的配置片段：

```yaml
apiManagerImage:
  registry: "hub.bktencent.com"
  repository: "blueking/bk-nodemgr-apigw-sync"
  tag: v3.0.1-alpha.40
  pullPolicy: IfNotPresent

apigwSync:
  enabled: true
  extraEnvVars:
    - name: BK_APIGW_NAME
      value: "bk-nodemgr"
    - name: BK_APP_CODE
      value: "bk-nodemgr"
    - name: BK_APP_SECRET
      value: "<bk_app_secret>"
    - name: BK_API_URL_TMPL
      value: "http://bkapi.example.com/api/{api_name}"
    - name: RELEASE_STAGES
      value: "stage"
    - name: NO_PUB
      value: "true"
  config:
    spec_version: 1
    release:
      version: "2.14.1"
      title: "bk-nodemgr"
      comment: "v3.0.1-alpha.40"
    apigateway:
      description: "蓝鲸节点管理"
      description_en: "bk-nodemgr"
      is_public: false
      maintainers:
        - "admin"
    stages:
      - name: "stage"
        description: "预发布环境"
        description_en: "stage"
        backends:
          - name: "default"
            config:
              timeout: 60
              loadbalance: "roundrobin"
              hosts:
                - host: "https://<bk-nodemgr-backend-host>"
                  weight: 100
    grant_permissions:
      - bk_app_code: "bk-nodemgr"
        grant_dimension: "resource"
        resource_names:
          - "<resource_name>"
```

`BK_APP_SECRET` 示例中只保留占位符，不要把真实 Secret 提交到仓库或长期保存在公开的 values 文件中。

## 2. 每个配置项分别影响什么？

| 配置项                                    | 作用                                                                                      |
| ----------------------------------------- | ----------------------------------------------------------------------------------------- |
| `apiManagerImage.registry/repository/tag` | 指定同步 Job 使用的 `bk-nodemgr-apigw-sync` 镜像。                                        |
| `apigwSync.enabled`                       | 是否创建 APIGateway 同步 Job。                                                            |
| `BK_APIGW_NAME`                           | 提供给 `apigw-manager` 的网关名配置；当前脚本命令显式使用 `bk-nodemgr`，建议保持一致。    |
| `BK_APP_CODE` / `BK_APP_SECRET`           | 调用 `bk-apigateway` 管理接口所需的应用身份。                                             |
| `BK_APP_TENANT_ID`                        | 调用 `bk-apigateway` 管理接口时透传的应用租户 ID；单租户传 `default`，多租户传 `system`。 |
| `BK_API_URL_TMPL`                         | 网关管理 API 地址模板，必须保留 `{api_name}` 占位符。                                     |
| `RELEASE_STAGES`                          | 创建版本后发布到哪个环境，应与 `apigwSync.config.stages[].name` 对齐。                    |
| `NO_PUB`                                  | 是否额外执行一次只生成版本不发布的命令。                                                  |
| `apigwSync.config.release`                | 渲染到 `definition.yaml` 的版本信息。                                                     |
| `apigwSync.config.apigateway`             | 渲染到 `definition.yaml` 的网关基础信息和维护人。                                         |
| `apigwSync.config.stages`                 | 渲染到 `definition.yaml` 的网关环境和后端回源地址。                                       |
| `apigwSync.config.grant_permissions`      | 渲染到 `definition.yaml` 的主动授权配置。                                                 |

`stages[].backends[].config.hosts[].host` 只写协议、域名或 IP，不包含 Path，例如 `https://bk-nodemgr.example.com`。

Helm 会根据 `backend.config.tenantMode` 自动注入 `BK_APP_TENANT_ID`：`single` 对应 `default`，`multiple` 对应
`system`，不要在 `apigwSync.extraEnvVars` 中重复配置。docker-compose 通过 `BK_NODEMGR_APP_TENANT_ID` 设置该值。
`apigw-manager` 会把它转换成请求头 `X-Bk-Tenant-Id`。如果同步日志出现 `1640302 Cross tenant forbidden`，先确认同步镜像
基于支持多租户的 `apigw-manager` 版本，再检查该变量是否按租户模式设置。

> notice: 一个环境的 `BK_API_URL_TMPL` 可以通过我们注册到 bk-apigw 后生成的 API 地址来获取。比如我们生成的 API 地址是
`https://bk-nodemgr.apigw.example.com/api/v1/nodemgr/nodes`，那么 BK_API_URL_TMPL 就是 `https://bk-nodemgr.apigw.example.com/api/{api_name}`。

## 3. `NO_PUB=true` 是否表示一定不会发布？

不是。

当前同步脚本会先读取：

```bash
release_stages="${RELEASE_STAGES:-}"
no_pub="${NO_PUB:-true}"
```

当 `NO_PUB=true` 时，脚本会先执行一次 `create_version_and_release_apigw --no-pub`，只生成版本不发布。

随后，只要 `RELEASE_STAGES` 非空，脚本仍会继续执行一次带 `--stage` 的发布命令：

```bash
create_version_and_release_apigw definition.yaml --gateway-name=bk-nodemgr --stage "${RELEASE_STAGES}"
```

因此，默认的 `NO_PUB=true` 与空 `RELEASE_STAGES` 组合表示：只生成一个未发布版本，不发布任何环境。
如需在生成版本后发布到指定环境，需要显式设置 `RELEASE_STAGES`。

## 4. `resources.yaml` 和接口文档从哪里来？

Helm values 只会渲染并挂载 `/data/definition.yaml`。

同步脚本还会读取 `/data/resources.yaml` 和 `/data/apidocs/`。这些内容不是由 `apigwSync.config` 生成的，而是在构建
`bk-nodemgr-apigw-sync` 镜像时从仓库的 `apigw/` 目录复制进镜像：

```bash
make docker-build-apigw-sync
```

如果只修改了 Helm values，通常只会影响网关定义、环境、后端地址和授权配置；如果需要更新网关资源列表或接口文档，需要更新
`apigw/resources.yaml`、`apigw/apidocs/`，并重新构建同步镜像。

这个同步镜像基于 BlueKing `apigw-manager`
，相关命令能力可参考 [apigw-manager](https://github.com/TencentBlueKing/bkpaas-python-sdk/tree/master/sdks/apigw-manager)。

## 5. 如何确认同步 Job 已经执行？

升级或安装 Chart 后，先确认 Job 是否创建：

```bash
kubectl get job -n <namespace> | grep apigw-sync
```

再查看 Pod 日志：

```bash
kubectl logs -n <namespace> job/<fullname>-apigw-sync-<revision>
```

如果需要确认 Helm 渲染出的网关定义，可以查看 ConfigMap：

```bash
kubectl get configmap -n <namespace> <fullname>-apigw-sync-definition -o yaml
```

重点检查 `definition.yaml` 里的 `stages[].name`、后端 `host`、`grant_permissions` 是否和预期一致。

## 6. 已设置 `BK_APP_TENANT_ID=system`，为什么还反复提示设置租户 ID？

### 现象与原因

同步日志反复出现以下提示，但请求头中的租户 ID 已经是 `system`：

```text
[warn] if the syncing to apigateway failed, and your app(bk-nodemgr) is a global tenant app, please set the environment variable BK_APP_TENANT_ID to `system` (or set django settings.BK_APP_TENANT_ID to `system`) and try again. the current value [X-Bk-Tenant-Id=system].
```

**这条 warning 本身不表示租户配置错误，也不表示同步失败。** 在 `apigw-manager:4.2.4` 中，未设置
`BKPAAS_APP_TENANT_ID` 时，SDK 会读取 `BK_APP_TENANT_ID`，并无条件输出这条提示；即使值已经是 `system` 也会输出。
每次构造 API 请求都会经过该逻辑，因此一次同步可能重复出现多次。

Helm 和 docker-compose 通过 `BK_APP_TENANT_ID` 传入应用租户 ID，属于该版本 SDK 支持的独立部署方式，
不需要因为这条提示改用 PaaS 的配置入口。

### 如何处理

- 如果 `X-Bk-Tenant-Id` 与应用所属租户一致，且最终同步成功，可以忽略这条提示，无需修改配置或重跑 Job。
- 不要仅为消除 warning 而添加 `BKPAAS_APP_TENANT_ID`：它会优先于 `BK_APP_TENANT_ID` 生效，且空字符串表示全租户应用，
  SDK 会将其转换为 `system`，并非单纯的日志开关。也不要全局屏蔽 WARNING，以免隐藏其他告警。
- 如果同步确实失败，应结合接口错误码和错误信息排查；遇到 `1640302 Cross tenant forbidden` 时，按第 2 节核对租户配置。
  不要只凭这条 warning 判断失败原因。
