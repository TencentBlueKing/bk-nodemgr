# Admin Client 运维命令

`bk-nodemgr-adminclient` 是独立的管理 CLI，在镜像中位于 `/bk-nodemgr/bin/bk-nodemgr-adminclient`。

## 子命令

- [backend](backend.md)：Backend 管理操作，目前记录租户初始化。
- [database](database.md)：存储连通性检查（check）与 etcd/redis/mongodb 客户端代理，支持 KMS 凭据覆盖。
