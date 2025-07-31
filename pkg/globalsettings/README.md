# globalsettings

## 设计意图
提供基于全局单例的globalsettings服务，主要用于存取全局配置

## 功能边界
1. 此包负责：
    - 实现globalsettings的storage抽象，并提供对应的单例
    - 提供全局配置的存取接口
2. 此包不负责：
    - DB的连接和管理，DB相关的由具体服务提供

## 使用限制
1. 所有的globalsettings相关的内容应限制在该包内

