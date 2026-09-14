{
  "system_id": {{ .system.id | toJson }},
  "operations": [
    {
      "operation": "upsert_role",
      "data": {
        "id": "plugin_manager",
        "name": "插件管理员",
        "description": "支持业务访问，查看和操作插件及查看其历史；资源范围以授权为准。",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "plugin_view", "resource_type_id": "biz"},
          {"id": "plugin_operate", "resource_type_id": "biz"},
          {"id": "plugin_history_view", "resource_type_id": "biz"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "plugin_viewer",
        "name": "插件只读人员",
        "description": "支持业务访问，查看插件及其历史；资源范围以授权为准。",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "plugin_view", "resource_type_id": "biz"},
          {"id": "plugin_history_view", "resource_type_id": "biz"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "config_policy_viewer",
        "name": "配置策略只读人员",
        "description": "支持业务访问，查看配置策略及其历史；资源范围以授权为准。",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "config_policy_view", "resource_type_id": "biz"},
          {"id": "config_policy_history_view", "resource_type_id": "biz"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "agent_manager",
        "name": "Agent 管理人员",
        "description": "支持业务访问，查看和操作 Agent 及查看其历史，查看管控区域和管控单元，并使用管控单元部署 Agent；资源范围以授权为准。",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "agent_view", "resource_type_id": "biz"},
          {"id": "agent_operate", "resource_type_id": "biz"},
          {"id": "agent_history_view", "resource_type_id": "biz"},
          {"id": "networkarea_view", "resource_type_id": "networkarea"},
          {"id": "networkunit_view", "resource_type_id": "networkunit"},
          {"id": "networkunit_use_for_agent", "resource_type_id": "networkunit"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "proxy_manager",
        "name": "Proxy 管理人员",
        "description": "支持业务访问，查看和操作 Proxy 及查看其历史，查看管控区域和管控单元，并使用管控单元部署 Proxy；资源范围以授权为准。",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "proxy_view", "resource_type_id": "biz"},
          {"id": "proxy_operate", "resource_type_id": "biz"},
          {"id": "proxy_history_view", "resource_type_id": "biz"},
          {"id": "networkarea_view", "resource_type_id": "networkarea"},
          {"id": "networkunit_view", "resource_type_id": "networkunit"},
          {"id": "networkunit_use_for_proxy", "resource_type_id": "networkunit"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "networkarea_creator",
        "name": "管控区域创建者",
        "description": "支持创建管控区域；创建后由系统自动授予创建者对新建管控区域的管理权限。",
        "actions": [
          {"id": "networkarea_create", "resource_type_id": ""}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "networkarea_manager",
        "name": "管控区域管理人员",
        "description": "支持查看、编辑和删除管控区域、查看其历史，以及在管控区域下创建管控单元；资源范围以授权为准。",
        "actions": [
          {"id": "networkarea_view", "resource_type_id": "networkarea"},
          {"id": "networkarea_edit", "resource_type_id": "networkarea"},
          {"id": "networkarea_delete", "resource_type_id": "networkarea"},
          {"id": "networkarea_history_view", "resource_type_id": "networkarea"},
          {"id": "networkunit_create", "resource_type_id": "networkarea"}
        ]
      }
    },
        {
          "operation": "upsert_role",
          "data": {
            "id": "networkunit_manager",
            "name": "管控单元管理人员",
            "description": "支持查看、编辑和删除管控单元、查看其历史，以及使用管控单元部署 Agent 和 Proxy；资源范围以授权为准。",
            "actions": [
              {"id": "networkunit_view", "resource_type_id": "networkunit"},
              {"id": "networkunit_edit", "resource_type_id": "networkunit"},
              {"id": "networkunit_delete", "resource_type_id": "networkunit"},
              {"id": "networkunit_history_view", "resource_type_id": "networkunit"},
              {"id": "networkunit_use_for_agent", "resource_type_id": "networkunit"},
              {"id": "networkunit_use_for_proxy", "resource_type_id": "networkunit"}
            ]
          }
        },
    {
      "operation": "upsert_role",
      "data": {
        "id": "config_policy_manager",
        "name": "配置策略管理人员",
        "description": "支持业务访问，查看和管理配置策略及查看其历史；资源范围以授权为准。",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "config_policy_view", "resource_type_id": "biz"},
          {"id": "config_policy_manage", "resource_type_id": "biz"},
          {"id": "config_policy_history_view", "resource_type_id": "biz"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "package_manager",
        "name": "资源包管理人员",
        "description": "支持按资源包类型上传资源包，以及查看、管理资源包和查看其历史；资源范围以授权为准。",
        "actions": [
          {"id": "package_type_upload", "resource_type_id": "package_type"},
          {"id": "package_view", "resource_type_id": "package"},
          {"id": "package_manage", "resource_type_id": "package"},
          {"id": "package_history_view", "resource_type_id": "package"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "agent_viewer",
        "name": "Agent 只读人员",
        "description": "支持业务访问，查看 Agent 及其历史；资源范围以授权为准。",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "agent_view", "resource_type_id": "biz"},
          {"id": "agent_history_view", "resource_type_id": "biz"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "proxy_viewer",
        "name": "Proxy 只读人员",
        "description": "支持业务访问，查看 Proxy 及其历史；资源范围以授权为准。",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "proxy_view", "resource_type_id": "biz"},
          {"id": "proxy_history_view", "resource_type_id": "biz"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "topo_viewer",
        "name": "网络拓扑只读人员",
        "description": "支持查看管控区域、管控单元及其历史；资源范围以授权为准。",
        "actions": [
          {"id": "networkarea_view", "resource_type_id": "networkarea"},
          {"id": "networkarea_history_view", "resource_type_id": "networkarea"},
          {"id": "networkunit_view", "resource_type_id": "networkunit"},
          {"id": "networkunit_history_view", "resource_type_id": "networkunit"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "package_viewer",
        "name": "资源包只读人员",
        "description": "支持查看资源包及其历史；资源范围以授权为准。",
        "actions": [
          {"id": "package_view", "resource_type_id": "package"},
          {"id": "package_history_view", "resource_type_id": "package"}
        ]
      }
    }
  ]
}
