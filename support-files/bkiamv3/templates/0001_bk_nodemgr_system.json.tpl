{
  "system_id": "{{ .system.id }}",
  "enabled": true,
  "operations": [
    {
      "operation": "upsert_system",
      "data": {
        "id": "{{ .system.id }}",
        "name": "{{ .system.name }}",
        "name_en": "{{ .system.name_en }}",
        "description": "蓝鲸节点管理系统, 提供节点管理, 插件管理等功能",
        "description_en": "BlueKing Node Manager, provides node management, plugin management and other functions",
        "clients": "{{ .system.clients }}",
        "provider_config": {
          "host": "{{ .provider.host }}",
          "auth": "{{ .provider.auth }}"
        }
      }
    }
  ]
}