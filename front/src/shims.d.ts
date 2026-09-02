declare interface Window {
  readonly BK_API_PREFIX: string
  readonly BK_SHARED_RES_BASE_JS_URL: string
  readonly BK_DAYU_HOST: string
  PROJECT_CONFIG: {
    BK_API_PREFIX: string,
    BK_SHARED_RES_BASE_JS_URL: string,
    BK_LOGIN_URL: string,
    BK_REQUEST_ID_HEADER_KEY: string,
    PASSWORD_VAULT_SWITCH: string,
    PASSWORD_VAULT_NAME: string,
    BK_DOMAIN: string,
    BKAPP_NAV_OPEN_SOURCE_URL: string,
    BK_DOCS_CENTER_URL: string,
    BK_TENANT: string,
    BK_TENANT_MODE: string,
    BK_USER_WEB_URL: string,
    ENABLE_NOTICE: string,
    APP_VERSION: string,
    LOGIN_NAME: string,
    WINDOWS_WMI_PORT_DEFAULT: string,
    WINDOWS_SSH_PORT_DEFAULT: string,
    UNIX_SSH_PORT_DEFAULT: string,
    BK_IAM_SYSTEM_ID_BK_NODEMGR: string,
    BK_IAM_SYSTEM_ID_BK_CMDB: string,
    BK_IAM_SAAS_HOST: string,
    BK_USER_SAAS_HOST: string,
    USER_TIMEZONE: string,
    BK_USERNAME: string,
    USER_EMAIL: string,
  }
  loginModal: Object
}

declare module '*.vue' {
  import type { DefineComponent } from 'vue';

  const component: DefineComponent<object, object, any>;
  export default component;
}

declare module '@blueking/platform-config'
