### 描述

- 该接口提供版本：v3.0.1+。
- 该接口所需权限：plugin_operate（操作插件）。
- 该接口功能描述：批量为指定主机安装插件，并可按需指定子配置文件和自定义渲染上下文。

### URL

POST /api/v3/plugin/install

### 输入参数

| 参数名称 | 参数类型     | 必选 | 描述                             |
| -------- | ------------ | ---- | -------------------------------- |
| plugin   | object array | 是   | 待安装插件列表，至少包含一个元素 |

#### plugin[n]

| 参数名称              | 参数类型     | 必选 | 描述                                                                     |
| --------------------- | ------------ | ---- | ------------------------------------------------------------------------ |
| bk_host_id            | int64        | 是   | 目标主机ID，必须为正整数                                                 |
| plugin_name           | string       | 是   | 插件名称                                                                 |
| version               | string       | 否   | 待安装插件版本；未传或传空时使用该插件当前默认发布版本                   |
| config_template_name  | string array | 否   | 需要下发的子配置文件名称列表；取值来自对应插件版本发布包中的配置模板名称 |
| custom_config_context | object       | 否   | 自定义配置渲染上下文；为自由结构对象，但禁止使用系统保留顶层键           |
| custom_spec           | object       | 否   | 自定义进程资源及监控策略；传入时整体覆盖默认配置和匹配的配置策略         |

#### plugin[n].custom_spec

| 参数名称       | 参数类型 | 必选 | 描述         |
| -------------- | -------- | ---- | ------------ |
| resource       | object   | 否   | 进程资源限制 |
| monitor_policy | object   | 否   | 进程监控策略 |

#### plugin[n].custom_spec.resource

| 参数名称          | 参数类型 | 必选 | 描述                       |
| ----------------- | -------- | ---- | -------------------------- |
| cpu_limit_percent | double   | 否   | CPU 使用率限制，单位百分比 |
| mem_limit_percent | double   | 否   | 内存使用率限制，单位百分比 |

#### plugin[n].custom_spec.monitor_policy

| 参数名称                | 参数类型 | 必选 | 描述                                               |
| ----------------------- | -------- | ---- | -------------------------------------------------- |
| restart_type            | string   | 否   | 重启方式：`auto`（自动重启）、`manual`（手动重启） |
| start_check_seconds     | int64    | 否   | 启动后检查进程存活的时间，单位秒                   |
| stop_check_seconds      | int64    | 否   | 停止后检查进程状态的时间，单位秒                   |
| operate_timeout_seconds | int64    | 否   | 进程操作超时时间，单位秒                           |

**参数说明**：

- `plugin`：至少需要传入一个待安装对象，否则请求会校验失败。
- `bk_host_id`：必须传入有效的正整数主机ID；省略或传入 `0`、负数会失败。
- `config_template_name`：不是固定枚举，实际可用值取决于目标插件版本发布包中的配置模板名称；未传时默认处理主配置文件。
- `custom_config_context`：可传入模板渲染所需的自定义变量。禁止使用保留顶层键，包括 `PluginInfo`、`NodeInfo`、
  `PreDefinitionConstants`、`CustomContext`、`plugin_path`、`nodeman`、`cmdb_instance`、`target`、`control_info`。
- `custom_spec`：未传时使用默认配置和匹配的配置策略。传入时整体覆盖资源及监控策略，不进行逐字段合并；省略的对象或数值字段按零值处理，省略的 `restart_type` 按空字符串处理。使用时应完整提供两组配置。
- 接口成功表示安装任务已提交，不表示插件已安装完成；使用 `workflow_id` 查询插件任务流及操作状态。模板名称与保留键校验在工作流执行阶段进行，相关错误需通过任务状态和日志查看。

### 调用示例

为两台主机安装 `bk-monitor-agent` 插件，其中一台指定子配置、自定义渲染变量和进程资源及监控策略。

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent",
      "version": "2.4.0",
      "config_template_name": ["bkmonitoragent.conf"],
      "custom_config_context": {
        "cluster_id": "prod-ap-guangzhou",
        "labels": {
          "env": "prod",
          "region": "ap-guangzhou"
        }
      },
      "custom_spec": {
        "resource": {
          "cpu_limit_percent": 10,
          "mem_limit_percent": 10
        },
        "monitor_policy": {
          "restart_type": "auto",
          "start_check_seconds": 5,
          "stop_check_seconds": 5,
          "operate_timeout_seconds": 5
        }
      }
    },
    {
      "bk_host_id": 1002,
      "plugin_name": "bk-monitor-agent",
      "version": "2.4.0"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260324-000001",
  "error": null,
  "data": {
    "workflow_id": "plugin-install-workflow-123456"
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

| 参数名称    | 参数类型 | 描述                                 |
| ----------- | -------- | ------------------------------------ |
| workflow_id | string   | 工作流ID，可用于查询插件安装任务状态 |

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
