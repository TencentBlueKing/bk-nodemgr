# NOTE: services, including third-party components and internal bk-nodemgr modules, are
# intended for quick experience and cannot be deployed as a final production
# environment due to their lack of high availability configuration.
#
name: bk-nodemgr

services:
  etcd:
    image: bitnami/etcd:3.5.5
    container_name: bk-nodemgr-etcd
    restart: always
    privileged: true
    ports:
      - "__BK_NODEMGR_ETCD_PORT__:__BK_NODEMGR_ETCD_PORT__"
      - "__BK_NODEMGR_ETCD_PEER_PORT__:__BK_NODEMGR_ETCD_PEER_PORT__"
    environment:
      - ETCD_LISTEN_CLIENT_URLS=http://0.0.0.0:__BK_NODEMGR_ETCD_PORT__
      - ETCD_ADVERTISE_CLIENT_URLS=http://etcd:__BK_NODEMGR_ETCD_PORT__
      - ETCD_LISTEN_PEER_URLS=http://0.0.0.0:__BK_NODEMGR_ETCD_PEER_PORT__
      - ETCD_INITIAL_ADVERTISE_PEER_URLS=http://etcd:__BK_NODEMGR_ETCD_PEER_PORT__
      - ETCD_INITIAL_CLUSTER_TOKEN=etcd-cluster-1
      - ETCD_INITIAL_CLUSTER_STATE=new
      - ETCD_INITIAL_CLUSTER=etcd=http://etcd:__BK_NODEMGR_ETCD_PEER_PORT__
      - ETCD_NAME=etcd
      - ETCD_ROOT_PASSWORD=__BK_NODEMGR_ETCD_PASSWORD__
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "5"

  redis:
    image: redis:6.2
    container_name: bk-nodemgr-redis
    restart: always
    privileged: true
    ports:
      - "__BK_NODEMGR_REDIS_PORT__:6379"
    volumes:
      - ./data/redis-data:/data
    command: /bin/sh -c "redis-server --requirepass __BK_NODEMGR_REDIS_PASSWORD__"
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "5"
    networks:
      - bk-nodemgr-network

  mongodb:
    image: mongo:6
    container_name: bk-nodemgr-mongodb
    restart: always
    privileged: true
    ports:
      - "__BK_NODEMGR_MONGODB_PORT__:27017"
    environment:
      - MONGO_INITDB_ROOT_USERNAME=root
      - MONGO_INITDB_ROOT_PASSWORD=__BK_NODEMGR_MONGODB_PASSWORD__
      - MONGO_INITDB_DATABASE=admin
    entrypoint:
      - bash
      - -c
      - |
          chmod 400 /mongo_keyfile
          chown 999:999 /mongo_keyfile
          exec docker-entrypoint.sh $$@
    volumes:
      - ./data/mongodb-data:/data/db
      - ./mongo_keyfile:/mongo_keyfile
    command: "mongod --bind_ip_all --replSet rs0 --keyFile /mongo_keyfile"
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "5"
    healthcheck:
      test: test $$(mongosh --username root --password __BK_NODEMGR_MONGODB_PASSWORD__ --eval "try {rs.initiate({_id:'rs0',members:[{_id:0,host:\"__BK_NODEMGR_ADVERTISE_IPV4__:__BK_NODEMGR_MONGODB_PORT__\"}]})} catch(e) {rs.status().ok}") -eq 1
      interval: 10s
      start_period: 30s
    networks:
      - bk-nodemgr-network

  nodemgr:
    image: __BK_NODEMGR_SERVICE_IMAGE__
    container_name: bk-nodemgr-service
    restart: always
    privileged: true
    ports:
      - "__BK_NODEMGR_APPLICATION_INFO_PORT__:__BK_NODEMGR_APPLICATION_INFO_PORT__"
      - "__BK_NODEMGR_APPLICATION_ADMIN_PORT__:__BK_NODEMGR_APPLICATION_ADMIN_PORT__"
      - "__BK_NODEMGR_APPLICATION_BASIC_PORT__:__BK_NODEMGR_APPLICATION_BASIC_PORT__"
      - "__BK_NODEMGR_BACKEND_INFO_PORT__:__BK_NODEMGR_BACKEND_INFO_PORT__"
      - "__BK_NODEMGR_BACKEND_ADMIN_PORT__:__BK_NODEMGR_BACKEND_ADMIN_PORT__"
      - "__BK_NODEMGR_BACKEND_BASIC_PORT__:__BK_NODEMGR_BACKEND_BASIC_PORT__"
      - "__BK_NODEMGR_BACKEND_CALLBACK_PORT__:__BK_NODEMGR_BACKEND_CALLBACK_PORT__"
      - "__BK_NODEMGR_BACKEND_PROXY_PORT__:__BK_NODEMGR_BACKEND_PROXY_PORT__"
      - "__BK_NODEMGR_FILE_SERVICE_PORT__:__BK_NODEMGR_FILE_SERVICE_PORT__"   
      - "__BK_NODEMGR_FILE_ADMIN_PORT__:__BK_NODEMGR_FILE_ADMIN_PORT__"
    volumes:
      - ./etc/:/bk-nodemgr/etc/
      - ./cert/:/bk-nodemgr/cert/
      - __BK_NODEMGR_FILE_MOUNT_HOST_DIR__:/bk-nodemgr/workspace
    command: "/bk-nodemgr/bin/serviced.sh"
    logging:
      driver: json-file
      options:
        max-size: "100m"
        max-file: "5"
    networks:
      - bk-nodemgr-network

networks:
  bk-nodemgr-network:
    driver: bridge
