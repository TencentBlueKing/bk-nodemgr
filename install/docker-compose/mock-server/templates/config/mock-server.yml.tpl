# basicServer defines mock-server http server settings.
basicServer:
  # listening IP and Port.
  bindIP: 0.0.0.0
  port: __MOCK_SERVER_BASIC_PORT__

  # defines the authentication mode, currently only none is supported.
  authIdentity: none

# log settings.
log:
  dir: /mock-server/log
  level: __MOCK_SERVER_LOG_LEVEL__
  maxNum: 10
  maxSizeMB: 200

# bkrepo mock storage settings.
bkrepoConfig:
  baseDir: __MOCK_SERVER_BKREPO_BASE_DIR__

# optional preset mock data.
mockData: __MOCK_SERVER_MOCK_DATA__
