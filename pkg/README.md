# pkg

## 功能边界

本目录主要用于存放各个 service 之间的公共代码，例如：

- 数据类型
- 第三方SDK（或者是包装成 SDK 的接口调用）
- 业务无关代码（如 SSH 封装、Redis 封装等，虽然不一定在各个服务中都会使用，但由于其业务无关性，适合放在 pkg 目录中）
- runtime

在将代码迁移到本目录时请考虑一下几个问题：

1. 对应的代码是否在各个 service 中都会用到？
2. 对应的代码是否与业务无关？

## pkg vs runtime

pkg 和 runtime 的区别在于：

- pkg 是用于存放各个 service 之间的公共代码，是允许接受系统内部约定好的概念，例如 networkAreaID, hostID 等
- runtime 是用于存放纯粹的运行逻辑，是可以直接抽离出系统的，例如 conv, gopool