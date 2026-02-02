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
        busybox \
        ${rpm_url} \
    ; \
    dnf -y clean all; \
    rm -rf /var/cache/dnf; \
    command -v mongo >/dev/null 2>&1; \
    ln -snf "/usr/share/zoneinfo/${TZ}" /etc/localtime || true

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
COPY scripts/manual/ /bk-nodemgr/script/manual
COPY tools/ /bk-nodemgr/file/tools
COPY dist/ /bk-nodemgr/static/

WORKDIR /bk-nodemgr/bin
