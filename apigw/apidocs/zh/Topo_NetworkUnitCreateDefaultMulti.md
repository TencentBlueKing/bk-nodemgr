### 描述

- 该接口提供版本：v3.0.1-alpha.67+。
- 该接口所需权限：`networkunit_create（创建管控单元）`、`networkunit_view（查看管控单元）`。
- 该接口功能描述：批量在空管控区域下创建同名、同上游接入点的空非直连管控单元，用于接管环境初始化。

### URL

POST /api/v3/topo/networkunit/create_default_multi

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_id | int64 array | 是 | 目标管控区域 ID 列表；必须唯一，数量为 1～100，且不允许包含 0 |
| bk_networkunit_name | string | 是 | 所有目标区域创建的管控单元名称 |
| upstream | object | 是 | 统一上游接入点 |

#### upstream

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkarea_id | int64 | 是 | 上游管控区域 ID |
| bk_networkunit_id | int64 | 是 | 上游管控单元 ID |
| accesspoint_id | int64 | 是 | 上游接入点 ID |

**创建规则**：

- 目标管控区域必须是普通管控区域，默认管控区域 `bk_networkarea_id=0` 不支持该操作。
- 目标管控区域在执行时必须为空，即不包含任何活跃管控单元。
- 创建结果为普通非直连管控单元，不配置下游接入点和自定义部署配置。
- `upstream` 同时用于 cluster、file、data 三个通道。
- 目标管控区域按项目 IAM 权限逐项校验 `networkunit_create`。
- 上游管控单元需要具备 `networkunit_view` 权限。
- 单次最多处理 100 个管控区域。
- 多个目标区域可以部分成功，失败项不会阻断其他目标区域。

### 调用示例

```json
{
  "bk_networkarea_id": [101, 102, 103],
  "bk_networkunit_name": "default",
  "upstream": {
    "bk_networkarea_id": 10,
    "bk_networkunit_id": 20,
    "accesspoint_id": 30
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
    "success_count": 2,
    "failed_count": 1,
    "items": [
      {
        "bk_networkarea_id": 101,
        "success": true,
        "bk_networkunit_id": 1001,
        "error_code": "",
        "message": "network unit created"
      },
      {
        "bk_networkarea_id": 102,
        "success": true,
        "bk_networkunit_id": 1002,
        "error_code": "",
        "message": "network unit created"
      },
      {
        "bk_networkarea_id": 103,
        "success": false,
        "error_code": "networkarea_not_empty",
        "message": "networkarea-id(103) is not empty"
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示请求处理完成；单项失败通过 data.items 返回 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 请求级错误信息 |
| permission | object | 请求级权限信息 |
| data | object | 批量创建结果 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| success_count | int64 | 创建成功数量 |
| failed_count | int64 | 创建失败数量 |
| items | object array | 每个目标管控区域的处理结果，顺序与请求一致 |

#### items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| bk_networkarea_id | int64 | 目标管控区域 ID |
| success | bool | 是否创建成功 |
| bk_networkunit_id | int64 | 创建成功时返回的管控单元 ID |
| error_code | string | 失败时的稳定错误码 |
| message | string | 处理结果说明 |

### 常见错误码

| 错误码 | 描述 |
|--------|------|
| default_networkarea_not_supported | 默认管控区域不支持批量初始化 |
| networkarea_permission_denied | 没有目标管控区域的管控单元创建权限 |
| networkarea_not_found | 目标管控区域不存在 |
| networkarea_not_empty | 目标管控区域已有管控单元 |
| networkarea_busy | 目标管控区域正在初始化，请稍后重试 |
| networkunit_create_failed | 管控单元创建失败 |
| upstream_permission_denied | 没有上游管控单元查看权限，请求未执行写入 |
| upstream_networkunit_not_found | 上游管控单元不存在，请求未执行写入 |
| upstream_networkarea_mismatch | 上游管控单元与上游管控区域不匹配，请求未执行写入 |
| upstream_accesspoint_not_found | 上游接入点不属于上游管控单元，请求未执行写入 |
