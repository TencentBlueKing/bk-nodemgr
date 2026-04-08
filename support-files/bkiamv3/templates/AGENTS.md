|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
|Required Tools:serena (semantic code ops)|context7 (3rd-party docs)|sequential-thinking (decisions)
|Language Policy:Chinese for Q&A|English for code/docs/tech discussions
|Compression Rule:Follow pipe-index format; keep concise; no prose/code blocks
|Scope:support-files/bkiamv3/templates
|Overview:IAM v3 migration template scope for bk-nodemgr|owns JSON.tpl model templates only
|Structure:support-files/bkiamv3/templates:{0001_bk_nodemgr_system.json.tpl,0002_bk_nodemgr_resource_type.json.tpl,0003_bk_nodemgr_actions.json.tpl,0004_bk_nodemgr_action_groups.json.tpl,0005_bk_nodemgr_common_actions.json.tpl,README.md}
|Where to look:template generation script:support-files/bkiamv3/new_template.sh:use for new numbered template files
|Where to look:render and output flow:support-files/bkiamv3:{iam-render,vars.yaml.example,README.md}:render templates before migration execution
|Where to look:migration execution:support-files/bkiamv3/do_migrate.py:validate rendered JSON against IAM behavior
|Where to look:model references:support-files/bkiamv3/templates/README.md:operation-to-model mapping is source of truth
|Conventions:template filenames must keep migration order and naming `{seq}_{APP_CODE}_{YYYYmmdd-HHMM}_iam.json.tpl` when creating new files
|Conventions:template content must match IAM migration operation payload contracts from README references; do not invent custom fields without upstream doc basis
|Conventions:prefer `upsert_*` operations for idempotent reruns unless one-time add/update/delete is explicitly required
|Conventions:system model updates must keep `clients` containing self `app_code` to avoid follow-up permission lockout
|Conventions:keep variable placeholders stable and explicit (`{{ .path.to.value }}`); defaults only when behavior is deterministic and documented
|Conventions:comments/docs can be Chinese; template keys/operation names stay English and align with IAM spec terms
|Anti-patterns:do not store secrets in templates or examples|do not hardcode environment hosts/app secrets in committed files
|Anti-patterns:do not break sequence ordering or rename historical templates already used in environments
|Anti-patterns:do not replace existing operation semantics by mutation; add next migration template file instead
|Dependencies:templates -> rendered JSON (iam-render) -> do_migrate.py -> IAM API gateway/endpoint
|Commands:render=cd support-files/bkiamv3 && ./iam-render -t templates -v vars.yaml -o output
|Commands:new-template=cd support-files/bkiamv3 && ./new_template.sh <operation_name>
|Commands:migrate=cd support-files/bkiamv3 && python do_migrate.py -t <iam_url> -f <rendered_json> -a <app_code> -s <app_secret>
