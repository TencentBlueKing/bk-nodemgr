{
  "system_id": "{{ .system.id }}",
  "operations": [
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "platform_resource",
        "name": "平台资源",
        "name_en": "Platform Resource",
        "description": "平台管理功能的固定枚举资源",
        "description_en": "Fixed enumerated resources for platform management capabilities",
        "parents": [],
        "provider_config": {
          "path": "{{ .provider.path }}"
        }
      }
    },
    {
      "operation": "upsert_instance_selection",
      "data": {
        "id": "platform_instance_selection",
        "name": "平台资源",
        "name_en": "Platform Resource",
        "resource_type_chain": [
          {
            "system_id": "{{ .system.id }}",
            "id": "platform_resource"
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "platform_manage",
        "name": "平台管理",
        "name_en": "Platform Management",
        "description": "管理所选平台资源对应的功能",
        "description_en": "Manage capabilities associated with the selected platform resources",
        "type": "manage",
        "related_actions": ["system_access"],
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "platform_resource",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "platform_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action_groups",
      "data": [
        {
          "name": "系统",
          "name_en": "System",
          "actions": [
            {
              "id": "system_access"
            }
          ]
        },
        {
          "name": "平台管理",
          "name_en": "Platform Management",
          "actions": [
            {
              "id": "platform_manage"
            }
          ]
        },
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
