# Init Package 初始化包功能脚本

本目录存放 bk-nodemgr 初始化包功能相关的脚本和工具。
提供了一个 `init_package.py` 脚本用于生成 JWT 令牌并上传初始化包文件到文件服务。
便于运维人员在部署 bk-nodemgr 时快速完成初始化包的准备工作。

## 文件说明

| 文件                  | 说明 |
|---------------------|------|
| `init_package.py`   | 初始化包功能主脚本 |
| `jwt_generator/`    | JWT 生成工具源码目录 |

## 使用方法

镜像构建后会在/bk-nodemgr/support-files/initpackage/目录下生成 `jwt-generator` 可执行文件和 `init_package.py` 脚本。

### jwt-generator
```bash
# jwt_key 为{{ .Values.file.config.basicServer.jwtServerConfig.symmetricKey }}的配置
kubectl exec -it <bk-nodemgr-file-pod> -- /bk-nodemgr/support-files/initpackage/jwt-generator \
    -k ${jwt_key}\
    -e 1
```

### init_package.py

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
