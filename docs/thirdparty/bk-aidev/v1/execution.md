# 运行参数与执行

CLI 与 Python 使用同一个 `bkai-init==0.1.6` 包及相同校验规则。Demo 用户可直接按 [Quick Start](quick-start.md) 使用容器，无需在宿主机安装 Python 包。

## 独立安装与生成模板

需要 Python 3.11+，并由平台提供当前网络可访问的可信包源；不假设 0.1.6 已发布到公共 PyPI。

```bash
export BKAI_PIP_INDEX_URL='<bkai-init 包源的 /simple 地址>'
export BKAI_DEPENDENCY_INDEX_URL='<公共依赖源的 /simple 地址>'

python3.11 -m venv .venv-bkai-init
. .venv-bkai-init/bin/activate
python -m pip install \
  --index-url "$BKAI_PIP_INDEX_URL" \
  --extra-index-url "$BKAI_DEPENDENCY_INDEX_URL" \
  "bkai-init==0.1.6"

python -m pip show bkai-init
bkai-init --help
```

上述包源变量仅用于安装示例，不是运行参数。包源不一定包含 Jinja2、PyYAML、requests 等依赖；如一个源已包含全部依赖，可省略 `--extra-index-url`。两个源均参与候选解析，不保证优先级，只使用平台确认的可信源，仓库访问权限由平台说明，不在文档中填写令牌。

生成本地模板并校验：

```bash
bkai-init init ./bk-job --agent-code ai-job --name "作业助手"
export BKAI_SPACE_ID='<目标空间ID>'
bkai-init validate -f ./bk-job/bkai.yaml --space "$BKAI_SPACE_ID"
```

`init` 只生成本地文件，不访问平台，不需要应用凭据，不同步或发布；目标目录必须不存在。需要依赖时增加 `--with-skill job_skill`、`--with-knowledgebase bk_job_docs` 或 `--with-subagent ai-job-exec`，知识库文档由接入方维护。

## 连接与运行参数

| 配置                                                    | 说明                                                                                                                                                                                          |
| ------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BK_API_URL_TMPL`                                       | 与 `BKAI_BASE_URL` 二选一，例如 `https://gateway.example.com/api/{api_name}/{stage}`；`api_name` 分别为 `bk-aidev`、`bk-user`，没有 stage 时补 `prod`。CLI 自动解析，无需再配 `BKAI_BASE_URL` |
| `BKAI_APP_CODE` / `BKAI_APP_SECRET`                     | 必填应用身份；未配置时分别读取 `BKPAAS_APP_ID` / `BKPAAS_APP_SECRET`。命令参数优先，真实凭据不入库                                                                                            |
| `BKPAAS_MULTI_TENANT_MODE`                              | 默认 `false`；多租户设为 `true`。Python 须显式传 `multi_tenant_mode=True`，类本身不自动读取该变量                                                                                             |
| `BKAI_BASE_URL` / `--base-url`                          | AIDEV 网关完整环境入口，命令行优先；未提供才使用模板。专用网关域名不能按通用模板猜测                                                                                                          |
| `BK_USER_BASE_URL` / `--bk-user-base-url`               | 仅多租户且未提供执行用户名时需要，可由模板推导；执行者与 admins 显式提供 `bk_username` 时不做转换                                                                                             |
| `BKAI_TENANT_ID` / `--tenant-id`                        | 仅多租户使用，写入 `X-Bk-Tenant-Id`；未设置时为 `system`，应填写应用实际租户，不自动推断。单租户忽略                                                                                          |
| `BKAI_ACCESS_TOKEN` / `--access-token` / `access_token` | 可选，通常无需配置；环境变量兼容 `ACCESS_TOKEN`，不替代应用、空间及 MCP 授权                                                                                                                  |
| `BKAI_USERNAME` / `--username` / `username`             | 可选；显式填写 `bk_username`，直接透传，不转换。未传默认 `bk_admin`，仅多租户下查询转换                                                                                                       |
| `-f`                                                    | 清单路径，Demo 镜像内为 `/app/package/bkai.yaml`，工作目录 `/app` 下相对路径为 `./package/bkai.yaml`                                                                                          |
| `--space` / `BKAI_SPACE_ID`                             | 默认 `bkaidev`；显式参数优先，不接受空命令行值。未传参数读取环境变量，空值回退默认空间。不自动拼接租户前缀，多租户需显式传完整 ID，如 `system-bkaidev`                                        |
| `--resource`                                            | 可重复，格式 `kind/code`；省略表示全包。类型为 `agent`、`collection`、`skill`、`knowledgebase`，仅接受清单内对象                                                                              |
| `--exclude-resource`                                    | 可重复，优先于选择列表，例如 `agent/ai-job`；适用于 `plan` / `sync`，`show` 不支持黑名单                                                                                                      |
| `--admins` / Make `ADMINS`                              | 空格分隔的 `bk_username`，应用于所有选中的主子 Agent；只追加管理员，不删除已有成员，YAML 禁止 `spec.admins`                                                                                   |
| `--confirm`                                             | 确认执行 `sync`；不传只读预览，不上传、不写入、不发布                                                                                                                                         |
| `--publish`                                             | 显式开启 Agent 发布，先子后主；不传仅同步草稿                                                                                                                                                 |
| `--publish_config_only`                                 | `0` 或 `1`，默认 `0`，须同时指定 `--publish`。`0` 请求发布到开发者中心，`1` 仅发布配置；平台可按代码版本一致策略跳过重新部署                                                                  |
| `--var KEY=VALUE`                                       | 可重复传部署变量；白名单和限制见[变量替换](#变量替换)                                                                                                                                         |

只有未提供执行用户名且开启多租户时，调用应用才需要 BK User 虚拟用户 lookup 权限，用于查询默认 `bk_admin` 对应的 `bk_username`。查询失败或结果不唯一时停止；平台不增加身份查询接口。

## 执行顺序

在已安装包或已有 Demo 镜像中，先加载带 `export` 的 `.env`：

```bash
export BK_API_URL_TMPL='https://gateway.example.com/api/{api_name}/{stage}'
export BKAI_APP_CODE='<应用编码>'
export BKAI_APP_SECRET='<应用密钥>'
export BKPAAS_MULTI_TENANT_MODE=true
export BKAI_TENANT_ID='<应用所在租户>'
export BKAI_SPACE_ID='<目标空间ID>'

# The Demo container working directory is /app.
source .env
bkai-init validate -f ./package/bkai.yaml --space "$BKAI_SPACE_ID"
bkai-init show -f ./package/bkai.yaml --space "$BKAI_SPACE_ID"
bkai-init plan -f ./package/bkai.yaml --space "$BKAI_SPACE_ID" --publish
bkai-init sync -f ./package/bkai.yaml --space "$BKAI_SPACE_ID" \
  --confirm --publish --publish_config_only=0
```

资源选择不会自动同步未选中的依赖；MCP 是 Agent 关联，不是独立同步资源。所有本地配置仍需通过校验。`plan` 只检查、不写入，就绪不保证写权限或服务端执行成功；`plan --format json` 或 `--format yaml` 输出结构化计划。未传 `--confirm` 的 `sync` 同样只预览。

### 发布规则

常规发布显式传递 `is_publish_to_paas_v3=true`。工具只调用发布接口，不轮询、不等待部署完成，不自动重试发布请求：

- `pending` / `running` 标记“已提交，后台处理中”；`successful` 按接口结果记录发布完成，实际部署结果仍到平台确认。
- 同包本轮子智能体提交发布后，主智能体可直接关联，来源指令使用本轮已同步配置。
- 仅当发布明确返回 HTTP 400，且原因是“智能体状态异常，无法发布，当前状态为：publishing”时，记录原因并继续，不计为失败。
- Skill 已有版本正在发布的明确响应也会跳过本次更新并继续；其它错误仍停止。
- 平台“代码版本一致时仅更新配置”的优化保持不变。

### 管理员追加

`--admins alice bob` 中的值均应替换为实际 `bk_username`。Agent 配置写入后、发布前追加管理员，仅支持插件智能体。单租户与多租户均直接透传，不查询用户管理，调用方自行转换 `login_name`。

追加失败或结果未确认时，输出异常并在最终报告保留原因，继续当前 Agent 发布和后续资源同步。已写入配置或权限不会自动回滚。`plan` 仅展示管理员与追加意图，不校验用户映射，不把无法回读的名单判为一致。

## 查看进度与失败信息

`plan` / `sync` 按 `bkai.yaml` 的资源事项展示进度、配置路径与步骤。正常接口只保留缩进后的 URL，不展开输入、响应正文或处理结果；MCP 查询成功合并成一行。以下计数与错误仅示意输出结构：

```text
[3/8] 正在处理：Skill/job_skill
  配置：skills/job_skill/skill.yaml
  ✓ 打包完成
  ✓ 上传完成
  → 正在同步 Skill
      接口请求：POST https://bkapi.example.com/api/bk-aidev/prod/openapi/aidev/app/v1/skills/upsert/
  ✓ 同步完成

      【正在查询 MCP】bk-job-apigw-mcp agent_code: ai-job 成功
```

失败时展开格式化、脱敏信息：

```text
[3/8] 正在处理：Skill/job_skill
  配置：skills/job_skill/skill.yaml
  ✓ 打包完成
  ✓ 上传完成
  → 正在同步 Skill
      接口请求：POST https://bkapi.example.com/api/bk-aidev/prod/openapi/aidev/app/v1/skills/upsert/
      请求参数：
          {
            "skill_code": "job_skill",
            "space_id": "<目标空间ID>"
          }
      接口响应：[HTTP 400]
          {
            "error": {
              "code": "INVALID_ARGUMENT",
              "message": "Dockerfile 基础镜像不在允许列表中: registry.example.com/teamskill:1.0"
            }
          }
  ✗ 同步 Skill失败：Dockerfile 基础镜像不在允许列表中

初始化已停止
  已完成 2 项，失败 1 项，未执行 5 项
  失败事项：Skill/job_skill
  失败步骤：同步 Skill
  排查建议：核对完整基础镜像名及目标环境允许列表
  反馈信息：执行时间、CLI 版本、trace_id 与服务端请求标识（如有）
```

查询请求遇到连接失败、超时、HTTP 429 或 5xx，分别等待 1、2、4 秒重试，最多请求 4 次。写入请求及其它业务错误不盲目重试。失败反馈保留资源、步骤、执行时间、CLI 版本和服务端请求标识，不粘贴原始凭据。

知识库异步导入显示“已提交，后台处理中”，不计为全部处理完成。已完成事项不自动回滚，重试前先用 `show` / `plan` 确认线上状态。

## 更新覆盖保护

- **提供完整配置**：同步 Agent 可能重建资源关联，省略字段不等于保留线上原值。
- **按资源保护**：不希望覆盖平台手工修改的 Agent 时，排除整个 Agent。仅排除 Skill 资源，不能阻止 Agent 中的 Skill 关联或 `envs` 更新。
- **检查环境变量**：`envs` 省略、空数组、缺少 `value`、空字符串语义不同；CLI 不把省略的 `value` 转为空字符串。`[]` 不代表清空全部变量，最终取值需满足 Skill 依赖声明。
- **确认引用与版本**：更新角色后需同步引用它的 Agent；Agent 关联 Skill 使用当前 Skill 版本。资源版本号不代表 Agent 已发布。
- **保持文档完整**：知识库有文档时以完整本地目录为准，不用不完整目录更新已有知识库。

## 变量替换

支持受限 Jinja 变量语法：`{{ KEY }}`、`{{ KEY | default("默认值") }}`、`{{ KEY | required }}`。不自动读取环境变量，替换结果不再次渲染。

| 文件白名单                                                          | 处理方式                                     |
| ------------------------------------------------------------------- | -------------------------------------------- |
| 入口 `bkai.yaml`、清单引用的 Agent / Collection YAML                | 解析后仅替换字符串值，不替换键名、结构或类型 |
| 已声明 Skill 根目录 `skill.yaml`、知识库根目录 `knowledgebase.yaml` | 同上，不扫描所有 YAML                        |
| 已声明 Skill 根目录已有 Dockerfile                                  | 仅替换文本并写入上传包，不修改本地源文件     |

```bash
bkai-init sync -f /app/package/bkai.yaml --space "$BKAI_SPACE_ID" --confirm \
  --var BKAI_AGENT_IMAGE_REPO=registry.example.com/team \
  --var SKILL_NAME=job_skill
```

- 变量名为 1–64 位英文、数字或下划线，不能以数字开头；重复 key 报错，按第一个等号分隔 key/value。
- `default` 只接受单个字符串参数，仅未传变量时生效；显式空字符串保持为空。
- `required` 不接受参数，未传、空字符串或纯空白均报错。
- 不支持表达式、属性访问、函数调用、其它过滤器、循环或条件。
- YAML 独立占位符须加引号，如 `name: "{{ SKILL_NAME | default('作业查询 Skill') }}"`。未传值且无默认值的 YAML 占位符保留，供运行时组件引用；部署变量不要与组件 key 重名。
- Dockerfile 不允许残留占位符，替换值不能含换行或 NUL；变量来源应可信，不通过 CLI 变量传密钥。
- `SKILL.md`（含 frontmatter）、脚本、知识库文档、普通 YAML、嵌套 Dockerfile 不处理；镜像开白与构建仍由平台校验。

## Python 调用

安装同一固定版本后，从包顶层导入 `BkaiInit`，无需另装其它包。先按 Quick Start 加载环境变量，再运行：

```python
import logging
import os
import sys

from bkai_init import BkaiInit

logging.basicConfig(level=logging.INFO, stream=sys.stdout)
api_url_template = os.environ.get("BK_API_URL_TMPL", "")
initializer = BkaiInit(
    base_url=os.environ.get("BKAI_BASE_URL") or api_url_template.format(api_name="bk-aidev", stage="prod"),
    app_code=os.environ["BKAI_APP_CODE"],
    app_secret=os.environ["BKAI_APP_SECRET"],
    bk_user_base_url=os.environ.get("BK_USER_BASE_URL") or api_url_template.format(api_name="bk-user", stage="prod"),
    username=os.environ.get("BKAI_USERNAME"),
    access_token=os.environ.get("BKAI_ACCESS_TOKEN"),
    multi_tenant_mode=os.environ.get("BKPAAS_MULTI_TENANT_MODE", "false").lower() == "true",
    tenant_id=os.environ.get("BKAI_TENANT_ID", "system"),
    space=os.environ["BKAI_SPACE_ID"],
    variables={"BKAI_AGENT_IMAGE_REPO": os.environ.get("BKAI_AGENT_IMAGE_REPO", "")},
)
package = "/app/package/bkai.yaml"
initializer.validate(package)
online = initializer.show(package)
preview = initializer.plan(package, publish=True, publish_config_only=False)

# Execute only after the caller has approved the plan.
report = initializer.sync(package, publish=True, publish_config_only=False)
```

本例模板须含完整 `{api_name}/{stage}`，如 `https://gateway.example.com/api/{api_name}/{stage}`。CLI 自动读取模板；Python 显式传地址和多租户开关，直接提供的 `BKAI_BASE_URL` / `BK_USER_BASE_URL` 分别优先。

`username`、`access_token` 参数可省略。Python 的 `sync()` 直接写入，没有 `confirm` 参数；只读预览使用 `plan()`。发布规则与 CLI 相同，类调用的 logging 由应用配置，CLI 默认输出到 stdout。
