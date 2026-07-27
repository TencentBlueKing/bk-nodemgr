### 描述

- 该接口提供版本：v3.0.1-alpha.60+。
- 该接口所需权限：networkunit_create（创建管控单元）。
- 该接口功能描述：创建管控单元，并配置关联接入点、链路与部署配置。

### URL

POST /api/v3/topo/networkunit/create

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkunit_name | string | 是 | 管控单元名称 |
| bk_networkarea_id | int64 | 是 | 管控区域 ID |
| accesspoints | object array | 否 | 接入点列表 |
| links | object | 否 | 链路配置 |
| is_direct | bool | 否 | 是否直连 |
| direct_endpoints | object | 否 | 直连端点配置 |
| generation | int64 | 是 | 代际 |
| custom_deploy_config | object | 否 | 自定义部署配置 |

#### accesspoints[n]

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| tenant_id | string | 否 | 租户 ID |
| accesspoint_id | int64 | 否 | 接入点 ID |
| accesspoint_name | string | 否 | 接入点名称 |
| bk_networkarea_id | int64 | 否 | 管控区域 ID |
| endpoints | object | 否 | 端点配置 |

#### accesspoints[n].endpoints

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| cluster | string array | 否 | 集群通道配置 |
| file | string array | 否 | 文件通道配置 |
| data | string array | 否 | 数据通道配置 |

#### links

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| cluster | object | 否 | 集群通道配置 |
| file | object | 否 | 文件通道配置 |
| data | object | 否 | 数据通道配置 |

#### links.cluster

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_id | int64 | 否 | 管控区域 ID |
| bk_networkunit_id | int64 | 否 | 管控单元 ID |
| accesspoint_id | int64 | 否 | 接入点 ID |

#### links.file

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_id | int64 | 否 | 管控区域 ID |
| bk_networkunit_id | int64 | 否 | 管控单元 ID |
| accesspoint_id | int64 | 否 | 接入点 ID |

#### links.data

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_id | int64 | 否 | 管控区域 ID |
| bk_networkunit_id | int64 | 否 | 管控单元 ID |
| accesspoint_id | int64 | 否 | 接入点 ID |

#### direct_endpoints

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| cluster | string array | 否 | 集群通道配置 |
| file | string array | 否 | 文件通道配置 |
| data | string array | 否 | 数据通道配置 |

#### custom_deploy_config.{key}

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| installer_runtime | object | 否 | 安装器运行时配置 |
| node_runtime | object | 否 | 节点运行时配置 |
| plugin_runtime | object | 否 | 插件运行时配置 |

#### custom_deploy_config.{key}.installer_runtime

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| base_work_dir | string | 否 | 基础工作目录 |

#### custom_deploy_config.{key}.node_runtime

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| base_deploy_dir | string | 否 | 基础部署目录 |
| data_ipc | string | 否 | Data 进程 IPC 路径 |
| plugin_ipc | string | 否 | Plugin 进程 IPC 路径 |
| log_dir | string | 否 | 日志目录 |
| zone_id | string | 否 | 区域 ID |
| city_id | string | 否 | 城市 ID |

#### custom_deploy_config.{key}.plugin_runtime

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| base_deploy_dir | string | 否 | 基础部署目录 |
| log_dir | string | 否 | 日志目录 |

### 调用示例

```json
{
  "bk_networkunit_name": "default",
  "bk_networkarea_id": 1,
  "accesspoints": [
    {
      "tenant_id": "id-001",
      "accesspoint_id": 1,
      "accesspoint_name": "default",
      "bk_networkarea_id": 1,
      "endpoints": {
        "cluster": [
          "string"
        ],
        "file": [
          "string"
        ],
        "data": [
          "string"
        ]
      }
    }
  ],
  "links": {
    "cluster": {
      "bk_networkarea_id": 1,
      "bk_networkunit_id": 1,
      "accesspoint_id": 1
    },
    "file": {
      "bk_networkarea_id": 1,
      "bk_networkunit_id": 1,
      "accesspoint_id": 1
    },
    "data": {
      "bk_networkarea_id": 1,
      "bk_networkunit_id": 1,
      "accesspoint_id": 1
    }
  },
  "is_direct": false,
  "direct_endpoints": {
    "cluster": [
      "string"
    ],
    "file": [
      "string"
    ],
    "data": [
      "string"
    ]
  },
  "generation": 1,
  "custom_deploy_config": {
    "default": {
      "installer_runtime": {
        "base_work_dir": "/var/lib/gse"
      },
      "node_runtime": {
        "base_deploy_dir": "/var/lib/gse",
        "data_ipc": "/var/run/ipc.sock",
        "plugin_ipc": "/var/run/ipc.sock",
        "log_dir": "/var/lib/gse",
        "zone_id": "id-001",
        "city_id": "id-001"
      },
      "plugin_runtime": {
        "base_deploy_dir": "/var/lib/gse",
        "log_dir": "/var/lib/gse"
      }
    }
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "error": null,
  "permission": null,
  "data": {
    "bk_networkunit_id": 1
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时为空 |
| permission | object | 权限信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| bk_networkunit_id | int64 | 管控单元 ID |
