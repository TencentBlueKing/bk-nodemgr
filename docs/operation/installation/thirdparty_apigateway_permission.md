# 第三方 APIGateway 资源使用清单

安装前请为 `bk-nodemgr` 按以下资源清单申请第三方 APIGateway 权限，申请期限选择永久。

`BKRepo` 不走 `APIGateway`，不在本清单范围内；其资源使用清单见 [BKRepo 资源使用说明](thirdparty_bkrepo_resource.md)。

## bk-gse

| resource                                    | 资源说明                                                                                                      | 备注 | 申请期限 |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------- | --- | -------- |
| `async_extensions_execute_script`           | 异步脚本执行, 可支持扩展型目标, 包括容器和主机                                                                | — | 永久     |
| `async_extensions_push_file`                | 通过任务通道直接发送文件内容到目标机; 文件内容大小不超过 10KB; 当前版本仅支持纯文本内容推送，不支持二进制内容 | — | 永久     |
| `async_extensions_terminate_execute_script` | 终止脚本任务执行, 可支持扩展型目标, 包括容器和主机                                                            | — | 永久     |
| `async_extensions_terminate_transfer_file`  | 终止文件分发任务, 可支持扩展型目标, 包括容器和主机                                                            | — | 永久     |
| `async_extensions_transfer_file`            | 启动文件分发任务, 可支持扩展型目标, 包括容器和主机                                                            | — | 永久     |
| `dispatch_message`                          | 消息槽下发消息                                                                                                | — | 永久     |
| `get_extensions_execute_script_result`      | 查询脚本执行的结果信息, 可支持扩展型目标, 包括容器和主机                                                      | — | 永久     |
| `get_extensions_transfer_file_result`       | 查询文件传输结果, 可支持扩展型目标, 包括容器和主机                                                            | — | 永久     |
| `get_proc_operate_result_v2`                | 进程操作                                                                                                      | — | 永久     |
| `list_agent_info`                           | 查询Agent详情列表信息                                                                                         | — | 永久     |
| `list_agent_state`                          | 查询Agent状态列表信息                                                                                         | — | 永久     |
| `operate_agent`                             | 操作控制Agent                                                                                                 | — | 永久     |
| `operate_proc_multi`                        | 批量进程操作                                                                                                  | — | 永久     |
| `operate_proc_v2`                           | 进程操作                                                                                                      | — | 永久     |

## bk-cmdb

| resource                                | 资源说明                                                 | 备注 | 申请期限 |
| --------------------------------------- | -------------------------------------------------------- | --- | -------- |
| `add_host_to_business_idle`             | 添加主机到业务空闲机                                     | 隐藏接口 | 永久     |
| `add_host_to_resource_pool`             | 添加主机到资源池                                         | 隐藏接口 | 永久     |
| `batch_update_host`                     | 批量更新主机属性                                         | 隐藏接口 | 永久     |
| `bind_host_agent`                       | 将agent绑定到主机上                                      | 隐藏接口 | 永久     |
| `create_biz_custom_field`               | 创建业务自定义模型属性                                   | 隐藏接口 | 永久     |
| `create_cloud_area`                     | 创建管控区域                                             | 隐藏接口 | 永久     |
| `create_dynamic_group`                  | 创建动态分组                                             | 隐藏接口 | 永久     |
| `delete_cloud_area`                     | 删除管控区域                                             | 隐藏接口 | 永久     |
| `delete_dynamic_group`                  | 删除动态分组                                             | 隐藏接口 | 永久     |
| `execute_dynamic_group`                 | 执行动态分组                                             | 隐藏接口 | 永久     |
| `find_host_biz_relations`               | 查询主机业务关系信息                                     | 隐藏接口 | 永久     |
| `find_host_by_service_template`         | 查询服务模板下的主机                                     | 隐藏接口 | 永久     |
| `find_host_by_set_template`             | 查询集群模板下的主机                                     | 隐藏接口 | 永久     |
| `find_host_by_topo`                     | 查询拓扑节点下的主机                                     | 隐藏接口 | 永久     |
| `find_host_identifier_push_result`      | 获取推送主机身份任务结果                                 | 隐藏接口 | 永久     |
| `find_host_relations_with_topo`         | 根据业务拓扑中的实例节点查询其下的主机关系信息           | 隐藏接口 | 永久     |
| `find_host_topo_relation`               | 获取主机与拓扑的关系                                     | 隐藏接口 | 永久     |
| `find_module_batch`                     | 批量查询某业务的模块详情                                 | 隐藏接口 | 永久     |
| `find_set_batch`                        | 批量查询某业务的集群详情                                 | 隐藏接口 | 永久     |
| `find_topo_node_paths`                  | 查询业务拓扑节点的拓扑路径                               | 隐藏接口 | 永久     |
| `get_biz_brief_cache_topo`              | 查询业务的简要拓扑树信息，包含所有层级的数据，不包含主机 | 隐藏接口 | 永久     |
| `get_biz_internal_module`               | 查询业务的空闲机/故障机/待回收模块                       | 隐藏接口 | 永久     |
| `get_dynamic_group`                     | 查询指定动态分组                                         | 隐藏接口 | 永久     |
| `get_mainline_object_topo`              | 查询主线模型的业务拓扑                                   | 隐藏接口 | 永久     |
| `list_biz_hosts`                        | 查询业务下的主机                                         | 隐藏接口 | 永久     |
| `list_biz_hosts_topo`                   | 查询业务下的主机和拓扑信息                               | 隐藏接口 | 永久     |
| `list_hosts_without_biz`                | 没有业务ID的主机查询                                     | 隐藏接口 | 永久     |
| `list_resource_pool_hosts`         | 查询资源池中的主机                                       | 隐藏接口 | 永久     |
| `list_proc_template`                    | 查询进程模板列表                                         | 隐藏接口 | 永久     |
| `list_process_instance`                 | 查询进程实例列表                                         | 隐藏接口 | 永久     |
| `list_service_instance`                 | 查询服务实例列表                                         | 隐藏接口 | 永久     |
| `list_service_instance_by_host`         | 通过主机查询关联的服务实例列表                           | 隐藏接口 | 永久     |
| `list_service_instance_by_set_template` | 通过集群模版查询关联的服务实例列表                       | 隐藏接口 | 永久     |
| `list_service_instance_detail`          | 获取服务实例详细信息                                     | 隐藏接口 | 永久     |
| `list_service_template`                 | 服务模板列表查询                                         | 隐藏接口 | 永久     |
| `list_set_template`                     | 查询集群模板                                             | 隐藏接口 | 永久     |
| `push_host_identifier`                  | 推送主机身份                                             | 隐藏接口 | 永久     |
| `resource_watch`                        | 监听资源变化事件                                         | 隐藏接口 | 永久     |
| `search_biz_inst_topo`                  | 查询业务实例拓扑                                         | 隐藏接口 | 永久     |
| `search_business`                       | 查询业务                                                 | 隐藏接口 | 永久     |
| `search_cloud_area`                     | 查询管控区域                                             | 隐藏接口 | 永久     |
| `search_dynamic_group`                  | 搜索动态分组                                             | 隐藏接口 | 永久     |
| `search_module`                         | 查询模块                                                 | 隐藏接口 | 永久     |
| `search_object_attribute`               | 查询对象模型属性                                         | 隐藏接口 | 永久     |
| `search_set`                            | 查询集群                                                 | 隐藏接口 | 永久     |
| `unbind_host_agent`                     | 将agent和主机解绑                                        | 隐藏接口 | 永久     |
| `update_cloud_area`                     | 更新管控区域                                             | 隐藏接口 | 永久     |
| `update_dynamic_group`                  | 更新动态分组                                             | 隐藏接口 | 永久     |
| `update_host_cloud_area_field`          | 更新主机的管控区域字段                                   | 隐藏接口 | 永久     |

## bk-monitor

| resource                            | 资源说明                                           | 备注 | 申请期限 |
| ----------------------------------- | -------------------------------------------------- | --- | -------- |
| `get_or_create_agent_event_data_id` | 获取或创建节点管理 Agent 告警事件数据 ID | 隐藏接口 | 永久     |

## bkiam（IAM V4）

| resource | 资源说明 | 备注 | 申请期限 |
| --- | --- | --- | --- |
| `direct_auth` | 单操作鉴权；支持无资源操作 | — | 永久 |
| `direct_auth_by_resources` | 同一操作的多资源批量鉴权 | — | 永久 |
| `direct_auth_by_actions` | 同一资源的多操作批量鉴权 | — | 永久 |
| `list_authorized_resource` | 查询有权限的资源范围 | 隐藏接口 | 永久 |
| `generate_perm_apply_url` | 生成权限申请链接 | — | 永久 |
| `retrieve_system_auth_token` | 获取系统回调认证凭据 | 隐藏接口 | 永久 |
| `add_authorization` | 授予管理角色 | — | 永久 |
| `retrieve_system` | 查询接入系统是否存在 | 隐藏接口 | 永久 |
| `create_system` | 创建接入系统 | 隐藏接口 | 永久 |
| `update_system` | 更新系统信息及回调配置 | 隐藏接口 | 永久 |
| `list_resource_type` | 查询资源类型列表 | 隐藏接口 | 永久 |
| `batch_create_resource_type` | 创建资源类型 | 隐藏接口 | 永久 |
| `update_resource_type` | 更新资源类型 | 隐藏接口 | 永久 |
| `list_action` | 查询操作列表 | 隐藏接口 | 永久 |
| `batch_create_action` | 创建操作 | 隐藏接口 | 永久 |
| `update_action` | 更新操作 | 隐藏接口 | 永久 |
| `list_role` | 查询角色及其操作绑定 | 隐藏接口 | 永久 |
| `batch_create_role` | 创建角色 | 隐藏接口 | 永久 |
| `update_role` | 更新角色信息 | 隐藏接口 | 永久 |
| `batch_create_role_action` | 增加角色的操作绑定 | 隐藏接口 | 永久 |
| `batch_delete_role_action` | 移除角色的操作绑定 | 隐藏接口 | 永久 |

## bk-notice

| resource                                 | 资源说明                     | 备注 | 申请期限 |
| ---------------------------------------- | ---------------------------- | --- | -------- |
| `announcement_get_current_announcements` | Get announcement list        | — | 永久     |
| `register_application`                   | register for the application | — | 永久     |
