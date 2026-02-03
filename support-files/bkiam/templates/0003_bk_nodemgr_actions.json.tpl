{
  "system_id": "{{ .system.id }}",
  "operations": [
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
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
        "description": "安装, 卸载, 升级, 重启等 Proxy 操作",
        "description_en": "Install, uninstall, upgrade, restart and other Proxy operations",
        "type": "manage",
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
        "description": "安装, 卸载, 升级等插件操作",
        "description_en": "Install, uninstall, upgrade and other Plugin operations",
        "type": "manage",
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
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
        "related_resource_types": [
          {
            "system_id": "{{ .cmdb.system_id }}",
            "id": "{{ .cmdb.resource.biz.resource_type_id }}",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .cmdb.system_id }}",
                "id": "{{ .cmdb.resource.biz.instance_selection_id }}"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkarea_view",
        "name": "查看管控区域",
        "name_en": "View Network Area",
        "description": "查看管控区域信息",
        "description_en": "View network area information",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkarea",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkarea_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkarea_create",
        "name": "创建管控区域",
        "name_en": "Create Network Area",
        "description": "创建新的管控区域",
        "description_en": "Create new network area",
        "type": "create",
        "related_resource_types": []
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkarea_edit",
        "name": "编辑管控区域",
        "name_en": "Edit Network Area",
        "description": "编辑管控区域信息",
        "description_en": "Edit network area information",
        "type": "edit",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkarea",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkarea_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkarea_delete",
        "name": "删除管控区域",
        "name_en": "Delete Network Area",
        "description": "删除管控区域",
        "description_en": "Delete network area",
        "type": "delete",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkarea",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkarea_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkarea_history_view",
        "name": "查看管控区域历史",
        "name_en": "View Network Area History",
        "description": "查看管控区域操作历史",
        "description_en": "View network area operation history",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkarea",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkarea_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_view",
        "name": "查看管控单元",
        "name_en": "View Network Unit",
        "description": "查看管控单元信息",
        "description_en": "View network unit information",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkunit",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkunit_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_create",
        "name": "创建管控单元",
        "name_en": "Create Network Unit",
        "description": "在管控区域下创建管控单元",
        "description_en": "Create network unit under network area",
        "type": "create",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkarea",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkarea_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_edit",
        "name": "编辑管控单元",
        "name_en": "Edit Network Unit",
        "description": "编辑管控单元信息",
        "description_en": "Edit network unit information",
        "type": "edit",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkunit",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkunit_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_delete",
        "name": "删除管控单元",
        "name_en": "Delete Network Unit",
        "description": "删除管控单元",
        "description_en": "Delete network unit",
        "type": "delete",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkunit",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkunit_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_use_for_agent",
        "name": "使用管控单元部署 Agent",
        "name_en": "Use Network Unit for Agent",
        "description": "使用管控单元部署 Agent",
        "description_en": "Use network unit to deploy Agent",
        "type": "use",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkunit",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkunit_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_use_for_proxy",
        "name": "使用管控单元部署 Proxy",
        "name_en": "Use Network Unit for Proxy",
        "description": "使用管控单元部署 Proxy",
        "description_en": "Use network unit to deploy Proxy",
        "type": "use",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkunit",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkunit_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_history_view",
        "name": "查看管控单元历史",
        "name_en": "View Network Unit History",
        "description": "查看管控单元操作历史",
        "description_en": "View network unit operation history",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkunit",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "networkunit_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "package_view",
        "name": "查看资源包",
        "name_en": "View Package",
        "description": "查看资源包信息",
        "description_en": "View package information",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "package",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "package_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "package_manage",
        "name": "管理资源包",
        "name_en": "Manage Package",
        "description": "管理资源包（上传, 发布, 删除等）",
        "description_en": "Manage package (upload, publish, delete, etc.)",
        "type": "manage",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "package",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "package_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "package_history_view",
        "name": "查看资源包历史",
        "name_en": "View Package History",
        "description": "查看资源包操作历史",
        "description_en": "View package operation history",
        "type": "view",
        "related_resource_types": [
          {
            "system_id": "{{ .system.id }}",
            "id": "package",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "package_instance_selection"
              }
            ]
          }
        ]
      }
    },
    {
      "operation": "upsert_resource_creator_actions",
      "data": {
        "config": [
          {
            "id": "networkarea",
            "actions": [
              {
                "id": "networkarea_edit",
                "required": false
              },
              {
                "id": "networkarea_view",
                "required": false
              },
              {
                "id": "networkarea_delete",
                "required": false
              }
            ]
          },
          {
            "id": "networkunit",
            "actions": [
              {
                "id": "networkunit_edit",
                "required": false
              },
              {
                "id": "networkunit_view",
                "required": false
              },
              {
                "id": "networkunit_delete",
                "required": false
              }
            ]
          }
        ]
      }
    }
  ]
}