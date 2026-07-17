### 描述

- 该接口提供版本：v3.0.0+。
- 该接口所需权限：无。
- 该接口功能描述：批量设置符合条件的 Agent 安装包标签。

### URL

POST /api/v3/package/release/agent/set_labels_many

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| generation | int64 | 是 | 安装包代次（枚举值：2） |
| exact_include_conditions | object | 是 | 精确匹配条件 |
| labels | string array | 是 | 要设置的标签列表 |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| platform | object array | 否 | 平台列表 |
| version | string array | 否 | 版本号列表 |
| as_default | bool array | 否 | 是否为默认版本 |
| enabled | bool array | 否 | 是否启用 |
| name | string array | 否 | 安装包名称列表 |
| file_name | string array | 否 | 文件名列表 |

#### platform

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| os_type | string | 是 | 操作系统类型（枚举值：linux、windows、darwin） |
| cpu_arch | string | 是 | CPU 架构（枚举值：386、arm、arm64、amd64） |

### 调用示例

```json
{
  "generation": 2,
  "exact_include_conditions": {
    "platform": [{"os_type": "linux", "cpu_arch": "amd64"}],
    "version": ["2.1.6-rc.5"]
  },
  "labels": ["stable"]
}
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
