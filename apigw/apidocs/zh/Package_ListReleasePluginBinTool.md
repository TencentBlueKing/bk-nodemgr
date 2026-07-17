### 描述

- 该接口提供版本：v3.0.1-alpha.15+。
- 该接口所需权限：package_view（查看资源包）。
- 该接口功能描述：查询插件工具包（PluginBinTool）安装包列表，接口仅返回当前用户有权限查看的插件工具包。

### URL

POST /api/v3/package/release/plugin_bintool/list

### 输入参数

| 参数名称 | 参数类型 | 必选 | 描述 |
|---------|----------|------|------|
| generation | int64 | 是 | 安装包代次（枚举值：2） |

**参数说明**：
- 接口根据用户的 `package_view` 权限范围自动过滤结果，仅返回有权查看的插件工具包数据。

### 调用示例

```json
{
  "generation": 2
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-1234567890",
  "data": {
    "total": 1,
    "items": [
      {
        "release": {
          "name": "bkmonitorbeat",
          "generation": 2,
          "release_type": "plugin_bintool",
          "os_type": "linux",
          "cpu_arch": "amd64",
          "version": "3.20.2305",
          "file_name": "bkmonitorbeat-bintool-3.20.2305.tgz",
          "labels": [],
          "enabled": true,
          "as_default": true,
          "md5": "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4",
          "updated_at": 1712000000000,
          "operator": "admin"
        }
      }
    ]
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

#### data

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| total | int64 | 符合条件的记录总数 |
| items | object array | 插件工具包列表 |

#### data.items[n]

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| release | object | 安装包基础信息 |

#### data.items[n].release

| 参数名称 | 参数类型 | 描述 |
|---------|----------|------|
| name | string | 插件名称 |
| generation | int64 | 安装包代次 |
| release_type | string | 安装包类型，固定为 plugin_bintool |
| os_type | string | 操作系统类型（与 cpu_arch 组成受支持的平台组合） |
| cpu_arch | string | CPU 架构（与 os_type 组成受支持的平台组合） |
| version | string | 版本号 |
| file_name | string | 安装包文件名 |
| labels | string array | 标签列表 |
| enabled | bool | 是否启用 |
| as_default | bool | 是否为默认版本 |
| md5 | string | 安装包 MD5 校验值 |
| updated_at | uint64 | 最后更新时间（Unix 毫秒级时间戳） |
| operator | string | 最后操作人 |
