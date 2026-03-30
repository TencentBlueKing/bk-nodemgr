{
  "system_id": "{{ .system.id }}",
  "operations": [
    {
      "operation": "upsert_common_actions",
      "data": [
        {
          "name": "业务节点管理员",
          "name_en": "Agent Manager",
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
              "id": "networkunit_use_for_agent"
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
          "name": "管控区域管理员",
          "name_en": "Network Area Manager",
          "actions": [
            {
              "id": "networkarea_view"
            },
            {
              "id": "networkarea_create"
            },
            {
              "id": "networkarea_edit"
            },
            {
              "id": "networkarea_delete"
            },
            {
              "id": "networkarea_history_view"
            },
            {
              "id": "networkunit_view"
            },
            {
              "id": "networkunit_create"
            },
            {
              "id": "networkunit_edit"
            },
            {
              "id": "networkunit_delete"
            },
            {
              "id": "networkunit_history_view"
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
              "id": "networkunit_use_for_proxy"
            }
          ]
        },
        {
          "name": "策略管理员",
          "name_en": "Policy Manager",
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
            },
            {
              "id": "package_view"
            },
            {
              "id": "package_manage"
            },
            {
              "id": "package_history_view"
            }
          ]
        },
        {
          "name": "资源包管理员",
          "name_en": "Package Manager",
          "actions": [
            {
              "id": "package_type_upload"
            },
            {
              "id": "package_view"
            },
            {
              "id": "package_manage"
            },
            {
              "id": "package_history_view"
            }
          ]
        }
      ]
    }
  ]
}