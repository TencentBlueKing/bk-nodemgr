declare interface Window {
  readonly BK_API_PREFIX: string
  readonly BK_SHARED_RES_BASE_JS_URL: string
  readonly BK_DAYU_HOST: string
  PROJECT_CONFIG: {
    BK_API_PREFIX: string,
    BK_USER_MANAGE: string,
    BK_SHARED_RES_BASE_JS_URL: string,
    BK_LOGIN_URL: string,
    SITE_URL: string,
    BK_REQUEST_ID_HEADER_KEY: string
  }
  loginModal: Object
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue';

  const component: DefineComponent<object, object, any>;
  export default component;
}

declare module '@blueking/platform-config'
