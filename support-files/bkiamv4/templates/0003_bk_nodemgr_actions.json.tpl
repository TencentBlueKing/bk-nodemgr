{
  "system_id": {{ .system.id | toJson }},
  "operations": [
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
