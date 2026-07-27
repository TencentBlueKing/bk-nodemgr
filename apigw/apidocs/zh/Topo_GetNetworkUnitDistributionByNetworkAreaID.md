### 描述

- 该接口提供版本：v3.0.1-alpha.60+。
- 该接口所需权限：无。
- 该接口功能描述：按管控区域 ID 统计管控单元分布。

### URL

POST /api/v3/topo/networkunit/get_networkunit_distribution_by_networkarea_id

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| exact_include_conditions | object | 否 | 精确匹配包含条件 |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_networkunit_id | int64 array | 否 | 管控单元 ID |
| bk_networkarea_id | int64 array | 否 | 管控区域 ID |
| is_direct | bool array | 否 | 是否直连 |
| generation | int64 array | 否 | 代际 |

### 调用示例

```json
{
  "exact_include_conditions": {
    "bk_networkunit_id": [
      1
    ],
    "bk_networkarea_id": [
      1
    ],
    "is_direct": [
      false
    ],
    "generation": [
      1
    ]
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
    "default": 1
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
