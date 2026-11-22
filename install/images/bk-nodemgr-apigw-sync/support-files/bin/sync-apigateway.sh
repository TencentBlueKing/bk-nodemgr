#!/bin/bash

# 加载 apigw-manager 原始镜像中的通用函数
source /apigw-manager/bin/functions.sh

# 待同步网关名，需修改为实际网关名；
# - 如在下面指令的参数中，指定了参数 --gateway-name=${gateway_name}，则使用该参数指定的网关名
# - 如在下面指令的参数中，未指定参数 --gateway-name，则使用 Django settings BK_APIGW_NAME
gateway_name="bk-nodemgr"

# 待同步网关、资源定义文件
definition_file="/data/definition.yaml"
resources_file="/data/resources.yaml"

# Prepare MCP input before any gateway writes; explicit stage definitions take precedence.
if [ "${ENABLE_SYNC_MCP_SERVERS:-false}" = "true" ]; then
	MCP_DEFINITION_FILE="${MCP_DEFINITION_FILE:-/data/mcp-definition.yaml}"
	mcp_args=(--definition "${definition_file}" --mcp-servers /data/mcp_servers.yaml --output "${MCP_DEFINITION_FILE}")
	if [ "${ENABLE_BUILTIN_MCP_SERVERS:-false}" = "true" ]; then
		mcp_args+=(--enable-builtin)
	fi
	python3 /data/bin/prepare-mcp-definition.py "${mcp_args[@]}" || exit 1
fi

title "begin to db migrate"
call_command_or_warning migrate apigw

title "syncing apigateway"
call_definition_command_or_exit sync_apigw_config "${definition_file}" --gateway-name=${gateway_name}
call_definition_command_or_exit sync_apigw_stage "${definition_file}" --gateway-name=${gateway_name}
call_definition_command_or_exit sync_apigw_resources "${resources_file}" --gateway-name=${gateway_name} --delete
call_definition_command_or_exit sync_resource_docs_by_archive "${definition_file}" --gateway-name=${gateway_name} --safe-mode
call_definition_command_or_exit grant_apigw_permissions "${definition_file}" --gateway-name=${gateway_name}

title "fetch apigateway public key"
apigw-manager.sh fetch_apigw_public_key --gateway-name=${gateway_name} --print >"apigateway.pub"

title "releasing"
# 创建资源版本并发布；指定参数 --generate-sdks 时，会同时生成资源版本对应的网关 SDK, 指定 --stage stage1 stage2 时会发布指定环境,不设置则发布所有环境
# 指定参数 --no-pub 则只生成版本，不发布

release_stages="${RELEASE_STAGES:-}"
no_pub="${NO_PUB:-true}"

if [ "$no_pub" = "true" ]; then
	call_definition_command_or_exit create_version_and_release_apigw "${definition_file}" --gateway-name=${gateway_name} --no-pub
fi

if [ "$release_stages" != "" ]; then
	call_definition_command_or_exit create_version_and_release_apigw "${definition_file}" --gateway-name=${gateway_name} --stage "${release_stages}"
fi

if [ "${ENABLE_SYNC_MCP_SERVERS:-false}" = "true" ]; then
	title "syncing stage MCP Servers"
	call_definition_command_or_exit sync_apigw_stage_mcp_servers "${MCP_DEFINITION_FILE:-${definition_file}}" --gateway-name=${gateway_name}
fi

title "done"
