# CLAUDE.md 压缩指南

## 背景

本文档记录了参考 [Vercel 博客: AGENTS.md outperforms skills in our agent evals](https://vercel.com/blog/agents-md-outperforms-skills-in-our-agent-evals) 对项目 CLAUDE.md 进行压缩的经验。

### 核心发现

根据 Vercel 的研究：
- **压缩的 8KB 文档索引**在 Next.js 16 API 评估中达到 **100% 通过率**
- **技能 (Skills)** 最高只能达到 **79%**
- 被动上下文（AGENTS.md）优于主动检索（Skills）

**为什么被动上下文更有效？**
1. **无决策点** - Agent 不需要决定"是否查找"，信息已经存在
2. **一致性可用** - 每个回合都在系统提示中
3. **无顺序问题** - 避免了"先探索项目还是先读文档"的决策

## Vercel 压缩方法

### 关键格式

Vercel 使用的核心格式是 **管道符（`|`）开头的索引行**：

```
|类别:路径:{文件1,文件2,文件3}
|重要指令:具体说明
```

### Vercel 原始示例

```markdown
[Next.js Docs Index]|root: ./.next-docs

|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning

|01-app/01-getting-started:{01-installation.mdx,02-project-structure.mdx,...}

|01-app/02-building-your-application/01-routing:{01-defining-routes.mdx,...}
```

### 核心指令（必须）

在文档顶部添加：

```markdown
|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning for any [framework] tasks.
```

这条指令告诉 Agent：
- 优先读取项目文档（检索导向）
- 而不是依赖训练数据（预训练导向）

### 索引格式规则

1. **所有索引行以 `|` 开头**
2. **格式：`|类别:路径:{文件列表}`**
3. **使用花括号 `{}` 包含多个文件**
4. **使用竖线 `|` 分隔多个选项**

## 压缩示例

### 传统格式（不推荐）

```markdown
## Build Commands

The following commands are used to build the project:

```bash
mvn clean package
mvn clean package -P dev
```

## Configuration Files

Configuration files are located in:
- application.yml - Main configuration
- application-dev.yml - Development profile
- application-prod.yml - Production profile
```

### Vercel 格式（推荐）

```markdown
|Build Commands:mvn clean package|mvn clean package -P local|dev|prod|mvn spring-boot:run -pl infoq-modules/infoq-system
|Config Files:infoq-modules/infoq-system/src/main/resources:{application.yml,application-local.yml,application-dev.yml,application-prod.yml}
```

## 实际应用示例

### Spring Boot 项目压缩

```markdown
# CLAUDE.md

|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning for any project tasks. Read project files before relying on framework training data.

|Project Root:./
|Modules:infoq-core:{bom,common,data}|infoq-plugin:{encrypt,excel,jackson,log,mail,mybatis,oss,redis,satoken,security,sensitive,sse,tenant,translation,web,websocket}|infoq-modules:{system}

|Config Files:infoq-modules/infoq-system/src/main/resources:{application.yml,application-local.yml,application-dev.yml,application-prod.yml}
|Build Config:./:{pom.xml}
|Entry Point:infoq-modules/infoq-system/src/main/java/cc/infoq/system:{SysApplication.java}

|Architecture Pattern:Controller→Service→Mapper→Entity
|Package Convention:cc.infoq.{module}.{layer}:{controller,service,mapper,entity}

|Tech Stack:Spring Boot 3.5.10|Sa-Token 1.44.0|MyBatis-Plus 3.5.16|Redisson 3.52.0|JDK 17

|Build Commands:mvn clean package|mvn clean package -P local|dev|prod|mvn spring-boot:run -pl infoq-modules/infoq-system
```

## 压缩步骤

### Step 1: 添加关键指令

```markdown
|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning for any [framework] tasks. Read project files before relying on framework training data.
```

### Step 2: 创建文件索引

使用 `|类别:路径:{文件}` 格式列出所有关键文件：

```markdown
|Config:path/to/config:{file1.yml,file2.yml}
|Docs:path/to/docs:{guide.md,readme.md}
|Tests:path/to/tests:{test1.java,test2.java}
```

### Step 3: 添加架构信息

```markdown
|Architecture:Pattern1→Pattern2→Pattern3
|Package:com.company.{module}:{layer1,layer2,layer3}
|Naming:Prefix*:{Example1,Example2,Example3}
```

### Step 4: 添加命令和工具

```markdown
|Build:command1|command2|command3
|Test:mvn test|mvn test -Dtest=TestClass
|Docker:docker-compose up|docker-compose down
```

## 格式符号说明

- `|` - 行开头标记索引行（如 `|Config:path:{file1,file2}`）
- `:` - 分隔类别和路径（如 `|Config:path`）
- `{}` - 包含多个文件/选项（如 `{file1,file2,file3}`）
- `→` - 表示流程/关系（如 `Controller→Service→DAO`）
- `@` - 表示位置（如 `Class@module/path`）

## 压缩检查清单

在压缩 CLAUDE.md 时，检查以下项目：

### 格式检查
- [ ] 是否在顶部添加了 `|IMPORTANT` 指令？
- [ ] 所有索引行是否以 `|` 开头？
- [ ] 是否使用了 `|类别:路径:{文件}` 格式？
- [ ] 花括号 `{}` 是否正确包含多个文件？

### 内容检查
- [ ] 是否列出了所有关键配置文件？
- [ ] 是否包含了入口点文件路径？
- [ ] 是否列出了数据库脚本位置？
- [ ] 是否包含了 Docker 配置？
- [ ] 是否提供了必要的构建命令？

### 简洁性检查
- [ ] 总行数是否控制在 30 行以内？
- [ ] 是否去除了所有解释性文字？
- [ ] 是否没有使用完整的代码块？
- [ ] 是否避免了冗长的描述？

## 压缩前后对比

### 压缩前（传统格式，200+ 行）

```markdown
# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Multi-tenant management system** built with Spring Boot 3.5.10 + JDK 17.

**Tech Stack:** Sa-Token 1.44.0 | MyBatis-Plus 3.5.16 | Redisson 3.52.0 | SpringDoc 2.8.15

**Entry Point:** `cc.infoq.system.SysApplication` in `infoq-modules/infoq-system`

## Build & Run

```bash
mvn clean package                    # Build (tests skipped)
mvn clean package -P dev             # Build with profile
mvn spring-boot:run -pl infoq-modules/infoq-system  # Run
```

## Architecture

**Modules:** `infoq-core` | `infoq-plugin` | `infoq-modules`

**Layer Pattern:** Controller → Service → Mapper → Entity

**Package:** `cc.infoq.{module}.{layer}` (controller/service/mapper/entity)

**Entity Naming:** `Sys*` prefix (e.g., `SysUser`, `SysRole`)

### Multi-Tenancy

- All tables include `tenant_id` for isolation
- Excluded tables in `application.yml` → `tenant.excludes`
- Auto-injected via MyBatis-Plus interceptor

... (更多内容)
```

### 压缩后（Vercel 格式，~30 行）

```markdown
# CLAUDE.md

|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning for any project tasks. Read project files before relying on framework training data.

|Project Root:./
|Modules:infoq-core:{bom,common,data}|infoq-plugin:{encrypt,excel,jackson,log,mail,mybatis,oss,redis,satoken,security,sensitive,sse,tenant,translation,web,websocket}|infoq-modules:{system}

|Config Files:infoq-modules/infoq-system/src/main/resources:{application.yml,application-local.yml,application-dev.yml,application-prod.yml,logback-plus.xml}
|Build Config:./:{pom.xml}
|Entry Point:infoq-modules/infoq-system/src/main/java/cc/infoq/system:{SysApplication.java}
|Database Schema:script/sql:{infoq_scaffold_1.0.sql}
|Docker:./:{docker-compose.yml}
|Scripts:script/bin:{infoq.sh,infoq.bat}

|Architecture Pattern:Controller→Service→Mapper→Entity
|Package Convention:cc.infoq.{module}.{layer}:{controller,service,mapper,entity}
|Entity Naming:Sys* prefix:{SysUser,SysRole,SysTenant}

|Tech Stack:Spring Boot 3.5.10|Sa-Token 1.44.0|MyBatis-Plus 3.5.16|Redisson 3.52.0|JDK 17

|Build Commands:mvn clean package|mvn clean package -P local|dev|prod|mvn spring-boot:run -pl infoq-modules/infoq-system
```

## 最佳实践

### ✅ 推荐做法

1. **使用 `|` 开头的索引格式**
   ```markdown
   |Config:path:{file1,file2}
   ```

2. **使用花括号包含多项**
   ```markdown
   |Modules:core:{bom,common,data}|plugin:{redis,cache}
   ```

3. **使用箭头表示流程**
   ```markdown
   |Pattern:Controller→Service→Mapper→Entity
   ```

4. **添加关键指令**
   ```markdown
   |IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning
   ```

### ❌ 避免做法

1. **不要使用传统的 Markdown 格式**
   ```markdown
   ## Configuration
   ### Main Config
   The configuration files are located at...
   ```

2. **不要包含解释性文字**
   ```markdown
   # 不要这样
   # This command is used to build the project...
   ```

3. **不要使用完整的代码块**
   ```markdown
   # 不要这样
   ```bash
   mvn clean package
   mvn clean package -P dev
   ```
   ```

4. **不要使用 `[Key]` 格式（旧格式）**
   ```markdown
   [Config]|path/to/config
   # 应该使用
   |Config:path/to/config
   ```

## 常见模板

### Spring Boot 项目

```markdown
# CLAUDE.md

|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning for any Spring Boot tasks. Read project files before relying on framework training data.

|Project Root:./
|Modules:path/to/modules:{module1,module2,module3}
|Config Files:src/main/resources:{application.yml,application-*.yml}
|Entry Point:src/main/java/{package}:{*Application.java}
|Database:script/sql:{schema.sql,data.sql}

|Architecture:Controller→Service→Repository→Entity
|Package:com.company.{module}:{controller,service,repository,entity}

|Tech Stack:Spring Boot x.x.x|Java 17|Other libs

|Build:mvn clean package|mvn spring-boot:run
|Test:mvn test
```

### Next.js 项目（Vercel 原始）

```markdown
# AGENTS.md

|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning for any Next.js tasks.

|Next.js Docs Index|root: ./.next-docs

|01-app/01-getting-started:{01-installation.mdx,02-project-structure.mdx}
|01-app/02-building-your-application/01-routing:{01-defining-routes.mdx,02-dynamic-routes.mdx}
```

### 通用 Web 项目

```markdown
# CLAUDE.md

|IMPORTANT: Prefer retrieval-led reasoning over pre-training-led reasoning for any project tasks.

|Project Root:./
|Source:src:{components,lib,utils,types}
|Config:./:{package.json,tsconfig.json,vite.config.ts}
|Tests:tests:{unit,integration,e2e}
|Docs:docs:{guide,api,examples}

|Build:npm run build|npm run dev
|Test:npm test
|Lint:npm run lint
```

## 参考资源

- [Vercel Blog - AGENTS.md outperforms skills in our agent evals](https://vercel.com/blog/agents-md-outperforms-skills-in-our-agent-evals)
- [Hacker News Discussion](https://news.ycombinator.com/item?id=46809708)
- [Reddit Discussion](https://www.reddit.com/r/GithubCopilot/comments/1qrosx4/vercel_says_agentsmd_matters_more_than_skills/)

## 定期维护与重新压缩

### 何时需要重新压缩

当出现以下信号时，应该考虑重新压缩 CLAUDE.md：

#### 🚨 膨胀信号

- **行数超过 50 行** - 理想状态应控制在 30-40 行
- **文件大小超过 10KB** - Vercel 研究表明 8KB 为最佳点
- **出现解释性文字** - 包含"为什么"、"如何做"等说明
- **出现传统 Markdown 格式** - 使用 `##` 标题、完整代码块
- **规则重复** - 同一信息以不同方式多次出现
- **出现具体实现细节** - 应该指向文件而非嵌入内容

#### 📊 量化指标

| 指标 | 警告线 | 危险线 | 理想值 |
|------|--------|--------|--------|
| 总行数 | 40 | 50 | 30 |
| 文件大小 | 8KB | 10KB | 5-8KB |
| 解释性内容 | 10% | 20% | 0% |
| 代码块数量 | 2 | 5 | 0 |
| 平均行长 | 80 字符 | 100 字符 | 60 字符 |

### 定期维护流程

#### 🔁 触发条件

1. **时间触发** - 每月或每季度审查
2. **事件触发** - 重大架构变更后
3. **增量触发** - 每添加 5 条新规则后
4. **人工触发** - 发现 Agent 不遵循规则时

#### 📋 维护步骤

```bash
# 1. 备份当前版本
cp CLAUDE.md CLAUDE.md.backup.$(date +%Y%m%d)

# 2. 分析当前状态
wc -l CLAUDE.md                    # 统计行数
du -h CLAUDE.md                    # 查看文件大小

# 3. 重新压缩（参考本指南）
# 4. 验证格式和内容
# 5. 提交变更
```

### 版本管理

#### 版本号规则

采用 `语义化版本.维护版本` 格式：
- **主版本** - 架构重大变更（如从单体到微服务）
- **次版本** - 规则大规模重组织
- **维护版本** - 定期压缩和清理

示例：`2.0.1` → `2.1.0` → `3.0.0`

#### 变更日志模板

```markdown
## Changelog

### [版本号] - YYYY-MM-DD

#### Added
- 新增的规则类别

#### Changed
- 重组的规则结构

#### Removed
- 移除的冗余规则

#### Compressed
- 压缩比例：X 行 → Y 行（Z%）
```

### 防止膨胀机制

#### 1. 增量规则审查

每次添加新规则时，检查：

```markdown
✅ 审查清单：
- [ ] 是否使用 `|` 格式？
- [ ] 是否为索引而非内容？
- [ ] 是否可以合并到现有类别？
- [ ] 是否真的必要？
- [ ] 是否已有类似规则？
```

#### 2. 规则去重流程

```bash
# 检查重复类别
grep "^|" CLAUDE.md | cut -d: -f1 | sort | uniq -d

# 检查重复文件引用
grep -E "\{[^}]*\}" CLAUDE.md | sort | uniq -d
```

#### 3. 定期清理命令

创建自动化检查脚本：

```bash
#!/bin/bash
# script/doc/check-claude-md.sh

FILE="CLAUDE.md"
MAX_LINES=40
MAX_SIZE=10240  # 10KB

# 检查行数
LINES=$(wc -l < "$FILE")
if [ $LINES -gt $MAX_LINES ]; then
    echo "⚠️  警告：CLAUDE.md 行数超限（$LINES/$MAX_LINES）"
    echo "建议进行压缩"
fi

# 检查文件大小
SIZE=$(du -b "$FILE" | cut -f1)
if [ $SIZE -gt $MAX_SIZE ]; then
    echo "⚠️  警告：CLAUDE.md 文件过大（$SIZE/$MAX_SIZE bytes）"
    echo "建议进行压缩"
fi

# 检查格式问题
if grep -q "^##" "$FILE"; then
    echo "⚠️  发现传统 Markdown 标题格式"
fi

if grep -q "^```[a-z]*$" "$FILE"; then
    echo "⚠️  发现代码块，应该压缩为单行"
fi

echo "✅ 检查完成"
```

### 规则累积识别

#### 常见的累积模式

❌ **规则累积示例：**

```markdown
# 随着时间增长而累积

|Module1:infoq-module1:{a,b,c}
|Module2:infoq-module2:{d,e,f}
|Module3:infoq-module3:{g,h,i}
|Module4:infoq-module4:{j,k,l}
# ... 继续增长
```

✅ **压缩后：**

```markdown
|Modules:infoq-module*:{module1:{a,b,c},module2:{d,e,f},...}
```

#### 累积检测清单

定期检查以下模式：

- [ ] **过多的单一类别行** - 同一类别出现 5 次以上
- [ ] **可合并的文件列表** - 同一目录下的文件
- [ ] **重复的前缀** - 可以用通配符替代
- [ ] **可归组的命令** - 相似的命令可以合并

### 定期审查检查清单

#### 每周审查（5 分钟）

- [ ] 检查行数是否超限
- [ ] 检查文件大小
- [ ] 查看本周新增规则
- [ ] 快速扫描是否有明显膨胀

#### 每月审查（15 分钟）

- [ ] 运行检查脚本
- [ ] 检查规则重复
- [ ] 验证所有文件路径是否仍然有效
- [ ] 测试 Agent 是否仍能遵循规则
- [ ] 更新变更日志

#### 每季度审查（30 分钟）

- [ ] 完整的格式审查
- [ ] 评估是否需要大规模重组
- [ ] 清理过时规则
- [ ] 考虑版本升级
- [ ] 更新本文档

### 自动化建议

#### Pre-commit Hook

```bash
#!/bin/bash
# .git/hooks/pre-commit

FILE="CLAUDE.md"
MAX_LINES=50

LINES=$(wc -l < "$FILE")
if [ $LINES -gt $MAX_LINES ]; then
    echo "⚠️  CLAUDE.md 已膨胀到 $LINES 行"
    echo "请参考 script/doc/CLAUDE.md-压缩指南.md 进行压缩"
    exit 1
fi
```

#### GitHub Action（可选）

```yaml
# .github/workflows/claude-md-check.yml
name: CLAUDE.md Size Check

on:
  pull_request:
    paths:
      - 'CLAUDE.md'

jobs:
  check-size:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Check CLAUDE.md size
        run: |
          LINES=$(wc -l < CLAUDE.md)
          if [ $LINES -gt 50 ]; then
            echo "::error::CLAUDE.md exceeds 50 lines ($LINES)"
            exit 1
          fi
```

### 压缩决策树

```
开始
  │
  ├─ 行数 < 40？
  │   └─ 是 → 无需压缩
  │   └─ 否 → 继续
  │
  ├─ 最近一次压缩 < 1 月？
  │   └─ 是 → 记录，计划下次
  │   └─ 否 → 继续
  │
  ├─ 有新增规则？
  │   └─ 是 → 评估是否必要
  │   └─ 否 → 继续
  │
  ├─ 可以合并类别？
  │   └─ 是 → 合并并压缩
  │   └─ 否 → 继续
  │
  └─ 有过时规则？
      └─ 是 → 删除并压缩
      └─ 否 → 执行压缩
```

### 压缩模板

#### 压缩前检查

```markdown
## 压缩前检查

- [ ] 当前行数：_____
- [ ] 当前大小：_____ KB
- [ ] 上次压缩日期：_____
- [ ] 新增规则数：_____
- [ ] 需要保留的规则：_____
- [ ] 可以删除的规则：_____
- [ ] 可以合并的规则：_____
```

#### 压缩后验证

```markdown
## 压缩后验证

- [ ] 压缩后行数：_____（减少 _____%）
- [ ] 压缩后大小：_____ KB（减少 _____%）
- [ ] 所有索引行以 `|` 开头
- [ ] 包含 `|IMPORTANT` 指令
- [ ] 无解释性文字
- [ ] 无完整代码块
- [ ] 所有文件路径有效
- [ ] 格式符合规范
```

## 总结

**核心理念：**
> "Prefer retrieval-led reasoning over pre-training-led reasoning"

**关键方法：**
1. 使用 `|` 开头的索引格式
2. 创建指向文件位置的索引，而非嵌入内容
3. 在顶部添加 `|IMPORTANT` 指令
4. 使用花括号 `{}` 和箭头 `→` 提高密度

**压缩目标：**
- 减少 80% 内容
- 保持 100% 准确性
- 优先检索而非预训练

**维护原则：**
- **定期压缩** - 每月或每季度
- **防止膨胀** - 增量审查，及时清理
- **版本管理** - 记录变更历史
- **自动化** - 使用脚本和 hook

**效果：**
- Vercel 测试：100% vs 79% 通过率
- Token 效率：8KB vs 40KB
- Agent 表现：显著提升
