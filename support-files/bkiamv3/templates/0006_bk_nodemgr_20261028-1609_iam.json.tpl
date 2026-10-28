{
  "system_id": "{{ .system.id }}",
  "operations": [
    {
      "operation": "upsert_action",
      "data": {
        "id": "system_access",
        "name": "系统访问",
        "name_en": "System Access",
        "description": "访问节点管理系统，不替代具体功能和资源权限",
        "description_en": "Access Node Manager without replacing feature or resource permissions",
        "type": "view",
        "related_resource_types": []
      }
    },
    {
      "operation": "update_action",
      "data": {"id": "biz_access", "related_actions": ["system_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "agent_view", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "agent_operate", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "agent_history_view", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "proxy_view", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "proxy_operate", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "proxy_history_view", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "plugin_view", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "plugin_operate", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "plugin_history_view", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "config_policy_view", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "config_policy_manage", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "config_policy_history_view", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "deploy_policy_view", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "deploy_policy_manage", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "deploy_policy_history_view", "related_actions": ["system_access", "biz_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkarea_view", "related_actions": ["system_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkarea_create", "related_actions": ["system_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkarea_edit", "related_actions": ["system_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkarea_delete", "related_actions": ["system_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkarea_history_view", "related_actions": ["system_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkunit_view", "related_actions": ["system_access", "networkarea_view"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkunit_create", "related_actions": ["system_access", "networkarea_view"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkunit_edit", "related_actions": ["system_access", "networkarea_view"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkunit_delete", "related_actions": ["system_access", "networkarea_view"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkunit_use_for_agent", "related_actions": ["system_access", "networkarea_view"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkunit_use_for_proxy", "related_actions": ["system_access", "networkarea_view"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "networkunit_history_view", "related_actions": ["system_access", "networkarea_view"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "package_type_upload", "related_actions": ["system_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "package_view", "related_actions": ["system_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "package_manage", "related_actions": ["system_access"]}
    },
    {
      "operation": "update_action",
      "data": {"id": "package_history_view", "related_actions": ["system_access"]}
    },
    {
      "operation": "upsert_action_groups",
      "data": [
        {
          "name": "系统",
          "name_en": "System",
          "actions": [{"id": "system_access"}]
        },
        {
          "name": "业务",
          "name_en": "Business",
          "actions": [{"id": "biz_access"}],
          "sub_groups": [
            {
              "name": "Agent",
              "name_en": "Agent",
              "actions": [{"id": "agent_view"}, {"id": "agent_operate"}, {"id": "agent_history_view"}]
            },
            {
              "name": "Proxy",
              "name_en": "Proxy",
              "actions": [{"id": "proxy_view"}, {"id": "proxy_operate"}, {"id": "proxy_history_view"}]
            },
            {
              "name": "插件",
              "name_en": "Plugin",
              "actions": [{"id": "plugin_view"}, {"id": "plugin_operate"}, {"id": "plugin_history_view"}]
            },
            {
              "name": "配置策略",
              "name_en": "Config Policy",
              "actions": [{"id": "config_policy_view"}, {"id": "config_policy_manage"}, {"id": "config_policy_history_view"}]
            },
            {
              "name": "部署策略",
              "name_en": "Deploy Policy",
              "actions": [{"id": "deploy_policy_view"}, {"id": "deploy_policy_manage"}, {"id": "deploy_policy_history_view"}]
            }
          ]
        },
        {
          "name": "管控区域",
          "name_en": "Network Area",
          "actions": [
            {"id": "networkarea_view"},
            {"id": "networkarea_create"},
            {"id": "networkarea_edit"},
            {"id": "networkarea_delete"},
            {"id": "networkarea_history_view"}
          ],
          "sub_groups": [
            {
              "name": "管控单元",
              "name_en": "Network Unit",
              "actions": [
                {"id": "networkunit_view"},
                {"id": "networkunit_create"},
                {"id": "networkunit_edit"},
                {"id": "networkunit_delete"},
                {"id": "networkunit_history_view"},
                {"id": "networkunit_use_for_agent"},
                {"id": "networkunit_use_for_proxy"}
              ]
            }
          ]
        },
        {
          "name": "资源包",
          "name_en": "Package Management",
          "actions": [
            {"id": "package_type_upload"},
            {"id": "package_view"},
            {"id": "package_manage"},
            {"id": "package_history_view"}
          ]
        }
      ]
    },
    {
      "operation": "upsert_common_actions",
      "data": [
        {
          "name": "业务节点管理员",
          "name_en": "Agent Manager",
          "actions": [
            {"id": "system_access"},
            {"id": "biz_access"},
            {"id": "agent_view"},
            {"id": "agent_operate"},
            {"id": "agent_history_view"},
            {"id": "networkunit_use_for_agent"},
            {"id": "plugin_view"},
            {"id": "plugin_operate"},
            {"id": "plugin_history_view"}
          ]
        },
        {
          "name": "管控区域管理员",
          "name_en": "Network Area Manager",
          "actions": [
            {"id": "system_access"},
            {"id": "biz_access"},
            {"id": "networkarea_view"},
            {"id": "networkarea_create"},
            {"id": "networkarea_edit"},
            {"id": "networkarea_delete"},
            {"id": "networkarea_history_view"},
            {"id": "networkunit_view"},
            {"id": "networkunit_create"},
            {"id": "networkunit_edit"},
            {"id": "networkunit_delete"},
            {"id": "networkunit_history_view"},
            {"id": "proxy_view"},
            {"id": "proxy_operate"},
            {"id": "proxy_history_view"},
            {"id": "networkunit_use_for_proxy"}
          ]
        },
        {
          "name": "策略管理员",
          "name_en": "Policy Manager",
          "actions": [
            {"id": "system_access"},
            {"id": "biz_access"},
            {"id": "config_policy_view"},
            {"id": "config_policy_manage"},
            {"id": "config_policy_history_view"},
            {"id": "deploy_policy_view"},
            {"id": "deploy_policy_manage"},
            {"id": "deploy_policy_history_view"}
          ]
        },
        {
          "name": "资源包管理员",
          "name_en": "Package Manager",
          "actions": [
            {"id": "system_access"},
            {"id": "package_type_upload"},
            {"id": "package_view"},
            {"id": "package_manage"},
            {"id": "package_history_view"}
          ]
        }
      ]
    }
  ]
}
