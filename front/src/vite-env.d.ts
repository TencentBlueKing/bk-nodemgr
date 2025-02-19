/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly BK_NODE_ENV: string
  readonly BK_SITE_URL: string
  readonly BK_STATIC_URL: string
  readonly BK_APP_HOST: string
  readonly BK_APP_PORT: string
  readonly BK_API_PREFIX: string
  readonly BK_SHARED_RES_BASE_JS_URL: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
