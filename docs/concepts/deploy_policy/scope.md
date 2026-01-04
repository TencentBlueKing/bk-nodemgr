## scope

### target

| 范围类型\目标类型        | host | service_instance |
|------------------|------|------------------|
| topo             | ✔    | ✔                |
| set_template     | ✔    | ✔                |
| service_template | ✔    | ✔                |
| instance         | ✔    | ✔                |
| dynamic_group    | ✔    | ✖                |

### host

#### topo

转换流程：

```
topo paths → 查询 host → 转换为 target
```

说明：

- 根据 topo paths 直接查询 CMDB 获取 host 列表
- 将 host 信息转换为 target 对象

#### set_template

转换流程：

```
set_template_ids + set_ids → 查询 host → 转换为 target
```

说明：

- 根据 set_template_id 和 set_id 查询 CMDB 获取 host 列表
- 将 host 信息转换为 target 对象

#### service_template

转换流程：

```
service_template_ids + module_ids → 查询 host → 转换为 target
```

说明：

- 根据 service_template_id 和 module_id 查询 CMDB 获取 host 列表
- 将 host 信息转换为 target 对象

#### instance

转换流程：

```
instance_ids (host_id) → 查询 host → 转换为 target
```

说明：

- 直接将 instance_id（即 host_id）作为条件查询 CMDB 获取 host 列表
- 如果指定了 `biz_id`，会在查询条件中增加 biz_id 过滤
- 将 host 信息转换为 target 对象

#### dynamic_group

转换流程：

```
dynamic_group_ids → 执行 dynamic_group 规则 → 查询 host → 转换为 target
```

说明：

- 根据 dynamic_group_id 执行 dynamic_group 规则查询 CMDB 获取 host 列表
- 将 host 信息转换为 target 对象

### service_instance

#### topo

转换流程（根据 topo 节点类型分为两种策略）：

**策略 1：module 节点**

```
topo paths (module) → module_ids → 查询 service_instance → 查询 host → 转换为 target
```

**策略 2：其他节点（非 module）**

```
topo paths → 查询 host → 查询 service_instance → 转换为 target
```

说明：

- 系统会先分离 module 节点和其他类型的节点
- 对于 module 节点，直接通过 module_id 查询 service_instance 详情，再查询关联的 host 信息
- 对于其他节点，先查询 host，再通过 host_id 查询 service_instance 详情
- 最后合并结果并去重，将 service_instance 详情和 host 信息转换为 target 对象

#### set_template

转换流程：

```
set_template_ids → 查询 module_id → 查询 service_instance → 查询 host → 转换为 target
```

说明：

- 首先通过 set_template_id 查询所有关联的 module_id
- 对每个 module_id，查询该 module 下的 service_instance 详情
- 收集所有 service_instance 关联的 host_id，批量查询 host 信息
- 将 service_instance 详情和 host 信息组合转换为 target 对象

#### service_template

转换流程：

```
service_template_ids → 查询 module_id → 查询 service_instance → 查询 host → 转换为 target
```

说明：

- 首先通过 service_template_id 查询所有关联的 module_id
- 对每个 module_id，查询该 module 下的 service_instance 详情
- 收集所有 service_instance 关联的 host_id，批量查询 host 信息
- 将 service_instance 详情和 host 信息组合转换为 target 对象

#### instance

转换流程：

```
instance_ids (service_instance_id) → 查询 service_instance → 查询 host → 转换为 target
```

说明：

- 直接将 instance_id（即 service_instance_id）作为条件查询 CMDB 获取 service_instance 详情
- 从 service_instance 详情中提取 host_id，批量查询 host 信息
- 如果指定了 `biz_id`，会在 host 查询条件中增加 biz_id 过滤
- 将 service_instance 详情和 host 信息组合转换为 target 对象

#### dynamic_group

service_instance 不支持 dynamic_group

### FAQ

#### **Q：为什么 `service_template` 会被转换为 `module_id`，而不是 `service_id`？**

**A：**

在 CMDB 中，`service_template` 只是一种模板，用来创建具体的实例。每当对 `service_template` 进行实例化时，就会产生一个 **module**，因此需要把模板的标识转换为对应的
**module_id**。

在同一个 module 中可以容纳多个 service_instance，每个实例均通过 module 内唯一的 service_id 进行标识。

module 与 host 的组合能够唯一确定对应的 service_instance。

所以，`service_template` 被转换为 `module_id`，而不是 `service_id`。