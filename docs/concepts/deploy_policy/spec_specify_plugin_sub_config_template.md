## specify_plugin_sub_config_template

`SpecifyPluginSubConfigTemplate` 根据用户指定的配置文件模板，为指定插件生成 deploy-policy 管理的子配置文件。

该 spec 只声明配置文件的期望状态，不负责安装插件、升级插件或启动插件。`plugin_name` 是用户指定的配置承载插件；plugin 落地到 host 后由 process 表达，所以运行态判断发生在该 host 上 `plugin_name` 对应的 process。只有当前部署范围内存在该 Running process，才会生成或保留对应配置文件。

### 决策输入

| 输入                 | 含义                                      |
| -------------------- | ----------------------------------------- |
| `deploy scope`       | 当前部署策略计算出的目标范围              |
| `plugin_name`        | 用户指定的配置承载插件                    |
| `config template`    | 用户指定的配置文件模板                    |
| `managed config set` | 当前 deploy policy 已管理的子配置文件集合 |

### 决策树

```mermaid
flowchart TD
    Start([开始]) --> BuildDesired[根据配置文件模板<br/>构造期望配置文件集合]
    BuildDesired --> ScanCurrent[扫描当前 deploy policy<br/>已管理的配置文件集合]

    subgraph Cleanup[阶段 1：收敛已管理配置文件集合]
        ScanCurrent --> CurrentInScope{配置文件所在主机<br/>是否在当前部署范围内？}
        CurrentInScope -->|否| NeedDelete[进入 delete_sub_config<br/>配置文件不属于期望集合]
        CurrentInScope -->|是| CurrentRunning{该 host 上是否存在<br/>指定 plugin 对应的 Running process？}
        CurrentRunning -->|否| NeedDelete
        CurrentRunning -->|是| CurrentDeclared{配置文件是否仍由<br/>当前模板声明？}
        CurrentDeclared -->|否| NeedDelete
        CurrentDeclared -->|是| KeepCurrent[no-op<br/>保留当前配置文件]

        NeedDelete --> DeleteProcessRunning{该 host 上是否存在<br/>指定 plugin 对应的 Running process？}
        DeleteProcessRunning -->|是| DeleteFileAndRecord[删除机器上的配置文件<br/>再删除 DB 配置记录]
        DeleteProcessRunning -->|否| DeleteRecordOnly[仅删除 DB 配置记录<br/>不下发机器删除]
    end

    DeleteFileAndRecord --> EnsureDesired
    DeleteRecordOnly --> EnsureDesired
    KeepCurrent --> EnsureDesired

    subgraph Ensure[阶段 2：补齐期望配置文件集合]
        EnsureDesired[遍历当前部署范围<br/>和模板配置项] --> DesiredRunning{该 host 上是否存在<br/>指定 plugin 对应的 Running process？}
        DesiredRunning -->|否| SkipStopped[no-op<br/>不生成配置文件]
        DesiredRunning -->|是| DesiredExists{期望配置文件<br/>是否已经存在？}
        DesiredExists -->|是| SkipExisting[no-op<br/>避免重复生成]
        DesiredExists -->|否| Apply[apply_sub_config<br/>根据模板生成配置文件]
    end

    SkipStopped --> Done([结束])
    SkipExisting --> Done
    Apply --> Done

    classDef start fill:#E6F4FF,stroke:#1769AA,stroke-width:2px,color:#0B3D5C
    classDef decision fill:#FFF4CC,stroke:#8A6D00,stroke-width:2px,color:#3D2F00
    classDef apply fill:#E8F5E9,stroke:#2E7D32,stroke-width:2px,color:#1B5E20
    classDef delete fill:#FFEBEE,stroke:#C62828,stroke-width:2px,color:#7F0000
    classDef noop fill:#F3F4F6,stroke:#4B5563,stroke-width:2px,color:#111827

    class Start,Done start
    class CurrentInScope,CurrentRunning,CurrentDeclared,DeleteProcessRunning,DesiredRunning,DesiredExists decision
    class BuildDesired,ScanCurrent,EnsureDesired,Apply apply
    class NeedDelete,DeleteFileAndRecord,DeleteRecordOnly delete
    class KeepCurrent,SkipStopped,SkipExisting noop
```

### 边界规则

| 场景                                                                                 | 决策                     |
| ------------------------------------------------------------------------------------ | ------------------------ |
| 当前部署范围内存在指定 plugin 对应的 Running process，且模板声明的配置文件不存在     | 生成配置文件             |
| 当前部署范围内存在指定 plugin 对应的 Running process，且配置文件已存在并仍由模板声明 | 保留配置文件，不重复生成 |
| host 上不存在指定 plugin 对应的 Running process                                      | 删除 DB 配置记录         |
| `managed config set` 中存在不在当前 `deploy scope` 内的插件配置文件                  | 进入删除分支             |
| `managed config set` 中存在当前模板不再声明的配置文件                                | 进入删除分支             |

### 删除分支

| 条件                                                                   | 行为                                     |
| ---------------------------------------------------------------------- | ---------------------------------------- |
| 需要删除配置文件，且该 host 上存在指定 plugin 对应的 Running process   | 删除机器上的配置文件，再删除 DB 配置记录 |
| 需要删除配置文件，但该 host 上不存在指定 plugin 对应的 Running process | 仅删除 DB 配置记录，不下发机器删除       |

### 结果语义

`SpecifyPluginSubConfigTemplate` 的收敛目标是：

```text
expected config files = Running process for specified plugin in deploy scope × config template items
```

任何不属于该集合、但仍在当前 deploy policy 管理集合中的子配置文件，都应被删除。
