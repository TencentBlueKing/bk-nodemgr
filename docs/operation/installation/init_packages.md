# 初始化包导入

`support-files/initpackage/init_package.py` 用于将本地初始化包上传到 `file` 服务，再将其发布为可用的 release 包。该工具适用于部署初始化场景，支持导入 V2 官方插件包、V2 外部插件包和 V3 插件包。

本文介绍 Helm 初始化 Job 以及手工导入插件包的方法。

## 工作原理

每个选中的初始化包依次经过以下两个阶段：

1. 将本地文件上传到对应的 `/api/v3/upload/origin/...` 接口。`file` 服务校验包内容并返回 `upload_id`。
2. 将 `upload_id` 传给对应的 `/api/v3/publish/release/...` 接口。`file` 服务解析初始化包，并将 release 产物发布到已配置的存储中。

脚本按照包类型顺序处理任务。某个包导入失败时，脚本会继续处理后续任务；只要存在失败任务，脚本最终就会返回非零退出码。

## 支持的初始化包类型

| 命令参数 | 包类型 | 上传目录 |
| -------- | ------ | -------- |
| `--init-plugin-v2` | V2 官方插件包 | `origin/v2/plugin` |
| `--init-external-plugin-v2` | V2 外部插件包 | `origin/v2/external_plugin` |
| `--init-plugin-v3` | V3 插件包 | `origin/v3/plugin` |
| `--init-plugin-bintool` | 插件二进制工具包 | `origin/plugin_bintool` |
| `--init-agent` | Agent 包 | `origin/agent` |
| `--init-proxy` | Proxy 包 | `origin/proxy` |
| `--init-server` | Server 包 | `origin/server` |
| `--init-cert` | GSE 证书包 | `origin/cert` |
| `--init-bintool` | 节点二进制工具包 | `origin/bintool` |

必须使用与参数类型对应的初始化包。V2 和 V3 插件包不能混用，`file` 服务会校验压缩包目录结构及 `project.yaml` 等元数据。

## 通过 Helm 自动初始化

只有同时满足以下条件时，Helm Chart 才会渲染 `templates/init-packages-job.yaml`：

- `file.enabled=true`
- `initPackages.enabled=true`

该 Job 以 Helm Hook 方式运行。它等待 `file` 服务健康后，使用与 `file` 服务相同的镜像，读取镜像内置的初始化包，并调用 `init_package.py` 连接集群内的 `file` 服务完成导入。

### 自动导入范围

当前 Helm Job 只会自动检查以下文件：

| 镜像内路径 | 来源 | 导入结果 |
| ---------- | ---- | -------- |
| `/bk-nodemgr/file/packages/bintool/bintool.tgz` | 构建镜像时从 `bintools/` 目录复制 | 导入节点二进制工具包 |
| `/bk-nodemgr/file/packages/bintool/plugin_bintool.tgz` | 构建镜像时从 `bintools/` 目录复制 | 导入插件二进制工具包 |
| `/bk-nodemgr/file/packages/cert/cert.tgz` | 八个 `gseCert` 文件全部存在时由 Job 生成 | 导入 GSE 证书包 |

该 Job **不会**自动检查或导入 V2/V3 插件包、V2 外部插件包、Agent 包、Proxy 包或 Server 包。Helm values 中没有将插件包路径映射到 `--init-plugin-v2`、`--init-external-plugin-v2` 或 `--init-plugin-v3` 的配置项。上述插件包需要按照[在 Kubernetes 中手工导入插件包](#在-kubernetes-中手工导入插件包)执行。

### Values 配置示例

在部署使用的 values 文件中增加以下配置：

```yaml
file:
  enabled: true

backend:
  config:
    file:
      jwtClientConfig:
        symmetricKey: "<jwt-symmetric-key>"
        tokenExpirationHour: 24

initPackages:
  enabled: true
  installOnly: true
  ttlSecondsAfterFinished: 86400
  backoffLimit: 3
  parallelism: 1
  bkUsername: "admin"
  loginName: "admin"
```

Job 会将 `backend.config.file.jwtClientConfig.symmetricKey` 和 `tokenExpirationHour` 直接传给 `init_package.py`。如果 `file` 上传接口启用了 JWT 鉴权，密钥必须与 `file` 服务接受的密钥一致；接口使用 `authIdentity: none` 时，密钥可以留空。

启用 JWT 鉴权时，应为 `bkUsername` 和 `loginName` 配置非空值。二者会写入 JWT 声明，用于标识本次初始化操作的执行用户。

### Helm 配置项

| 配置项 | 默认值 | 行为 |
| ------ | ------ | ---- |
| `initPackages.enabled` | `false` | 是否启用初始化 Hook Job；同时要求 `file.enabled=true`。 |
| `initPackages.installOnly` | `true` | 是否只在 `helm install` 后运行；设为 `false` 时也会在 `helm upgrade` 后运行。 |
| `initPackages.ttlSecondsAfterFinished` | `86400` | Job 完成后由 Kubernetes 清理的等待秒数；成功的 Hook Job 可能被 Helm 提前删除。配置为空可关闭 TTL 清理。 |
| `initPackages.backoffLimit` | `3` | Job Pod 失败后的最大重试次数；重试时可能遇到前一次部分成功后已经发布的包。 |
| `initPackages.parallelism` | `1` | Job 并行度。应保持为 `1`；当前 Job 只有一个完成任务，且脚本按包类型串行导入。 |
| `initPackages.bkUsername` | 空 | 设置 JWT 中的 `bk_username`。 |
| `initPackages.loginName` | 空 | 设置 JWT 中的 `login_name`。 |
| `initPackages.gseCert.*` | 空 | 设置经过 Base64 编码的 GSE 证书和密钥文件内容。 |
| `initContainerResources` | Chart 默认值 | 设置初始化容器的 CPU 和内存资源。 |

默认情况下，Job 是 `post-install` Hook。设置 `installOnly=false` 后，Hook 类型变为 `post-install,post-upgrade`。其删除策略为 `before-hook-creation,hook-succeeded`：成功的 Job 通常会在完成后被 Helm 删除；下一次执行 Hook 前，Helm 也会删除同名的旧 Job。

### 可选的 GSE 证书初始化

Job 可以使用以下经过 Base64 编码的 values 生成 `cert.tgz`：

| Helm 配置项 | 压缩包内文件 |
| ----------- | ------------ |
| `initPackages.gseCert.ca` | `cert/gseca.crt` |
| `initPackages.gseCert.cert` | `cert/gse_server.crt` |
| `initPackages.gseCert.key` | `cert/gse_server.key` |
| `initPackages.gseCert.certEncryptKey` | `cert/cert_encrypt.key` |
| `initPackages.gseCert.apiClient.cert` | `cert/gse_api_client.crt` |
| `initPackages.gseCert.apiClient.key` | `cert/gse_api_client.key` |
| `initPackages.gseCert.agent.cert` | `cert/gse_agent.crt` |
| `initPackages.gseCert.agent.key` | `cert/gse_agent.key` |

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

删除本次不需要导入的包参数。每种类型在一次命令中只能指定一个文件；需要导入多个同类型插件包时，应分别执行命令。

插件类型由对应的命令参数决定。`--generation` 只作用于 Agent、Proxy、Server 和节点二进制工具包，使用 `--init-plugin-v2`、`--init-external-plugin-v2` 或 `--init-plugin-v3` 时不需要设置。

可选参数 `--overwrite` 会在上传元数据中加入 `overwrite=true`。不要依赖该参数替换已经发布的同名同版本插件，重复包的处理结果由 `file` 服务和包内容决定；删除或替换已有 release 前应先确认命令结果。

### 多租户部署

`init_package.py` 使用 `--tenant-id` 设置 `X-Bk-Tenant-Id` 请求头。Helm Job 没有传入该参数，因此会使用脚本默认值 `default`。

多租户部署需要针对每个目标租户分别执行插件导入命令，并设置正确的 `--tenant-id`。BKRepo 项目名称转换由 `file` 服务配置处理，不要在插件包路径中自行添加租户前缀。存储配置说明参考 [BKRepo 资源使用说明](thirdparty_bkrepo_resource.md)。

## 确认导入结果

手工导入成功时，脚本输出以下汇总并以状态码 `0` 退出：

```text
summary: <count> success, 0 failed
```

在将插件用于安装或升级策略前，还应进入节点管理的插件包页面，确认对应的插件名称和版本已经出现。

对于 Helm Job，可以在 Hook Pod 运行期间或失败后查看状态和日志：

```bash
kubectl -n "${NAMESPACE}" get pod -l app=nodemgr-file-init-packages
kubectl -n "${NAMESPACE}" logs -l app=nodemgr-file-init-packages --all-containers=true
```

成功安装后找不到该 Pod 属于正常现象，因为 Hook 使用了 `hook-succeeded` 删除策略。

## 常见问题

| 现象 | 检查项 |
| ---- | ------ |
| 没有执行初始化 Job | 确认 `file.enabled=true` 和 `initPackages.enabled=true`。在升级场景还需确认 `initPackages.installOnly=false`。 |
| Job 输出 `No packages to upload` | 镜像中不存在 `bintool.tgz` 或 `plugin_bintool.tgz`，同时也没有提供完整的证书文件。普通插件包不会被该 Job 自动发现。 |
| 输出 `file service not ready within 300s` | 检查 `file` Deployment、Service、`file.config.infoServer.port` 和 `readinessProbe.path`。 |
| 输出 `jwt-generator binary not found` | 使用通过项目 Makefile 完整构建的 bk-nodemgr 镜像，构建产物必须包含 `/bk-nodemgr/support-files/initpackage/jwt-generator`。 |
| 鉴权失败 | 对比导入脚本的 JWT 密钥、过期时间与已部署的 `file` 鉴权配置，并检查 `bkUsername`、`loginName` 和租户 ID。 |
| 上传或发布失败 | 确认压缩包符合对应的 V2、V2 外部插件或 V3 包格式，且包内元数据有效；通过 `file` 服务日志查看具体校验错误。 |
| 部分包成功但命令仍返回非零状态 | 脚本会在单个任务失败后继续处理，并在最后返回失败。根据各任务输出，只重试失败的插件包。 |

脚本完整参数以及 Docker、本地执行示例还可参考 [`support-files/initpackage/README.md`](../../../support-files/initpackage/README.md)。
