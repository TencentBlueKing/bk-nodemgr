### 描述

- 该接口提供版本：v3.0.1-alpha.72+。
- 该接口所需权限：⚠️ 待确认（接口尚未实现 handler，权限待补充）。
- 该接口功能描述：启动一个插件的一次性调试工作流。调试单元包含目标范围（主机或服务实例）、插件、版本、配置模板及自定义配置上下文。

### URL

POST /api/v3/plugin/start_debug

### 输入参数

| 参数名称   | 参数类型 | 必选 | 描述                                                           |
| ---------- | -------- | ---- | -------------------------------------------------------------- |
| debug_info | object   | 是   | 插件调试信息，描述单个调试单元的目标范围、插件、版本与配置信息 |

#### debug_info

| 参数名称              | 参数类型     | 必选 | 描述                                                                                                |
| --------------------- | ------------ | ---- | --------------------------------------------------------------------------------------------------- |
| scope                 | object       | 是   | 调试目标范围（主机或服务实例）                                                                      |
| plugin_name           | string       | 是   | 插件名称                                                                                            |
| version               | string       | 是   | 插件版本号                                                                                          |
| config_template_name  | string array | 否   | 配置模板名称列表。配置模板记录在对应插件版本的包（release）中，通过 插件名+版本号+模板名 可唯一定位 |
| custom_config_context | object       | 否   | 自定义配置上下文，对本次调试的所有配置模板通用，用于覆盖模板渲染变量                                |

#### debug_info.scope

| 参数名称     | 参数类型    | 必选 | 描述                                                                               |
| ------------ | ----------- | ---- | ---------------------------------------------------------------------------------- |
| granularity  | string      | 是   | 目标粒度（枚举值：host、service_instance）                                         |
| bk_biz_id    | int64       | 是   | 业务ID                                                                             |
| filter       | object      | 否   | 目标过滤条件（当前为占位结构，预留）                                               |
| instance_ids | int64 array | 是   | 目标实例ID列表；granularity 为 host 时为主机ID，为 service_instance 时为服务实例ID |

### 调用示例

启动一个插件的调试：插件针对一台主机。

```json
{
  "debug_info": {
    "scope": {
      "granularity": "host",
      "bk_biz_id": 2,
      "instance_ids": [1001]
    },
    "plugin_name": "basereport",
    "version": "1.0.0",
    "config_template_name": ["basereport.conf"],
    "custom_config_context": {
      "report_path": "/usr/local/gse/plugins/etc"
    }
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "workflow_id": "workflow-abc123def456"
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                   |
| ---------- | -------- | ---------------------- |
| code       | int32    | 状态码，0表示成功      |
| message    | string   | 请求信息               |
| request_id | string   | 请求ID                 |
| error      | object   | 错误信息，成功时为null |
| data       | object   | 响应数据               |

#### data

| 参数名称    | 参数类型 | 描述                                   |
| ----------- | -------- | -------------------------------------- |
| workflow_id | string   | 调试工作流ID，可用于停止或查询调试状态 |

#### error

| 参数名称 | 参数类型 | 描述         |
| -------- | -------- | ------------ |
| system   | string   | 错误系统标识 |
| message  | string   | 错误消息     |
| details  | array    | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述     |
| -------- | -------- | -------- |
| code     | string   | 错误代码 |
| message  | string   | 错误消息 |
