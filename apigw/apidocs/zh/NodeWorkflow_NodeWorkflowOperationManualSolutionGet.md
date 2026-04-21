### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：agent_operate（操作Agent）、proxy_operate（操作Proxy）。
  - 权限说明：根据任务流涉及的节点角色动态收敛。若任务流仅涉及 Agent 节点，则只需 agent_operate；若仅涉及 Proxy 节点，则只需 proxy_operate；若涉及两者或角色未知，则需要两者。
- 该接口功能描述：根据任务流 ID 和操作 ID 获取手动处理方案，返回基于可执行命令的手动安装步骤。

### URL

POST /api/v3/node/workflow/operation/manual/solution/get

### 输入参数

| 参数名称     | 参数类型 | 必选 | 描述      |
| ------------ | -------- | ---- | --------- |
| workflow_id  | string   | 是   | 任务流 ID |
| operation_id | string   | 是   | 操作 ID   |

### 调用示例

查询任务流下某个操作的手动处理方案。

```json
{
  "workflow_id": "wf-20240601-0001",
  "operation_id": "op-001"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": [
    {
      "type": "bash",
      "description_en": "Use bash to manually install node.",
      "description_zh": "使用 bash 手动安装节点.",
      "steps": [
        {
          "type": "command",
          "name_en": "Execute bash command",
          "name_zh": "执行 bash 命令",
          "content_en": "bash /tmp/install.sh",
          "content_zh": "bash /tmp/install.sh"
        }
      ]
    }
  ]
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                     |
| ---------- | -------- | ------------------------ |
| code       | int32    | 状态码，`0` 表示成功     |
| message    | string   | 请求信息                 |
| request_id | string   | 请求 ID                  |
| error      | object   | 错误信息，成功时通常为空 |
| permission | object   | 权限信息                 |
| data       | array    | 手动处理方案列表         |

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

#### data[n]

| 参数名称       | 参数类型 | 描述                                    |
| -------------- | -------- | --------------------------------------- |
| type           | string   | 手动方案类型，当前可见值：`bash`、`bat` |
| description_en | string   | 英文方案描述                            |
| description_zh | string   | 中文方案描述                            |
| steps          | array    | 手动执行步骤列表                        |

#### data[n].steps[m]

| 参数名称   | 参数类型 | 描述                         |
| ---------- | -------- | ---------------------------- |
| type       | string   | 步骤类型，当前为 `command`   |
| name_en    | string   | 英文步骤名称                 |
| name_zh    | string   | 中文步骤名称                 |
| content_en | string   | 英文步骤内容，通常为命令文本 |
| content_zh | string   | 中文步骤内容，通常为命令文本 |
