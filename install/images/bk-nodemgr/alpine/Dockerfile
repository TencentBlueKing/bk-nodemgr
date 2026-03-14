FROM alpine:3.21

# install dependencies.
RUN apk add jq vim busybox-extras redis etcd-ctl tzdata

# set timezone.
ENV TZ Asia/Shanghai

# install mongodb with old package cause it doesn’t exist for Alpine after version 3.9.
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/main' >> /etc/apk/repositories
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/community' >> /etc/apk/repositories
RUN apk add mongodb && rm -f /usr/bin/mongod /usr/bin/mongos
RUN apk add py3-requests

RUN mkdir -p \
    /bk-nodemgr/bin \
    /bk-nodemgr/etc \
    /bk-nodemgr/cert \
    /bk-nodemgr/log \
    /bk-nodemgr/script \
    /bk-nodemgr/static \
    /bk-nodemgr/bkiam/v3 \
    /bk-nodemgr/bkiam/v4 \
    /bk-nodemgr/file/tools \
    /bk-nodemgr/file/cache \
    /bk-nodemgr/file/local \
    /bk-nodemgr/file/temp \
    /bk-nodemgr/file/packages/cert \
    /bk-nodemgr/file/packages/bintool

COPY bk-nodemgr-backend /bk-nodemgr/bin/
COPY bk-nodemgr-application /bk-nodemgr/bin/
COPY bk-nodemgr-file /bk-nodemgr/bin/
COPY serviced.sh /bk-nodemgr/bin/
COPY scripts/manual/ /bk-nodemgr/script/manual
COPY scripts/ops/ /bk-nodemgr/script/ops
COPY tools/ /bk-nodemgr/file/tools
COPY bintools/ /bk-nodemgr/file/packages/bintool
COPY dist/ /bk-nodemgr/static/
COPY support-files/ /bk-nodemgr/support-files

WORKDIR /bk-nodemgr/bin