#!/bin/sh

set -eu

name="bk-nodemgr-deploy-policy-probe"
version="1.0.1"
base_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
package_dir="${base_dir}/package"
artifact="${name}-${version}.tgz"
staging_dir="$(mktemp -d)"

cleanup() {
	rm -rf "$staging_dir"
}
trap cleanup EXIT

cp -R "${package_dir}/${name}" "$staging_dir/"
rm -rf "${staging_dir}/${name}/plugins_linux_x86_64/run"
(
	cd "$base_dir"
	GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" \
		-o "${staging_dir}/${name}/plugins_linux_x86_64/bin/${name}" "./cmd/${name}"
)
chmod +x "${staging_dir}/${name}/plugins_linux_x86_64/bin/${name}" \
	"${staging_dir}/${name}/plugins_linux_x86_64/bin/start.sh" \
	"${staging_dir}/${name}/plugins_linux_x86_64/bin/stop.sh" \
	"${staging_dir}/${name}/plugins_linux_x86_64/bin/restart.sh" \
	"${staging_dir}/${name}/plugins_linux_x86_64/bin/reload.sh"

tar -C "$staging_dir" -czf "${base_dir}/${artifact}" "$name"
(
	cd "$base_dir"
	sha256sum "$artifact" >"${artifact}.sha256"
)

printf 'built %s\n' "$artifact"
printf 'updated %s.sha256\n' "$artifact"
