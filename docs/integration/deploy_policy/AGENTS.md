|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:中文主述|English technical terms/code identifiers/paths/API names
|Compression Rule:pipe-index format|concise|no prose sections|no code blocks
|Scope:docs/integration/deploy_policy
|Overview:deploy_policy 第三方接入指南|cohesive function folder|curl-first|desired-state oriented
|Structure:docs/integration/deploy_policy:{README.md,specify_plugin.md,specify_plugin_pkg.md,specify_plugin_sub_config.md}
|Source of truth:spec semantics=docs/concepts/deploy_policy/spec.md|scope semantics=docs/concepts/deploy_policy/scope.md|API endpoints/envelopes=docs/api/swagger/backend/api/v3/deploy_policy.swagger.json|variant fields=proto/backend/api/v3/deploy_policy.proto
|Conventions:README.md 是 flow entrypoint|按 integration flow 组织|spec modes 放在声明 desired state 下|详情链接到 mode pages
|Conventions:mode page sections={目的与适用场景,输入,最小 payload 和 curl template,系统解释,即时输出,最终或机器侧可见产物,重复行为,失败情况与限制,Contract 参考}
|Conventions:每个 executable mode page 必须包含 curl template|初始化 shell variables|复用返回 ID 时先提取|不发明 auth/tenant headers|不包含 secrets/真实 identifiers
|Conventions:分离 API accepted/created IDs 与 eventual machine-visible effects|公开 contract 缺口标为 待确认 或 not specified
|specify_plugin_sub_config:仅用于已安装插件的 config-only 声明|不安装/升级 plugin|当前 deploy_policy execute 不支持该 mode|不要写 execute curl
|Unsupported mode:specify_plugin_pkg_sub_config 出现在 concept docs 中，但当前 proto/swagger/types 缺失|文档中标记为 unavailable
|Anti-patterns:不完整复制 Swagger|不写 internal call chains|不发明 operation/status endpoint|不发明 machine paths/reload/health/idempotency/retry/timing guarantees
