## 鉴权说明

1. 添加鉴权中间件：
   ```go
   rg.Use(restserver.MiddlewareAuth(authIdentity))
   ```

2. 使用规范：
    - 仅在一级路由添加鉴权中间件
    - 中间件会自动解析用户信息并绑定到 `rest.Context` 中
