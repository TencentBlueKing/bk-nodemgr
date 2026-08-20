# runMode debug/release.
runMode: debug

# tenantMode single/multiple.
tenantMode: single

# front settings.
front:
  # password vault related settings.
  passwordVaultSwitch: false
  passwordVaultName: "password_vault"
  # bk user web url.
  bkUserWebURL: "__BK_NODEMGR_APPLICATION_USER_WEB_URL__"
  # bk domain.
  bkDomain: "__BK_NODEMGR_APPLICATION_DOMAIN__"
  # bk docs center url.
  bkDocsCenterURL: "__BK_NODEMGR_APPLICATION_DOCS_CENTER_URL__"
  # bk app nav open source url.
  bkAppNavOpenSourceURL: "__BK_NODEMGR_APPLICATION_NAV_OPEN_SOURCE_URL__"
  # default port for Windows WMI connection.
  windowsWMIPortDefault: __BK_NODEMGR_APPLICATION_WINDOWS_WMI_PORT_DEFAULT__
  # default port for Unix-like OS (Linux, AIX, Darwin, etc.) SSH connection.
  unixSSHPortDefault: __BK_NODEMGR_APPLICATION_UNIX_SSH_PORT_DEFAULT__
  # bk iam saas host.
  bkIamSaaSHost: __BK_NODEMGR_APPLICATION_IAM_SAAS_HOST__
  # bk user saas host.
  bkUserSaaSHost: __BK_NODEMGR_APPLICATION_USER_SAAS_HOST__

# infoServer defines self info http server settings.
infoServer:
  # listening IP and Port.
  bindIP: __BK_NODEMGR_APPLICATION_INFO_BIND_IP__
  port: __BK_NODEMGR_APPLICATION_INFO_PORT__

  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__
  traceServiceName: "application-server-info"
  traceSampleRate: 0

# adminServer defines self admin http server settings.
adminServer:
  # listening IP and Port.
  bindIP: 127.0.0.1
  port: __BK_NODEMGR_APPLICATION_ADMIN_PORT__

  # defines the authentication mode, currently only rest-server and none is supported.
  authIdentity: __BK_NODEMGR_APPLICATION_ADMIN_AUTH_IDENTITY__
  traceServiceName: "application-server-admin"
  traceSampleRate: 0

  # defines the JWT server configuration for authentication
  jwtServerConfig:
    # JWT encryption type: symmetric or asymmetric
    cryptoType: symmetric
    # symmetric key for JWT HMAC algorithms (HS256, HS384, HS512)
    symmetricKey: "__BK_NODEMGR_APPLICATION_ADMINSERVER_JWT_SYMMETRIC_KEY__"

# basicServer defines self basic http server settings.
basicServer:
  # listening IP and Port.
  bindIP: 0.0.0.0
  port: __BK_NODEMGR_APPLICATION_BASIC_PORT__

  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__
  traceServiceName: "application-server-basic"
  traceSampleRate: 0

# backend settings.
backend:
  # backend endpoints.
  endpoints:
    - __BK_NODEMGR_SERVICE_ENDPOINT__

  # if backend is behind bk-apigw, set appCode and appSecret.
  appCode: __BK_NODEMGR_APPCODE__
  appSecret: __BK_NODEMGR_APPSECRET__
  authMode: "un"
  traceServiceName: "application-client-backend"
  traceSampleRate: 0

# log settings.
log:
  dir: /bk-nodemgr/log/
  level: __BK_NODEMGR_APPLICATION_LOG_LEVEL__
  maxNum: 10
  maxSizeMB: 200

# etcd settings.
etcd:
  endpoints:
    - __BK_NODEMGR_ADVERTISE_IPV4__:__BK_NODEMGR_ETCD_PORT__
  username: root
  password: __BK_NODEMGR_ETCD_PASSWORD__

# mongodb settings.
mongodb:
  appName: __BK_NODEMGR_MONGODB_APPNAME__-application
  hosts:
    - __BK_NODEMGR_ADVERTISE_IPV4__:__BK_NODEMGR_MONGODB_PORT__
  username: root
  password: __BK_NODEMGR_MONGODB_PASSWORD__
  authSource: admin
  authMechanism: SCRAM-SHA-256
  database: bk_nodemgr
  traceServiceName: "bk_nodemgr_mongo"
  traceSampleRate: 0

# file settings.
file:
  # defines the JWT client configuration for file service
  jwtClientConfig:
    # JWT encryption type: symmetric or asymmetric
    cryptoType: symmetric
    # symmetric key for JWT HMAC algorithms (HS256, HS384, HS512)
    symmetricKey: "__BK_NODEMGR_FILE_BASICSERVER_JWT_SYMMETRIC_KEY__"
    # token expiration duration (e.g., 1h, 24h)
    tokenExpirationHour: 24
  traceServiceName: "application-client-file"
  traceSampleRate: 0

# notice settings.
notice:
  # 是否启用通知功能
  # true: 启用 bk-notice 集成，需要确保 bk-notice 服务可访问
  # false: 禁用通知功能，使用 NoopHandler，不会向 bk-notice 发送请求
  # 默认: false（需要显式设置为 true 才启用）
  enabled: false

  # notice API Gateway 配置（仅在 enabled: true 时需要）
  # endpoints:
  #   - __BK_NODEMGR_NOTICE_ENDPOINT__
  # appCode: __BK_NODEMGR_APPCODE__
  # appSecret: __BK_NODEMGR_APPSECRET__
  # authMode: "un"
  # traceServiceName: "application-client-notice"
  # traceSampleRate: 0

# bkSaaS saas settings.
bkSaaS:
  bkLogin:
    loginURL: __BK_NODEMGR_APPLICATION_LOGIN_URL__
    endpoints:
      - __BK_NODEMGR_APPLICATION_BKLOGIN_ENDPOINT__
    appCode: __BK_NODEMGR_APPCODE__
    appSecret: __BK_NODEMGR_APPSECRET__
    user: __BK_NODEMGR_VIRTUAL_USER__
    authMode: "un"
    accessToken: ""
    authType: __BK_NODEMGR_APPLICATION_AUTH_TYPE__
    # trace service name for BK login client
    traceServiceName: "application-client-bklogin"
    # trace sample rate for BK login client
    traceSampleRate: 0
    # defines tls related options.
    tls:
      # server should be accessed without verifying the TLS certificate.
      insecureSkipVerify: true

# tracing settings.
tracing:
  exporterType: "stdout"
  otlpEndpoint: ""
  otlpInsecure: false
  otlpHeaders: {}

# config policy option settings.
configPolicyOption:
  filePath: /bk-nodemgr/support-files/configpolicy
