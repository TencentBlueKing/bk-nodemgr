{
  "system_id": "{{ .system.id  }}",
  "enabled": true,
  "operations": [
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "networkarea",
        "name": "管控区域",
        "name_en": "Network Area",
        "description": "管控区域, 用于隔离不同的云区域或IP池",
        "description_en": "Network Area, used to isolate different cloud regions or IP pools",
        "parents": [],
        "provider_config": {
          "path": "{{ .provider.path }}"
        }
      }
    },
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "networkunit",
        "name": "管控单元",
        "name_en": "Network Unit",
        "description": "管控单元, 节点所属的基本单位, 共享同一组上游服务, 归属于管控区域",
        "description_en": "Network Unit, the basic unit of nodes, shares the same upstream services, belongs to Network Area",
        "parents": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkarea"
          }
        ],
        "provider_config": {
          "path": "{{ .provider.path }}"
        }
      }
    },
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "package_type",
        "name": "资源包类型",
        "name_en": "Package Type",
        "description": "资源包类型",
        "description_en": "Package Type",
        "parents": [],
        "provider_config": {
          "path": "{{ .provider.path }}"
        }
      }
    },
    {
      "operation": "upsert_resource_type",
      "data": {
        "id": "package",
        "name": "资源包",
        "name_en": "Package",
        "description": "资源包",
        "description_en": "Package",
        "parents": [
          {
            "system_id": "{{ .system.id }}",
            "id": "package_type"
          }
        ],
        "provider_config": {
          "path": "{{ .provider.path }}"
        }
      }
    },
    {
      "operation": "upsert_instance_selection",
      "data": {
        "id": "networkarea_instance_selection",
        "name": "管控区域",
        "name_en": "Network Area",
        "resource_type_chain": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkarea"
          }
        ]
      }
    },
    {
      "operation": "upsert_instance_selection",
      "data": {
        "id": "networkunit_instance_selection",
        "name": "管控单元",
        "name_en": "Network Unit",
        "resource_type_chain": [
          {
            "system_id": "{{ .system.id }}",
            "id": "networkarea"
          },
          {
            "system_id": "{{ .system.id }}",
            "id": "networkunit"
          }
        ]
      }
    },
    {
      "operation": "upsert_instance_selection",
      "data": {
        "id": "package_type_instance_selection",
        "name": "资源包类型",
        "name_en": "Package Type",
        "resource_type_chain": [
          {
            "system_id": "{{ .system.id }}",
            "id": "package_type"
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
            "system_id": "{{ .system.id }}",
            "id": "package_type"
          },
          {
            "system_id": "{{ .system.id }}",
            "id": "package"
          }
        ]
      }
    }
  ]
}