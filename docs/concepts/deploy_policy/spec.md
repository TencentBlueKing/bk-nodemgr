## spec

spec 定义了期望的最终状态，用于指定目标最终应达到的状态。

### specify_plugin

指定插件版本，确保目标主机上安装指定名称和版本的插件。如果插件不存在则安装，版本不匹配则升级。

### specify_plugin_sub_config

指定插件子配置，用于更新已安装插件的配置文件内容。仅更新配置，不涉及插件版本的安装或升级。

### specify_plugin_pkg

指定插件包版本，确保目标主机/服务实例上安装指定插件包名称和版本的插件。插件名称会根据部署策略 ID 和模块 ID 自动生成。如果插件不存在则安装，版本不匹配则升级。

### specify_plugin_pkg_sub_config

指定插件包子配置，用于更新已安装插件包的配置文件内容。仅更新配置，不涉及插件包版本的安装或升级。插件名称会根据部署策略 ID 和模块 ID 自动生成。

### project_plugin_pkg_to_hosts

投射插件包实例到指定承载主机。scope 计算出的目标主机/服务实例只作为 remote target，用于派生插件实例身份；插件进程实际运行在 spec 参数 `placement_host_ids` 指定的承载主机上。

对每一个 scope remote target 和每一个 `placement_host_id`，系统都应确保承载主机上存在一个对应插件进程。插件名称根据插件包名、部署策略 ID、remote target 的模块 ID 和 remote target 的主机 ID 自动生成：`{plugin_pkg_name}_{deploy_policy_id}_{remote_module_id}_{remote_host_id}`。当 remote target 是主机粒度时，`remote_module_id` 保持现有 scope target 的零值语义，即 `0`。

该 spec 表达的是 `scope remote targets × placement_host_ids` 的声明式期望状态：期望存在但实际不存在时安装，版本不匹配时升级，实际存在但不再属于期望集合时卸载。`placement_host_ids` 是承载位置，不是第二套 scope，也不表达调度、分片、主备角色或按主机差异化配置。
