# 初始化包导入

`support-files/initpackage/init_package.py` 用于将本地初始化包上传到 `file` 服务，再将其发布为可用的 release 包。该工具适用于部署初始化场景，支持导入 V2 官方插件包、V2 外部插件包和 V3 插件包。

本文介绍 Helm 初始化 Job 以及手工导入插件包的方法。

## 工作原理

每个选中的初始化包依次经过以下两个阶段：

1. 将本地文件上传到对应的 `/api/v3/upload/origin/...` 接口。`file` 服务校验包内容并返回 `upload_id`。
2. 将 `upload_id` 传给对应的 `/api/v3/publish/release/...` 接口。`file` 服务解析初始化包，并将 release 产物发布到已配置的存储中。

脚本按照包类型顺序处理任务。某个包在上传、发布或 JWT 生成阶段失败时，脚本会继续处理后续任务；最终汇总会逐文件列出成功和失败清单，只要存在失败任务，脚本最终就会返回非零退出码。

## 连接配置与参数优先级

脚本支持 `--config-file <path>`。未指定时依次尝试 `/bk-nodemgr/etc/file_conf.yaml`、`/bk-nodemgr/etc/bk-nodemgr-file.yml`，并使用 PyYAML `safe_load` 读取。命令行显式参数始终优先于配置文件；未显式指定配置文件时，配置文件不存在或解析失败会使用安全回落值；显式指定 `--config-file` 失败时脚本会报错退出。脚本会打印配置来源（如有）及最终 tenant/host/port，不会打印 JWT key。

| 参数                            | 配置值与回落值                                                                                                                        |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------- |
| `--host`                        | CLI，否则 `localhost`                                                                                                                 |
| `--port`                        | CLI，否则读取 File 配置 `file.config.basicServer.port`，再回落 `28202`                                                                |
| `--jwt-key`                     | CLI，否则仅当 File 配置 `file.config.basicServer.authIdentity` 不为 `none` 时读取其 `jwtServerConfig.symmetricKey`；未启用 JWT 时为空 |
| `--expire-hours`                | 仅使用 CLI 显式设置；File 配置没有 JWT 有效期字段，未设置时默认为 `24` 小时                                                           |
| `--bk-username`、`--login-name` | CLI，否则 `admin`                                                                                                                     |
| `--tenant-id`                   | CLI，否则读取 File 配置 `file.config.tenantMode`；`multiple` 使用 `system`，其他情况使用 `default`                                    |

单租户默认命令可以只指定包路径：

```bash
python3 /bk-nodemgr/support-files/initpackage/init_package.py --init-agent /tmp/agent.tgz
```

多租户未传 `--tenant-id` 时目标租户为 `system`；导入其他租户时显式传入对应的 `--tenant-id`。多租户初始化要求 File 的 `config.basicServer.authIdentity` 为 `rest-server`，且 `config.basicServer.jwtServerConfig.symmetricKey` 非空并与生成 JWT 使用的密钥匹配。`authIdentity: none` 会使请求在 `default` tenant 处理，不能作为 `tenantMode: multiple` 的初始化配置。

## 支持的初始化包类型

| 命令参数                    | 包类型           | 上传目录                    |
| --------------------------- | ---------------- | --------------------------- |
| `--init-plugin-v2`          | V2 官方插件包    | `origin/v2/plugin`          |
| `--init-external-plugin-v2` | V2 外部插件包    | `origin/v2/external_plugin` |
| `--init-plugin-v3`          | V3 插件包        | `origin/v3/plugin`          |
| `--init-plugin-bintool`     | 插件二进制工具包 | `origin/plugin_bintool`     |
| `--init-agent`              | Agent 包         | `origin/agent`              |
| `--init-proxy`              | Proxy 包         | `origin/proxy`              |
| `--init-server`             | Server 包        | `origin/server`             |
| `--init-cert`               | GSE 证书包       | `origin/cert`               |
| `--init-bintool`            | 节点二进制工具包 | `origin/bintool`            |

必须使用与参数类型对应的初始化包。V2 和 V3 插件包不能混用，`file` 服务会校验压缩包目录结构及 `project.yaml` 等元数据。

## 通过 Helm 自动初始化

只有同时满足以下条件时，Helm Chart 才会渲染 `templates/init-packages-job.yaml`：

- `file.enabled=true`
- `initPackages.enabled=true`

该 Job 以 Helm Hook 方式运行。它等待 `file` 服务健康后，使用与 `file` 服务相同的镜像，读取镜像内置的初始化包，并调用 `init_package.py` 连接集群内的 `file` 服务完成导入。Job 会将 `{{ .Values.common.etcPath }}/file_conf.yaml` 挂载为 `{{ template "bk-nodemgr.fullname" . }}-file-config` ConfigMap 中的 `file_conf.yaml`。命令通过 `--config-file` 读取该配置，因此连接端口、JWT 密钥和 tenant 来自 File 配置；JWT 有效期没有 File 配置字段，默认 24 小时，也可由 `--expire-hours` 显式覆盖。Job 仍通过 `--host` 使用 File Service DNS 覆盖配置中的 host，因为 ConfigMap 不包含该 Service DNS。该挂载的 ConfigMap 内容是 File 配置模板，不应理解为与 File Pod 中经过 advertise/tracing 占位符替换后的运行时文件逐字节相同。

### 自动导入范围

`--set-as-default` 会在所有包逐个完成 upload → publish 后，统一启用目标包并设为默认。Agent 单独一组，Proxy/Server 共同一组，各取现有遍历顺序中最后一个输入包；三种 Plugin 入口分别按插件 name 分组，各取最后一个 upload 成功的包。显式参数在自动扫描文件之前，不另按 version 排序，也不限制同版本文件。目标 publish 失败不回退；其他文件失败不阻止目标操作。Plugin upload 失败只记失败，不参与分组。

状态操作覆盖目标整包实际产出的全部 platform，不使用旧版本补齐平台。Cert、BinTool、Plugin BinTool 仅上传和发布。单个平台 enable 失败时跳过该平台的 set_as_default，继续其他平台；状态操作失败将目标文件记入 failed files，最终非零退出，已完成操作不回滚。手动初始化需要显式传入该 flag；不传时仅上传和发布。

当前 Helm Job 调用 `init_package.py --auto-select --set-as-default` 扫描以下目录，并按脚本固定顺序导入其中的普通文件：

| 镜像内目录                                      | 导入类型                                                         |
| ----------------------------------------------- | ---------------------------------------------------------------- |
| `/bk-nodemgr/file/packages/cert/`               | cert                                                             |
| `/bk-nodemgr/file/packages/bintool/`            | bintool；文件名以 `plugin_bintool` 开头的文件归为 plugin_bintool |
| `/bk-nodemgr/file/packages/agent/`              | agent                                                            |
| `/bk-nodemgr/file/packages/proxy/`              | proxy                                                            |
| `/bk-nodemgr/file/packages/server/`             | server                                                           |
| `/bk-nodemgr/file/packages/plugin-v2/`          | plugin_v2                                                        |
| `/bk-nodemgr/file/packages/external-plugin-v2/` | external_plugin_v2                                               |
| `/bk-nodemgr/file/packages/plugin-v3/`          | plugin_v3                                                        |

其中 `cert/` 下的 `cert.tgz` 仍由 Job 在八个 `gseCert` 文件全部存在时生成；生成后由脚本自动发现并导入。Job 按 cert、bintool、plugin_bintool、agent、proxy、server、plugin_v2、external_plugin_v2、plugin_v3 的固定顺序处理，同一类型可导入多个普通文件。没有找到包时脚本输出 `No packages to upload`，Job 仍成功结束。

### Values 配置示例

在部署使用的 values 文件中增加以下配置：

```yaml
file:
  enabled: true
  config:
    # 仅当 File Basic Server 启用 JWT 鉴权时配置
    basicServer:
      authIdentity: rest-server
      jwtServerConfig:
        cryptoType: symmetric
        symmetricKey: "<jwt-symmetric-key>"

initPackages:
  enabled: true
  installOnly: true
  ttlSecondsAfterFinished: 86400
  backoffLimit: 3
  parallelism: 1
  bkUsername: "admin"
  loginName: "admin"
```

Job 不再从 Helm values 单独传入端口、JWT 密钥或 tenant；这些值均由挂载的 File 配置读取。JWT 有效期没有 File 配置字段，默认 24 小时，也可通过 `--expire-hours` 显式覆盖。`--host` 仍显式设置为 Job 使用的 File Service DNS。若 `file` Basic Server 启用了 JWT 鉴权，`file.config.basicServer.jwtServerConfig.symmetricKey` 必须与生成 JWT 使用的密钥一致；使用 `authIdentity: none` 时请求会按 `default` tenant 处理，密钥可以留空。File Pod 会在启动时替换配置中的 advertise/tracing 占位符，因此 Job 挂载的 ConfigMap 与 File Pod 的运行时配置文件不是逐字节相同的文件。

启用 JWT 鉴权时，应为 `bkUsername` 和 `loginName` 配置非空值。二者会写入 JWT 声明，用于标识本次初始化操作的执行用户。

### Helm 配置项

| 配置项                                 | 默认值       | 行为                                                                                                    |
| -------------------------------------- | ------------ | ------------------------------------------------------------------------------------------------------- |
| `initPackages.enabled`                 | `false`      | 是否启用初始化 Hook Job；同时要求 `file.enabled=true`。                                                 |
| `initPackages.installOnly`             | `true`       | 是否只在 `helm install` 后运行；设为 `false` 时也会在 `helm upgrade` 后运行。                           |
| `initPackages.ttlSecondsAfterFinished` | `86400`      | Job 完成后由 Kubernetes 清理的等待秒数；成功的 Hook Job 可能被 Helm 提前删除。配置为空可关闭 TTL 清理。 |
| `initPackages.backoffLimit`            | `3`          | Job Pod 失败后的最大重试次数；重试时可能遇到前一次部分成功后已经发布的包。                              |
| `initPackages.parallelism`             | `1`          | Job 并行度。应保持为 `1`；当前 Job 只有一个完成任务，且脚本按包类型串行导入。                           |
| `initPackages.bkUsername`              | 空           | 设置 JWT 中的 `bk_username`。                                                                           |
| `initPackages.loginName`               | 空           | 设置 JWT 中的 `login_name`。                                                                            |
| `initPackages.gseCert.*`               | 空           | 设置经过 Base64 编码的 GSE 证书和密钥文件内容。                                                         |
| `initContainerResources`               | Chart 默认值 | 设置初始化容器的 CPU 和内存资源。                                                                       |

默认情况下，Job 是 `post-install` Hook。设置 `installOnly=false` 后，Hook 类型变为 `post-install,post-upgrade`。其删除策略为 `before-hook-creation,hook-succeeded`：成功的 Job 通常会在完成后被 Helm 删除；下一次执行 Hook 前，Helm 也会删除同名的旧 Job。

### 可选的 GSE 证书初始化

Job 可以使用以下经过 Base64 编码的 values 生成 `cert.tgz`：

| Helm 配置项                           | 压缩包内文件              |
| ------------------------------------- | ------------------------- |
| `initPackages.gseCert.ca`             | `cert/gseca.crt`          |
| `initPackages.gseCert.cert`           | `cert/gse_server.crt`     |
| `initPackages.gseCert.key`            | `cert/gse_server.key`     |
| `initPackages.gseCert.certEncryptKey` | `cert/cert_encrypt.key`   |
| `initPackages.gseCert.apiClient.cert` | `cert/gse_api_client.crt` |
| `initPackages.gseCert.apiClient.key`  | `cert/gse_api_client.key` |
| `initPackages.gseCert.agent.cert`     | `cert/gse_agent.crt`      |
| `initPackages.gseCert.agent.key`      | `cert/gse_agent.key`      |

只有八个文件全部存在时，Job 才会生成并导入 `cert.tgz`。只提供部分 `gseCert` 配置会跳过证书包，但不影响其他已找到的初始化包。不要将生产环境证书或私钥提交到代码仓库。

### 安装或升级

使用准备好的 values 文件执行正常的 Chart 安装命令：

```bash
helm upgrade --install <release-name> install/helm/bk-nodemgr \
  --namespace <namespace> \
  --create-namespace \
  --values <values-file>
```

Job 最多等待 `file` 健康接口 300 秒，随后导入镜像中存在的全部自动初始化包。自动初始化文件不存在时只会跳过，不会将其视为失败。

## 在 Kubernetes 中手工导入插件包

部署镜像在 `/bk-nodemgr/support-files/initpackage` 下包含 `init_package.py` 和 `jwt-generator`。建议在 `file` Pod 中运行脚本，以便直接访问服务并复用镜像中的 Python 依赖。

### 1. 获取 File Pod

```bash
export NAMESPACE="<namespace>"
export FILE_POD="$(kubectl -n "${NAMESPACE}" get pod \
  -l app=nodemgr-file \
  -o jsonpath='{.items[0].metadata.name}')"
```

### 2. 将插件包复制到 Pod

只需复制本次需要导入的包：

```bash
kubectl -n "${NAMESPACE}" cp ./plugin-v2.tgz "${FILE_POD}:/tmp/plugin-v2.tgz"
kubectl -n "${NAMESPACE}" cp ./external-plugin-v2.tgz "${FILE_POD}:/tmp/external-plugin-v2.tgz"
kubectl -n "${NAMESPACE}" cp ./plugin-v3.tgz "${FILE_POD}:/tmp/plugin-v3.tgz"
```

### 3. 执行导入

`FILE_PORT` 应使用 `file.config.basicServer.port` 的实际配置，Chart 默认值为 `28202`。JWT 密钥和执行用户应与部署配置一致；`file` 接口未启用 JWT 鉴权时，`JWT_KEY` 可以留空。

```bash
export FILE_PORT="28202"
export JWT_KEY="<jwt-symmetric-key>"
export TENANT_ID="default"
export BK_USERNAME="admin"
export LOGIN_NAME="admin"

kubectl -n "${NAMESPACE}" exec "${FILE_POD}" -- \
  python3 /bk-nodemgr/support-files/initpackage/init_package.py \
    --host 127.0.0.1 \
    --port "${FILE_PORT}" \
    --jwt-key "${JWT_KEY}" \
    --expire-hours 24 \
    --tenant-id "${TENANT_ID}" \
    --bk-username "${BK_USERNAME}" \
    --login-name "${LOGIN_NAME}" \
    --init-plugin-v2 /tmp/plugin-v2.tgz \
    --init-external-plugin-v2 /tmp/external-plugin-v2.tgz \
    --init-plugin-v3 /tmp/plugin-v3.tgz
```

删除本次不需要导入的包参数。同一个 `--init-*` 参数可以重复指定，以便一次命令导入多个同类型文件。

插件类型由对应的命令参数决定。`--generation` 只作用于 Agent、Proxy、Server 和节点二进制工具包，使用 `--init-plugin-v2`、`--init-external-plugin-v2` 或 `--init-plugin-v3` 时不需要设置。

可选参数 `--overwrite` 会在上传元数据中加入 `overwrite=true`。不要依赖该参数替换已经发布的同名同版本插件，重复包的处理结果由 `file` 服务和包内容决定；删除或替换已有 release 前应先确认命令结果。

### 多租户部署

`init_package.py` 使用配置中的 tenantMode 决定 `X-Bk-Tenant-Id` 请求头。Helm Job 通过挂载的 File ConfigMap 读取该配置：File 服务为 `tenantMode: single` 或其他非 `multiple` 值时使用 `default`，为 `tenantMode: multiple` 时使用 `system`；Job 的 `--host` 仍覆盖配置中的 host 为 File Service DNS。

多租户 Helm Job 默认导入 `system` 租户；要导入其他租户，需在 File Pod 内手动执行脚本并显式指定对应的 `--tenant-id`。BKRepo 项目名称转换由 `file` 服务配置处理，不要在插件包路径中自行添加租户前缀。存储配置说明参考 [BKRepo 资源使用说明](thirdparty_bkrepo_resource.md)。

### 自动选择目录中的初始化包

`--auto-select` 默认扫描 `/bk-nodemgr/file/packages`，也可以使用 `--packages-dir <path>` 覆盖。扫描不解压、不检测压缩包，只选择普通文件，不删除源文件；同一目录的多个文件按文件名字典序执行。

| 子目录                          | 导入类型           |
| ------------------------------- | ------------------ |
| `cert/`                         | cert               |
| `agent/`                        | agent              |
| `proxy/`                        | proxy              |
| `server/`                       | server             |
| `plugin-v2/`                    | plugin_v2          |
| `external-plugin-v2/`           | external_plugin_v2 |
| `plugin-v3/`                    | plugin_v3          |
| `bintool/` 中 `plugin_bintool*` | plugin_bintool     |
| `bintool/` 中其他普通文件       | bintool            |

发现的文件会追加到显式 `--init-*` 文件，并按 cert、bintool、plugin_bintool、agent、proxy、server、plugin_v2、external_plugin_v2、plugin_v3 的固定类型顺序执行。同类型多文件会作为独立任务逐个处理。

```bash
kubectl -n "${NAMESPACE}" cp ./packages/. "${FILE_POD}:/bk-nodemgr/file/packages"
kubectl -n "${NAMESPACE}" exec "${FILE_POD}" -- \
  python3 /bk-nodemgr/support-files/initpackage/init_package.py --auto-select
```

没有显式包且扫描目录没有发现包时，脚本输出 `No packages to upload` 并以状态码 0 退出。单个文件失败后会继续其他文件，最终按文件显示成功/失败清单；存在失败时返回非零状态码。

## 确认导入结果

手工导入成功时，脚本输出以下汇总并以状态码 `0` 退出：

```text
summary: <count> success, 0 failed
successful files:
  - <file-path>
failed files:
```

在将插件用于安装或升级策略前，还应进入节点管理的插件包页面，确认对应的插件名称和版本已经出现。

对于 Helm Job，可以在 Hook Pod 运行期间或失败后查看状态和日志：

```bash
kubectl -n "${NAMESPACE}" get pod -l app=nodemgr-file-init-packages
kubectl -n "${NAMESPACE}" logs -l app=nodemgr-file-init-packages --all-containers=true
```

成功安装后找不到该 Pod 属于正常现象，因为 Hook 使用了 `hook-succeeded` 删除策略。

Helm Job 和运行中 File Pod 的手工命令均可复用 `--auto-select` 目录扫描能力；该功能不新增或修改 Helm values 配置。

## 常见问题

| 现象                                      | 检查项                                                                                                                     |
| ----------------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| 没有执行初始化 Job                        | 确认 `file.enabled=true` 和 `initPackages.enabled=true`。在升级场景还需确认 `initPackages.installOnly=false`。             |
| Job 输出 `No packages to upload`          | 扫描目录中没有普通初始化包，且没有提供完整的证书文件；这是成功的空操作，不表示 Job 失败。                                  |
| 输出 `file service not ready within 300s` | 检查 `file` Deployment、Service、`file.config.infoServer.port` 和 `readinessProbe.path`。                                  |
| 输出 `jwt-generator binary not found`     | 使用通过项目 Makefile 完整构建的 bk-nodemgr 镜像，构建产物必须包含 `/bk-nodemgr/support-files/initpackage/jwt-generator`。 |
| 鉴权失败                                  | 对比导入脚本的 JWT 密钥、过期时间与已部署的 `file` 鉴权配置，并检查 `bkUsername`、`loginName` 和租户 ID。                  |
| 上传或发布失败                            | 确认压缩包符合对应的 V2、V2 外部插件或 V3 包格式，且包内元数据有效；通过 `file` 服务日志查看具体校验错误。                 |
| 部分包成功但命令仍返回非零状态            | 脚本会在单个任务失败后继续处理，并在最后返回失败。根据各任务输出，只重试失败的插件包。                                     |

脚本完整参数以及 Docker、本地执行示例还可参考 [`support-files/initpackage/README.md`](../../../support-files/initpackage/README.md)。
