{
  "system_id": {{ .system.id | toJson }},
  "operations": [
    {
      "operation": "upsert_system",
      "data": {
        "id": {{ .system.id | toJson }},
        "name": {{ .system.name | toJson }},
        "description": {{ .system.description | toJson }},
        "clients": {{ .system.clients | toJson }},
        "callback_url": {{ .system.callback_url | toJson }}
      }
    }
  ]
}
