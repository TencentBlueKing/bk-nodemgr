# contextx

## 设计意图

本包提供全局唯一的通用Context, 任何跨大范围流通的Context都必须用此包内的定义和实现.

## 功能边界
1. 若是某个特定场景中有私有流通的Context, 可以自行定义自己的上下文, 如restserver
2. message_id: 特定业务逻辑内的唯一标识, 如从外部请求进来的数据, 其message_id为request_id, 而在workflow中的数据, 其message_id为operinst_id+action_name
3. tracing_id: 横跨多个区块的追溯键

## 概念解释
1. New和From的都是从一个context生成一个新的context(copy), 区别在于New是从golang原生的Context生成, From是从一个已存在的IContext生成.