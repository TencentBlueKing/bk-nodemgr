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

# Install MongoDB Shell from the official tarball and keep the mongo command
# used by Helm init containers and ops scripts.
ARG MONGOSH_VERSION=2.12.0

RUN set -eux; \
    mongosh_arch=""; \
    mongosh_sha256=""; \
    case "$(uname -m)" in \
        x86_64|amd64) \
            mongosh_arch="x64"; \
            mongosh_sha256="aa42cb826b7b8e655c5481293f5365ddaf1a23e43af07520326e5bbc957838ad"; \
            ;; \
        aarch64|arm64) \
            mongosh_arch="arm64"; \
            mongosh_sha256="99f8d38d41c510e7d4c685b71022ea1db113245968d5d933b5333fb3835e64ae"; \
            ;; \
        *) echo "unsupported uname -m=$(uname -m)" >&2; exit 1 ;; \
    esac; \
    dnf -y install tar gzip; \
    mongosh_url="https://downloads.mongodb.com/compass/mongosh-${MONGOSH_VERSION}-linux-${mongosh_arch}.tgz"; \
    curl -fsSL "${mongosh_url}" -o /tmp/mongosh.tgz; \
    echo "${mongosh_sha256}  /tmp/mongosh.tgz" | sha256sum -c -; \
    mkdir -p /opt/mongosh; \
    tar -xzf /tmp/mongosh.tgz -C /opt/mongosh --strip-components=1; \
    ln -sf /opt/mongosh/bin/mongosh /usr/bin/mongosh; \
    ln -sf /opt/mongosh/bin/mongosh /usr/bin/mongo; \
    rm -f /tmp/mongosh.tgz; \
    dnf -y clean all; \
    rm -rf /var/cache/dnf; \
    mongosh --version; \
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
