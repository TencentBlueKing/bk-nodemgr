{
  "system_id": "{{ .system.id  }}",
  "operations": [
    {
      "operation": "upsert_action_groups",
      "data": [
        {
          "name": "业务",
          "name_en": "Business",
          "actions": [
            {
              "id": "biz_access"
            }
          ],
          "sub_groups": [
            {
              "name": "Agent",
              "name_en": "Agent",
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
              "name": "Proxy",
              "name_en": "Proxy",
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
              "name": "插件",
              "name_en": "Plugin",
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
              "name": "配置策略",
              "name_en": "Config Policy",
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
              "name": "部署策略",
              "name_en": "Deploy Policy",
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
            }
          ]
        },
        {
          "name": "管控区域",
          "name_en": "Network Area",
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
            }
          ],
          "sub_groups": [
            {
              "name": "管控单元",
              "name_en": "Network Unit",
              "actions": [
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
                  "id": "networkunit_use_for_agent"
                },
                {
                  "id": "networkunit_use_for_proxy"
                }
              ]
            }
          ]
        },
        {
          "name": "资源包",
          "name_en": "Package Management",
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
