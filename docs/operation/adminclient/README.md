# Admin Client 运维命令

`bk-nodemgr-adminclient` 是节点管理的统一运维 CLI，面向运维人员、运维脚本和安装/升级 Job。在服务镜像中位于 `/bk-nodemgr/bin/bk-nodemgr-adminclient`。

## 定位与职责边界

- **复用服务配置与凭据加载**：从已有服务配置中获取连接信息，并复用 KMS 解密能力，避免运维脚本各自解析配置、解密信封或维护另一套明文凭据。
- **提供可直接调用的运维能力**：支持服务管理、存储检查与客户端代理，以及安装/升级所需的应用凭据获取。脚本和 Helm Job 调用对应子命令，负责外围流程编排与重试。
- **绑定容器内的数据库客户端**：`database etcd`、`database redis`、`database mongodb` 分别调用容器内的 `etcdctl`、`redis-cli`、`mongo`，依赖这些客户端在 `PATH` 中可用及其版本支持的命令和参数。adminclient 负责加载配置、注入连接信息和凭据并调用客户端，不替代数据库客户端本身；脱离服务镜像运行时需自行准备对应客户端。`database check` 是例外，使用内置 Go 客户端探测，不依赖这些外部命令。
- **保持服务端业务归属**：Backend 管理操作通过 admin API 执行，业务校验和状态变更仍由 Backend 负责；本地存储检查、客户端代理和 KMS 凭据读取不依赖 Backend admin API。
- **按需运行，不承担常驻服务职责**：adminclient 是独立命令行工具，不替代 Backend 等服务，也不负责凭据持久化或轮换。执行时需具备目标配置、相关密钥和网络访问条件；具体依赖见各子命令文档。

`bkapp secret` 会显式输出明文凭据，供调用脚本捕获；不要将输出写入日志或开启 shell 命令追踪。存储代理的凭据可见性限制见 [database](database.md)。

## 子命令

- [backend](backend.md)：Backend 管理操作，目前记录租户初始化。
- [database](database.md)：存储连通性检查（check）与 etcd/redis/mongodb 客户端代理，支持 KMS 凭据覆盖。
- [bkapp](bkapp.md)：从 KMS global envelope 输出蓝鲸应用 appCode/appSecret，供安装/升级 Job 使用。
- [jwt_token](jwt_token.md)：用 KMS global envelope 中的 JWT 密钥为指定服务的 HTTP server 签发 token，供安装/升级脚本调用需要认证的内部 API。
