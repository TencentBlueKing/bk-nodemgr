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

FROM tencentos/tencentos4-minimal

ENV TZ=Asia/Shanghai

RUN set -eux; \
    command -v dnf >/dev/null 2>&1; \
    dnf -y install \
        curl \
        ca-certificates \
        tzdata \
        jq \
        vim-minimal \
        redis \
        etcd \
        busybox \
        python3-pip \
    ; \
    dnf -y clean all; \
    rm -rf /var/cache/dnf; \
    command -v etcdctl >/dev/null 2>&1; \
    ln -snf "/usr/share/zoneinfo/${TZ}" /etc/localtime || true

# Install mongosh and keep the mongo command used by Helm init containers and ops scripts.
RUN dnf -y install nodejs npm && \
    npm install --global --no-audit --no-fund mongosh@2.12.0 && \
    ln -s "$(command -v mongosh)" /usr/bin/mongo && \
    npm cache clean --force && \
    dnf -y --setopt=clean_requirements_on_remove=False remove npm && \
    dnf -y clean all && \
    rm -rf /var/cache/dnf && \
    mongo --version

RUN pip3 install requests PyYAML

RUN mkdir -p \
    /bk-nodemgr/bin \
    /bk-nodemgr/etc \
    /bk-nodemgr/cert \
    /bk-nodemgr/log \
    /bk-nodemgr/script \
    /bk-nodemgr/static \
    /bk-nodemgr/workspace \
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
