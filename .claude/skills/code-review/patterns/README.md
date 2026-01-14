# 模块特定代码审查规范

本目录包含针对不同模块的特定代码审查规范和模式。

## 已有模块规范

### Backend - Deploy Policy Manager
- **文件**: [dpmgr-executor.md](dpmgr-executor.md)
- **适用范围**: `internal/backend/dpmgr/executor.go`
- **关键模式**: Executor 函数对称模式（Install/Uninstall/Upgrade）

## 添加新模块规范

当需要为新模块添加特定审查规范时：

1. **创建模块文档**
   - 文件命名：`{模块名}-{组件名}.md`
   - 例如：`workflow-operation.md`、`plugin-manager.md`

2. **文档结构建议**
   ```markdown
   # {模块名} 审查规范
   
   ## 适用范围
   明确列出适用的文件或目录
   
   ## 代码模式
   描述该模块特有的代码模式
   
   ## 审查检查清单
   具体的检查项列表
   
   ## 常见问题
   该模块常见的问题和解决方案
   ```

3. **在主 SKILL.md 中引用**
   - 在"逻辑一致性检查"部分添加条件引用
   - 使用清晰的条件判断（如文件路径匹配）

## 模块组织建议

按功能域组织：
- `backend/` - 后端服务相关
- `application/` - 应用服务相关
- `file/` - 文件服务相关
- `workflow/` - 工作流相关
- `thirdparty/` - 第三方集成相关

## 示例模块规范模板

参考 `dpmgr-executor.md` 的结构：
1. 明确适用范围
2. 列出代码模式和对称函数
3. 提供检查清单表格
4. 包含常见问题和解决方案
5. 提供审查步骤指南
