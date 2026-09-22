FROM node:20-alpine AS frontend-builder

WORKDIR /app
RUN corepack enable

COPY front/package.json front/.npmrc ./
COPY front/scripts/ ./scripts/
RUN CYPRESS_INSTALL_BINARY=0 pnpm install

COPY front/ .
RUN pnpm build

FROM tencentos/tencentos4-minimal

ENV TZ=Asia/Shanghai

# Install legacy MongoDB shell binary `mongo` (for Helm readiness/liveness checks).
# Notes:
# - `mongo` is the legacy shell; newer MongoDB versions default to `mongosh`.
# - This installs only the shell package (no `mongod`).
# - Uses MongoDB official RPM download URLs and auto-selects arch.
ARG MONGO_SHELL_MAJOR=4.4
ARG MONGO_SHELL_PKG_VERSION=4.4.30-1.el8

RUN set -eux; \
    rpm_arch=""; \
    case "$(uname -m)" in \
        x86_64|amd64) rpm_arch="x86_64" ;; \
        aarch64|arm64) rpm_arch="aarch64" ;; \
        *) echo "unsupported uname -m=$(uname -m)" >&2; exit 1 ;; \
    esac; \
    rpm_url="https://repo.mongodb.org/yum/redhat/8/mongodb-org/${MONGO_SHELL_MAJOR}/${rpm_arch}/RPMS/mongodb-org-shell-${MONGO_SHELL_PKG_VERSION}.${rpm_arch}.rpm"; \
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
        ${rpm_url} \
    ; \
    dnf -y clean all; \
    rm -rf /var/cache/dnf; \
    command -v mongo >/dev/null 2>&1; \
    command -v etcdctl >/dev/null 2>&1; \
    ln -snf "/usr/share/zoneinfo/${TZ}" /etc/localtime || true

RUN pip3 install requests PyYAML

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
