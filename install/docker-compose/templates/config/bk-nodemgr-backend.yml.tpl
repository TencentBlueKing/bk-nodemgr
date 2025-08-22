# runMode debug/release.
runMode: debug

# tenantMode single/multiple.
tenantMode: single

# httpServer defines self http server settings.
httpServer:
  # listening IP and Port.
  bindIP: 0.0.0.0
  port: __BK_NODEMGR_BACKEND_SERVICE_PORT__
  
  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__

# adminServer defines self admin http server settings.
adminServer:
  # listening IP and Port.
  bindIP: 127.0.0.1
  port: __BK_NODEMGR_BACKEND_ADMIN_PORT__

  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__

# callbackServer defines self callback http server settings.
callbackServer:
  # listening IP and Port.
  bindIP: 0.0.0.0
  port: __BK_NODEMGR_BACKEND_CALLBACK_PORT__

  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__

# proxyServer defines self proxy http server settings.
proxyServer:
  # listening IP and Port.
  bindIP: 0.0.0.0
  port: __BK_NODEMGR_BACKEND_PROXY_PORT__

  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__

# workflow defines the backend workflow settings.
workflow:
  workerNum: 8

# system defines the gse environs and edition.
system:
  env: __BK_NODEMGR_BACKEND_SYSTEM_ENV__
  edition: __BK_NODEMGR_BACKEND_SYSTEM_EDITION__

# gseDeployConfs defines the gse deploy related settings.
gseDeployConfs:
  - generation: 2
    osType: "linux"
    baseWorkDir: "/tmp/bknm/"
    baseDeployDir: "/usr/local/"
  - generation: 2
    osType: "windows"
    baseWorkDir: "c:\\tmp\\bknm\\"
    baseDeployDir: "c:\\"
  - generation: 2
    osType: "darwin"
    baseWorkDir: "/tmp/bknm/"
    baseDeployDir: "/usr/local/"

# encryptKey: define the key used to encrypt the sensitive data.
encryptKey: "__BK_NODEMGR_BACKEND_ENCRYPT_KEY__"

# installerFileGroup settings.
installerFileGroup:
  fullPath: /bk-nodemgr/file/tools

# defines the API gateway related settings.
apiGateWayServer:
  publickeyPem: "__BK_NODEMGR_BACKEND_APIGW_PUBLIC_PEM__"

# defines the CMDB (bk-apigw) related settings.
cmdb:
  supplierAccount: "0"
  endpoints:
    - "__BK_NODEMGR_CMDB_ENDPOINT__"
  appCode: __BK_NODEMGR_APPCODE__
  appSecret: __BK_NODEMGR_APPSECRET__
  user: admin
  authMode: "un"

# defines the GSE (bk-apigw) related settings.
gse:
  # endpoints is a seed list of host:port addresses of esb nodes.
  endpoints:
    - "__BK_NODEMGR_GSE_ENDPOINT__"
  appCode: __BK_NODEMGR_APPCODE__
  appSecret: __BK_NODEMGR_APPSECRET__
  user: admin
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

# redis settings.
redis:
  host: __BK_NODEMGR_ADVERTISE_IPV4__
  port: __BK_NODEMGR_REDIS_PORT__
  password: __BK_NODEMGR_REDIS_PASSWORD__
  db: 0
