# Redis Cluster 跨环境消费 Machinery 任务

## 现象

backend worker 在进入具体 action 前，无法根据 `oper-inst-id` 获取 operation instance：

```text
failed to get operation instance brief data. oper-inst-id(...): failed to get operation instance data, oper-inst-id(...): not found
```

错误稳定出现，但并非每个任务都会失败。相同操作可能部分成功、部分失败。

## 结论

当多套 bk-nodemgr 环境共用一个 Redis Cluster，且使用相同的 Machinery 队列命名空间时，不同环境的 worker 会竞争消费同一批任务。

失败任务并不是因为 MongoDB 写入后传播较慢，而是因为任务被错误环境消费。该环境使用任务中的 `oper-inst-id` 查询自己的 MongoDB，自然无法在 `oper_inst_data` 中找到另一套环境创建的数据。

## 根因链路

bk-nodemgr workflow 默认使用固定的 Machinery 队列名：

```text
operation_inst_engine_queue
```

Redis Cluster 只支持 database 0，不支持 `SELECT`。因此，Cluster 模式下配置不同的 `redis.db` 不能隔离环境。

完整链路如下：

1. 环境 A 创建 workflow 和 operation instance，并将 `oper_inst_data` 写入 Mongo A。
2. 环境 A 将 action task 投递到共用的 Redis Cluster。
3. 环境 A 和环境 B 的 worker 同时监听 `operation_inst_engine_queue`。
4. 如果环境 A 的 worker 消费任务，它能从 Mongo A 读取对应数据，任务正常执行。
5. 如果环境 B 的 worker 消费任务，它会用环境 A 生成的 `oper-inst-id` 查询 Mongo B。
6. Mongo B 不存在对应的 `oper_inst_data`，worker 返回 `not found`。

因此，错误是否出现取决于哪套环境的 worker 抢到任务。worker 数量、调度时机和 Redis 出队顺序都会影响失败比例。

## 为什么不是 MongoDB 传播延迟

backend MongoDB client 使用以下配置：

```go
ReadPreference: readpref.Primary(),
WriteConcern:   writeconcern.W1(),
```

写操作由 primary 确认后返回，后续查询也读取 primary。同一 MongoDB、同一 database 和同一 primary 上，写入成功后不需要等待 secondary 复制完成才能读取。

如果创建和查询由不同环境完成，即使两边 MongoDB 都工作正常，错误环境仍然查不到对应数据。

## 确认方法

### 1. 查询 operation instance 明细

在报错环境查询：

```js
db.getCollection("oper_inst_data").find(
  {
    "data.oper_inst_id": "oper-inst:<id>",
  },
  {
    basic: 1,
    "data.oper_inst_id": 1,
    "data.operation_id": 1,
    "data.trigger_id": 1,
    "data.life_cycle": 1,
  },
);
```

再在任务所属环境执行同一查询。如果仅任务所属环境存在记录，说明该 task 已跨环境消费。

### 2. 对齐 backend 日志

按同一个 `oper-inst-id` 对齐以下日志，并确认日志来自哪个环境和 pod：

```text
created operation instance
try to launch operation instance
send chain to machinery
failed to get operation instance brief data
```

如果前三条来自环境 A，而最后一条来自环境 B，即可确认跨环境消费。

### 3. 对比 Redis 配置

检查两套环境的以下配置：

- Redis `type` 是否均为 `cluster`
- Redis `addrs` 是否指向同一集群
- workflow queue 是否均为 `operation_inst_engine_queue`
- 是否误以为不同 `redis.db` 能隔离 Cluster 模式

## 处理原则

### 首选方案

每套 bk-nodemgr 环境使用独立的 Redis Cluster。该方案同时隔离 Machinery queue、result 和 lock 等 Redis key，边界最清晰。

### 共用 Redis 的前提

如果必须共用 Redis Cluster，需要先为每套环境提供完整的应用级 namespace。隔离范围不能只覆盖 queue，还必须覆盖 Machinery result、lock 及其他 workflow key。

当前系统没有提供完整的环境 namespace 配置前，不应让多套 bk-nodemgr 环境共用同一个 Redis Cluster。

### 禁止方案

不要使用不同的 `redis.db` 隔离 Redis Cluster。Cluster 模式只使用 database 0，该配置不能形成环境边界。

## 恢复检查

- [ ] 两套环境不再共用同一个 Redis Cluster queue namespace。
- [ ] 每套环境的 backend pod 使用一致的 MongoDB `hosts`、`database`、`replicaSet` 和 `authSource`。
- [ ] 在任务所属环境的 `oper_inst_data` 中能查到失败的 `oper-inst-id`。
- [ ] 在错误消费环境的 `oper_inst_data` 中查不到该 `oper-inst-id`。
- [ ] Redis 隔离后，在确认没有其他环境继续使用旧 namespace 的前提下，清理残留 Machinery task，再启动 worker。

## 代码定位

| 路径 | 作用 |
| --- | --- |
| `pkg/workflow/manager.go` | 设置 Machinery 默认队列 `operation_inst_engine_queue`，创建 Redis broker 和 backend。 |
| `pkg/config/config.go` | 定义 Redis 模式，并说明 `redis.db` 在 Cluster 模式下不生效。 |
| `internal/backend/manager/manager.go` | 将 workflow storage 和 Redis 配置注入 workflow manager。 |
| `pkg/workflow/controller.go` | 创建 operation instance、写入 `oper_inst_data` 并向 Machinery 投递任务。 |
| `pkg/workflow/worker.go` | 消费 action task，并在执行 action 前查询 operation instance 数据。 |

相关问题：[GitHub Issue #2983](https://github.com/TencentBlueKing/bk-nodemgr/issues/2983)。
