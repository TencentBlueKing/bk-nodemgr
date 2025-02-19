declare interface Window {
  readonly BK_API_PREFIX: string
  readonly BK_SHARED_RES_BASE_JS_URL: string
  readonly BK_DAYU_HOST: string
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue';

  const component: DefineComponent<object, object, any>;
  export default component;
}

declare module '@blueking/platform-config'
