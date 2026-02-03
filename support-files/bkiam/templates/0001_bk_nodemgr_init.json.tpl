{
    "system_id": "{{ .system.id | default "bk-nodemgr" }}",
    "operations": [
        {
            "operation": "upsert_system",
            "data": {
                "id": "{{ .system.id | default "bk-nodemgr" }}",
                "name": "{{ .system.name | default "节点管理" }}",
                "name_en": "{{ .system.name_en | default "BlueKing Node Manager" }}",
                "description": "{{ .system.description | default "蓝鲸节点管理，用于管理和监控节点" }}",
                "description_en": "{{ .system.description_en | default "BlueKing Node Manager for managing and monitoring nodes" }}",
                "clients": "{{ .system.clients | default "bk-nodemgr" }}",
                "provider_config": {
                    "host": "{{ .provider.host | default "" }}",
                    "auth": "{{ .provider.auth | default "basic" }}"
                }
            }
        }
    ]
}
