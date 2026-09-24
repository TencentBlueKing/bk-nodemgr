FROM node:20-alpine AS frontend-builder

ARG MODE=production
# Skip the Cypress binary download during image builds.
ENV CYPRESS_INSTALL_BINARY=0

WORKDIR /app
RUN corepack enable

COPY front/.npmrc front/package.json front/pnpm-lock.yaml ./
COPY front/scripts/ ./scripts/
# Point pnpm's store at the cache mount to reuse downloaded dependencies.
RUN --mount=type=cache,id=pnpm-store,target=/root/.pnpm-store \
    pnpm config set store-dir /root/.pnpm-store --global && \
    pnpm install --frozen-lockfile

COPY front/ .
RUN pnpm build --mode ${MODE}

FROM alpine:3.24

# install dependencies.
RUN apk add bash curl jq yq-go vim busybox-extras redis etcd-ctl tzdata

# set timezone.
ENV TZ Asia/Shanghai

# Install mongosh without the legacy Alpine 3.9 OpenSSL dependencies.
# Keep the mongo command used by Helm init containers and ops scripts.
RUN apk add --no-cache nodejs npm && \
    npm install --global --no-audit --no-fund mongosh@2.12.0 && \
    ln -s "$(command -v mongosh)" /usr/bin/mongo && \
    npm cache clean --force && \
    apk del npm && \
    mongo --version
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
COPY docs/adminclient/ /bk-nodemgr/docs/adminclient/
COPY tools/ /bk-nodemgr/file/tools
COPY bintools/ /bk-nodemgr/file/packages/bintool
COPY --from=frontend-builder /app/dist/ /bk-nodemgr/static/
COPY support-files/ /bk-nodemgr/support-files

WORKDIR /bk-nodemgr/bin
