{
  "system_id": "{{ .system.id | default "bk_nodemgr" }}",
  "operations": [
    {
      "operation": "upsert_common_actions",
      "data": [
        {
          "name": "Agent 管理",
          "name_en": "Agent Management",
          "actions": [
            {
              "id": "agent_view"
            },
            {
              "id": "agent_operate"
            },
            {
              "id": "agent_history_view"
            }
          ]
        },
        {
          "name": "Proxy 管理",
          "name_en": "Proxy Management",
          "actions": [
            {
              "id": "proxy_view"
            },
            {
              "id": "proxy_operate"
            },
            {
              "id": "proxy_history_view"
            }
          ]
        },
        {
          "name": "插件管理",
          "name_en": "Plugin Management",
          "actions": [
            {
              "id": "plugin_view"
            },
            {
              "id": "plugin_operate"
            },
            {
              "id": "plugin_history_view"
            }
          ]
        },
        {
          "name": "配置策略管理",
          "name_en": "Config Policy Management",
          "actions": [
            {
              "id": "config_policy_view"
            },
            {
              "id": "config_policy_manage"
            },
            {
              "id": "config_policy_history_view"
            }
          ]
        },
        {
          "name": "部署策略管理",
          "name_en": "Deploy Policy Management",
          "actions": [
            {
              "id": "deploy_policy_view"
            },
            {
              "id": "deploy_policy_manage"
            },
            {
              "id": "deploy_policy_history_view"
            }
          ]
        },
        {
          "name": "节点完整操作权限",
          "name_en": "Full Node Operations",
          "actions": [
            {
              "id": "agent_view"
            },
            {
              "id": "agent_operate"
            },
            {
              "id": "agent_history_view"
            },
            {
              "id": "proxy_view"
            },
            {
              "id": "proxy_operate"
            },
            {
              "id": "proxy_history_view"
            },
            {
              "id": "plugin_view"
            },
            {
              "id": "plugin_operate"
            },
            {
              "id": "plugin_history_view"
            }
          ]
        },
        {
          "name": "策略完整管理权限",
          "name_en": "Full Policy Management",
          "actions": [
            {
              "id": "config_policy_view"
            },
            {
              "id": "config_policy_manage"
            },
            {
              "id": "config_policy_history_view"
            },
            {
              "id": "deploy_policy_view"
            },
            {
              "id": "deploy_policy_manage"
            },
            {
              "id": "deploy_policy_history_view"
            }
          ]
        }
      ]
    }
  ]
}
