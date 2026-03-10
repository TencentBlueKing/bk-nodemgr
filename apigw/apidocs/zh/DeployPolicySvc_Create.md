### 描述

- 该接口提供版本：v3.0.1+
- 该接口所需权限：
- 该接口功能描述：创建部署策略。

### URL

POST /api/v3/deploy_policy/create

### 输入参数

| 参数名称        | 参数类型   | 必选 | 描述     |
|-------------|--------|----|--------|
| name        | string | 是  | 部署策略名称 |
| description | string | 否  | 部署策略描述 |
| enabled     | bool   | 是  | 是否启用   |
| specs       | array  | 是  | 部署规范列表 |
| scopes      | array  | 是  | 部署范围列表 |

#### specs[n]

部署规范定义了期望的最终状态。

| 参数名称  | 参数类型   | 必选 | 描述                                                                    |
|-------|--------|----|-----------------------------------------------------------------------|
| type  | string | 是  | 规范类型（枚举值：specify_plugin、specify_plugin_pkg、specify_plugin_sub_config） |
| param | object | 是  | 规范参数，结构根据type字段而定                                                     |

**type 为 specify_plugin 时，param 结构：**

指定插件版本，确保目标节点上安装指定名称和版本的插件。如果插件不存在则安装，版本不匹配则升级。

| 参数名称                  | 参数类型   | 必选 | 描述       |
|-----------------------|--------|----|----------|
| plugin_name           | string | 是  | 插件名称     |
| version               | string | 是  | 插件版本     |
| custom_config_context | object | 否  | 自定义配置上下文 |

**type 为 specify_plugin_pkg 时，param 结构：**

指定插件包版本，确保目标节点上安装指定插件包名称和版本的插件。插件名称会根据部署策略ID和模块ID自动生成。如果插件不存在则安装，版本不匹配则升级。

| 参数名称                  | 参数类型   | 必选 | 描述       |
|-----------------------|--------|----|----------|
| plugin_pkg_name       | string | 是  | 插件包名称    |
| version               | string | 是  | 插件包版本    |
| custom_config_context | object | 否  | 自定义配置上下文 |

**type 为 specify_plugin_sub_config 时，param 结构：**

指定插件子配置，用于更新已安装插件的配置文件内容。仅更新配置，不涉及插件版本的安装或升级。

| 参数名称                  | 参数类型   | 必选 | 描述       |
|-----------------------|--------|----|----------|
| plugin_name           | string | 是  | 插件名称     |
| config_files_detail   | array  | 是  | 配置文件详情列表 |
| custom_config_context | object | 否  | 自定义配置上下文 |

**config_files_detail[n] 结构：**

| 参数名称           | 参数类型   | 必选 | 描述       |
|----------------|--------|----|----------|
| name           | string | 是  | 配置文件名    |
| content        | string | 是  | 配置文件内容   |
| is_main_config | bool   | 是  | 是否为主配置文件 |

#### scopes[n]

部署范围定义了策略应用的目标范围。

| 参数名称  | 参数类型   | 必选 | 描述                                                                  |
|-------|--------|----|---------------------------------------------------------------------|
| type  | string | 是  | 范围类型（枚举值：topo、service_template、set_template、instance、dynamic_group） |
| scope | object | 是  | 范围详情，结构根据type字段而定                                                   |

**type 为 topo 时，scope 结构：**

根据拓扑路径指定目标范围。

| 参数名称        | 参数类型   | 必选 | 描述                              |
|-------------|--------|----|---------------------------------|
| granularity | string | 是  | 目标粒度（枚举值：host、service_instance） |
| bk_biz_id   | int64  | 是  | 业务ID                            |
| filter      | object | 否  | 目标过滤器                           |
| paths       | array  | 是  | 拓扑路径列表                          |

**paths[n] 结构：**

| 参数名称         | 参数类型   | 必选 | 描述     |
|--------------|--------|----|--------|
| topo_obj_id  | string | 是  | 拓扑对象ID |
| topo_inst_id | int64  | 是  | 拓扑实例ID |

**type 为 service_template 时，scope 结构：**

根据服务模板指定目标范围。

| 参数名称                 | 参数类型        | 必选 | 描述                              |
|----------------------|-------------|----|---------------------------------|
| granularity          | string      | 是  | 目标粒度（枚举值：host、service_instance） |
| bk_biz_id            | int64       | 是  | 业务ID                            |
| filter               | object      | 否  | 目标过滤器                           |
| service_template_ids | int64 array | 否  | 服务模板ID列表                        |
| module_ids           | int64 array | 否  | 模块ID列表                          |

**type 为 set_template 时，scope 结构：**

根据集群模板指定目标范围。

| 参数名称             | 参数类型        | 必选 | 描述                              |
|------------------|-------------|----|---------------------------------|
| granularity      | string      | 是  | 目标粒度（枚举值：host、service_instance） |
| bk_biz_id        | int64       | 是  | 业务ID                            |
| filter           | object      | 否  | 目标过滤器                           |
| set_template_ids | int64 array | 否  | 集群模板ID列表                        |
| set_ids          | int64 array | 否  | 集群ID列表                          |

**type 为 instance 时，scope 结构：**

根据实例ID直接指定目标范围。

| 参数名称         | 参数类型        | 必选 | 描述                                  |
|--------------|-------------|----|-------------------------------------|
| granularity  | string      | 是  | 目标粒度（枚举值：host、service_instance）     |
| bk_biz_id    | int64       | 是  | 业务ID                                |
| filter       | object      | 否  | 目标过滤器                               |
| instance_ids | int64 array | 是  | 实例ID列表（host_id或service_instance_id） |

**type 为 dynamic_group 时，scope 结构：**

根据动态分组指定目标范围。注意：dynamic_group 类型仅支持 host 粒度。

| 参数名称              | 参数类型         | 必选 | 描述             |
|-------------------|--------------|----|----------------|
| granularity       | string       | 是  | 目标粒度（仅支持：host） |
| bk_biz_id         | int64        | 是  | 业务ID           |
| filter            | object       | 否  | 目标过滤器          |
| dynamic_group_ids | string array | 是  | 动态分组ID列表       |

### 调用示例

#### 示例1：使用拓扑路径（topo）指定范围

创建一个部署策略，用于在整个业务的所有主机上部署监控采集插件。

```json
{
  "name": "生产环境监控插件部署策略",
  "description": "用于生产环境统一部署监控采集插件",
  "enabled": true,
  "specs": [
    {
      "type": "specify_plugin",
      "param": {
        "plugin_name": "bkmonitorbeat",
        "version": "3.60.3066"
      }
    }
  ],
  "scopes": [
    {
      "type": "topo",
      "scope": {
        "granularity": "host",
        "bk_biz_id": 100,
        "filter": {},
        "paths": [
          {
            "topo_obj_id": "biz",
            "topo_inst_id": 100
          }
        ]
      }
    }
  ]
}
```

#### 示例2：使用服务模板（service_template）指定范围

创建一个部署策略，用于在指定服务模板下的所有服务实例上部署插件。

```json
{
  "name": "Web服务监控插件部署策略",
  "description": "针对Web服务模板下的所有服务实例部署监控插件",
  "enabled": true,
  "specs": [
    {
      "type": "specify_plugin",
      "param": {
        "plugin_name": "bkmonitorbeat",
        "version": "3.60.3066"
      }
    }
  ],
  "scopes": [
    {
      "type": "service_template",
      "scope": {
        "granularity": "service_instance",
        "bk_biz_id": 100,
        "filter": {},
        "service_template_ids": [1001, 1002]
      }
    }
  ]
}
```

#### 示例3：使用集群模板（set_template）指定范围

创建一个部署策略，用于在指定集群模板下的所有主机上部署插件。

```json
{
  "name": "测试集群监控插件部署策略",
  "description": "针对测试集群模板下的所有主机部署监控插件",
  "enabled": true,
  "specs": [
    {
      "type": "specify_plugin",
      "param": {
        "plugin_name": "bkmonitorbeat",
        "version": "3.60.3066"
      }
    }
  ],
  "scopes": [
    {
      "type": "set_template",
      "scope": {
        "granularity": "host",
        "bk_biz_id": 100,
        "filter": {},
        "set_template_ids": [2001, 2002]
      }
    }
  ]
}
```

#### 示例4：使用实例ID（instance）直接指定范围

创建一个部署策略，用于在指定的主机列表上部署插件。

```json
{
  "name": "特定主机监控插件部署策略",
  "description": "针对指定主机列表部署监控插件",
  "enabled": true,
  "specs": [
    {
      "type": "specify_plugin",
      "param": {
        "plugin_name": "bkmonitorbeat",
        "version": "3.60.3066"
      }
    }
  ],
  "scopes": [
    {
      "type": "instance",
      "scope": {
        "granularity": "host",
        "bk_biz_id": 100,
        "filter": {},
        "instance_ids": [10001, 10002, 10003]
      }
    }
  ]
}
```

#### 示例5：使用动态分组（dynamic_group）指定范围

创建一个部署策略，用于在动态分组中的所有主机上部署插件。

```json
{
  "name": "动态分组监控插件部署策略",
  "description": "针对动态分组中的主机部署监控插件",
  "enabled": true,
  "specs": [
    {
      "type": "specify_plugin",
      "param": {
        "plugin_name": "bkmonitorbeat",
        "version": "3.60.3066"
      }
    }
  ],
  "scopes": [
    {
      "type": "dynamic_group",
      "scope": {
        "granularity": "host",
        "bk_biz_id": 100,
        "filter": {},
        "dynamic_group_ids": ["group-abc123", "group-def456"]
      }
    }
  ]
}
```

### 响应示例

```json
{
  "code": 0,
  "message": "ok",
  "request_id": "req-123456",
  "data": {
    "deploy_policy_id": 1001
  }
}
```

### 响应参数说明

| 参数名称       | 参数类型   | 描述          |
|------------|--------|-------------|
| code       | int32  | 状态码，0表示成功   |
| message    | string | 请求信息        |
| request_id | string | 请求ID        |
| error      | object | 错误信息（成功时为空） |
| data       | object | 响应数据        |

#### data

| 参数名称             | 参数类型  | 描述          |
|------------------|-------|-------------|
| deploy_policy_id | int64 | 创建成功的部署策略ID |
