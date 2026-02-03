# IAM Migrator

IAM Migrator 是一个用于管理蓝鲸 IAM 权限模型的 Python 脚本。通过 API Gateway 直接注册权限模型（系统、资源类型、操作等）到 IAM 系统。

该脚本来自蓝鲸官方 [iam-python-sdk](https://github.com/TencentBlueKing/iam-python-sdk/blob/master/iam/contrib/iam_migration/utils/do_migrate.py)。

## 功能特性

- **权限模型迁移**: 将 JSON 格式的迁移文件应用到 IAM 系统
- **Upsert 语义**: 自动判断新增或更新
- **完整操作支持**: system、resource_type、action、instance_selection、action_groups 等

## 安装

```bash
cd script_tools/bkiam
pip install -r requirements.txt
```

## 使用方法

```bash
python do_migrate.py \
  -t "https://bkapi.example.com/api/bk-iam/prod/" \
  -f "./0001_init.json" \
  -a "your-app-code" \
  -s "your-app-secret"
```

### 参数说明

| 参数 | 说明 | 必需 |
|------|------|------|
| `-t` | IAM API Gateway URL | 是 |
| `-f` | 迁移文件路径 | 是 |
| `-a` | 应用代码 (app_code) | 是 |
| `-s` | 应用密钥 (app_secret) | 是 |
| `--bk_tenant_id` | 蓝鲸租户 ID（多租户场景） | 否 |

## 迁移文件格式

迁移文件使用 JSON 格式：

```json
{
  "system_id": "bk-nodemgr",
  "operations": [
    {
      "operation": "upsert_system",
      "data": {
        "id": "bk-nodemgr",
        "name": "蓝鲸节点管理",
        "name_en": "BlueKing Node Manager",
        "description": "蓝鲸节点管理系统",
        "description_en": "BlueKing Node Manager"
      }
    }
  ]
}
```

### 支持的操作类型

| 操作类型 | 说明 |
|---------|------|
| `upsert_system` | 创建或更新系统 |
| `add_system` / `update_system` | 添加/更新系统 |
| `upsert_resource_type` | 创建或更新资源类型 |
| `add_resource_type` / `update_resource_type` / `delete_resource_type` | 资源类型操作 |
| `upsert_instance_selection` | 创建或更新实例选择 |
| `add_instance_selection` / `update_instance_selection` / `delete_instance_selection` | 实例选择操作 |
| `upsert_action` | 创建或更新操作 |
| `add_action` / `update_action` / `delete_action` | 操作管理 |
| `upsert_action_groups` | 创建或更新操作组 |
| `upsert_resource_creator_actions` | 创建或更新资源创建者操作 |
| `upsert_common_actions` | 创建或更新通用操作 |
| `upsert_feature_shield_rules` | 创建或更新功能屏蔽规则 |

## 退出状态码

| 状态码 | 说明 |
|-------|------|
| 0 | 成功 |
| 1 | 失败（服务不可用、配置错误、迁移失败） |

## 故障排查

### 服务不可用

```
iam service is not available: https://...
```

检查：
1. 确认 API Gateway URL 正确
2. 确认网络连接正常
3. 确认 IAM 服务正在运行

### 认证失败

```
_call_iam_api fail. error: ...
```

检查：
1. 确认 `app_code` 和 `app_secret` 正确
2. 确认应用已在 IAM 系统注册

### 迁移文件解析失败

```
parser json data file error: ...
```

检查：
1. 确认 JSON 格式正确
2. 确认文件路径正确

## 许可证

本项目遵循 BlueKing 开源许可协议。
