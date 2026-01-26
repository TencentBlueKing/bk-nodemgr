# 渐进式提案实际示例

本文档提供不同场景的完整示例,展示如何判断、设计和实施渐进式提案。

---

## 示例索引

1. **API 重构** - 大规模破坏性变更 (REST → gRPC, 20 文件)
2. **数据库迁移** - 字段移除 (5 个 collection, 数据迁移脚本)
3. **缓存策略变更** - 性能优化 (Redis → 直接查询, 8 个工作流)
4. **认证系统重构** - 架构变更 (Session → JWT, 15 文件)
5. **小规模变更** - 不需要渐进式 (单文件修改)

---

## 示例 1: API 重构 (大规模破坏性变更)

### 输入 (用户需求)

"需要将现有的 REST API 重构为 gRPC,提升性能和类型安全性。影响约 20 个文件,包括服务端实现和客户端调用。需要确保迁移期间服务不中断。"

### 步骤 1: 判断是否需要渐进式策略

**分析**:
- ✅ 影响 > 5 个文件 (20 个文件)
- ✅ 破坏性 API 变更 (协议从 REST 改为 gRPC)
- ✅ 需要外部依赖升级 (客户端需要同步升级)
- ✅ 服务可用性要求 (不能中断服务)

**结论**: **必须使用渐进式策略**

### 步骤 2: 6 阶段设计

#### 阶段 1: 准备 gRPC 服务 (无破坏性)

**目标**: 实现 gRPC 版本的 API,与 REST API 并行存在

**任务**:
- 编写 proto 定义文件 (`proto/api/v1/service.proto`)
- 实现 gRPC server handlers
- 添加 gRPC 服务的单元测试
- 保持 REST API 完全不变

**验证**:
```bash
make build          # 编译通过
make test           # 新增测试通过,现有测试不受影响
grpcurl localhost:9090 api.v1.Service/GetUser  # gRPC 可独立调用
```

**提交**: `feat: add gRPC API implementation`

---

#### 阶段 2: 双轨运行 (新旧共存)

**目标**: gRPC 和 REST API 同时提供服务

**任务**:
- 配置 gRPC 和 REST 监听不同端口
- 添加 feature flag 控制 gRPC 启用/禁用
- 更新文档,说明两种 API 的使用方式
- 更新监控,分别追踪 gRPC 和 REST 流量

**验证**:
```bash
# REST API 正常工作
curl http://localhost:8080/api/v1/users/123

# gRPC API 正常工作
grpcurl localhost:9090 api.v1.Service/GetUser

# 两者返回相同数据
diff <(curl ...) <(grpcurl ... | jq)
```

**提交**: `feat: enable dual-track mode for gRPC and REST`

---

#### 阶段 3: 逐个迁移客户端 (独立切换)

**目标**: 将客户端逐个从 REST 切换到 gRPC

**迁移清单**:
| 客户端 | 文件 | 状态 | Commit | 测试 |
|--------|------|------|--------|------|
| Web 前端 | `web/src/api/client.ts` | ✅ | abc123 | ✅ E2E 通过 |
| Mobile App | `mobile/api/service.dart` | ✅ | def456 | ✅ 集成测试通过 |
| Admin CLI | `cmd/admin/client.go` | ✅ | ghi789 | ✅ 功能测试通过 |
| Worker Service | `services/worker/rpc.go` | ⚠️ | - | 进行中 |
| Cron Jobs | `jobs/sync/api.go` | ❌ | - | 待开始 |

**进度**: 3/5 客户端已迁移 (60%)

**每个客户端的迁移步骤**:
```bash
# 1. 更新客户端代码,使用 gRPC
# 2. 运行客户端测试
make test-client

# 3. 部署到测试环境验证
# 4. 提交 (每个客户端一个独立 commit)
git commit -m "refactor(web): migrate to gRPC API"

# 5. 部署到生产环境
# 6. 监控 24 小时,确认无问题
```

**回滚策略**:
- 如果客户端出现问题,修改 feature flag 切回 REST
- 或者回滚该客户端的代码部署

**提交**: 每个客户端一个 commit
- `refactor(web): migrate to gRPC API`
- `refactor(mobile): migrate to gRPC API`
- `refactor(admin): migrate to gRPC API`
- ...

---

#### 阶段 4: 标记 REST API 废弃 (保留代码)

**目标**: 标记 REST API 为废弃,但保留代码

**任务**:
- 在 REST API 文档中添加废弃声明
- 在 REST handlers 中添加 `Deprecated` 注释
- 在响应 header 中添加 `X-API-Deprecated: true`
- 发送通知给所有 API 用户

**代码示例**:
```go
// Deprecated: Use gRPC API instead (proto/api/v1/service.proto)
// This REST API will be removed in v2.0 (2026-06-01)
func (h *Handler) GetUser(c *gin.Context) {
    c.Header("X-API-Deprecated", "true")
    c.Header("X-API-Deprecated-Replacement", "grpc://api.v1.Service/GetUser")
    // ... 保持现有实现
}
```

**验证**:
```bash
# 确认废弃标记存在
curl -I http://localhost:8080/api/v1/users/123 | grep Deprecated

# REST API 仍然正常工作
make test-rest
```

**提交**: `deprecate: mark REST API as deprecated`

---

#### 阶段 5: 观察期 (监控稳定)

**目标**: 生产环境运行 2-4 周,监控 gRPC 稳定性

**监控指标**:
- gRPC 请求成功率 > 99.9%
- gRPC 平均响应时间 < REST 的 50%
- gRPC 错误率 < 0.1%
- REST 流量下降到 < 1% (确认所有客户端已迁移)

**观察结果** (2 周后):
```
gRPC 成功率: 99.95% ✅
gRPC 平均响应时间: 15ms (REST: 45ms) ✅
gRPC 错误率: 0.03% ✅
REST 流量: 0.5% (仅遗留的内部工具) ✅
```

**提交**: 无代码变更,仅更新文档

---

#### 阶段 6: 清理 REST API (安全移除)

**目标**: 删除 REST API 代码和依赖

**任务**:
- 删除 REST handlers
- 删除 REST 路由配置
- 删除 REST 相关的中间件
- 删除 REST API 文档
- 更新 CI/CD 配置

**验证**:
```bash
# 确认 REST 代码已删除
grep -r "gin.Context" internal/api/  # 应该没有结果

# 确认编译和测试通过
make build && make test

# 确认二进制文件大小减小
ls -lh bin/server  # 应该比之前小
```

**提交**: `cleanup: remove deprecated REST API`

---

### 预期效果

**时间线**:
- 阶段 1-2: 2 周 (准备 + 双轨)
- 阶段 3: 3 周 (逐个迁移 5 个客户端)
- 阶段 4: 1 周 (标记废弃)
- 阶段 5: 2 周 (观察期)
- 阶段 6: 1 周 (清理)
- **总计**: 9 周完成迁移

**风险控制**:
- ✅ 每个客户端独立迁移,风险隔离
- ✅ Feature flag 可快速回滚
- ✅ 双轨运行期间零风险
- ✅ 观察期充分验证稳定性

**业务价值**:
- ✅ API 响应时间降低 66% (45ms → 15ms)
- ✅ 类型安全性提升 (Protocol Buffers)
- ✅ 迁移过程无服务中断
- ✅ 代码库减小 (移除 REST 相关代码)

---

## 示例 2: 数据库字段移除 (数据迁移)

### 输入 (用户需求)

"需要移除 MongoDB 中废弃的 `old_status` 字段,该字段已被新的 `status` 字段替代。影响 5 个 collection,涉及约 1000 万条数据。需要确保数据不丢失,迁移过程中服务可用。"

### 步骤 1: 判断是否需要渐进式策略

**分析**:
- ✅ 影响 > 5 个文件 (5 个 collection 对应的代码文件)
- ✅ 数据结构变更 (删除字段)
- ✅ 需要数据迁移脚本 (1000 万条数据)
- ✅ 服务可用性要求 (迁移期间不能中断服务)

**结论**: **必须使用渐进式策略**

### 步骤 2: 6 阶段设计

#### 阶段 1: 准备新字段查询方法

**目标**: 添加使用新字段 `status` 的查询方法,不影响现有代码

**任务**:
- 实现 `GetUsersByStatus(status string)` (使用新字段)
- 保持 `GetUsersByOldStatus(oldStatus string)` 不变
- 添加单元测试验证新方法

**代码示例**:
```go
// 新方法 - 使用新字段
func (d *DAO) GetUsersByStatus(ctx context.Context, status string) ([]*User, error) {
    filter := bson.M{"status": status}  // 使用新字段
    return d.find(ctx, filter)
}

// 旧方法 - 保持不变
func (d *DAO) GetUsersByOldStatus(ctx context.Context, oldStatus string) ([]*User, error) {
    filter := bson.M{"old_status": oldStatus}  // 使用旧字段
    return d.find(ctx, filter)
}
```

**验证**:
```bash
make test  # 新旧方法都能正常工作
```

**提交**: `feat: add query method using new status field`

---

#### 阶段 2: 双字段写入

**目标**: 代码同时写入 `status` 和 `old_status` 两个字段

**任务**:
- 更新所有写入操作,同时写两个字段
- 确保数据一致性

**代码示例**:
```go
func (d *DAO) UpdateUserStatus(ctx context.Context, userID string, newStatus string) error {
    update := bson.M{
        "$set": bson.M{
            "status":     newStatus,      // 新字段
            "old_status": newStatus,      // 旧字段 (同步写入)
            "updated_at": time.Now(),
        },
    }
    return d.updateOne(ctx, userID, update)
}
```

**验证**:
```bash
# 验证新数据同时包含两个字段
mongo --eval 'db.users.findOne({_id: "test-user"})'
# 输出: { status: "active", old_status: "active", ... }
```

**提交**: `feat: enable dual-field write for status`

---

#### 阶段 3: 逐个 collection 迁移历史数据

**目标**: 将历史数据的 `old_status` 复制到 `status`

**迁移脚本**:
```javascript
// scripts/migrate-status-field.js
db.users.updateMany(
    { status: { $exists: false } },  // 只更新没有新字段的文档
    [{ $set: { status: "$old_status" } }]  // 复制旧字段到新字段
);
```

**迁移清单**:
| Collection | 文档数 | 状态 | 耗时 | Commit |
|-----------|--------|------|------|--------|
| users | 5,000,000 | ✅ | 10 min | abc123 |
| organizations | 100,000 | ✅ | 30 sec | def456 |
| projects | 3,000,000 | ✅ | 6 min | ghi789 |
| tasks | 1,500,000 | ⚠️ | 3 min | 进行中 |
| comments | 500,000 | ❌ | - | 待开始 |

**进度**: 3/5 collections 已迁移 (60%)

**每个 collection 的迁移步骤**:
```bash
# 1. 在测试环境执行迁移脚本
mongo test-db < scripts/migrate-users-status.js

# 2. 验证数据完整性
node scripts/verify-migration.js users

# 3. 在生产环境执行 (低峰期)
mongo prod-db < scripts/migrate-users-status.js

# 4. 验证生产数据
node scripts/verify-migration.js users --prod

# 5. 提交记录
git commit -m "data: migrate status field for users collection"
```

**验证**:
```javascript
// 验证脚本
db.users.count({ status: { $exists: false } })  // 应该为 0
db.users.count({ old_status: { $exists: true } })  // 应该 > 0
```

**提交**: 每个 collection 一个 commit

---

#### 阶段 4: 逐个切换代码查询

**目标**: 将代码从使用 `old_status` 切换到 `status`

**任务**:
- 将所有 `GetUsersByOldStatus` 调用替换为 `GetUsersByStatus`
- 每个调用方独立切换,独立测试

**迁移清单**:
| 模块 | 文件 | 状态 | Commit |
|------|------|------|--------|
| User Service | `services/user/handler.go` | ✅ | abc123 |
| Project Service | `services/project/query.go` | ✅ | def456 |
| Report Service | `services/report/stats.go` | ⚠️ | 进行中 |
| Workflow Engine | `workflow/executor.go` | ❌ | 待开始 |

**提交**: 每个模块一个 commit

---

#### 阶段 5: 标记 `old_status` 废弃

**目标**: 在代码中标记旧字段为废弃

**代码示例**:
```go
type User struct {
    Status    string `bson:"status" json:"status"`
    // Deprecated: Use Status field instead. Will be removed in v3.0
    OldStatus string `bson:"old_status,omitempty" json:"old_status,omitempty"`
}
```

**提交**: `deprecate: mark old_status field as deprecated`

---

#### 阶段 6: 删除 `old_status` 字段

**目标**: 从数据库和代码中完全删除旧字段

**任务**:
- 从 struct 定义中删除 `OldStatus`
- 从数据库中删除字段

**数据库清理脚本**:
```javascript
// 在低峰期执行,逐个 collection 清理
db.users.updateMany({}, { $unset: { old_status: "" } });
db.organizations.updateMany({}, { $unset: { old_status: "" } });
// ...
```

**验证**:
```bash
# 确认代码中无引用
grep -r "old_status" internal/  # 应该没有结果

# 确认数据库中字段已删除
mongo --eval 'db.users.findOne({old_status: {$exists: true}})'  # 应该为 null
```

**提交**: `cleanup: remove deprecated old_status field`

---

### 预期效果

**时间线**:
- 阶段 1-2: 1 周 (准备查询方法 + 双字段写入)
- 阶段 3: 1 周 (迁移历史数据,每天 1-2 个 collection)
- 阶段 4: 1 周 (切换代码查询)
- 阶段 5: 1 天 (标记废弃)
- 阶段 6: 1 周 (观察期 + 清理)
- **总计**: 4 周完成迁移

**风险控制**:
- ✅ 双字段写入期间数据不丢失
- ✅ 逐个 collection 迁移,风险隔离
- ✅ 每个模块独立切换,可独立回滚
- ✅ 历史数据迁移在低峰期执行

**业务价值**:
- ✅ 数据库存储优化 (移除冗余字段)
- ✅ 代码简化 (移除旧字段相关逻辑)
- ✅ 迁移过程零数据丢失
- ✅ 服务无中断

---

## 示例 3: 缓存策略变更 (性能优化)

### 输入 (用户需求)

"当前使用 Redis 缓存 relay 信息,导致数据一致性问题。需要改为按需查询 MongoDB,通过优化查询性能来补偿缓存移除的影响。影响 8 个工作流。"

### 步骤 1: 判断是否需要渐进式策略

**分析**:
- ✅ 影响 > 5 个文件 (8 个工作流)
- ✅ 架构变更 (缓存 → 直接查询)
- ✅ 性能影响风险 (需要验证查询性能)
- ✅ 数据一致性要求 (不能出现缓存不一致)

**结论**: **必须使用渐进式策略**

### 步骤 2: 6 阶段设计

#### 阶段 1: 准备高效查询方法

**目标**: 实现优化的直接查询方法,性能接近缓存

**优化策略**:
- 使用 MongoDB projection (只查询需要的字段)
- 添加合适的索引
- 使用连接池优化

**代码示例**:
```go
// 新方法 - 高效的直接查询
func (d *DAO) GetRelayListByFilter(ctx context.Context, filter *RelayFilter) ([]*Relay, error) {
    // 使用 projection 只查询需要的字段
    projection := bson.M{
        "relay_id": 1,
        "host":     1,
        "port":     1,
        "status":   1,
    }

    opts := options.Find().SetProjection(projection)
    return d.find(ctx, filter.ToBSON(), opts)
}

// 旧方法 - 从 Redis 缓存读取
func (d *DAO) GetRelayInfoFromCache(ctx context.Context, relayID string) (*Relay, error) {
    // 保持不变,继续使用 Redis
    return d.cache.Get(ctx, "relay:"+relayID)
}
```

**性能测试**:
```bash
# 基准测试
go test -bench=BenchmarkGetRelay -benchmem

# 结果:
# GetRelayFromCache:     5ms   ✅
# GetRelayListByFilter:  8ms   ✅ (目标 < 10ms)
```

**验证**:
```bash
make test  # 新方法测试通过
```

**提交**: `feat: add optimized relay query method`

---

#### 阶段 2: 双轨运行

**目标**: 同时支持缓存和直接查询,通过 feature flag 控制

**代码示例**:
```go
func (w *Workflow) GetRelays(ctx context.Context, filter *RelayFilter) ([]*Relay, error) {
    if config.UseDirectQuery {  // Feature flag
        return w.dao.GetRelayListByFilter(ctx, filter)  // 新方法
    }
    return w.dao.GetRelayInfoFromCache(ctx, filter)  // 旧方法
}
```

**验证**:
```bash
# 测试两种模式都正常
USE_DIRECT_QUERY=false make test  # 缓存模式
USE_DIRECT_QUERY=true make test   # 直接查询模式
```

**提交**: `feat: enable dual-track mode for relay query`

---

#### 阶段 3: 逐个工作流切换

**目标**: 将工作流逐个切换到直接查询,验证性能

**迁移清单**:
| 工作流 | 文件 | QPS | 响应时间 | 状态 | Commit |
|--------|------|-----|---------|------|--------|
| Node Install | `workflow/node_install.go` | 100 | 8ms | ✅ | abc123 |
| Node Update | `workflow/node_update.go` | 50 | 7ms | ✅ | def456 |
| Batch Deploy | `workflow/batch_deploy.go` | 200 | 9ms | ✅ | ghi789 |
| Health Check | `workflow/health_check.go` | 500 | 6ms | ⚠️ | 进行中 |
| Auto Scale | `workflow/auto_scale.go` | 10 | - | ❌ | 待开始 |

**进度**: 3/8 工作流已迁移 (37.5%)

**每个工作流的切换步骤**:
```bash
# 1. 在测试环境切换 feature flag
kubectl set env deployment/backend USE_DIRECT_QUERY=true

# 2. 运行负载测试
k6 run tests/load/node-install.js

# 3. 验证性能指标
# - 响应时间 < 10ms ✅
# - 错误率 < 0.1% ✅
# - QPS 满足要求 ✅

# 4. 部署到生产环境
# 5. 监控 24 小时
# 6. 提交记录
git commit -m "refactor(node-install): switch to direct relay query"
```

**回滚策略**:
- 如果性能不达标,立即切回 feature flag
- 或者优化查询后重试

**提交**: 每个工作流一个 commit

---

#### 阶段 4: 标记缓存废弃

**目标**: 标记 Redis 缓存为废弃

**代码示例**:
```go
// Deprecated: Use GetRelayListByFilter instead
// Redis cache will be removed in v2.0
func (d *DAO) GetRelayInfoFromCache(ctx context.Context, relayID string) (*Relay, error) {
    log.Warn("GetRelayInfoFromCache is deprecated, use GetRelayListByFilter")
    return d.cache.Get(ctx, "relay:"+relayID)
}
```

**提交**: `deprecate: mark relay cache as deprecated`

---

#### 阶段 5: 观察期

**目标**: 监控 2 周,确认直接查询稳定且性能达标

**监控指标**:
```
查询成功率: 99.98% ✅
平均响应时间: 7.5ms (目标 < 10ms) ✅
P99 响应时间: 15ms ✅
Redis 流量: < 1% (确认几乎无缓存查询) ✅
MongoDB 负载: 增加 5% (在可接受范围) ✅
```

**提交**: 无代码变更

---

#### 阶段 6: 清理 Redis 缓存

**目标**: 删除 Redis 缓存相关代码和配置

**任务**:
- 删除 `GetRelayInfoFromCache` 方法
- 删除 Redis 连接配置
- 删除缓存刷新逻辑
- 删除缓存相关的监控

**验证**:
```bash
# 确认 Redis 代码已删除
grep -r "cache.Get" internal/dao/  # 应该没有 relay 相关结果

# 确认编译和测试通过
make build && make test
```

**提交**: `cleanup: remove deprecated relay cache`

---

### 预期效果

**时间线**:
- 阶段 1-2: 1 周 (准备查询方法 + 双轨)
- 阶段 3: 2 周 (逐个切换 8 个工作流)
- 阶段 4: 1 天 (标记废弃)
- 阶段 5: 2 周 (观察期)
- 阶段 6: 3 天 (清理)
- **总计**: 5 周完成迁移

**风险控制**:
- ✅ 双轨运行期间零风险
- ✅ 每个工作流独立切换,可独立回滚
- ✅ 性能指标实时监控
- ✅ Feature flag 可快速回滚

**业务价值**:
- ✅ 数据一致性提升 (无缓存不一致问题)
- ✅ 架构简化 (移除 Redis 依赖)
- ✅ 运维成本降低 (无需维护 Redis 集群)
- ✅ 查询性能依然优秀 (< 10ms)

---

## 示例 4: 认证系统重构 (架构变更)

### 输入 (用户需求)

"当前使用 Session 认证,需要改为 JWT 认证以支持分布式部署和移动端。影响约 15 个文件,包括认证中间件、用户服务、权限验证等。"

### 步骤 1: 判断是否需要渐进式策略

**分析**:
- ✅ 影响 > 5 个文件 (15 个文件)
- ✅ 架构变更 (Session → JWT)
- ✅ 破坏性变更 (认证机制变化)
- ✅ 用户影响 (所有用户需要重新登录)

**结论**: **必须使用渐进式策略**

### 6 阶段设计概览

#### 阶段 1: 准备 JWT 认证实现
- 实现 JWT token 生成和验证
- 添加 JWT 中间件
- 保持 Session 认证不变

#### 阶段 2: 双认证支持
- 同时支持 Session 和 JWT
- 登录接口返回两种 token
- 两种认证方式都能访问 API

#### 阶段 3: 逐步引导用户切换
- Web 端先切换到 JWT (feature flag)
- Mobile App 切换到 JWT
- Admin 后台切换到 JWT

#### 阶段 4: 标记 Session 废弃
- 文档说明 Session 将被移除
- 登录接口添加废弃警告

#### 阶段 5: 观察期 (4 周)
- 监控 JWT 使用率 > 99%
- Session 使用率 < 1%

#### 阶段 6: 清理 Session
- 删除 Session 存储 (Redis)
- 删除 Session 中间件
- 更新文档

---

## 示例 5: 小规模变更 (不需要渐进式)

### 输入 (用户需求)

"修改配置文件,将日志级别从 Info 改为 Warn,减少日志输出。只影响 1 个文件 `config/logger.yaml`。"

### 判断结果

**分析**:
- ❌ 仅影响 1 个文件
- ❌ 无破坏性变更
- ❌ 无数据迁移需求
- ❌ 配置变更,风险极低

**结论**: **不需要渐进式策略**,直接修改即可

### 实施方式

**直接修改**:
```yaml
# config/logger.yaml
logger:
  level: warn  # 从 info 改为 warn
  format: json
```

**验证**:
```bash
# 重启服务,确认日志级别生效
make deploy
kubectl logs deployment/backend | grep "level=warn"
```

**提交**: `config: change log level to warn`

**耗时**: < 10 分钟

---

## 总结: 何时使用渐进式策略

### 必须使用 ✅

| 场景 | 示例 | 风险 |
|------|------|------|
| 大规模变更 (> 5 文件) | API 重构, 架构变更 | 高 |
| 破坏性变更 | 协议变更, 接口签名变化 | 高 |
| 数据迁移 | 字段移除, 类型变更 | 中 |
| 性能影响 | 缓存策略变更 | 中 |
| 多方协作 | 客户端需同步升级 | 中 |

### 可选使用 ⚠️

| 场景 | 示例 | 风险 |
|------|------|------|
| 中等规模 (2-5 文件) | 功能增强 | 低-中 |
| 向后兼容变更 | 新增可选参数 | 低 |

### 不需要使用 ❌

| 场景 | 示例 | 风险 |
|------|------|------|
| 单文件修改 | 配置变更, 小 bug 修复 | 极低 |
| 文档更新 | README, API 文档 | 无 |
| 测试添加 | 新增单元测试 | 无 |

---

## 关键成功因素

从以上示例中提炼的关键要素:

1. **明确的判断标准** - 使用 5 个文件作为阈值
2. **标准 6 阶段模式** - 准备 → 双轨 → 迁移 → 废弃 → 观察 → 清理
3. **独立提交** - 每个单元一个 commit,可独立回滚
4. **快速验证** - 每个阶段 < 30 分钟验证
5. **实时监控** - 关键指标持续追踪
6. **Feature Flag** - 支持快速回滚
7. **详细文档** - 每个阶段都有清晰的任务和验证标准
8. **进度追踪** - 使用表格追踪迁移进度
