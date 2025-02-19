# 项目文档

## 项目介绍

这是一个基于Vue 3的前端项目，使用了现代化的开发工具和最佳实践。本文档旨在帮助新手快速入门并了解项目结构和开发规范。

## 快速开始

1. 克隆项目仓库
2. 安装依赖：`pnpm install`
3. 启动开发服务器：`pnpm dev`
4. 在浏览器中打开 `http://localhost:5008`（或控制台输出的URL）

## 项目结构

- `src/`: 源代码目录
  - `components/`: 全局组件，会自动注册
  - `layouts/`: 通用界面布局相关
  - `modules/`: 全局模块，实现install方法，会自动安装
  - `styles/`: 样式文件
  - `types.ts`: 类型定义文件
  - `App.vue`: 根组件
  - `main.ts`: 入口文件

## 开发规范

### API定义

- API名称定义跟后端接口定义保持一致`proto`文件夹下
- API定义最好自定义下请求和返回数据结构
- 路径参数用`$`开头做一下区分

### 文件命名

- 文件命名统一使用*连字符*格式，例如：
  - `user.ts`
  - `request-queue.ts`
  - `home-page.vue`

### 组件开发

- 组件文件放置在 `src/components/` 目录下
- 组件文件名使用连字符格式，如 `user-profile.vue`
- 组件名称使用大写驼峰形式，如 `UserProfile`
- 组件引用统一使用大写驼峰形式，例如：
  - `<FlexRow />`

### 布局使用

- 布局文件放置在 `src/layouts/` 目录下
- 可通过配置`meta`下`layout`属性定制特定的布局
- 默认布局为`default`，所有页面都使用该布局
- 示例：
  ```js
  export default {
    meta: {
      layout: 'custom'
    }
  }
  ```

### 模块开发

- 模块文件放置在 `src/modules/` 目录下
- 模块应实现`install`方法，例如：
  ```ts
  import type { UserModule } from '@/types'

  export const install: UserModule = ({ app, router, isClient }) => {
    // 在这里实现模块的安装逻辑
  }
  ```

### 样式开发

- 全局样式文件放置在 `src/styles/` 目录下
- 使用 Tailwind CSS 进行样式开发
- 可以在 `src/styles/main.css` 中添加自定义样式

## 构建和部署

1. 运行构建命令：`pnpm build`
2. 构建后的文件将生成在 `dist/` 目录下
3. 将 `dist/` 目录下的文件部署到您的服务器或静态文件托管服务

## 常见问题

如果您在开发过程中遇到任何问题，请查看项目的 `FAQ.md` 文件或提交 issue 到项目仓库。

## 贡献指南

我们欢迎并感谢任何形式的贡献。如果您想为项目做出贡献，请查看 `CONTRIBUTING.md` 文件了解详细信息。
