# stopoperinst

## 设计意图

使用 MongoDB 短期标记记录需要停止的操作实例，并由 workflow storage 定向轮询当前 backend 正在等待的操作实例。

## 功能边界

1. 此包负责：
    - `stopping_operation_inst` 的存储结构与索引
    - 停止标记的 Upsert、全量查询和按操作实例 ID 查询
    - 通过 TTL 索引清理过期停止标记
2. 此包不负责：
    - 轮询任务调度
    - 本地订阅管理和停止事件通知
    - operation instance 的具体停止逻辑

## 设计考量

停止标记是短期状态，不是工作流历史。写入时会刷新过期时间，相同 `oper_inst_id` 只保留一条记录。
workflow storage 每秒只查询本实例当前订阅的 `oper_inst_id`，避免 MongoDB Change Stream 在高写入环境持续扫描大量无关 oplog。

## 使用限制

1. 热路径必须使用 `FindByIDs` 定向查询，不能每秒执行全表查询
2. `FindAll` 仅用于低频全量同步兜底
3. TTL 必须覆盖轮询和全量同步的恢复窗口
4. 复杂业务编排应放在 `internal/backend/storage/workflow`
