# filecache

## 设计意图

1. 提供通用的本地文件缓存能力，消除多服务重复从 file 服务下载同一制品（release 包、installer 等）的开销。
2. 以 MD5 值命名隔离子目录（`{cacheBaseDir}/{md5}/`），保证不同版本的文件互不干扰，旧版引用在新版写入期间仍然有效。
3. 解耦下载逻辑与缓存逻辑：调用方只需提供 `fetchFn`，缓存层负责命中判断、原子写入、并发控制和过期清理。

## 功能边界

此包负责：
- 基于 MD5 的缓存命中判断（filename + MD5 双键索引）
- 下载过程的并发序列化（per-filename 锁 + double-check）
- 原子写入：写入隔离子目录，写完后更新索引
- 启动恢复：可选扫描磁盘重建内存索引（`RestoreOnStart`）
- 后台 GC：定期清理过期条目和孤立 MD5 目录
- 安全防护：MD5 格式校验（防路径穿越）、`safeRemoveAll`（防误删 baseDir 外路径）

此包不负责：
- 具体的下载实现（由调用方通过 `fetchFn` 提供）
- 文件内容的业务解析
- 跨进程或分布式缓存

## 设计考量

1. **目录名即校验值**：以 MD5 hex 串作为子目录名，restore 时无需重新计算 MD5，只需读目录名。
2. **两层锁设计**：全局 `sync.RWMutex` 保护内存索引（短暂持有），per-filename `sync.Mutex`（存于 `sync.Map`）保护下载过程，避免同一文件被并发重复下载。
3. **GC 与下载的 TOCTOU 防护**：`pendingDirs` 标记正在写入的目录，GC 扫描孤立目录时会跳过，避免删除未完成的写入。
4. **`RestoreOnStart` 默认关闭**：大多数场景服务重启后制品会重新下载，启用恢复会延长启动时间；仅在文件较大且重启频繁时建议开启。

## 使用限制

1. `GetOrFetch` 返回的 `fileiface.File` 代表缓存中的文件；调用方通过 `Content()` 获取 `io.ReadCloser` 并负责关闭，**不得对同一 `io.ReadCloser` 关闭两次**。
2. `fetchFn` 返回的 `io.ReadCloser` 由缓存下载函数通过 defer 关闭（`LocalDir.Store` 仅借用），调用方不应在传入后再次关闭。
3. `expectedMD5` 必须是合法的 32 字符十六进制字符串；否则 `GetOrFetch` 直接返回错误。
4. 缓存的 baseDir 及其所有子目录由此包独占管理，不应在外部直接创建或删除其中的文件和目录。

## 演进方向

1. 支持多文件同时预热（批量 `GetOrFetch`）。
2. 支持基于 LRU 或容量上限的淘汰策略，作为时间过期策略的补充。
3. 提供 metrics 埋点（cache hit rate、GC 清理数量、下载耗时）。
