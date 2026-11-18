# 数据迁移

存放 bk-nodemgr 相关的跨系统数据迁移手册。每个迁移场景一个子目录，包含操作手册和所需脚本。

## [nodeman 插件进程迁移](nodeman_process_to_nodemgr/README.md)

蓝鲸 7.2 升级至 7.3 时，将 nodeman 安装的主插件进程与主配置导出为 mongosh 导入文件，回填到 nodemgr 的 processV2 集合，并触发状态同步。
