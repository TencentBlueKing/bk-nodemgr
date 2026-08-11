# 第三方 APIGateway 权限申请

安装前需要为 `bk-nodemgr` 申请以下第三方系统权限。以下权限均按普通权限、永久期限申请。

## bk-gse

| 权限项                                      | 权限说明                                                                                                      | 权限级别 | 申请期限 |
| ------------------------------------------- | ------------------------------------------------------------------------------------------------------------- | -------- | -------- |
| `async_extensions_execute_script`           | 异步脚本执行, 可支持扩展型目标, 包括容器和主机                                                                | 普通     | 永久     |
| `async_extensions_push_file`                | 通过任务通道直接发送文件内容到目标机; 文件内容大小不超过 10KB; 当前版本仅支持纯文本内容推送，不支持二进制内容 | 普通     | 永久     |
| `async_extensions_terminate_execute_script` | 终止脚本任务执行, 可支持扩展型目标, 包括容器和主机                                                            | 普通     | 永久     |
| `async_extensions_terminate_transfer_file`  | 终止文件分发任务, 可支持扩展型目标, 包括容器和主机                                                            | 普通     | 永久     |
| `async_extensions_transfer_file`            | 启动文件分发任务, 可支持扩展型目标, 包括容器和主机                                                            | 普通     | 永久     |
| `dispatch_message`                          | 消息槽下发消息                                                                                                | 普通     | 永久     |
| `dispatch_multi_message`                    | 消息槽下发组合消息                                                                                            | 普通     | 永久     |
| `get_extensions_execute_script_result`      | 查询脚本执行的结果信息, 可支持扩展型目标, 包括容器和主机                                                      | 普通     | 永久     |
| `get_extensions_transfer_file_result`       | 查询文件传输结果, 可支持扩展型目标, 包括容器和主机                                                            | 普通     | 永久     |
| `get_proc_operate_result_v2`                | 进程操作                                                                                                      | 普通     | 永久     |
| `list_agent_info`                           | 查询Agent详情列表信息                                                                                         | 普通     | 永久     |
| `list_agent_state`                          | 查询Agent状态列表信息                                                                                         | 普通     | 永久     |
| `operate_agent`                             | 操作控制Agent                                                                                                 | 普通     | 永久     |
| `operate_proc_multi`                        | 批量进程操作                                                                                                  | 普通     | 永久     |
| `operate_proc_v2`                           | 进程操作                                                                                                      | 普通     | 永久     |

## bk-cmdb

| 权限项                                  | 权限说明                                                 | 权限级别 | 申请期限 |
| --------------------------------------- | -------------------------------------------------------- | -------- | -------- |
| `add_host_to_business_idle`             | 添加主机到业务空闲机                                     | 普通     | 永久     |
| `batch_update_host`                     | 批量更新主机属性                                         | 普通     | 永久     |
| `bind_host_agent`                       | 将agent绑定到主机上                                      | 普通     | 永久     |
| `create_cloud_area`                     | 创建管控区域                                             | 普通     | 永久     |
| `create_dynamic_group`                  | 创建动态分组                                             | 普通     | 永久     |
| `delete_cloud_area`                     | 删除管控区域                                             | 普通     | 永久     |
| `delete_dynamic_group`                  | 删除动态分组                                             | 普通     | 永久     |
| `execute_dynamic_group`                 | 执行动态分组                                             | 普通     | 永久     |
| `find_host_biz_relations`               | 查询主机业务关系信息                                     | 普通     | 永久     |
| `find_host_by_service_template`         | 查询服务模板下的主机                                     | 普通     | 永久     |
| `find_host_by_set_template`             | 查询集群模板下的主机                                     | 普通     | 永久     |
| `find_host_by_topo`                     | 查询拓扑节点下的主机                                     | 普通     | 永久     |
| `find_host_identifier_push_result`      | 获取推送主机身份任务结果                                 | 普通     | 永久     |
| `find_host_relations_with_topo`         | 根据业务拓扑中的实例节点查询其下的主机关系信息           | 普通     | 永久     |
| `find_host_topo_relation`               | 获取主机与拓扑的关系                                     | 普通     | 永久     |
| `find_module_batch`                     | 批量查询某业务的模块详情                                 | 普通     | 永久     |
| `find_set_batch`                        | 批量查询某业务的集群详情                                 | 普通     | 永久     |
| `find_topo_node_paths`                  | 查询业务拓扑节点的拓扑路径                               | 普通     | 永久     |
| `get_biz_brief_cache_topo`              | 查询业务的简要拓扑树信息，包含所有层级的数据，不包含主机 | 普通     | 永久     |
| `get_biz_internal_module`               | 查询业务的空闲机/故障机/待回收模块                       | 普通     | 永久     |
| `get_dynamic_group`                     | 查询指定动态分组                                         | 普通     | 永久     |
| `get_mainline_object_topo`              | 查询主线模型的业务拓扑                                   | 普通     | 永久     |
| `list_biz_hosts`                        | 查询业务下的主机                                         | 普通     | 永久     |
| `list_biz_hosts_topo`                   | 查询业务下的主机和拓扑信息                               | 普通     | 永久     |
| `list_hosts_without_biz`                | 没有业务ID的主机查询                                     | 普通     | 永久     |
| `list_proc_template`                    | 查询进程模板列表                                         | 普通     | 永久     |
| `list_process_instance`                 | 查询进程实例列表                                         | 普通     | 永久     |
| `list_service_instance`                 | 查询服务实例列表                                         | 普通     | 永久     |
| `list_service_instance_by_host`         | 通过主机查询关联的服务实例列表                           | 普通     | 永久     |
| `list_service_instance_by_set_template` | 通过集群模版查询关联的服务实例列表                       | 普通     | 永久     |
| `list_service_instance_detail`          | 获取服务实例详细信息                                     | 普通     | 永久     |
| `list_service_template`                 | 服务模板列表查询                                         | 普通     | 永久     |
| `list_set_template`                     | 查询集群模板                                             | 普通     | 永久     |
| `push_host_identifier`                  | 推送主机身份                                             | 普通     | 永久     |
| `resource_watch`                        | 监听资源变化事件                                         | 普通     | 永久     |
| `search_biz_inst_topo`                  | 查询业务实例拓扑                                         | 普通     | 永久     |
| `search_business`                       | 查询业务                                                 | 普通     | 永久     |
| `search_cloud_area`                     | 查询管控区域                                             | 普通     | 永久     |
| `search_dynamic_group`                  | 搜索动态分组                                             | 普通     | 永久     |
| `search_module`                         | 查询模块                                                 | 普通     | 永久     |
| `search_object_attribute`               | 查询对象模型属性                                         | 普通     | 永久     |
| `search_set`                            | 查询集群                                                 | 普通     | 永久     |
| `unbind_host_agent`                     | 将agent和主机解绑                                        | 普通     | 永久     |
| `update_cloud_area`                     | 更新管控区域                                             | 普通     | 永久     |
| `update_dynamic_group`                  | 更新动态分组                                             | 普通     | 永久     |
| `update_host_cloud_area_field`          | 更新主机的管控区域字段                                   | 普通     | 永久     |

## bk-notice

| 权限项                                   | 权限说明                     | 权限级别 | 申请期限 |
| ---------------------------------------- | ---------------------------- | -------- | -------- |
| `announcement_get_current_announcements` | Get announcement list        | 普通     | 永久     |
| `register_application`                   | register for the application | 普通     | 永久     |
