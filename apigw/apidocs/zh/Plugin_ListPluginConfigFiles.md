### 描述

- 该接口提供版本：v3.0.1-alpha.77+。
- 该接口所需权限：plugin_view（查看插件）。
- 该接口功能描述：根据主机 ID 和插件名称查询该插件在目标主机上的配置文件列表。

### URL

POST /api/v3/plugin/list_config_files

### 输入参数

| 参数名称    | 参数类型 | 必选 | 描述                                   |
| ----------- | -------- | ---- | -------------------------------------- |
| bk_host_id  | int64    | 是   | 主机 ID                                |
| plugin_name | string   | 是   | 插件名称；用于匹配配置记录中的进程名称 |

**参数说明**：

- `bk_host_id`：必须传入非 0 值，否则请求校验失败。
- `plugin_name`：必须传入非空字符串，否则请求校验失败。
- 权限按目标主机所属业务校验，需要具备该业务下的插件查看权限。

### 调用示例

查询主机 `1001` 上 `bk-monitor-agent` 插件的配置文件列表。

```json
{
  "bk_host_id": 1001,
  "plugin_name": "bk-monitor-agent"
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20260828-000001",
  "error": null,
  "data": {
    "items": [
      {
        "name": "bkmonitoragent.conf",
        "template_name": "bkmonitoragent.conf",
        "process_name": "bk-monitor-agent",
        "bk_host_id": 1001,
        "set": "default",
        "is_main_config": true,
        "content": "dataid: 100001\nperiod: 60\n",
        "md5": "d9d21a34e52e49dff8e6f5aaeb5c3d51",
        "file_path": "etc/bkmonitoragent.conf",
        "custom_config_context": {
          "cluster_id": "prod-ap-guangzhou",
          "labels": {
            "env": "prod"
          }
        }
      }
    ]
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

| 参数名称 | 参数类型     | 描述             |
| -------- | ------------ | ---------------- |
| items    | object array | 插件配置文件列表 |

#### data.items[n]

| 参数名称              | 参数类型 | 描述                                           |
| --------------------- | -------- | ---------------------------------------------- |
| name                  | string   | 配置文件名称                                   |
| template_name         | string   | 原始插件配置模板名称                           |
| process_name          | string   | 配置所属插件进程名称                           |
| bk_host_id            | int64    | 主机 ID                                        |
| set                   | string   | 配置记录中的 set 字段                          |
| is_main_config        | bool     | 是否为主配置文件                               |
| content               | string   | 配置文件内容                                   |
| md5                   | string   | 配置文件内容 MD5                               |
| file_path             | string   | 配置文件相对插件部署目录的路径                 |
| custom_config_context | object   | 插件进程的自定义配置渲染上下文；为自由结构对象 |

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
