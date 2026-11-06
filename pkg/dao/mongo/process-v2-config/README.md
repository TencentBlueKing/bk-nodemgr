# processv2config

## 设计意图

提供节点管理 V2 兼容模式插件配置（`plugin_v2`）的 DAO 层操作，主要用于对 MongoDB 的操作。
表 `process_v2_config_<tenant_id>` 与 V3 的 `process_config` 表结构一致，但当前只写入兼容插件的主配置文件。

## 功能边界

1. 此包负责：
    - V2 兼容插件配置的 DAO 层操作
    - V2 兼容插件配置的存储结构设计
2. 此包不负责：
    - 具体的业务关系和逻辑
    - 主配置与子配置的来源划分（由工作流 action 决定）

## 设计考量

该包将数据库操作和表结构屏蔽在内，对外提供基于 pkg/types 类型的接口。
唯一键为 `{name, process_name, host_id, is_main_config}`，与 V3 配置表保持一致，
因此后续扩展子配置存储时无需变更表结构。

## 使用限制

1. 调用 `DeleteMany` 必须携带明确的过滤条件；该接口为硬删除
2. `Get` 按 `{host_id, process_name, name, is_main_config}` 唯一键读取单条配置，`Count`/`List` 按 `types.ProcessConfigCondition` 过滤（与 V3 `processconfig` 同形，不支持 fuzzy 条件）
3. 仅提供原子接口，复杂的业务场景逻辑查询应放在上层处理

## 演进方向

1. 按需补充子配置写入
2. 按需丰富数据字段
