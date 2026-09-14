{
  "system_id": {{ .system.id | toJson }},
  "operations": [
    {
      "operation": "upsert_role",
      "data": {
        "id": "agent_manager",
        "name": "Agent 管理员",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "agent_view", "resource_type_id": "biz"},
          {"id": "agent_operate", "resource_type_id": "biz"},
          {"id": "agent_history_view", "resource_type_id": "biz"},
          {"id": "networkarea_view", "resource_type_id": "networkarea"},
          {"id": "networkunit_view", "resource_type_id": "networkunit"},
          {"id": "networkunit_use_for_agent", "resource_type_id": "networkunit"},
          {"id": "plugin_view", "resource_type_id": "biz"},
          {"id": "plugin_operate", "resource_type_id": "biz"},
          {"id": "plugin_history_view", "resource_type_id": "biz"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "proxy_manager",
        "name": "Proxy 管理员",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "proxy_view", "resource_type_id": "biz"},
          {"id": "proxy_operate", "resource_type_id": "biz"},
          {"id": "proxy_history_view", "resource_type_id": "biz"},
          {"id": "networkarea_view", "resource_type_id": "networkarea"},
          {"id": "networkunit_view", "resource_type_id": "networkunit"},
          {"id": "networkunit_use_for_proxy", "resource_type_id": "networkunit"},
          {"id": "plugin_view", "resource_type_id": "biz"},
          {"id": "plugin_operate", "resource_type_id": "biz"},
          {"id": "plugin_history_view", "resource_type_id": "biz"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "networkarea_manager",
        "name": "管控区域管理员",
        "actions": [
          {"id": "networkarea_view", "resource_type_id": "networkarea"},
          {"id": "networkarea_create", "resource_type_id": ""},
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
            "name": "管控单元管理员",
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
        "name": "配置策略管理员",
        "actions": [
          {"id": "biz_access", "resource_type_id": "biz"},
          {"id": "config_policy_view", "resource_type_id": "biz"},
          {"id": "config_policy_manage", "resource_type_id": "biz"},
          {"id": "config_policy_history_view", "resource_type_id": "biz"},
          {"id": "deploy_policy_view", "resource_type_id": "biz"},
          {"id": "deploy_policy_manage", "resource_type_id": "biz"},
          {"id": "deploy_policy_history_view", "resource_type_id": "biz"}
        ]
      }
    },
    {
      "operation": "upsert_role",
      "data": {
        "id": "package_manager",
        "name": "资源包管理员",
        "actions": [
          {"id": "package_type_upload", "resource_type_id": "package_type"},
          {"id": "package_view", "resource_type_id": "package"},
          {"id": "package_manage", "resource_type_id": "package"},
          {"id": "package_history_view", "resource_type_id": "package"}
        ]
      }
    }
  ]
}
