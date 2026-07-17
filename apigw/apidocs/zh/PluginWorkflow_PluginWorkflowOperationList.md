### 描述

- 该接口提供版本：v3.0.1-alpha.1+。
- 版本变更：`v3.0.1-alpha.6+` 将总数字段调整为 `total`；`v3.0.1-alpha.13+` 增加权限响应字段；`v3.0.1-alpha.18+` 增加业务 ID 条件字段；`v3.0.1-alpha.19+` 增加 `plugin_operate` 鉴权。
- 该接口所需权限：plugin_operate（操作插件）。
- 该接口功能描述：查询指定插件任务流下的操作列表，支持分页、精确条件过滤和仅返回总数。

### URL

POST /api/v3/plugin/workflow/operation/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| only_count | bool | 否 | 是否只返回总数，不返回操作详情 |
| page | object | 否 | 分页配置 |
| workflow_id | string | 是 | 插件任务流 ID，不可为空 |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件，当前对象为空 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| offset | int32 | 否 | 分页起始位置，起始值为 `0` |
| limit | int32 | 否 | 每页条数，最大值为 `500` |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
| --- | --- | --- | --- |
| bk_host_id | int64 array | 否 | 主机 ID 列表 |
| plugin_name | string array | 否 | 插件名称列表 |
| plugin_version | string array | 否 | 插件版本列表 |
| state | string array | 否 | 操作状态列表，可选值：`init`、`launched`、`running`、`success`、`failed`、`timeout`、`terminated` |
| bk_biz_id | int64 array | 否 | 业务 ID 列表；协议已声明，但当前后端转换器不使用该字段进行过滤 |

#### fuzzy_include_conditions

当前对象为空，传空对象 `{}` 或不传。

### 调用示例

```json
{
  "only_count": false,
  "page": {
    "offset": 0,
    "limit": 20
  },
  "workflow_id": "wf-plugin-0001",
  "exact_include_conditions": {
    "bk_host_id": [1001],
    "plugin_name": ["bkmonitorbeat"],
    "plugin_version": ["3.6.0"],
    "state": ["running"]
  },
  "fuzzy_include_conditions": {}
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 1,
    "operations": [
      {
        "operation_id": "op-plugin-0001",
        "instance_ids": ["inst-plugin-0001"],
        "operator": "admin",
        "create_time": 1719820000000,
        "plugin_deployment_info": {
          "bk_host_id": 1001,
          "bk_biz_id": 2,
          "bk_networkarea_id": 0,
          "bk_networkunit_id": 1,
          "bk_host_innerip_list": ["10.0.0.1"],
          "bk_host_innerip_v6_list": ["fe80::1"],
          "plugin_name": "bkmonitorbeat",
          "plugin_version": "3.6.0"
        },
        "latest_oper_inst_brief_data": {
          "life_cycle": {
            "state": "running",
            "create_time": 1719820000000,
            "start_time": 1719820010000,
            "end_time": 0
          },
          "latest_action_inst_brief_data": {
            "name": "upgrade_plugin",
            "tags": ["main"]
          }
        }
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| operations | array | 操作列表 |
| total | int64 | 当前过滤条件命中的总记录条数 |

#### data.operations[n]

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| operation_id | string | 操作 ID |
| instance_ids | string array | 操作实例 ID 列表 |
| operator | string | 操作人 |
| create_time | int64 | 操作创建时间，Unix 毫秒时间戳 |
| plugin_deployment_info | object | 插件部署信息 |
| latest_oper_inst_brief_data | object | 最新操作实例摘要 |

#### data.operations[n].plugin_deployment_info

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| bk_host_id | int64 | 主机 ID |
| bk_biz_id | int64 | 业务 ID |
| bk_networkarea_id | int64 | 管控区域 ID |
| bk_networkunit_id | int64 | 管控单元 ID |
| bk_host_innerip_list | string array | 主机内网 IPv4 列表 |
| bk_host_innerip_v6_list | string array | 主机内网 IPv6 列表 |
| plugin_name | string | 插件名称 |
| plugin_version | string | 插件版本 |

#### data.operations[n].latest_oper_inst_brief_data

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| life_cycle | object | 生命周期摘要 |
| latest_action_inst_brief_data | object | 最新动作实例摘要 |

#### data.operations[n].latest_oper_inst_brief_data.life_cycle

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| state | string | 生命周期状态，可选值：`init`、`launched`、`running`、`success`、`failed`、`timeout`、`terminated` |
| create_time | int64 | 创建时间，Unix 毫秒时间戳 |
| start_time | int64 | 开始时间，Unix 毫秒时间戳 |
| end_time | int64 | 结束时间，Unix 毫秒时间戳 |
| stop_time | int64 | 停止时间，Unix 毫秒时间戳；协议已声明，但当前后端转换器不填充 |

#### data.operations[n].latest_oper_inst_brief_data.latest_action_inst_brief_data

| 参数名称 | 参数类型 | 描述 |
| --- | --- | --- |
| name | string | 最新动作名称 |
| tags | string array | 最新动作标签 |

### 处理规则说明

- 后端根据任务流关联的 `trigger_id` 查询操作，并根据操作 token 关联插件部署信息。
- `bk_biz_id` 已在请求 Proto 中声明，但当前转换器不将其传入操作或插件部署查询条件，因此该字段当前不产生过滤效果。
- 当 `only_count=true` 时，返回 `data.total`，`data.operations` 为空数组。
- 最新操作实例的 `life_cycle.stop_time` 已在共享 Proto 中声明，但当前响应转换器未设置该字段，因此成功示例不包含它。
