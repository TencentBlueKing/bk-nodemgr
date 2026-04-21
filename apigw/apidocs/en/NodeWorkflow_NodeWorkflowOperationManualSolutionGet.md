### Description

- API Version: v3.0.1-alpha.18+.
- Required Permission: agent_operate (Operate Agent), proxy_operate (Operate Proxy).
  - Permission Note: Dynamically narrowed based on node roles in the workflow. If the workflow only involves Agent nodes, only agent_operate is required; if only Proxy nodes, only proxy_operate is required; if both or unknown roles, both are required.
- Function: Get manual solution by workflow ID and operation ID, returning executable command-based manual install steps.

### URL

POST /api/v3/node/workflow/operation/manual/solution/get

### Request Parameters

| Parameter    | Type   | Required | Description  |
| ------------ | ------ | -------- | ------------ |
| workflow_id  | string | Yes      | Workflow ID  |
| operation_id | string | Yes      | Operation ID |

### Request Example

Query manual solution for one operation in a workflow.

```json
{
  "workflow_id": "wf-20240601-0001",
  "operation_id": "op-001"
}
```

### Response Example

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

### Response Parameters

| Parameter  | Type   | Description                                 |
| ---------- | ------ | ------------------------------------------- |
| code       | int32  | Status code, `0` means success              |
| message    | string | Request message                             |
| request_id | string | Request ID                                  |
| error      | object | Error information, usually empty on success |
| permission | object | Permission information                      |
| data       | array  | Manual solution list                        |

#### error

| Parameter | Type   | Description             |
| --------- | ------ | ----------------------- |
| system    | string | Error system identifier |
| message   | string | Error message           |
| details   | array  | Error detail list       |

#### error.details[n]

| Parameter | Type   | Description   |
| --------- | ------ | ------------- |
| code      | string | Error code    |
| message   | string | Error message |

#### permission

| Parameter   | Type   | Description            |
| ----------- | ------ | ---------------------- |
| system      | string | Permission system ID   |
| system_name | string | Permission system name |
| apply_url   | string | Permission apply URL   |
| actions     | array  | Related action list    |

#### permission.actions[n]

| Parameter              | Type   | Description            |
| ---------------------- | ------ | ---------------------- |
| id                     | string | Action ID              |
| name                   | string | Action name            |
| related_resource_types | array  | Related resource types |

#### data[n]

| Parameter      | Type   | Description                                     |
| -------------- | ------ | ----------------------------------------------- |
| type           | string | Manual solution type, currently `bash` or `bat` |
| description_en | string | English solution description                    |
| description_zh | string | Chinese solution description                    |
| steps          | array  | Manual execution step list                      |

#### data[n].steps[m]

| Parameter  | Type   | Description                                |
| ---------- | ------ | ------------------------------------------ |
| type       | string | Step type, currently `command`             |
| name_en    | string | English step name                          |
| name_zh    | string | Chinese step name                          |
| content_en | string | English step content, usually command text |
| content_zh | string | Chinese step content, usually command text |
