### 描述

- 该接口提供版本：v3.0.1-alpha.60+。
- 该接口所需权限：networkunit_view（查看管控单元）。
- 该接口功能描述：根据管控单元 ID 查询管控单元详情。

### URL

POST /api/v3/topo/networkunit/get

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkunit_id | int64 | 否 | 管控单元 ID |

### 调用示例

```json
{
  "bk_networkunit_id": 1
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
    "tenant_id": "id-001",
    "bk_networkunit_id": 1,
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
    }
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
| tenant_id | string | 租户 ID |
| bk_networkunit_id | int64 | 管控单元 ID |
| bk_networkunit_name | string | 管控单元名称 |
| bk_networkarea_id | int64 | 管控区域 ID |
| accesspoints | object array | 接入点列表 |
| links | object | 链路配置 |
| is_direct | bool | 是否直连 |
| direct_endpoints | object | 直连端点配置 |
| generation | int64 | 代际 |
| custom_deploy_config | object | 自定义部署配置 |

#### data.accesspoints[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| tenant_id | string | 租户 ID |
| accesspoint_id | int64 | 接入点 ID |
| accesspoint_name | string | 接入点名称 |
| bk_networkarea_id | int64 | 管控区域 ID |
| endpoints | object | 端点配置 |

#### data.links

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| cluster | object | 集群通道配置 |
| file | object | 文件通道配置 |
| data | object | 数据通道配置 |

#### data.direct_endpoints

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| cluster | string array | 集群通道配置 |
| file | string array | 文件通道配置 |
| data | string array | 数据通道配置 |

#### data.custom_deploy_config.{key}

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| installer_runtime | object | 安装器运行时配置 |
| node_runtime | object | 节点运行时配置 |
| plugin_runtime | object | 插件运行时配置 |
