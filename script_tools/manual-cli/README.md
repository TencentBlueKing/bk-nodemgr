# Manual Install CLI

`script_tools/manual-cli` 提供通过蓝鲸 API 网关（apigw）触发 bk-nodemgr **手动安装（manual install）** 并获取一键安装命令的命令行工具。

单个 Python 文件（`manual_cli.py`），仅依赖 Python 3 标准库，无需安装第三方包。

## 调用链路

```text
manual_cli.py -> bk-apigateway -> bk-nodemgr backend
```

涉及的网关接口：

| 用途 | 接口 |
| ---- | ---- |
| 安装预检（区分新装/重装） | `POST /api/v3/node/agent/install_check` |
| 触发手动安装 | `POST /api/v3/node/agent/install`（`is_manual=true`） |
| 按任务流查操作 | `POST /api/v3/node/workflow/operation/list` |
| 获取一键安装命令 | `POST /api/v3/node/workflow/operation/manual/info/get` |

## 认证与前置条件

认证方式与 `pkg/thirdparty/apigw/client` 的 `un` 模式一致，通过 `X-Bkapi-Authorization` 头传递：

```json
{"bk_app_code": "...", "bk_app_secret": "...", "bk_username": "..."}
```

前置条件：

1. 调用方 `bk_app_code` 已被授予 bk-nodemgr 网关资源权限（网关侧 `allowApplyPermission=false`，需运维提前授权）。
2. 对应身份在 IAM 具备 `agent_operate`（操作 Agent）和 `networkunit_use_for_agent`（使用网络单元部署 Agent）权限。
3. `bk_username` 为租户作用域用户名：单租户环境一般就是登录名（如 `admin`）；多租户（system 租户）通常是 `bk_<login_name>` 形式（如 `bk_admin`）。后端要求 apigw 签发的 JWT 中 `username` claim 非空，缺失会报 `401 username is required`。

## 公共参数

| 参数 | 说明 |
| ---- | ---- |
| `--apigw-url` | 网关环境地址，形如 `http://bkapi.example.com/api/bk-nodemgr/prod` |
| `--app-code` | 应用 bk_app_code |
| `--app-secret` | 应用 bk_app_secret；**推荐改用 `BK_APP_SECRET` 环境变量**，避免泄露到 shell history / `ps` |
| `--bk-username` | 租户作用域用户名 |
| `--http-timeout` | 单次 HTTP 请求超时秒数，默认 30 |

## 子命令

### check — 安装预检

只调 `install_check`，输出判定结果（JSON），`category=error` 时退出码为 1：

```bash
python3 manual_cli.py \
    --apigw-url http://bkapi.example.com/api/bk-nodemgr/prod \
    --app-code bk-demo --bk-username admin \
    check --biz-id 100 --networkunit-id 1 --ip 10.0.0.1
```

### install — 触发手动安装

默认先自动执行 `install_check` 判定新装/重装，再触发安装：

- 匹配到已有主机（`matched.bk_host_id >= 0`）→ **重装**，请求自动带上该 `bk_host_id`；未指定 `--os-type` 时沿用已有主机的 os_type
- 无匹配（`register_to_cmdb_and_install`）→ **新装**，不带 host_id
- `need_confirm`（如 IP 冲突）→ 默认中止，加 `--force` 后继续并复用冲突主机
- `error` → 直接退出

```bash
export BK_APP_SECRET="your-app-secret"

python3 manual_cli.py \
    --apigw-url http://bkapi.example.com/api/bk-nodemgr/prod \
    --app-code bk-demo --bk-username admin \
    install \
    --biz-id 100 \
    --networkunit-id 1 \
    --ip 10.0.0.1 \
    --os-type linux \
    --wait \
    --output /tmp/manual_install_cmd.txt
```

常用可选参数：

| 参数 | 说明 |
| ---- | ---- |
| `--host-id` | 显式指定 bk_host_id（跳过自动识别时也可单独配合 `--skip-check` 使用） |
| `--re-register` | 重装时重新注册主机 |
| `--skip-check` | 跳过 `install_check` 直接安装 |
| `--force` | `install_check` 需要确认时强制继续 |
| `--wait` | 阻塞轮询直到一键安装命令生成（默认间隔 5s、超时 300s，`--poll-interval`/`--poll-timeout` 可调） |
| `--output` | 把一键安装命令原文写入文件（不带注释行，可直接执行）；缺省打印到 stdout |
| `--addressing` / `--login-*` | 寻址方式与登录字段；manual 模式不使用凭据，这些字段仅为通过服务端校验，有默认值 |

不加 `--wait` 时，install 成功拿到 `workflow_id` 后立即退出。

### manual-info — 查询已有任务的一键安装命令

```bash
# 已知 workflow_id + operation_id，查一次
python3 manual_cli.py ... manual-info --workflow-id wf-xxx --operation-id op-yyy

# 只有 workflow_id，按 IP（--ip/--ipv6 至少其一）定位 operation，并阻塞轮询
python3 manual_cli.py ... manual-info --workflow-id wf-xxx --ip 10.0.0.1 --wait
```

省略 `--operation-id` 时必须提供 `--ip` 或 `--ipv6` 作为定位条件；定位结果必须恰好命中一条 operation，命中多条（如批量任务流）会直接报错并提示显式传 `--operation-id`，避免拿到其他主机的安装命令。

## 输出约定

- 一键安装命令原文（或 `workflow_id` JSON）输出到 **stdout**，可用管道/重定向消费
- 诊断日志（判定结果、轮询进度、保存路径）输出到 **stderr**
- 成功退出码 0，失败 1

输出示例：

```text
# type: bash
/bin/bash -c "$(curl -fsSL http://<callback_addr>/api/v3/callback/workflow/node_install/get_manual_script/linux/<oper_inst_id>?action=gen_manual_bootstrap_command)"
```

把该命令拷贝到目标主机执行即可完成手动安装（Windows 主机对应 `# type: bat` 命令）。

## 轮询行为说明

`--wait` 轮询时，每轮先等待一个 `--poll-interval` 再发起请求。操作实例由 workflow 引擎异步拉起，期间 `manual/info/get` 可能返回 `operation has no instances` 或 `private data does not contain bootstrap command`，均属正常瞬态，脚本会自动重试；超过 `--poll-timeout` 仍未就绪才报错退出。若长时间未就绪，应排查 backend workflow worker 是否正常运行。
