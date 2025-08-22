# runMode debug/release.
runMode: debug

# tenantMode single/multiple.
tenantMode: single

# httpServer defines self http server settings.
httpServer:
  # listening IP and Port.
  bindIP: 0.0.0.0
  port: __BK_NODEMGR_APPLICATION_SERVICE_PORT__

  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__

# adminServer defines self admin http server settings.
adminServer:
  # listening IP and Port.
  bindIP: 127.0.0.1
  port: __BK_NODEMGR_APPLICATION_ADMIN_PORT__

# backend settings.
backend:
  # backend endpoints. 
  endpoints:
    - __BK_NODEMGR_SERVICE_ENDPOINT__

  # if backend is behind bk-apigw, set appCode and appSecret.
  appCode: __BK_NODEMGR_APPCODE__
  appSecret: __BK_NODEMGR_APPSECRET__

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

# bkSaaS saas settings.
bkSaaS:
  bkLogin:
    loginURL: __BK_NODEMGR_APPLICATION_LOGIN_URL__
    # defines tls related options.
    tls:
      # server should be accessed without verifying the TLS certificate.
      insecureSkipVerify: true