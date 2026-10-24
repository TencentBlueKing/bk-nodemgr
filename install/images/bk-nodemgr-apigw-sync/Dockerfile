FROM hub.bktencent.com/blueking/apigw-manager:4.2.4

COPY support-files /data/

# adminclient resolves kms-managed credentials (e.g. bkapp code/secret) on the
# command line when the sync job runs with kms.enabled.
COPY bk-nodemgr-adminclient /usr/local/bin/bk-nodemgr-adminclient

WORKDIR /data
