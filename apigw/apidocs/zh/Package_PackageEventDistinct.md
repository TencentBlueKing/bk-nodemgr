### 描述

- 该接口提供版本：v3.0.1-alpha.13+。
- 该接口所需权限：无。
- 该接口功能描述：查询资源包操作事件各字段的去重值。

### URL

POST /api/v3/package/event/distinct

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| exact_include_conditions | object | 否 | 精确匹配条件 |
| fuzzy_include_conditions | object | 否 | 模糊匹配条件，当前没有可用字段 |
| operate_time_range | object | 否 | 操作时间范围；不传时默认为最近 365 天，最大跨度为 365 天 |

#### exact_include_conditions

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| generation | int64 array | 否 | 安装包代次列表（有效值：2） |
| os_type | string array | 否 | 操作系统类型列表（有效值：linux、windows、darwin） |
| cpu_arch | string array | 否 | CPU 架构列表（有效值：386、arm、arm64、amd64） |
| release_type | string array | 否 | 资源包类型列表（有效值：agent、proxy、cert、bintool、plugin_bintool、plugin） |
| operator | string array | 否 | 操作人列表 |
| event_type | string array | 否 | 事件类型列表（有效值：publish、delete、enable、disable、set_as_default、cancel_as_default、upload） |
| version | string array | 否 | 版本号列表 |

#### operate_time_range

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| start_timestamp_sec | int64 | 是 | 起始时间，Unix 秒级时间戳 |
| end_timestamp_sec | int64 | 是 | 结束时间，Unix 秒级时间戳 |

### 调用示例

```json
{
  "exact_include_conditions": {"generation": [2]},
  "fuzzy_include_conditions": {},
  "operate_time_range": {
    "start_timestamp_sec": 1748736000,
    "end_timestamp_sec": 1751327999
  }
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "release_type": ["agent", "proxy", "plugin"],
    "event_type": ["publish", "enable", "disable"],
    "os_type": ["linux", "windows"],
    "cpu_arch": ["amd64", "arm64"],
    "version": ["2.1.6-rc.5"],
    "operator": ["admin"]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，0 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| data | object | 响应数据 |
| data.release_type | string array | 资源包类型去重值 |
| data.event_type | string array | 事件类型去重值 |
| data.os_type | string array | 操作系统类型去重值 |
| data.cpu_arch | string array | CPU 架构去重值 |
| data.version | string array | 版本号去重值 |
| data.operator | string array | 操作人去重值 |
