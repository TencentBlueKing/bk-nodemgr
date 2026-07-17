### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：package_manage（管理资源包）。
- 该接口功能描述：删除指定代次的二进制工具安装包。

### URL

POST /api/v3/package/release/bintool/delete

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| generation | int64 | 是 | 安装包代次（枚举值：2） |

### 调用示例

```json
{"generation": 2}
```

### 响应示例

```json
{"code": 0, "message": "ok", "request_id": "req-1234567890", "data": {}}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| data | object | 空对象，表示操作成功 |
