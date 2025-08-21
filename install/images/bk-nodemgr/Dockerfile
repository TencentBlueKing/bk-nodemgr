FROM alpine:3.21

# install dependencies.
RUN apk add jq vim busybox-extras redis

# install mongodb with old package cause it doesn’t exist for Alpine after version 3.9.
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/main' >> /etc/apk/repositories
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/community' >> /etc/apk/repositories
RUN apk add mongodb && rm -f /usr/bin/mongod /usr/bin/mongos

RUN mkdir -p \
    /bk-nodemgr/bin \
    /bk-nodemgr/etc \
    /bk-nodemgr/cert \
    /bk-nodemgr/log \
    /bk-nodemgr/script \
    /bk-nodemgr/static \
    /bk-nodemgr/file/tools \
    /bk-nodemgr/file/cache \
    /bk-nodemgr/file/local \
    /bk-nodemgr/file/temp

COPY bk-nodemgr-backend /bk-nodemgr/bin/
COPY bk-nodemgr-application /bk-nodemgr/bin/
COPY bk-nodemgr-file /bk-nodemgr/bin/
COPY serviced.sh /bk-nodemgr/bin/
COPY tools/ /bk-nodemgr/file/tools
COPY dist/ /bk-nodemgr/static/

WORKDIR /bk-nodemgr/bin