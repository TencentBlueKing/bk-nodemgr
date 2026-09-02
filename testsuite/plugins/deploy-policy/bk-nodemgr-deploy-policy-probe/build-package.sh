#!/bin/sh

set -eu

name="bk-nodemgr-deploy-policy-probe"
version="1.0.0"
base_dir="$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)"
package_dir="${base_dir}/package"
artifact="${name}-${version}.tgz"

tar -C "$package_dir" -czf "${base_dir}/${artifact}" "$name"
(
	cd "$base_dir"
	sha256sum "$artifact" >"${artifact}.sha256"
)

printf 'built %s\n' "$artifact"
printf 'updated %s.sha256\n' "$artifact"
