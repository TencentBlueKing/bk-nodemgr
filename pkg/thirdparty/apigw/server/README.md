# server

## 设计意图

1. 提供统一的 apigw server 的交互逻辑，统一处理来自网关接口请求。

## 功能边界

1. 此包负责：
   - 负责解析网关接口请求的 header
      - auth 中间件
      - requestid 中间件
      - tenantid 中间件
