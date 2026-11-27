### 描述

- 该接口提供版本：v3.0.1-alpha.122+。
- 该接口所需权限：对请求中的每个插件均需 `plugin_operate`（操作插件）权限。
- 该接口功能描述：批量向指定主机上的插件配置目录下发适用的依赖文件，并重启插件进程。

工作流按目标插件进程的操作系统、CPU 架构和当前插件包版本匹配依赖文件。每个非空的适用范围都必须匹配。若适用文件的目标文件名重复（忽略大小写），该目标的下发步骤失败。

每个目标在适用文件全部下发成功后记录本次下发的依赖，再重启进程。没有匹配文件时跳过文件下发，仍会重启进程。

### URL

POST /api/v3/plugin/apply_requirements

### 输入参数

| 参数名称 | 参数类型     | 必选 | 描述                           |
| -------- | ------------ | ---- | ------------------------------ |
| plugin   | object array | 是   | 目标插件进程列表，至少包含一项 |

#### plugin[n]

| 参数名称    | 参数类型 | 必选 | 描述                                             |
| ----------- | -------- | ---- | ------------------------------------------------ |
| bk_host_id  | int64    | 是   | 目标主机 ID，必须为正整数                        |
| plugin_name | string   | 是   | 目标主机上已有的插件进程名称，不可为空或全为空白 |

同一请求中重复的主机 ID 与插件名称组合只处理一次。

### 调用示例

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bk-monitor-agent"
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20261127-000001",
  "error": null,
  "data": {
    "workflow_id": "wf-plugin-0001"
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                         |
| ---------- | -------- | ---------------------------- |
| code       | int32    | 状态码，`0` 表示工作流已启动 |
| message    | string   | 请求信息                     |
| request_id | string   | 请求 ID                      |
| error      | object   | 错误信息，成功时为 `null`    |
| permission | object   | 权限信息                     |
| data       | object   | 响应数据                     |

#### data

| 参数名称    | 参数类型 | 描述                                              |
| ----------- | -------- | ------------------------------------------------- |
| workflow_id | string   | 工作流 ID，用于查询各目标的依赖下发和进程重启结果 |

使用 `workflow_id` 调用 `POST /api/v3/plugin/workflow/operation/list` 查询各目标的执行状态。

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

#### permission

| 参数名称    | 参数类型 | 描述         |
| ----------- | -------- | ------------ |
| system      | string   | 权限系统 ID  |
| system_name | string   | 权限系统名称 |
| apply_url   | string   | 权限申请链接 |
| actions     | array    | 关联动作列表 |

#### permission.actions[n]

| 参数名称               | 参数类型 | 描述         |
| ---------------------- | -------- | ------------ |
| id                     | string   | 动作 ID      |
| name                   | string   | 动作名称     |
| related_resource_types | array    | 关联资源类型 |
