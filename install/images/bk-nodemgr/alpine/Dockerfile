FROM node:20-alpine AS frontend-builder

WORKDIR /app
RUN corepack enable

COPY front/package.json front/.npmrc ./
COPY front/scripts/ ./scripts/
RUN CYPRESS_INSTALL_BINARY=0 pnpm install

COPY front/ .
RUN pnpm build

FROM alpine:3.21

# install dependencies.
RUN apk add bash jq vim busybox-extras redis etcd-ctl tzdata

# set timezone.
ENV TZ Asia/Shanghai

# install mongodb with old package cause it doesn’t exist for Alpine after version 3.9.
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/main' >> /etc/apk/repositories
RUN echo 'http://dl-cdn.alpinelinux.org/alpine/v3.9/community' >> /etc/apk/repositories
RUN apk add mongodb && rm -f /usr/bin/mongod /usr/bin/mongos
RUN apk add py3-requests py3-yaml

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
COPY bk-nodemgr-adminclient /bk-nodemgr/bin/
COPY serviced.sh /bk-nodemgr/bin/
COPY scripts/manual/ /bk-nodemgr/script/manual
COPY scripts/ops/ /bk-nodemgr/script/ops
COPY tools/ /bk-nodemgr/file/tools
COPY bintools/ /bk-nodemgr/file/packages/bintool
COPY --from=frontend-builder /app/dist/ /bk-nodemgr/static/
COPY support-files/ /bk-nodemgr/support-files

WORKDIR /bk-nodemgr/bin
