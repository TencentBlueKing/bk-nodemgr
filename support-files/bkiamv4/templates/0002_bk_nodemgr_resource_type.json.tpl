{
  "system_id": {{ .system.id | toJson }},
  "operations": [
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "biz",
        "name": "业务",
        "ancestors": []
      }
    },
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "networkarea",
        "name": "管控区域",
        "ancestors": []
      }
    },
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "networkunit",
        "name": "管控单元",
        "ancestors": ["networkarea"]
      }
    },
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "package_type",
        "name": "资源包类型",
        "ancestors": []
      }
    },
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "package",
        "name": "资源包",
        "ancestors": ["package_type"]
      }
    }
  ]
}
