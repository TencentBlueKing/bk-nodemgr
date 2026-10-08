### 描述

- 该接口提供版本：v3.0.1-alpha.91+
- 该接口所需权限：无新增 IAM 权限；保留现有身份认证和租户隔离。
- 该接口功能描述：同步预览一个已启用部署策略的当前状态，按原始 Spec 返回每个 Target 的身份信息和满足、不满足、非管理状态。
- 首期仅提供 Backend 接口。

### URL

POST /api/v3/deploy_policy/preview

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| deploy_policy_id | int64 | 是 | 当前租户内的部署策略 ID，必须非负；0 表示具体策略 ID，不表示全部策略 |

### 调用示例

```json
{"deploy_policy_id": 1001}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-example",
  "data": {
    "items": [
      {
        "spec": {
          "type": "specify_plugin",
          "param": {
            "plugin_name": "example",
            "version": "2.0",
            "custom_config_context": {}
          }
        },
        "results": [
          {
            "target": {
              "host": {
                "bk_host_id": 1
              },
              "service_instance": null
            },
            "status": "satisfied"
          },
          {
            "target": {
              "host": {
                "bk_host_id": 2
              },
              "service_instance": null
            },
            "status": "unsatisfied"
          },
          {
            "target": {
              "host": {
                "bk_host_id": 3
              },
              "service_instance": null
            },
            "status": "unmanaged"
          },
          {
            "target": {
              "host": {
                "bk_host_id": 4
              },
              "service_instance": {
                "id": 101,
                "bk_module_id": 20
              }
            },
            "status": "satisfied"
          }
        ]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息 |
| permission | object | 权限信息 |
| data | object | 完整计算结果 |
| data.items | array | 按请求策略原始 Spec 顺序返回；空 Specs 返回空数组 |
| data.items[].spec | object | 原始 Spec，包含 type 和 param，定义与策略查询接口一致 |
| data.items[].results | array | 统一的目标结果列表；无目标时返回 [] |
| data.items[].results[].target | object | 本次 Scope 计算保留的目标身份信息 |
| data.items[].results[].target.host | object | 主机身份；主机目标和服务实例目标均返回 |
| data.items[].results[].target.host.bk_host_id | int64 | CMDB 主机 ID |
| data.items[].results[].target.service_instance | object/null | 服务实例身份；主机目标返回 null |
| data.items[].results[].target.service_instance.id | int64 | CMDB 服务实例 ID |
| data.items[].results[].target.service_instance.bk_module_id | int64 | 服务实例所属的 CMDB 模块 ID |
| data.items[].results[].status | string | satisfied：冲突后保留且该 Spec 无变更任务；unsatisfied：冲突后保留且该 Spec 有变更任务；unmanaged：被现有策略冲突处理排除 |

### 计算语义与限制

- 每次请求都重新执行关联策略发现、Scope 计算、冲突处理和状态分析；不分页，不读取历史执行结果。
- 关联策略参与计算，但只返回被请求策略的结果。同一策略各 Spec 共享冲突后保留的目标范围。
- 满足状态完全复用 execute 的现有变更任务判定，包括其跳过行为；不增加额外的配置内容、进程健康或主机文件检查。
- 目标集合沿用现有 Scope 计算与去重规则。每个 Spec 的 results 按主机 ID、模块 ID 升序排列；空列表返回 []。身份信息只取自本次计算结果，不额外查询 CC。
- 只展示当前 Scope 计算出的目标。分析器产生的 Scope 外清理任务不会增加响应中的实例。
- 查询只读：不执行变更任务，不创建 trigger、operation、workflow，不更新执行时间或 dsu_id。
- 参数无效、策略不存在、策略未启用、关联策略发现失败、CC 查询失败或分析失败时，整次请求返回错误，不返回部分结果。
- 数据以查询时现有数据源为准，不主动探测主机，也不提供跨数据源的原子快照。耗时受 Scope 大小、关联策略和依赖接口影响，沿用请求取消和超时限制。
