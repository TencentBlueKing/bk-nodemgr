# start_plugin

## 目的与适用场景

`start_plugin` 用于批量启动指定主机上已安装的插件进程。

该能力由 `POST /api/v3/plugin/start` 提供，异步启动已安装的插件进程并返回 `workflow_id`。

该接口只启动已安装的插件进程，不负责安装或停止其他进程。

## 输入

请求字段以 [plugin.proto `PluginStartReq`](../../../proto/backend/api/v3/plugin.proto) 为准：

| 字段                   | Required | 含义                             |
| ---------------------- | -------- | -------------------------------- |
| `plugin`               | yes      | 待启动插件列表，至少包含一个元素 |
| `plugin[].bk_host_id`  | yes      | 目标主机 ID，必须为正整数        |
| `plugin[].plugin_name` | yes      | 插件名称，不能为空               |

## 最小 payload 和 curl template

需要 `curl` 和 `jq`。先设置当前部署的 API base URL、目标主机和插件名称。

```bash
export BK_NODEMGR_API_BASE="https://bk-nodemgr.example.com"
export BK_HOST_ID=1001
export PLUGIN_NAME="bkunifylogbeat"

START_RESPONSE="$(curl -sS -X POST "${BK_NODEMGR_API_BASE}/api/v3/plugin/start" \
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

START_WORKFLOW_ID="$(printf '%s' "${START_RESPONSE}" | jq -r '.data.workflow_id')"
```

## 系统解释

平台把请求解释为一次插件启动：

- 按 `bk_host_id` 和 `plugin_name` 定位已安装的插件进程。
- 通过 GSE 托管并启动进程。
- 等待进程进入 `running` 且 `auto_start=true`。

## 即时输出

响应包含 `data.workflow_id`，表示启动 workflow 已创建。

`workflow_id` 不证明进程已经启动；必须等待 workflow 进入成功终态。

## 最终或机器侧可见产物

启动期望产生的可观察效果：

| 产物       | 说明                                                         |
| ---------- | ------------------------------------------------------------ |
| 进程运行   | 插件进程状态为 `running`                                     |
| 进程自启动 | `auto_start = true`                                          |
| 启动结果   | 通过 `plugin/workflow/*` 读取启动 workflow，确认进入成功终态 |

## 重复行为

公开 contract 不定义重复 `start` 调用的合并、替换或去重行为，也不定义 idempotency key 或 retry window。

重复调用会创建新的启动 workflow；进程已经运行时的具体重复启动语义未在 contract 中定义。

## 失败情况与限制

- `plugin` 至少包含一个元素；`bk_host_id` 必须为正整数；`plugin_name` 不能为空。
- 目标主机上没有对应进程记录时，启动 workflow 无法完成；应先安装插件。
- `start` 是异步 workflow；成功响应只表示启动任务已提交。
- 公开 contract 不定义启动最大等待时长、失败重试窗口或推荐轮询间隔。

## Contract 参考

- [Plugin Swagger](../../api/swagger/backend/api/v3/plugin.swagger.json)
- [Plugin Proto](../../../proto/backend/api/v3/plugin.proto)
- [Plugin Start API](../../../apigw/apidocs/zh/Plugin_StartPlugin.md)
