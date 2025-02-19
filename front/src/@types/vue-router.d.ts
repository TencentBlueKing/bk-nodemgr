
import 'vue-router';
import type { MainMenuNames } from '@/composables/use-menu';

export {};

declare module 'vue-router' {
  interface RouteMeta {
    layout?: 'default' | 'content' // 当前路由采用的布局
    title?: string // 当前路由是否有title
    mainMenu?: MainMenuNames // 当前路由对应的主菜单
    back?: boolean // 是否显示返回按钮
  }
}
