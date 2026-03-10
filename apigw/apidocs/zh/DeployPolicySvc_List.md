### 描述

- 该接口提供版本：v3.0.1+
- 该接口所需权限：
- 该接口功能描述：查询部署策略列表，支持分页和条件过滤。

### URL

POST /api/v3/deploy_policy/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| page | object | 否 | 分页配置 |
| only_count | bool | 否 | 是否只返回总数，不返回详情 |
| exact_include_conditions | object | 否 | 精确匹配包含条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配包含条件 |
| exact_exclude_conditions | object | 否 | 精确匹配排除条件 |
| fuzzy_exclude_conditions | object | 否 | 模糊匹配排除条件 |
| executed_time_range | object | 否 | 执行时间范围 |

#### page

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| count | bool | 是 | 是否返回总记录条数 |
| start | uint32 | 否 | 记录开始位置，起始值为0 |
| limit | uint32 | 否 | 每页限制条数，最大500 |
| sort | string | 否 | 排序字段 |
| order | string | 否 | 排序顺序（ASC、DESC） |

#### exact_include_conditions

精确匹配包含条件，满足任一条件即匹配。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| deploy_policy_id | int64 array | 否 | 部署策略ID列表 |
| dsu_id | int64 array | 否 | DSU ID列表 |
| deploy_policy_name | string array | 否 | 部署策略名称列表 |
| operator | string array | 否 | 操作人列表 |
| enabled | bool array | 否 | 启用状态列表 |

#### fuzzy_include_conditions

模糊匹配包含条件，满足任一条件即匹配。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| deploy_policy_name | string array | 否 | 部署策略名称列表（模糊匹配） |
| operator | string array | 否 | 操作人列表（模糊匹配） |

#### exact_exclude_conditions

精确匹配排除条件，满足任一条件即排除。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| deploy_policy_id | int64 array | 否 | 部署策略ID列表 |
| dsu_id | int64 array | 否 | DSU ID列表 |
| deploy_policy_name | string array | 否 | 部署策略名称列表 |
| operator | string array | 否 | 操作人列表 |
| enabled | bool array | 否 | 启用状态列表 |

#### fuzzy_exclude_conditions

模糊匹配排除条件，满足任一条件即排除。

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| deploy_policy_name | string array | 否 | 部署策略名称列表（模糊匹配） |
| operator | string array | 否 | 操作人列表（模糊匹配） |

#### executed_time_range

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| start | string | 否 | 开始时间，标准格式：2006-01-02T15:04:05Z |
| end | string | 否 | 结束时间，标准格式：2006-01-02T15:04:05Z |

### 调用示例

查询启用状态的部署策略，按创建时间倒序排列。

```json
{
  "page": {
    "count": true,
    "start": 0,
    "limit": 10,
    "sort": "created_at",
    "order": "DESC"
  },
  "exact_include_conditions": {
    "enabled": [true]
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "total": 25,
    "items": [
      {
        "deploy_policy_id": 1001,
        "dsu_id": 2001,
        "meta": {
          "name": "生产环境监控插件部署策略",
          "description": "用于生产环境统一部署监控采集插件"
        },
        "specs": [
          {
            "type": "specify_plugin",
            "param": {
              "plugin_name": "bkmonitorbeat",
              "version": "3.60.3066"
            }
          }
        ],
        "scopes": [
          {
            "type": "topo",
            "scope": {
              "granularity": "host",
              "bk_biz_id": 100,
              "filter": {},
              "paths": [
                {
                  "topo_obj_id": "biz",
                  "topo_inst_id": 100
                }
              ]
            }
          }
        ],
        "operator": "admin",
        "enabled": true
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求ID |
| error | object | 错误信息（成功时为空） |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| total | int64 | 当前规则能匹配到的总记录条数 |
| items | array | 查询返回的数据 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| deploy_policy_id | int64 | 部署策略ID |
| dsu_id | int64 | DSU ID |
| meta | object | 部署策略元信息 |
| specs | array | 部署规范列表 |
| scopes | array | 部署范围列表 |
| operator | string | 操作人 |
| enabled | bool | 是否启用 |

#### data.items[n].meta

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| name | string | 部署策略名称 |
| description | string | 部署策略描述 |

#### data.items[n].specs[n]

部署规范定义了期望的最终状态。

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| type | string | 规范类型（枚举值：specify_plugin、specify_plugin_pkg、specify_plugin_sub_config） |
| param | object | 规范参数，结构根据type字段而定 |

**type 为 specify_plugin 时，param 结构：**

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| plugin_name | string | 插件名称 |
| version | string | 插件版本 |
| custom_config_context | object | 自定义配置上下文 |

**type 为 specify_plugin_pkg 时，param 结构：**

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| plugin_pkg_name | string | 插件包名称 |
| version | string | 插件包版本 |
| custom_config_context | object | 自定义配置上下文 |

**type 为 specify_plugin_sub_config 时，param 结构：**

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| plugin_name | string | 插件名称 |
| config_files_detail | array | 配置文件详情列表 |
| custom_config_context | object | 自定义配置上下文 |

**config_files_detail[n] 结构：**

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| name | string | 配置文件名 |
| content | string | 配置文件内容 |
| is_main_config | bool | 是否为主配置文件 |

#### data.items[n].scopes[n]

部署范围定义了策略应用的目标范围。

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| type | string | 范围类型（枚举值：topo、service_template、set_template、instance、dynamic_group） |
| scope | object | 范围详情，结构根据type字段而定 |

**type 为 topo 时，scope 结构：**

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| granularity | string | 目标粒度（枚举值：host、service_instance） |
| bk_biz_id | int64 | 业务ID |
| filter | object | 目标过滤器 |
| paths | array | 拓扑路径列表 |

**paths[n] 结构：**

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| topo_obj_id | string | 拓扑对象ID |
| topo_inst_id | int64 | 拓扑实例ID |

**type 为 service_template 时，scope 结构：**

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| granularity | string | 目标粒度（枚举值：host、service_instance） |
| bk_biz_id | int64 | 业务ID |
| filter | object | 目标过滤器 |
| service_template_ids | int64 array | 服务模板ID列表 |
| module_ids | int64 array | 模块ID列表 |

**type 为 set_template 时，scope 结构：**

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| granularity | string | 目标粒度（枚举值：host、service_instance） |
| bk_biz_id | int64 | 业务ID |
| filter | object | 目标过滤器 |
| set_template_ids | int64 array | 集群模板ID列表 |
| set_ids | int64 array | 集群ID列表 |

**type 为 instance 时，scope 结构：**

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| granularity | string | 目标粒度（枚举值：host、service_instance） |
| bk_biz_id | int64 | 业务ID |
| filter | object | 目标过滤器 |
| instance_ids | int64 array | 实例ID列表（host_id或service_instance_id） |

**type 为 dynamic_group 时，scope 结构：**

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| granularity | string | 目标粒度（仅支持：host） |
| bk_biz_id | int64 | 业务ID |
| filter | object | 目标过滤器 |
| dynamic_group_ids | string array | 动态分组ID列表 |

**注意**：dynamic_group 类型不支持 service_instance 粒度。
