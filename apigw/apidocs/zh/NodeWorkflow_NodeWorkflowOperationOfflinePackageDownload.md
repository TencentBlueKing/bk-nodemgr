### 描述

- 该接口提供版本：v3.0.1-alpha.18+。
- 该接口所需权限：无。
- 该接口功能描述：按操作 ID 下载离线安装包，返回 tar.gz 二进制流。

### URL

POST /api/v3/node/workflow/operation/offline/download

### 输入参数

| 参数名称     | 参数类型 | 必选 | 描述                            |
| ------------ | -------- | ---- | ------------------------------- |
| operation_id | string   | 是   | 操作 ID，用于定位对应离线安装包 |

### 调用示例

```json
{
  "operation_id": "op-20260421-0001"
}
```

### 响应示例

```http
HTTP/1.1 200 OK
Content-Type: application/gzip
Content-Disposition: attachment; filename=<package_name>.tar.gz

<binary gzip stream>
```

### 响应参数说明

| 参数名称            | 参数类型 | 描述                                             |
| ------------------- | -------- | ------------------------------------------------ |
| Content-Type        | string   | 固定为 `application/gzip`                        |
| Content-Disposition | string   | 附件下载头，文件名格式为 `<package_name>.tar.gz` |
| body                | binary   | tar.gz 二进制内容流                              |

### 说明

- 请求体为空或 `operation_id` 为空时，返回参数错误。
- 接口为流式下载，建议客户端按流读取并保存文件。
