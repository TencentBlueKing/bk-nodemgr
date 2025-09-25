# contextx

## 设计意图

本包提供全局唯一的通用Context, 任何跨大范围流通的Context都必须用此包内的定义和实现.

## 功能边界
1. 若是某个特定场景中有私有流通的Context, 可以自行定义自己的上下文, 如restserver