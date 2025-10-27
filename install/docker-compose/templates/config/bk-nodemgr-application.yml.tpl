# runMode debug/release.
runMode: debug

# tenantMode single/multiple.
tenantMode: single

# infoServer defines self info http server settings.
infoServer:
  # listening IP and Port.
  bindIP: 127.0.0.1
  port: __BK_NODEMGR_APPLICATION_INFO_PORT__

  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__

# adminServer defines self admin http server settings.
adminServer:
  # listening IP and Port.
  bindIP: 127.0.0.1
  port: __BK_NODEMGR_APPLICATION_ADMIN_PORT__

  # defines the authentication mode, currently only rest-server and none is supported.
  authIdentity: rest-server

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

# backend settings.
backend:
  # backend endpoints.
  endpoints:
    - __BK_NODEMGR_SERVICE_ENDPOINT__

  # if backend is behind bk-apigw, set appCode and appSecret.
  appCode: __BK_NODEMGR_APPCODE__
  appSecret: __BK_NODEMGR_APPSECRET__
  authMode: "un"

# log settings.
log:
  dir: /bk-nodemgr/log/
  level: INFO
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
  hosts:
    - __BK_NODEMGR_ADVERTISE_IPV4__:__BK_NODEMGR_MONGODB_PORT__
  username: root
  password: __BK_NODEMGR_MONGODB_PASSWORD__
  authSource: admin
  authMechanism: SCRAM-SHA-256
  database: bk_nodemgr

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

# bkSaaS saas settings.
bkSaaS:
  bkLogin:
    loginURL: __BK_NODEMGR_APPLICATION_LOGIN_URL__
    authType: __BK_NODEMGR_APPLICATION_AUTH_TYPE__
    # defines tls related options.
    tls:
      # server should be accessed without verifying the TLS certificate.
      insecureSkipVerify: true