{
  "system_id": {{ .system.id | toJson }},
  "operations": [
    {
      "operation": "upsert_action",
      "data": {
        "id": "biz_access",
        "name": "业务访问",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "agent_view",
        "name": "查看 Agent",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "agent_operate",
        "name": "操作 Agent",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "agent_history_view",
        "name": "查看 Agent 历史",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "proxy_view",
        "name": "查看 Proxy",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "proxy_operate",
        "name": "操作 Proxy",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "proxy_history_view",
        "name": "查看 Proxy 历史",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "plugin_view",
        "name": "查看插件",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "plugin_operate",
        "name": "操作插件",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "plugin_history_view",
        "name": "查看插件历史",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "config_policy_view",
        "name": "查看配置策略",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "config_policy_manage",
        "name": "管理配置策略",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "config_policy_history_view",
        "name": "查看配置策略历史",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "deploy_policy_view",
        "name": "查看部署策略",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "deploy_policy_manage",
        "name": "管理部署策略",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "deploy_policy_history_view",
        "name": "查看部署策略历史",
        "resource_type_id": "biz"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkarea_view",
        "name": "查看管控区域",
        "resource_type_id": "networkarea"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkarea_create",
        "name": "创建管控区域",
        "resource_type_id": ""
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkarea_edit",
        "name": "编辑管控区域",
        "resource_type_id": "networkarea"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkarea_delete",
        "name": "删除管控区域",
        "resource_type_id": "networkarea"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkarea_history_view",
        "name": "查看管控区域历史",
        "resource_type_id": "networkarea"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_view",
        "name": "查看管控单元",
        "resource_type_id": "networkunit"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_create",
        "name": "创建管控单元",
        "resource_type_id": "networkarea"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_edit",
        "name": "编辑管控单元",
        "resource_type_id": "networkunit"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_delete",
        "name": "删除管控单元",
        "resource_type_id": "networkunit"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_use_for_agent",
        "name": "使用管控单元部署 Agent",
        "resource_type_id": "networkunit"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_use_for_proxy",
        "name": "使用管控单元部署 Proxy",
        "resource_type_id": "networkunit"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "networkunit_history_view",
        "name": "查看管控单元历史",
        "resource_type_id": "networkunit"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "package_type_upload",
        "name": "上传资源包",
        "resource_type_id": "package_type"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "package_view",
        "name": "查看资源包",
        "resource_type_id": "package"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "package_manage",
        "name": "管理资源包",
        "resource_type_id": "package"
      }
    },
    {
      "operation": "upsert_action",
      "data": {
        "id": "package_history_view",
        "name": "查看资源包历史",
        "resource_type_id": "package"
      }
    }
  ]
}
