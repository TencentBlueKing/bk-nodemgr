#!/bin/bash

set -euo pipefail

BIN_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
HELM_VERSION="${HELM_VERSION:-v3.21.3}"
HELMFILE_VERSION="${HELMFILE_VERSION:-v1.8.0}"
HELM_DIFF_VERSION="${HELM_DIFF_VERSION:-v3.15.13}"
PLATFORM="linux-amd64"

cd "$BIN_DIR"

download() {
	local url=$1
	local output=$2

	curl -L --fail --retry 3 --retry-delay 2 -o "$output" "$url"
}

echo "Installing helm ${HELM_VERSION} (${PLATFORM})"
download "https://get.helm.sh/helm-${HELM_VERSION}-${PLATFORM}.tar.gz" "helm-${HELM_VERSION}-${PLATFORM}.tar.gz"
tar -xzf "helm-${HELM_VERSION}-${PLATFORM}.tar.gz"
mv "${PLATFORM}/helm" ./helm
rm -rf "${PLATFORM}" "helm-${HELM_VERSION}-${PLATFORM}.tar.gz"
chmod +x ./helm

echo "Installing helmfile ${HELMFILE_VERSION} (${PLATFORM})"
download "https://github.com/helmfile/helmfile/releases/download/${HELMFILE_VERSION}/helmfile_${HELMFILE_VERSION#v}_linux_amd64.tar.gz" "helmfile-${HELMFILE_VERSION}-${PLATFORM}.tar.gz"
tar -xzf "helmfile-${HELMFILE_VERSION}-${PLATFORM}.tar.gz" helmfile
rm -f "helmfile-${HELMFILE_VERSION}-${PLATFORM}.tar.gz"
chmod +x ./helmfile

echo "Installing helm-diff ${HELM_DIFF_VERSION} (${PLATFORM})"
rm -rf helm-plugins
mkdir -p helm-plugins
download "https://github.com/databus23/helm-diff/releases/download/${HELM_DIFF_VERSION}/helm-diff-linux-amd64.tgz" "helm-diff-${HELM_DIFF_VERSION}-${PLATFORM}.tgz"
tar -xzf "helm-diff-${HELM_DIFF_VERSION}-${PLATFORM}.tgz" -C helm-plugins
rm -f "helm-diff-${HELM_DIFF_VERSION}-${PLATFORM}.tgz"

echo "Installed tools:"
./helm version --short
./helmfile --version
HELM_PLUGINS="$BIN_DIR/helm-plugins" ./helm plugin list
