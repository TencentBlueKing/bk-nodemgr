# install_v3_no_start

## 目的与适用场景

`install_v3_no_start` 用于向指定主机下发并安装 V3 插件包和配置，但不启动插件进程。

该能力由 `POST /api/v3/plugin/install_no_start` 提供：

- 下发并安装 V3 插件包。
- 完成 V3 配置渲染。
- 不启动 V3 插件进程。
- 如果目标主机上的 V3 进程正在运行，会先停止该进程。
- 任务成功结束后，V3 进程记录为 `stopped`。

该接口只负责 V3 下发和安装，不负责启动进程。

## 输入

请求字段以 [plugin.proto `PluginInstallNoStartReq`](../../../proto/backend/api/v3/plugin.proto) 为准：

| 字段                             | Required | 含义                                                             |
| -------------------------------- | -------- | ---------------------------------------------------------------- |
| `plugin`                         | yes      | 待安装插件列表，至少包含一个元素                                 |
| `plugin[].bk_host_id`            | yes      | 目标主机 ID，必须为正整数                                        |
| `plugin[].plugin_name`           | yes      | 插件名称                                                         |
| `plugin[].version`               | no       | V3 插件版本；未传或传空时使用当前默认发布版本                    |
| `plugin[].custom_config_context` | no       | 自定义配置渲染上下文                                             |
| `plugin[].custom_spec`           | no       | 自定义进程资源及监控策略；传入时整体覆盖默认配置和匹配的配置策略 |

`custom_config_context` 禁止使用系统保留顶层键，包括 `PluginInfo`、`NodeInfo`、`PreDefinitionConstants`、`CustomContext`、`plugin_path`、`nodeman`、`cmdb_instance`、`target`、`control_info`。

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL、目标主机和插件名称。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_HOST_ID=1001
export PLUGIN_NAME="bkunifylogbeat"

INSTALL_NO_START_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/install_no_start" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
  {
    "plugin": [
      {
        "bk_host_id": ${BK_HOST_ID},
        "plugin_name": "${PLUGIN_NAME}"
      }
    ]
  }
EOF
)"

INSTALL_NO_START_WORKFLOW_ID="$(printf '%s' "${INSTALL_NO_START_RESPONSE}" | jq -r '.data.workflow_id')"
```

如目标主机需要自定义配置，增加 `custom_config_context`：

```json
{
  "plugin": [
    {
      "bk_host_id": 1001,
      "plugin_name": "bkunifylogbeat",
      "custom_config_context": {
        "cluster_id": "prod-ap-guangzhou"
      }
    }
  ]
}
```

## 系统解释

平台把请求解释为一次 V3 插件下发和安装：

- `bk_host_id` 决定目标主机。
- `plugin_name` 决定安装的插件。
- `version` 未传或传空时由服务端选择默认发布版本；若指定版本，调用方需要确认目标主机的 OS 和 CPU 架构有对应包。
- `custom_config_context` 只影响本次 V3 配置渲染。
- `custom_spec` 传入时整体覆盖资源及监控策略，不进行逐字段合并；省略的对象或数值字段按零值处理。

安装 workflow 会先停止正在运行的 V3 进程，再下发并安装插件包和配置，最后把进程记录标记为 `stopped`。

## 即时输出

响应包含 `data.workflow_id`，表示下发（不启动）workflow 已创建。

`workflow_id` 不证明：

- V3 插件包已经下载、校验、传输并安装完成。
- V3 配置已经渲染完成。
- V3 进程记录已经变成 `stopped`。

## 最终或机器侧可见产物

安装期望产生的可观察效果：

| 产物        | 说明                                                         |
| ----------- | ------------------------------------------------------------ |
| V3 进程记录 | 目标主机存在对应 `plugin_name` 的 V3 进程记录                |
| V3 进程状态 | 进程记录状态为 `stopped`                                     |
| 安装结果    | 通过 `plugin/workflow/*` 读取安装 workflow，确认进入成功终态 |

## 重复行为

公开 contract 不定义重复 `install_no_start` 调用的合并、替换或去重行为，也不定义 idempotency key 或 retry window。

重复调用会创建新的 install-no-start workflow；如果 V3 进程正在运行，新 workflow 会先停止该进程，再重新下发和安装。

## 失败情况与限制

- `plugin` 至少包含一个元素；`bk_host_id` 必须为正整数；`plugin_name` 不能为空。
- `custom_config_context` 的保留顶层键会被拒绝；具体校验错误需通过 workflow 状态和日志查看。
- `custom_spec` 传入时整体覆盖，不会与默认配置逐字段合并。
- `install_no_start` 不启动进程。
- 安装成功响应只表示 workflow 已创建，不证明安装已经完成。

## Contract 参考

- [Plugin Swagger](../../api/swagger/backend/api/v3/plugin.swagger.json)
- [Plugin Proto](../../../proto/backend/api/v3/plugin.proto)
- [Plugin Install No Start API](../../../apigw/apidocs/zh/Plugin_InstallPluginNoStart.md)
