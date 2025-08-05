### 描述

安装 agent

### 输入参数

| 参数名称                           | 参数类型         | 必选 | 描述       |
|--------------------------------|--------------|----|----------|
| host                           | Object Array | 是  | 主机信息     |
| disable_default_target_version | Boolean      | 否  | 是否禁用默认版本 |
| target_version                 | Object Array | 否  | 目标版本     |

> 只有 disable_default_target_version 为 true 时，target_version 才有效

#### host

| 参数名称              | 参数类型   | 必选 | 描述               |
|-------------------|--------|----|------------------|
| bk_host_innerip   | string | 是  | 主机内网 IP          |
| os_type           | string | 是  | 操作系统类型(用于选择连接方式) |
| login_ip          | string | 是  |                  |            
| login_user        | string | 是  |                  |
| login_port        | int64  | 是  |                  |
| login_mode        | string | 是  |                  |
| login_password    | string | 是  |                  |
| bk_biz_id         | int64  | 是  |                  |
| bk_networkunit_id | int64  | 是  |                  |
| node_role         | string | 是  |                  |
| tenant_id         | string | 是  |                  |
| bk_addressing     | string | 是  |                  |

#### target_version

| 参数名称     | 参数类型   | 必选 | 描述     |
|----------|--------|----|--------|
| os_type  | string | 是  | 操作系统类型 |
| cpu_arch | string | 是  | CPU 架构 |
| version  | string | 是  | 版本     |

### 响应示例

| 参数名称     | 参数类型   | 必选 | 描述     |
|----------|--------|----|--------|
| os_type  | string | 是  | 操作系统类型 |
| cpu_arch | string | 是  | CPU 架构 |
| version  | string | 是  | 版本     |
