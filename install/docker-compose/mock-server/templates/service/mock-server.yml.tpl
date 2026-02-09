# NOTE: mock-server is intended for development and testing purposes only,
# providing mock API endpoints for third-party services (CMDB, BK-Repo, etc.).
#
name: mock-server

services:
  mock-server:
    image: __MOCK_SERVER_IMAGE__
    container_name: mock-server
    restart: always
    ports:
      - "__MOCK_SERVER_BASIC_PORT__:__MOCK_SERVER_BASIC_PORT__"
    volumes:
      - ./etc/mock-server.yml:/mock-server/etc/mock-server.yml
    command: "/mock-server/bin/mock-server -f /mock-server/etc/mock-server.yml"
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "5"
