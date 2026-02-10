# Init Package 初始化包功能脚本

本目录存放 bk-nodemgr 初始化包功能相关的脚本和工具。
提供了一个 `init_package.py` 脚本用于生成 JWT 令牌并上传初始化包文件到文件服务。
便于运维人员在部署 bk-nodemgr 时快速完成初始化包的准备工作。

## 文件说明

| 文件                  | 说明 |
|---------------------|------|
| `init_package.py`   | 初始化包功能主脚本 |
| `jwt_generator/`    | JWT 生成工具源码目录 |

## init_package.py 参数

| 参数               | 说明 |
|------------------|------|
| `--jwt-key`      | JWT 对称加密密钥，需与后端配置文件中 `basicServer.file.jwtServerConfig.symmetricKey` 的值保持一致 |
| `--expire-hours` | JWT 令牌过期时间，单位为小时，需与后端配置文件中 `basicServer.file.jwtServerConfig.expireHours` 的值保持一致 |
| `--host`         | 文件服务地址，需与后端配置文件中 `basicServer.file.host` 的值保持一致，或使用文件服务的路由地址 |
| `--port`         | 文件服务端口，需与后端配置文件中 `basicServer.file.port` 的值保持一致 |
| `--tenant-id`    | 租户 ID，用于区分不同租户的初始化包 |
| `--init-cert`    | 证书初始化包文件路径，需上传到文件服务的 `cert` 目录下 |
| `--init-bintool` | 二进制工具初始化包文件路径，需上传到文件服务的 `bintool` 目录下 |
| `--init-plugin-bintool` | 插件二进制工具初始化包文件路径，需上传到文件服务的 `plugin_bintool` 目录下 |
| `--init-agent` | Agent 初始化包文件路径，需上传到文件服务的 `agent` 目录下（如果有） |
| `--init-proxy` | Proxy 初始化包文件路径，需上传到文件服务的 `proxy` 目录下（如果有） |
| `--init-server` | Server 初始化包文件路径，需上传到文件服务的 `server` 目录下（如果有） |
| `--init-plugin-v2` | V2 插件初始化包文件路径，需上传到文件服务的 `v2/plugin` 目录下 |
| `--init-external-plugin-v2` | V2 外部插件初始化包文件路径，需上传到文件服务的 `v2/external_plugin` 目录下 |
| `--init-plugin-v3` | V3 插件初始化包文件路径，需上传到文件服务的 `v3/plugin` 目录下 |
| `--generation` | 生成版本（默认：2），用于指定初始化包的版本号 |
| `--overwrite` | 是否覆盖已存在的文件（布尔值，指定即为覆盖） |

## 工作流程说明

`init_package.py` 脚本执行时遵循以下两个阶段的工作流程：

1. **Upload 阶段 (上传)**: 脚本首先将指定的本地初始化包文件上传到文件服务的临时区域。上传成功后，文件服务会返回一个唯一的 `upload_id`。
2. **Publish 阶段 (发布)**: 脚本使用上一步获得的 `upload_id` 调用发布接口，将文件正式移动到目标目录下并完成初始化。

**注意事项**:
- 脚本会按顺序处理每个指定的初始化任务。
- 如果任一阶段（上传或发布）失败，脚本会记录错误并跳过该项，继续执行下一个初始化任务。
- 所有任务执行完毕后，脚本会输出成功与失败的汇总信息。

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
# jwt_key 可以参考 backend 配置文件中 basicServer.file.jwtServerConfig.symmetricKey 的值
docker exec -it <container_name> /bk-nodemgr/support-files/initpackage/jwt-generator \
    -k ${jwt_key} \
    -e 1
```

#### init_package.py

```bash
# jwt_key 可以参考 backend 配置文件中 basicServer.file.jwtServerConfig.symmetricKey 的值
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
# jwt_key 可以参考backend配置文件中basicServer.file.jwtServerConfig.symmetricKey的值
# expire_hours 过期时间，单位小时 可以参考backend配置文件中basicServer.file.jwtServerConfig.expireHours的值
./jwt-generator -k ${jwt_key} -e ${expire_hours}
```

#### init_package.py

```bash
# jwt_key 可以参考backend配置文件中basicServer.file.jwtServerConfig.symmetricKey的值
# expire_hours 过期时间，单位小时 可以参考backend配置文件中basicServer.file.jwtServerConfig.expireHours的值
# host 可以参考file配置文件中basicServer.file.host的值,或file service的路由地址
# port 可以参考file配置文件中basicServer.file.port的值
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
