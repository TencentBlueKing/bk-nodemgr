# etcdiscover

## 设计意图
提供基于etcd的服务发现能力

## 功能边界
1. 此包负责：
    - 实现runtime/discover/IProvider
    - 与etcd交互、处理etcd的请求和返回数据
2. 此包不负责：
    - 生产具体的服务发现要注册的内容，内容应由调用方生成和消费

## 使用限制
1. 所有的etcd相关的内容应限制在该包内
