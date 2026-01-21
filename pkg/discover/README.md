# discover

## 设计意图
提供服务发现相关的定义、接口。

## 功能边界
1. 此包负责：
    - 定义通用的服务发现元数据结构
    - 定义通用的服务发现、注册、管理接口、错误
    - 自带一个默认的基于内存的服务发现实现，用作示例
    - 提供端点选择器（selector）能力，包括：
      - `Selector` 接口：定义端点选择的标准接口，用于从多个 endpoints 中选择一个 endpoint
      - `RandomSelector`：随机选择器实现，使用随机方式选择 endpoint
      - `RoundRobinSelector`：轮询选择器实现，使用轮询方式选择 endpoint，确保负载均衡
      - `SelectEndpoints` 函数：从多个 endpoints 中选择指定数量的 endpoints，使用传入的 selector 参数决定选择策略
2. 此包不负责：
    - 具体的第三方服务发现实现

## 使用限制
1. 不应包含第三方服务发现实现的逻辑和数据结构
2. 当有一个新的服务或端口要添加时，应该在该包里声明ServiceName或EndpointName
