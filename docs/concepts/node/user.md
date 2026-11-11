## User（主机账号）

主机账号需要按用途区分：`LoginUser` 回答“安装时用谁登录主机”，`OperateUser` 回答“GSE 用谁操作插件进程和文件”，默认管理员账号回答“登录用户为空时选谁”。插件模板中的 `account` 则是配置渲染值，不等于 GSE 操作用户。

这些名称可能取相同的值，但不能因此互相替代。

## 账号定义

| 名称 / 字段                    | 常用名称             | 所属模型或来源                          | 含义                                                                                       |
| ------------------------------ | -------------------- | --------------------------------------- | ------------------------------------------------------------------------------------------ |
| `login_user` / `LoginUser`     | 登录用户             | `Host.Dynamic.LoginUser`                | 安装及主机信息探测时，通过 SSH/WMI 认证登录目标主机的操作系统账号。                        |
| `operate_user` / `OperateUser` | GSE 操作用户         | `Host.Dynamic.OperateUser`              | GSE 通道操作插件进程和文件时使用的操作系统身份，包括插件进程身份、配置文件推送的 `Owner`。 |
| `AdminUser`                    | 默认系统管理员账号   | `criteria.DefaultAdminUser`             | 按操作系统选择的默认管理员账号，用于补全空的 `LoginUser`；不是独立的 Host 字段。           |
| `account`                      | 插件配置中的账号变量 | 配置来源主机的 `Host.Dynamic.LoginUser` | 插件配置渲染上下文中的兼容变量，仍取登录用户，不随 GSE 操作身份切换。                      |

这些账号都描述目标主机上的身份，但生效位置不同：

- `LoginUser` 面向安装链路的远程登录认证，与 `login_ip`、`login_port`、认证方式及凭证配合使用。
- `OperateUser` 面向 GSE 通道中的插件管理，不使用安装时的登录凭证来代替操作身份。
- `AdminUser` 是默认值规则，不代表每台主机的登录用户必须固定为管理员账号。
- `account` 面向插件配置内容，不能用来判断 GSE 实际使用的进程身份或文件属主。

> 蓝鲸请求用户、workflow 的 `Operator` 和 CMDB 主机维护人也不是主机登录账号。它们描述请求发起者或资产维护人，不应直接填入 `LoginUser` 或 `OperateUser`。

## 默认值与补全规则

| 操作系统                                    | 默认登录用户（`DefaultAdminUser`） | 默认 GSE 操作用户（`DefaultOperateUser`） |
| ------------------------------------------- | ---------------------------------- | ----------------------------------------- |
| Linux 及默认值函数支持的其他 Unix-like 系统 | `root`                             | `root`                                    |
| Windows                                     | `Administrator`                    | `system`                                  |

默认值函数支持的操作系统范围以各自实现为准，不能仅凭 OS 名称推断所有平台均受支持。

- **主机同步**：检查并补全空的 `LoginUser`；`OperateUser` 的值完全由操作系统决定，每次同步都按操作系统默认值复写（不复用 DB 存量值），将结果写入主机动态属性；不是把 `LoginUser` 复制给 `OperateUser`。
- **插件运行时解析**：`EnsureHostLoginUser` 在历史主机登录用户为空时按节点操作系统类型补全；`EnsureHostOperateUser` 总是按节点当前操作系统类型解析操作用户，不沿用任何存量值（避免重装操作系统后残留旧值）。这一步本身不负责写回数据库。
- **操作系统不受支持**：同步补全过程记录警告并跳过对应默认值；插件运行时兜底返回错误，不任意选择 `root` 或 `system`。

默认值补全只选择账号名称，不代表账号、凭证或安装权限已经验证。

`OperateUser` 由系统按操作系统推导（同步时复写、运行时按当前操作系统解析），当前没有面向用户的修改入口；它不是安装页面上可自由选择的登录账号。

### Windows 的特殊语义

Windows 的 `Administrator` 和 `system` 是不同身份：

- `Administrator` 用于登录认证，也是 Windows 默认登录用户。
- `system` 是 GSE 操作用户使用的值。Windows 下以指定用户身份执行 GSE 操作需要该用户密码，因此默认使用 `system`，而不是复用登录用户。
- 插件模板 `account` 仍取 `LoginUser`。部分插件配置需要 `Administrator`，不能因为 GSE 操作身份是 `system` 就统一替换模板中的账号。

## 典型场景

| 场景                                       | 登录用户        | GSE 操作用户 | 模板 `account`  | 应如何理解                                                    |
| ------------------------------------------ | --------------- | ------------ | --------------- | ------------------------------------------------------------- |
| Linux 使用默认账号                         | `root`          | `root`       | `root`          | 值相同，但登录认证、GSE 操作和配置渲染仍是不同职责。          |
| Linux 使用具备所需安装权限的 `ubuntu` 登录 | `ubuntu`        | `root`       | `ubuntu`        | 登录账号不决定后续 GSE 插件进程和配置文件使用的身份。         |
| Windows 使用默认账号                       | `Administrator` | `system`     | `Administrator` | 登录认证和插件模板仍使用登录用户，GSE 插件管理使用 `system`。 |

上表假定模板使用当前目标主机作为配置来源；使用其他配置来源主机时，`account` 取该来源主机的 `LoginUser`。

## 使用边界

- **安装登录失败**：检查登录用户、地址、认证方式、凭证和安装所需权限，不通过修改 GSE 操作用户解决。
- **GSE 插件进程或配置文件权限异常**：检查 `OperateUser`、进程身份及文件属主，不因为某个账号能 SSH 登录就将其作为 GSE 操作用户。
- **插件配置中的账号不符合预期**：检查模板变量及配置来源主机；`account` 和 Go template 的 `NodeInfo.Dynamic.LoginUser` 仍表达登录用户。
- **需要判断插件实际进程身份**：查看插件部署生成的 `ProcessSpec.Identity.User`，不能仅根据模板变量 `account` 推断。

## 实现位置

| 语义               | 实现位置                                                                                                                   |
| ------------------ | -------------------------------------------------------------------------------------------------------------------------- |
| 主机账号模型       | `pkg/types/host.go`                                                                                                        |
| 默认登录用户       | `pkg/runtime/criteria/admin_user.go`                                                                                       |
| 默认 GSE 操作用户  | `pkg/thirdparty/gse/operate_user.go`                                                                                       |
| 同步补全与持久化   | `internal/backend/manager/workflowdef/syncdata/action_sync_host.go`                                                        |
| 插件运行时账号兜底 | `internal/backend/manager/workflowdef/plugin/utils/user.go`、`internal/backend/manager/workflowdef/pluginv2/utils/user.go` |
| 标准插件进程身份   | `internal/backend/manager/workflowdef/plugin/process_spec.go`                                                              |
| 标准插件配置渲染   | `internal/backend/manager/workflowdef/plugin/action_ensure_and_update_plugin_config_details.go`                            |

登录用户与 GSE 操作用户的分离见 [PR #3945](https://github.com/TencentBlueKing/bk-nodemgr/pull/3945)。

## 相关文档

- [Host（主机）：模型与 IP 字段](host.md)
- [节点操作：传输模式与状态回写](control_mode.md)
- [插件开发指南](../../developer/plugin_pkg_build_guide.md)
