### 描述

- 该接口提供版本：v3.0.1-alpha.106+。
- 该接口所需权限：plugin_view（查看插件）。
- 该接口功能描述：按业务和主机查询主机上插件的进程状态，返回内容兼容节点管理 V2 的插件检索。
- 使用范围：仅监控平台使用，不对外开放申请。

### URL

POST /api/v3/process/list_v2_compatible

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| bk_biz_id | int64 array | 是 | 业务 ID 列表 |
| bk_host_id | int64 array | 否 | 主机 ID 列表；不传则返回业务内全部主机 |

**参数说明**：

- `bk_biz_id` 必传，缺少或为空数组时接口返回参数错误。
- 接口不分页，一次返回授权范围内匹配的全部主机，`data.total` 为匹配到的主机数量。
- `bk_biz_id` 按调用方 `plugin_view` 的业务权限收敛：调用方拥有全部业务权限时按请求的业务查询；仅有部分业务权限时只查询其中已授权的业务；请求的业务均不在授权范围内时返回 `total=0` 与空 `list`，不返回无权限错误。调用方没有任何业务实例授权时，接口按请求的业务做权限校验并返回无权限错误。
- `bk_host_id` 在收敛后的业务范围内过滤主机，未命中任何主机时返回空 `list`。
- 目前仅返回 `bkmonitorbeat` 一种插件的进程状态，与节点管理 V2 的插件检索范围一致；`plugin_status` 每台主机固定包含一项，不会返回其他插件的进程。

### 调用示例

查询业务 `2` 下主机 `101`、`102` 的插件进程状态。

```json
{
  "bk_biz_id": [2],
  "bk_host_id": [101, 102]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-20261106-000001",
  "error": null,
  "permission": null,
  "data": {
    "total": 2,
    "list": [
      {
        "bk_host_id": 101,
        "inner_ip": "10.0.0.1",
        "inner_ipv6": "",
        "bk_cloud_id": 0,
        "setup_path": "/usr/local/gse/gse_agent",
        "plugin_status": [
          {
            "name": "bkmonitorbeat",
            "version": "3.6.0",
            "status": "running"
          }
        ]
      },
      {
        "bk_host_id": 102,
        "inner_ip": "10.0.0.2,10.0.1.2",
        "inner_ipv6": "fd00::2",
        "bk_cloud_id": 0,
        "setup_path": "/usr/local/gse/gse_agent",
        "plugin_status": [
          {
            "name": "bkmonitorbeat",
            "version": "",
            "status": "unregister"
          }
        ]
      }
    ]
  }
}
```

### 响应参数说明

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | int32 | 状态码，`0` 表示成功 |
| message | string | 请求信息 |
| request_id | string | 请求 ID |
| error | object | 错误信息，成功时通常为空 |
| permission | object | 权限申请信息 |
| data | object | 响应数据 |

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| total | int64 | 匹配条件的主机总数 |
| list | object array | 主机及其插件进程状态列表；无匹配主机时为空数组 |

#### data.list[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| bk_host_id | int64 | 主机 ID |
| inner_ip | string | 内网 IPv4，多个地址以逗号分隔 |
| inner_ipv6 | string | 内网 IPv6，多个地址以逗号分隔；没有 IPv6 时为空 |
| bk_cloud_id | int64 | 云区域 ID |
| setup_path | string | 主机上 Agent 的安装路径，按部署配置与云区域自定义部署配置计算；主机没有节点安装信息时为空；上报的节点代际或操作系统类型缺失时，按第二代（generation=2）与 `linux` 的部署配置计算 |
| plugin_status | object array | 插件进程状态列表，当前固定包含一项 |

#### plugin_status[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| name | string | 插件名，当前固定为 `bkmonitorbeat` |
| version | string | 插件版本；无该插件进程记录时为空 |
| status | string | 插件进程状态，可选值：`init`、`running`、`stopped`、`unknown`、`unregister` |

**status 取值说明**：

- `running`：进程运行中。
- `stopped`：进程已停止。
- `unknown`：进程状态无法判定。
- `init`：进程记录已创建，但尚未同步到进程的实际状态。
- `unregister`：主机上没有该插件的进程记录。

#### error

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| system | string | 错误系统标识 |
| message | string | 错误消息 |
| details | object array | 错误详情列表 |

#### error.details[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| code | string | 错误代码 |
| message | string | 错误消息 |
