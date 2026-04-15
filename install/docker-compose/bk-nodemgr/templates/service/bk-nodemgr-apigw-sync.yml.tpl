# NOTE: apigw-sync used to sync api resources to bk-apigw.
name: bk-nodemgr-apigw-sync

services:
  image: mirrors.tencent.com/bk-nodemgr/bk-nodemgr-apigw-sync:stage
  network_mode: host
  environment:
    - BK_APIGW_NAME=bk-nodemgr
    - BK_APP_CODE=__BK_NODEMGR_APPCODE__
    - BK_APP_SECRET=__BK_NODEMGR_APPSECRET__
    - BK_API_URL_TMPL=__BK_NODEMGR_APIGW_URL_TMPL__
    - RELEASE_STAGES=__BK_NODEMGR_APIGW_RELEASE_STAGES__
    - NO_PUB=false
  volumes:
    - ./etc/definition.yaml:/data/definition.yaml
  command: "/bin/bash /data/bin/sync-apigateway.sh"
  restart: "no"