# runMode debug/release.
runMode: debug

# tenantMode single/multiple.
tenantMode: single

# infoServer defines self info http server settings.
infoServer:
  # listening IP and Port.
  bindIP: 127.0.0.1
  port: __BK_NODEMGR_FILE_INFO_PORT__

  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__

# adminServer defines self admin http server settings.
adminServer:
  # listening IP and Port.
  bindIP: 127.0.0.1
  port: __BK_NODEMGR_FILE_ADMIN_PORT__

  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__

# basicServer defines self basic http server settings.
basicServer:
  # listening IP and Port.
  bindIP: 0.0.0.0
  port: __BK_NODEMGR_FILE_BASIC_PORT__
  
  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__

# downloadServer defines self node http server settings.
downloadServer:
  # listening IP and Port.
  bindIP: 127.0.0.1
  port: __BK_NODEMGR_FILE_DOWNLOAD_PORT__

  # advertiseIP advertise ip for external access.
  advertiseIPV4: __BK_NODEMGR_ADVERTISE_IPV4__

# repo defines the bkrepo related settings.
repo:
  endpoint: "__BK_NODEMGR_REPO_ENDPOINT__"
  projectID: "__BK_NODEMGR_REPO_PROJECT_ID__"
  repoName: "__BK_NODEMGR_REPO_REPO_NAME__"
  accessKey: "__BK_NODEMGR_REPO_ACCESS_KEY__"
  secretKey: "__BK_NODEMGR_REPO_SECRET_KEY__"

# workspaceFileGroup defines the workspace file group settings.
workspaceFileGroup:
  fullPath: /bk-nodemgr/workspace/

# mountHostDir defines the mount host directory settings.
mountHostDir: "__BK_NODEMGR_FILE_MOUNT_HOST_DIR__"

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
