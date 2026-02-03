{
  "system_id": "{{ .system.id | default "bk-nodemgr" }}",
  "enabled": true,
  "operations": [
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "networkarea",
        "name": "网络区域",
        "name_en": "Network Area",
        "description": "网络区域, 用于隔离不同的云区域或IP池",
        "description_en": "Network Area, used to isolate different cloud regions or IP pools",
        "parents": [],
        "provider_config": {
          "path": "/api/v3/iam/v3/resource"
        }
      }
    },
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "networkunit",
        "name": "网络单元",
        "name_en": "Network Unit",
        "description": "网络单元, 节点所属的基本单位, 共享同一组上游服务, 归属于网络区域",
        "description_en": "Network Unit, the basic unit of nodes, shares the same upstream services, belongs to Network Area",
        "parents": [
          {
            "system_id": "{{ .system.id | default "bk-nodemgr" }}",
            "id": "networkarea"
          }
        ],
        "provider_config": {
          "path": "/api/v3/iam/v3/resource"
        }
      }
    },
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "package",
        "name": "资源包",
        "name_en": "Package",
        "description": "资源包管理",
        "description_en": "Package Management",
        "parents": [],
        "provider_config": {
          "path": "/api/v3/iam/v3/resource"
        }
      }
    },
    {
      "operation": "upsert_instance_selection",
      "data": {
        "id": "networkarea_instance_selection",
        "name": "网络区域",
        "name_en": "Network Area",
        "resource_type_chain": [
          {
            "system_id": "{{ .system.id | default "bk-nodemgr" }}",
            "id": "networkarea"
          }
        ]
      }
    },
    {
      "operation": "upsert_instance_selection",
      "data": {
        "id": "networkunit_instance_selection",
        "name": "网络单元",
        "name_en": "Network Unit",
        "resource_type_chain": [
          {
            "system_id": "{{ .system.id | default "bk-nodemgr" }}",
            "id": "networkarea"
          },
          {
            "system_id": "{{ .system.id | default "bk-nodemgr" }}",
            "id": "networkunit"
          }
        ]
      }
    },
    {
      "operation": "upsert_instance_selection",
      "data": {
        "id": "package_instance_selection",
        "name": "资源包",
        "name_en": "Package",
        "resource_type_chain": [
          {
            "system_id": "{{ .system.id | default "bk-nodemgr" }}",
            "id": "package"
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "agent_view",
        "name": "查看 Agent",
        "name_en": "View Agent",
        "description": "查看业务下的 Agent 信息",
        "description_en": "View Agent information under business",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "agent_operate",
        "name": "操作 Agent",
        "name_en": "Operate Agent",
        "description": "安装, 卸载, 升级, 重启等 Agent 操作",
        "description_en": "Install, uninstall, upgrade, restart and other Agent operations",
        "type": "manage",
        "related_actions": [
          "agent_view"
        ],
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "agent_history_view",
        "name": "查看 Agent 历史",
        "name_en": "View Agent History",
        "description": "查看 Agent 操作历史记录",
        "description_en": "View Agent operation history",
        "type": "view",
        "related_actions": [
          "agent_view"
        ],
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "proxy_view",
        "name": "查看 Proxy",
        "name_en": "View Proxy",
        "description": "查看业务下的 Proxy 信息",
        "description_en": "View Proxy information under business",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "proxy_operate",
        "name": "操作 Proxy",
        "name_en": "Operate Proxy",
        "description": "安装, 卸载, 重启等 Proxy 操作",
        "description_en": "Install, uninstall, restart and other Proxy operations",
        "type": "manage",
        "related_actions": [
          "proxy_view"
        ],
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "proxy_history_view",
        "name": "查看 Proxy 历史",
        "name_en": "View Proxy History",
        "description": "查看 Proxy 操作历史记录",
        "description_en": "View Proxy operation history",
        "type": "view",
        "related_actions": [
          "proxy_view"
        ],
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "plugin_view",
        "name": "查看插件",
        "name_en": "View Plugin",
        "description": "查看业务下的插件信息",
        "description_en": "View Plugin information under business",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "plugin_operate",
        "name": "操作插件",
        "name_en": "Operate Plugin",
        "description": "安装, 卸载, 更新等插件操作",
        "description_en": "Install, uninstall, update and other Plugin operations",
        "type": "manage",
        "related_actions": [
          "plugin_view"
        ],
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "plugin_history_view",
        "name": "查看插件历史",
        "name_en": "View Plugin History",
        "description": "查看插件操作历史记录",
        "description_en": "View Plugin operation history",
        "type": "view",
        "related_actions": [
          "plugin_view"
        ],
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "config_policy_view",
        "name": "查看配置策略",
        "name_en": "View Config Policy",
        "description": "查看业务下的配置策略",
        "description_en": "View config policy under business",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "config_policy_manage",
        "name": "管理配置策略",
        "name_en": "Manage Config Policy",
        "description": "创建, 编辑, 删除配置策略",
        "description_en": "Create, edit, delete config policy",
        "type": "manage",
        "related_actions": [
          "config_policy_view"
        ],
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "config_policy_history_view",
        "name": "查看配置策略历史",
        "name_en": "View Config Policy History",
        "description": "查看配置策略操作历史",
        "description_en": "View config policy operation history",
        "type": "view",
        "related_actions": [
          "config_policy_view"
        ],
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "deploy_policy_view",
        "name": "查看部署策略",
        "name_en": "View Deploy Policy",
        "description": "查看业务下的部署策略",
        "description_en": "View deploy policy under business",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "deploy_policy_manage",
        "name": "管理部署策略",
        "name_en": "Manage Deploy Policy",
        "description": "创建, 编辑, 删除部署策略",
        "description_en": "Create, edit, delete deploy policy",
        "type": "manage",
        "related_actions": [
          "deploy_policy_view"
        ],
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "deploy_policy_history_view",
        "name": "查看部署策略历史",
        "name_en": "View Deploy Policy History",
        "description": "查看部署策略操作历史",
        "description_en": "View deploy policy operation history",
        "type": "view",
        "related_actions": [
          "deploy_policy_view"
        ],
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
            "id": "biz",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id | default "bk-cmdb" }}",
                "id": "business"
              }
            ]
          }
        ]
      }
    }
  ]
}
