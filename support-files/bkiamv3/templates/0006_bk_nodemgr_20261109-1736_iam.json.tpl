{
  "system_id": "{{ .system.id }}",
  "operations": [
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
            "id": "package_type",
            "selection_mode": "instance",
            "related_instance_selections": [
              {
                "system_id": "{{ .system.id }}",
                "id": "package_type_instance_selection"
              }
            ]
          }
        ]
      }
    }
  ]
}
