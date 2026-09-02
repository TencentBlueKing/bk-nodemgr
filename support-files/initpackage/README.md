# Init Package 初始化包功能脚本

本目录存放 bk-nodemgr 初始化包功能相关的脚本和工具。
提供了一个 `init_package.py` 脚本用于生成 JWT 令牌并上传初始化包文件到文件服务。
便于运维人员在部署 bk-nodemgr 时快速完成初始化包的准备工作。

## 文件说明

| 文件              | 说明                 |
| ----------------- | -------------------- |
| `init_package.py` | 初始化包功能主脚本   |
| `jwt_generator/`  | JWT 生成工具源码目录 |

## init_package.py 参数

命令行显式参数优先于 File 配置。`--config-file` 可指定配置路径；未指定时依次尝试 `/bk-nodemgr/etc/file_conf.yaml` 和 `/bk-nodemgr/etc/bk-nodemgr-file.yml`，使用 PyYAML `safe_load` 读取。未显式指定配置文件时，配置读取失败或不存在会使用安全回落值；显式指定 `--config-file` 失败时脚本会报错退出，脚本会打印配置来源（如有）以及最终 tenant/host/port，但不会打印 JWT key。

| 参数                            | 说明                                                                                                                                               |
| ------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `--config-file`                 | File YAML 配置路径；默认依次尝试上述两个路径                                                                                                       |
| `--host`                        | CLI 优先，否则 `localhost`                                                                                                                         |
| `--port`                        | CLI 优先，否则读取 File 配置 `file.config.basicServer.port`，再回落 `28202`                                                                        |
| `--jwt-key`                     | CLI 优先；未显式指定时仅在 File 配置 `file.config.basicServer.authIdentity` 不为 `none` 时读取其 `jwtServerConfig.symmetricKey`，未启用 JWT 时为空 |
| `--expire-hours`                | 仅使用 CLI 显式设置；File 配置没有 JWT 有效期字段，未设置时默认为 `24` 小时                                                                        |
| `--bk-username`、`--login-name` | CLI 优先，否则 `admin`                                                                                                                             |
| `--tenant-id`                   | CLI 优先，否则读取 File 配置 `file.config.tenantMode`；`multiple` 使用 `system`，其他情况使用 `default`                                            |
| `--init-cert`                   | 证书初始化包文件路径，可重复指定                                                                                                                   |
| `--init-bintool`                | 二进制工具初始化包文件路径                                                                                                                         |
| `--init-plugin-bintool`         | 插件二进制工具初始化包文件路径                                                                                                                     |
| `--init-agent`                  | Agent 初始化包文件路径                                                                                                                             |
| `--init-proxy`                  | Proxy 初始化包文件路径                                                                                                                             |
| `--init-server`                 | Server 初始化包文件路径                                                                                                                            |
| `--init-plugin-v2`              | V2 插件初始化包文件路径                                                                                                                            |
| `--init-external-plugin-v2`     | V2 外部插件初始化包文件路径                                                                                                                        |
| `--init-plugin-v3`              | V3 插件初始化包文件路径                                                                                                                            |
| `--generation`                  | 生成版本（默认：2）                                                                                                                                |
| `--overwrite`                   | 是否覆盖已存在的文件                                                                                                                               |

单租户下可以只指定包路径；多租户未传 `--tenant-id` 时使用 `system`，其他租户须显式传入。多租户初始化要求 File 的 `config.basicServer.authIdentity` 为 `rest-server`，且 `config.basicServer.jwtServerConfig.symmetricKey` 非空并与生成 JWT 使用的密钥匹配。`authIdentity: none` 会使请求在 `default` tenant 处理，不能作为 `tenantMode: multiple` 的初始化配置。

`--auto-select` 会扫描 `/bk-nodemgr/file/packages`，也可用 `--packages-dir` 覆盖。只处理普通文件，不解压、不检测压缩包、不删除源文件；同一目录中的多个文件按字典序执行。目录映射为：`cert→cert`、`agent→agent`、`proxy→proxy`、`server→server`、`plugin-v2→plugin_v2`、`external-plugin-v2→external_plugin_v2`、`plugin-v3→plugin_v3`；`bintool/` 中 `plugin_bintool*`→`plugin_bintool`，其他普通文件→`bintool`。发现文件会追加到显式包，任务按 cert、bintool、plugin_bintool、agent、proxy、server、plugin_v2、external_plugin_v2、plugin_v3 固定顺序执行，同类型文件分别作为任务处理。

Helm 初始化 Job 会将 File ConfigMap 中的 `file_conf.yaml` 挂载到 `{{ .Values.common.etcPath }}/file_conf.yaml`，并通过 `--config-file` 使用该配置。连接端口、JWT 密钥和 tenant 均从 File 配置读取；JWT 有效期没有 File 配置字段，默认 24 小时，也可由 `--expire-hours` 显式覆盖。Job 仍通过 `--host` 显式使用 File Service DNS 覆盖配置中的 host，因为配置中没有该 Service DNS。File Pod 启动时会替换 `advertise` 和 `tracing` 占位符，因此 Job 挂载的 ConfigMap 内容不是与 File Pod 运行时文件逐字节相同的副本。

```bash
python3 /bk-nodemgr/support-files/initpackage/init_package.py --auto-select
```

没有显式包且没有发现包时输出 `No packages to upload` 并退出 0。上传、发布或 JWT 生成的单任务异常会记录失败并继续；最终逐文件列出成功/失败清单，存在失败时退出非零。

## 自动选择示例

```bash
kubectl exec -it <bk-nodemgr-file-pod> -- \
  python3 /bk-nodemgr/support-files/initpackage/init_package.py --auto-select
```

## 工作流程说明

`init_package.py` 脚本执行时遵循以下两个阶段的工作流程：

1. **Upload 阶段 (上传)**: 脚本首先将指定的本地初始化包文件上传到文件服务的临时区域。上传成功后，文件服务会返回一个唯一的 `upload_id`。
2. **Publish 阶段 (发布)**: 脚本使用上一步获得的 `upload_id` 调用发布接口，将文件正式移动到目标目录下并完成初始化。

**注意事项**:

- 脚本会按顺序处理每个指定的初始化任务。
- 如果任一阶段（上传、发布或 JWT 生成）失败，脚本会记录该文件失败并继续执行下一个任务。
- 最终汇总会逐文件列出成功与失败清单；存在失败时返回非零退出码。

## 使用方法

镜像构建后会在/bk-nodemgr/support-files/initpackage/目录下生成 `jwt-generator` 可执行文件和 `init_package.py` 脚本。

### k8s 部署方式

#### jwt-generator

```bash
# jwt_key 为{{ .Values.file.config.basicServer.jwtServerConfig.symmetricKey }}的配置
kubectl exec -it <bk-nodemgr-file-pod> -- /bk-nodemgr/support-files/initpackage/jwt-generator \
    -k ${jwt_key}\
    -e 1
```

#### init_package.py

```bash
# jwt_key 为{{ .Values.file.config.basicServer.jwtServerConfig.symmetricKey }}的配置
kubectl exec -it <bk-nodemgr-file-pod> -- python3 /bk-nodemgr/support-files/initpackage/init_package.py \
                    --jwt-key "${jwt_key}" \
                    --expire-hours 1 \
                    --host bk-nodemgr-file \
                    --port {{ .Values.file.config.basicServer.port }} \
                    --tenant-id default \
                    --init-cert /bk-nodemgr/file/packages/cert/cert.tgz\
                    --init-bintool /bk-nodemgr/file/packages/bintool/bintool.tgz \
                    --init-plugin-bintool /bk-nodemgr/file/packages/plugin_bintool/plugin_bintool.tgz
```

### Docker 部署方式

#### jwt-generator

```bash
# jwt_key 使用 File 配置中 file.config.basicServer.jwtServerConfig.symmetricKey 的值
docker exec -it <container_name> /bk-nodemgr/support-files/initpackage/jwt-generator \
    -k ${jwt_key} \
    -e 1
```

#### init_package.py

```bash
# jwt_key 使用 File 配置中 file.config.basicServer.jwtServerConfig.symmetricKey 的值
docker exec -it <container_name> python3 /bk-nodemgr/support-files/initpackage/init_package.py \
                    --jwt-key "${jwt_key}" \
                    --expire-hours 1 \
                    --host localhost \
                    --port 28202 \
                    --tenant-id default \
                    --init-cert /bk-nodemgr/file/packages/cert/cert.tgz \
                    --init-bintool /bk-nodemgr/file/packages/bintool/bintool.tgz \
                    --init-plugin-bintool /bk-nodemgr/file/packages/plugin_bintool/plugin_bintool.tgz
```

### 本地执行方式

#### jwt-generator

```bash
# jwt_key 使用 File 配置中 file.config.basicServer.jwtServerConfig.symmetricKey 的值
# expire_hours 过期时间，单位小时；File 配置没有 JWT 有效期字段，默认 24
./jwt-generator -k ${jwt_key} -e ${expire_hours}
```

#### init_package.py

```bash
# jwt_key 使用 File 配置中 file.config.basicServer.jwtServerConfig.symmetricKey 的值
# expire_hours 过期时间，单位小时；File 配置没有 JWT 有效期字段，默认 24
# host 使用 File Service 的路由地址
# port 使用 File 配置中 file.config.basicServer.port 的值
# tenant_id 为租户id
python3 ./init_package.py \
        --jwt-key "${jwt_key}" \
        --expire-hours ${expire_hours} \
        --host ${host} \
        --port ${port} \
        --tenant-id ${tenant_id} \
        --init-cert ${cert_path} \
        --init-bintool ${bintool_path} \
        --init-plugin-bintool ${plugin_bintool_path}
```
