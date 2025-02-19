## 模块

全局模块，实现install方法，会自动安装

```ts
import type { UserModule } from '@/types'

export const install: UserModule = ({ app }) => {
  // do something
}
```

### pinia.ts

pinia.ts 模块用于设置 Pinia 状态管理库，并实现了持久化存储功能。

主要功能：
1. 创建 Pinia 实例
2. 实现自定义的 Pinia 插件，用于持久化存储
3. 配置持久化存储选项，包括存储键名、版本和需要持久化的 store ID

关键点：
- 使用 `createPinia()` 创建 Pinia 实例
- 自定义 `installPiniaStorage` 插件，实现状态持久化
- 通过 `STORAGE_KEY` 和 `STORAGE_VERSION` 常量管理存储键和版本
- 目前配置为只持久化 'user' store

### i18n.ts

i18n.ts 模块用于设置国际化（i18n）功能，支持多种语言的切换。

主要功能：
1. 创建 i18n 实例
2. 配置语言列表和默认语言
3. 实现语言切换功能
