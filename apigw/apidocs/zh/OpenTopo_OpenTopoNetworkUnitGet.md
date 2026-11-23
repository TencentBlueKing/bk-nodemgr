### 描述

- 该接口提供版本：v3.0.1-alpha.107+。
- 该接口所需权限：system_access（系统访问）、networkunit_view（查看管控单元）。
- 该接口功能描述：通过用户态入口按 ID 查询管控单元及其接入点、链路和部署配置。查看权限按目标管控单元校验。
- 用户身份及租户信息由 API 网关传递，无需在请求体中提供用户名或租户 ID。

### URL

POST /api/v3/open/topo/networkunit/get

### 输入参数

| 参数名称          | 参数类型 | 必选 | 描述                                             |
| ----------------- | -------- | ---- | ------------------------------------------------ |
| bk_networkunit_id | int64    | 否   | 管控单元 ID，必须大于等于 `0`；省略时查询 ID `0` |

### 调用示例

```json
{ "bk_networkunit_id": 10 }
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "tenant_id": "default",
    "bk_networkunit_id": 10,
    "bk_networkunit_name": "unit-prod",
    "bk_networkarea_id": 1,
    "accesspoints": [
      {
        "tenant_id": "default",
        "accesspoint_id": 1001,
        "accesspoint_name": "ap-prod",
        "bk_networkarea_id": 1,
        "endpoints": {
          "cluster": ["10.0.0.10:48668"],
          "file": ["10.0.0.10:48669"],
          "data": ["10.0.0.10:48670"]
        },
        "cloud_security_group_pool_id": ""
      }
    ],
    "links": {},
    "is_direct": false,
    "direct_endpoints": null,
    "generation": 2,
    "custom_deploy_config": {}
  }
}
```

### 响应参数说明

| 参数名称   | 参数类型 | 描述                             |
| ---------- | -------- | -------------------------------- |
| code       | int32    | 状态码，`0` 表示成功             |
| message    | string   | 请求结果信息                     |
| request_id | string   | 请求 ID                          |
| error      | object   | 错误信息，成功时为空             |
| permission | object   | 权限申请信息，权限不足时可能返回 |
| data       | object   | 管控单元详情                     |

#### data

| 参数名称             | 参数类型     | 描述                                            |
| -------------------- | ------------ | ----------------------------------------------- |
| tenant_id            | string       | 租户 ID                                         |
| bk_networkunit_id    | int64        | 管控单元 ID                                     |
| bk_networkunit_name  | string       | 管控单元名称                                    |
| bk_networkarea_id    | int64        | 所属管控区域 ID                                 |
| accesspoints         | object array | 接入点完整信息，不是 ID 数组                    |
| links                | object       | 上级链路，按 `cluster`、`file`、`data` 通道配置 |
| is_direct            | bool         | 是否使用直连端点                                |
| direct_endpoints     | object       | 直连端点，未配置时可为空                        |
| generation           | int64        | 节点代际                                        |
| custom_deploy_config | object       | 按操作系统类型为键的自定义部署配置，如 `linux`  |

#### data.accesspoints[n]

| 参数名称                     | 参数类型 | 描述                                            |
| ---------------------------- | -------- | ----------------------------------------------- |
| tenant_id                    | string   | 租户 ID                                         |
| accesspoint_id               | int64    | 接入点 ID                                       |
| accesspoint_name             | string   | 接入点名称                                      |
| bk_networkarea_id            | int64    | 所属管控区域 ID                                 |
| endpoints                    | object   | 端点配置，结构与下表相同                        |
| cloud_security_group_pool_id | string   | 绑定的全局安全组池 ID，空表示未启用云安全组管理 |

#### data.direct_endpoints / data.accesspoints[n].endpoints

| 参数名称 | 参数类型     | 描述         |
| -------- | ------------ | ------------ |
| cluster  | string array | 集群通道地址 |
| file     | string array | 文件通道地址 |
| data     | string array | 数据通道地址 |

#### data.links

| 参数名称 | 参数类型 | 描述             |
| -------- | -------- | ---------------- |
| cluster  | object   | 集群通道上级链路 |
| file     | object   | 文件通道上级链路 |
| data     | object   | 数据通道上级链路 |

各通道链路使用以下结构；未配置的通道可为空。

| 参数名称          | 参数类型 | 描述            |
| ----------------- | -------- | --------------- |
| bk_networkarea_id | int64    | 上级管控区域 ID |
| bk_networkunit_id | int64    | 上级管控单元 ID |
| accesspoint_id    | int64    | 上级接入点 ID   |

#### data.custom_deploy_config.{os_type}

| 参数名称          | 参数类型 | 描述           |
| ----------------- | -------- | -------------- |
| installer_runtime | object   | 安装器运行配置 |
| node_runtime      | object   | 节点运行配置   |
| plugin_runtime    | object   | 插件运行配置   |

#### installer_runtime

| 参数名称            | 参数类型     | 描述                                                           |
| ------------------- | ------------ | -------------------------------------------------------------- |
| base_work_dir       | string       | 安装器工作根目录                                               |
| firewall            | object       | 防火墙配置；未配置时继承部署默认值，空对象表示不执行防火墙配置 |
| firewall.remote_ips | string array | 远端 IP 配置                                                   |
| firewall.features   | string array | 防火墙功能配置                                                 |

#### node_runtime

| 参数名称        | 参数类型 | 描述           |
| --------------- | -------- | -------------- |
| base_deploy_dir | string   | 节点部署根目录 |
| data_ipc        | string   | 数据 IPC 配置  |
| plugin_ipc      | string   | 插件 IPC 配置  |
| log_dir         | string   | 节点日志目录   |
| zone_id         | string   | 区域标识       |
| city_id         | string   | 城市标识       |

#### plugin_runtime

| 参数名称        | 参数类型 | 描述           |
| --------------- | -------- | -------------- |
| base_deploy_dir | string   | 插件部署根目录 |
| log_dir         | string   | 插件日志目录   |
