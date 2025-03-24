# backend

## 设计意图
提供与backend模块的标准化交互接口。

## 功能边界
1. 此包负责：
    - backend模块API调用封装
    - 错误处理与重试逻辑
2. 此包不负责：
    - 具体的backend业务流程

## 设计考量
数据交互上使用pkg/types中的类型，屏蔽backend的protocol信息。

## 使用限制
1. backend提供最基础的原子接口，场景层逻辑不应该放到backend这里处理。